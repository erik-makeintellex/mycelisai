package mcp

import (
	"strings"
	"testing"
)

// RED: RedactToolText forms found by the MCPS-QA and MCPA-QA passes.
func TestRedRedactToolTextForms(t *testing.T) {
	cases := []struct {
		name, in, secret, kept string
	}{
		{"json token", `{"token": "redsecret01"}`, "redsecret01", `"token": "[REDACTED]"`},
		{"json api_key tight", `{"api_key":"redsecret02"}`, "redsecret02", `"api_key":"[REDACTED]"`},
		{"json password spaced", `{"password" : "redsecret03"}`, "redsecret03", `"password" : "[REDACTED]"`},
		{"json escaped in string", `body={\"access_token\":\"redsecret04\"}`, "redsecret04", `access_token`},
		{"camel apiKey", `apiKey=redsecret05 next`, "redsecret05", "apiKey=[REDACTED] next"},
		{"camel accessToken json", `{"accessToken":"redsecret06"}`, "redsecret06", `"accessToken":"[REDACTED]"`},
		{"camel clientSecret", `clientSecret: redsecret07`, "redsecret07", "clientSecret: [REDACTED]"},
		{"camel privateKey", `privateKey='redsecret08'`, "redsecret08", "privateKey='[REDACTED]'"},
		{"header x-api-key", "X-Api-Key: redsecret09\n", "redsecret09", "X-Api-Key: [REDACTED]"},
		{"header basic", "Authorization: Basic dXNlcjpyZWRzZWNyZXQxMA==", "dXNlcjpyZWRzZWNyZXQxMA", "Authorization: Basic [REDACTED]"},
		{"header bearer", "Authorization: Bearer redsecret11.x", "redsecret11", "Authorization: Bearer [REDACTED]"},
		{"free basic", "login with Basic dXNlcjpwYXNz failed", "dXNlcjpwYXNz", "Basic [REDACTED] failed"},
		{"bare ghp", "invalid credential ghp_redsecret-12x", "redsecret-12x", "invalid credential ghp_[REDACTED]"},
		{"bare gho", "gho_RedSecret13abcdef", "RedSecret13", "gho_[REDACTED]"},
		{"bare github_pat", "github_pat_11ABCDEFG0redsecret14", "redsecret14", "github_pat_[REDACTED]"},
		{"bare sk", "key sk-redsecret15abcdefgh used", "redsecret15", "key sk-[REDACTED] used"},
		{"bare sk-proj", "sk-proj-redsecret16abcdefgh", "redsecret16", "sk-proj-[REDACTED]"},
		{"bare xoxb", "slack xoxb-1234-redsecret17", "redsecret17", "xoxb-[REDACTED]"},
		{"bare xoxp", "xoxp-9876-redsecret18", "redsecret18", "xoxp-[REDACTED]"},
		{"bare aws", "id AKIAREDSECRET1900XYZ ok", "REDSECRET1900", "id [REDACTED] ok"},
		{"bare jwt", "jwt eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiJyZWQyMCJ9.redsecret20sig done", "redsecret20sig", "jwt [REDACTED] done"},
		{"url userinfo", "dial postgres://red:redsecret21@db:5432/x", "redsecret21", "postgres://[REDACTED]@db:5432/x"},
	}
	for _, tc := range cases {
		got := RedactToolText(tc.in)
		if strings.Contains(got, tc.secret) {
			t.Errorf("%s: secret %q leaked: %s", tc.name, tc.secret, got)
		}
		if !strings.Contains(got, tc.kept) {
			t.Errorf("%s: want %q in %q", tc.name, tc.kept, got)
		}
	}
}

// RED: the MCPA-QA leaky text as one string.
func TestRedRedactToolTextQALeakyText(t *testing.T) {
	text := `denied {"token": "zzqa-json-secret-1"} X-Api-Key: zzqa-hdr-secret-2 apiKey=zzqa-camel-secret-3 invalid credential ghp_zzqa-bare-ghp-4`
	got := RedactToolText(text)
	for _, s := range []string{"zzqa-json-secret-1", "zzqa-hdr-secret-2", "zzqa-camel-secret-3", "zzqa-bare-ghp-4"} {
		if strings.Contains(got, s) {
			t.Errorf("secret %q leaked: %s", s, got)
		}
	}
	if !strings.HasPrefix(got, "denied ") || !strings.Contains(got, "invalid credential") {
		t.Errorf("prose lost: %s", got)
	}
}

// RED: ordinary prose must pass unchanged.
func TestRedRedactToolTextKeepsProse(t *testing.T) {
	for _, in := range []string{
		"the tokenizer split the skeleton into a task-list",
		"max_tokens: 1200 and monkey: banana",
		"basic usage of Basic Authentication is documented",
		"a secret of success; the token bucket refills",
		"ask-me-anything sk-learn eye.jpg eyJ short",
		"see https://example.com/path?q=1 and user@example.com",
		"KEYBOARD layout AKIA short",
	} {
		if got := RedactToolText(in); got != in {
			t.Errorf("over-redacted:\n in: %s\nout: %s", in, got)
		}
	}
}
