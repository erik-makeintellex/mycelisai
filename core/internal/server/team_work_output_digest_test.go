package server

import (
	"net/url"
	"testing"

	"github.com/mycelis/core/pkg/protocol"
)

func TestTeamWorkValidationLaunchURLNormalizesEntrypointForms(t *testing.T) {
	t.Setenv("MYCELIS_API_URL", "http://127.0.0.1:8081")
	storage := "groups/delivery-team/generated/app"
	for name, entrypoint := range map[string]string{
		"workspace relative": storage + "/index.html",
		"folder relative":    "index.html",
	} {
		t.Run(name, func(t *testing.T) {
			launchURL, err := teamWorkValidationLaunchURL([]protocol.TeamOutputRef{{
				StorageRef: storage,
				Entrypoint: entrypoint,
			}})
			if err != nil {
				t.Fatal(err)
			}
			parsed, err := url.Parse(launchURL)
			if err != nil {
				t.Fatal(err)
			}
			if got := parsed.Query().Get("path"); got != storage+"/index.html" {
				t.Fatalf("launch path = %q, want %q", got, storage+"/index.html")
			}
		})
	}
}
