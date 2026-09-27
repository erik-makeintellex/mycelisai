package exchange

import "testing"

// Handoff notes are written by Core's handoff recorder only. No agent role may
// publish into the lane directly, or a model could forge a handoff.
func TestTeamHandoffChannelIsCoreWrittenAndLeadReadable(t *testing.T) {
	var channel *Channel
	for i := range SeedChannels {
		if SeedChannels[i].Name == "organization.team.handoffs" {
			channel = &SeedChannels[i]
		}
	}
	if channel == nil {
		t.Fatal("organization.team.handoffs is not seeded")
	}
	if _, ok := SchemaByID(channel.SchemaID); !ok || channel.SchemaID != "HandoffNote" {
		t.Fatalf("channel schema = %q", channel.SchemaID)
	}
	for _, role := range []string{"soma", "team_lead", "specialist", "automation", "mcp"} {
		if canWriteChannel(Actor{Role: role}, channel) {
			t.Fatalf("role %s can write the handoff lane", role)
		}
	}
	if !canReadChannel(Actor{Role: "team_lead"}, channel) || canReadChannel(Actor{Role: "specialist"}, channel) {
		t.Fatal("handoff lane must be readable by team leads only among agent roles")
	}
	schema, _ := SchemaByID("HandoffNote")
	for _, field := range append(append([]string{}, schema.RequiredFields...), schema.OptionalFields...) {
		if _, ok := FieldByName(field); !ok {
			t.Fatalf("HandoffNote field %q is not registered", field)
		}
	}
	for _, capability := range schema.RequiredCapabilities {
		if _, ok := CapabilityByID(capability); !ok {
			t.Fatalf("HandoffNote capability %q is not registered", capability)
		}
	}
}

// A lead sees only HandoffNotes its own team sent or received; admin
// (the operator API) keeps seeing every note.
func TestTeamHandoffNotesAreScopedToSenderAndReceiver(t *testing.T) {
	var channel Channel
	for _, seeded := range SeedChannels {
		if seeded.Name == "organization.team.handoffs" {
			channel = seeded
		}
	}
	note := &ExchangeItem{SourceTeam: "research", TargetTeam: "marketing", SourceRole: "team_lead", TargetRole: "team_lead", AllowedConsumers: []string{"team_lead"}, SensitivityClass: "role_scoped"}
	for _, tc := range []struct {
		actor Actor
		want  bool
	}{
		{Actor{Role: "team_lead", Team: "research"}, true},
		{Actor{Role: "team_lead", Team: "marketing"}, true},
		{Actor{Role: "team_lead", Team: "sales"}, false},
		{Actor{Role: "team_lead"}, false},
		{Actor{Role: "soma", Team: "admin-core"}, false},
		{Actor{Role: "admin"}, true},
	} {
		if got := canReadItem(tc.actor, &channel, note); got != tc.want {
			t.Fatalf("%#v read = %v, want %v", tc.actor, got, tc.want)
		}
	}
}
