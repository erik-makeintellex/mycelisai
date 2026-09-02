package auth

import "testing"

func TestCredentialConfigurationRejectsUnsafeTokens(t *testing.T) {
	for _, token := range []string{"short", " 0123456789abcdef0123456789abcdef", "0123456789abcdef0123456789abcdef "} {
		if _, err := NewCredential("core", token, "runs:api"); err == nil {
			t.Fatalf("accepted unsafe token %q", token)
		}
	}
	if _, err := NewCredential("core", testCredentialToken(), "runs:api"); err != nil {
		t.Fatal(err)
	}
}

func testCredentialToken() string { return "abcdef0123456789abcdef0123456789" }
