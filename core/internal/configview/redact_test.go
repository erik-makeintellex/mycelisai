package configview

import (
	"reflect"
	"strings"
	"testing"
)

const leak = "sk-live-0123456789abcdefSECRET"

func TestRedact_Table(t *testing.T) {
	cases := []struct {
		name      string
		doc       any
		want      any
		redacted  []string
		mustShow  []string
		mustntSee []string
	}{
		{
			name:     "top-level api_key with env sibling becomes env reference",
			doc:      map[string]any{"api_key": leak, "api_key_env": "OPENAI_API_KEY", "model_id": "gpt-4o"},
			want:     map[string]any{"api_key": "${env:OPENAI_API_KEY}", "api_key_env": "OPENAI_API_KEY", "model_id": "gpt-4o"},
			redacted: []string{"api_key"},
		},
		{
			name: "nested maps and lists",
			doc: map[string]any{"providers": map[string]any{"openai": map[string]any{
				"auth": map[string]any{"client_secret": leak, "client_id": "visible-id"},
			}}, "servers": []any{map[string]any{"env": map[string]any{"GITHUB_TOKEN": leak, "LOG_LEVEL": "debug"}}}},
			redacted:  []string{"providers.openai.auth.client_secret", "servers[0].env.GITHUB_TOKEN"},
			mustShow:  []string{"visible-id", "debug", "GITHUB_TOKEN"},
			mustntSee: []string{leak},
		},
		{
			name:      "mixed case and separators",
			doc:       map[string]any{"ApiKey": leak, "PASSWORD": leak, "Private-Key": leak, "x.Authorization": leak, "BearerToken": leak, "OPENAIAPIKEY": leak, "Credentials": leak, "passwd": leak},
			redacted:  []string{"ApiKey", "BearerToken", "Credentials", "OPENAIAPIKEY", "PASSWORD", "Private-Key", "passwd", "x.Authorization"},
			mustntSee: []string{leak},
		},
		{
			name:     "env and ref references stay visible",
			doc:      map[string]any{"api_key_env": "OPENAI_API_KEY", "tokenEnv": "GH_TOKEN", "secret_ref": "vault:mycelis/openai", "token_env_var": "SLACK_BOT_TOKEN"},
			want:     map[string]any{"api_key_env": "OPENAI_API_KEY", "tokenEnv": "GH_TOKEN", "secret_ref": "vault:mycelis/openai", "token_env_var": "SLACK_BOT_TOKEN"},
			redacted: []string{},
		},
		{
			name:      "reference holding a literal secret is redacted",
			doc:       map[string]any{"api_key_env": leak, "secret_ref": "Bearer abcdefgh12345678"},
			redacted:  []string{"api_key_env", "secret_ref"},
			mustntSee: []string{leak, "abcdefgh12345678"},
		},
		{
			name:     "non-string secret values",
			doc:      map[string]any{"password": 123456, "token": 3.5, "secret": map[string]any{"a": "b"}, "credentials": []any{"x", "y"}},
			want:     map[string]any{"password": Marker, "token": Marker, "secret": Marker, "credentials": Marker},
			redacted: []string{"credentials", "password", "secret", "token"},
		},
		{
			name:     "flags, empties, and look-alike keys stay visible",
			doc:      map[string]any{"secret": true, "api_key": "", "max_tokens": 4096, "max_output_tokens": 2048, "token_budget_profile": "standard", "tokens": 5, "keywords": "ops"},
			want:     map[string]any{"secret": true, "api_key": "", "max_tokens": 4096, "max_output_tokens": 2048, "token_budget_profile": "standard", "tokens": 5, "keywords": "ops"},
			redacted: []string{},
		},
		{
			name:      "secret-shaped values under ordinary keys",
			doc:       map[string]any{"note": "use " + leak + " here", "header": "Bearer abcdefgh12345678", "dsn": "postgres://mycelis:hunter2pass@db:5432/x", "hook": "https://x.test/h?token=zzsecretzz&team=a", "cmd": "GITHUB_TOKEN=ghx-literal-value run", "jwt": "eyJhbGciOiJI.eyJzdWIiOiIxMjM0.c2lnbmF0dXJlMTIz", "pem": "-----BEGIN RSA PRIVATE KEY-----\nabc\n-----END RSA PRIVATE KEY-----"},
			redacted:  []string{"cmd", "dsn", "header", "hook", "jwt", "note", "pem"},
			mustShow:  []string{"postgres://mycelis:", "team=a", "GITHUB_TOKEN=" + Marker},
			mustntSee: []string{leak, "abcdefgh12345678", "hunter2pass", "zzsecretzz", "ghx-literal-value", "c2lnbmF0dXJlMTIz", "abc\n"},
		},
		{
			name:      "sensitive CLI flag hides the next argument",
			doc:       map[string]any{"args": []any{"serve", "--api-key", "plainliteral", "--port", "8080", "--token=inline-literal"}},
			want:      map[string]any{"args": []any{"serve", "--api-key", Marker, "--port", "8080", "--token=" + Marker}},
			redacted:  []string{"args[2]", "args[5]"},
			mustntSee: []string{"plainliteral", "inline-literal"},
		},
		{
			name:     "prose is not mistaken for a secret",
			doc:      map[string]any{"system_prompt": "Basic instructions: ask-for-approval before any token is spent. Bearer of news."},
			want:     map[string]any{"system_prompt": "Basic instructions: ask-for-approval before any token is spent. Bearer of news."},
			redacted: []string{},
		},
		{
			name: "typed structs and map[any]any are normalized and walked",
			doc: map[any]any{"team": struct {
				APIKey string `yaml:"api_key"`
				Name   string `yaml:"name"`
			}{leak, "core"}, 7: "seven"},
			redacted:  []string{"team.api_key"},
			mustShow:  []string{"core", "seven"},
			mustntSee: []string{leak},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			res, err := Redact(tc.doc)
			if err != nil {
				t.Fatalf("redact: %v", err)
			}
			if tc.want != nil && !reflect.DeepEqual(res.Document, tc.want) {
				t.Fatalf("document = %#v\nwant %#v", res.Document, tc.want)
			}
			if !reflect.DeepEqual(res.Redactions, tc.redacted) {
				t.Fatalf("redactions = %v, want %v", res.Redactions, tc.redacted)
			}
			for _, format := range []string{FormatYAML, FormatJSON} {
				out, err := Render(tc.doc, format)
				if err != nil {
					t.Fatalf("render %s: %v", format, err)
				}
				for _, s := range tc.mustntSee {
					if strings.Contains(out.Document, s) {
						t.Fatalf("%s leaked %q:\n%s", format, s, out.Document)
					}
				}
				for _, s := range tc.mustShow {
					if format == FormatYAML && !strings.Contains(out.Document, s) {
						t.Fatalf("%s hid %q:\n%s", format, s, out.Document)
					}
				}
			}
		})
	}
}

func TestIsSensitiveKey(t *testing.T) {
	for key, want := range map[string]bool{
		"api_key": true, "apikey": true, "apiKey": true, "OPENAI_API_KEY": true, "private_key": true,
		"client_secret": true, "db_password": true, "Authorization": true, "bearer": true,
		"GITHUB_PERSONAL_ACCESS_TOKEN": true, "access-token": true, "credentials": true, "passwd": true,
		"ssh_passphrase": true, "keyPassphrase": true, "session_cookie": true,
		"api_key_env": false, "secret_ref": false, "max_tokens": false, "token_budget_profile": false,
		"model_id": false, "name": false, "monkey": false, "env": false, "approval_policy_ref": false,
	} {
		if got := IsSensitiveKey(key); got != want {
			t.Errorf("IsSensitiveKey(%q) = %v, want %v", key, got, want)
		}
	}
}

func TestParseFormat(t *testing.T) {
	for raw, want := range map[string]string{"": FormatYAML, "YAML": FormatYAML, "yml": FormatYAML, "json": FormatJSON} {
		if got, err := ParseFormat(raw); err != nil || got != want {
			t.Errorf("ParseFormat(%q) = %q, %v", raw, got, err)
		}
	}
	if _, err := ParseFormat("xml"); err == nil {
		t.Fatal("xml must be rejected")
	}
}
