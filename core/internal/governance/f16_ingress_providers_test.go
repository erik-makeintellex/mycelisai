package governance

import "testing"

func TestF16ValidateIngressAllowedProviders(t *testing.T) {
	g := &Guard{}
	g.SetIngressProviders([]string{"whatsapp", "a.b", "*", ">", "Slack", "a b", ""})
	if err := g.ValidateIngress("swarm.global.input.whatsapp", []byte("x")); err != nil {
		t.Fatalf("allowed provider rejected: %v", err)
	}
	for _, subject := range []string{
		"swarm.global.input.a.b", "swarm.global.input.*", "swarm.global.input.>",
		"swarm.global.input.Slack", "swarm.global.input.slack", "swarm.global.input.a b",
		"swarm.global.input.", "swarm.global.input.whatsapp.x",
	} {
		if err := g.ValidateIngress(subject, []byte("x")); err == nil {
			t.Errorf("subject %q admitted after invalid provider tokens were offered", subject)
		}
	}
	if !ValidIngressProviderToken("whatsapp") || !ValidIngressProviderToken("my_hook-2") {
		t.Fatalf("valid provider tokens rejected")
	}
	for _, bad := range []string{"", "a.b", "*", ">", "A", "a/b", "a%2Fb", "a b", "a\n", "abcdefghijklmnopqrstuvwxyz0123456"} {
		if ValidIngressProviderToken(bad) {
			t.Errorf("provider token %q accepted", bad)
		}
	}
}
