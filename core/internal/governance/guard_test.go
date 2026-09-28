package governance

import (
	"testing"
)

func TestValidateIngress(t *testing.T) {
	// Setup Guard (mock engine not needed for Ingress check)
	g := &Guard{}

	tests := []struct {
		name    string
		subject string
		data    []byte
		wantErr bool
	}{
		{
			name:    "Valid User Input",
			subject: "swarm.global.input.user",
			data:    []byte("hello"),
			wantErr: false,
		},
		{
			// F16: exact match only; deep subjects are no longer admitted.
			name:    "Deep Sensor Subject Rejected",
			subject: "swarm.global.input.sensor.temp",
			data:    []byte("25c"),
			wantErr: true,
		},
		{
			name:    "Invalid Subject Prefix",
			subject: "swarm.internal.attack",
			data:    []byte("rm -rf"),
			wantErr: true,
		},
		{
			name:    "Payload Too Large",
			subject: "swarm.global.input.cli",
			data:    make([]byte, 2*1024*1024), // 2MB
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := g.ValidateIngress(tt.subject, tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateIngress() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
