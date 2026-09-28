package swarm

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestRecRecallQuery(t *testing.T) {
	long := strings.Repeat("é", 200) // 400 bytes, 2 bytes per rune
	cases := []struct{ name, input, want string }{
		{"chat-wrapped turn ranks on the request", recWrappedChatTurn(recPromoAsk), recPromoAsk},
		{"plain input is kept", "  " + recPromoAsk + "\n", recPromoAsk},
		{"bracketed text without a marker is kept", "[note] cardamom knot", "[note] cardamom knot"},
		{"marker inside plain text is not unwrapped", "Weekend plan\nOriginal request:\ncardamom", "Weekend plan\nOriginal request:\ncardamom"},
		{"wrapper around a bare marker yields empty", "[DIRECT ANSWER ROUTE]\nAnswer.\nOriginal request:\n", ""},
	}
	for _, tc := range cases {
		if got := recallQuery(tc.input, 240); got != tc.want {
			t.Fatalf("%s: got %q want %q", tc.name, got, tc.want)
		}
	}
	got := recallQuery(recWrappedChatTurn(long), 241)
	if len(got) != 240 || !utf8.ValidString(got) {
		t.Fatalf("cap must cut on a rune boundary: len=%d valid=%t", len(got), utf8.ValidString(got))
	}
}
