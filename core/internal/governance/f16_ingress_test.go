package governance

import "testing"

// F16: ValidateIngress admits only an exact set of global-input subjects
// (protocol constants plus explicitly allowed provider tokens).
func TestF16ValidateIngressExactMatch(t *testing.T) {
	g := &Guard{}
	cases := []struct {
		subject string
		admit   bool
	}{
		{"swarm.global.input.user", true},
		{"swarm.global.input", false},
		{"swarm.global.inputX", false},
		{"swarm.global.input_evil", false},
		{"swarm.global.input.", false},
		{"swarm.global.input.a.b.c", false},
		{"swarm.global.input.user.x", false},
		{"swarm.global.input.*", false},
		{"swarm.global.input.>", false},
		{"swarm.global.input.cli.command", false},
		{"swarm.global.input.sensor.x", false},
		{"swarm.global.input.a b", false},
		{"swarm.global.input.whatsapp", false}, // not allowed until configured
		{"swarm.global.inpu", false},
		{"swarm.team.admin-core.internal.command", false},
		{"", false},
	}
	for _, tc := range cases {
		err := g.ValidateIngress(tc.subject, []byte("x"))
		if got := err == nil; got != tc.admit {
			t.Errorf("subject %q: admitted=%v, want %v (err=%v)", tc.subject, got, tc.admit, err)
		}
	}
	if err := g.ValidateIngress("swarm.global.input.user", make([]byte, 1024*1024+1)); err == nil {
		t.Errorf("oversize payload admitted")
	}
}
