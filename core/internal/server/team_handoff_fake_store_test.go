package server

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/mycelis/core/internal/dispatchoutbox"
	"github.com/mycelis/core/pkg/protocol"
)

// memTeamHandoffStore is an in-memory teamHandoffStore for recorder and API
// tests. The SQL store is covered separately against sqlmock.
type memTeamHandoffStore struct {
	mu           sync.Mutex
	artifacts    map[string]teamHandoffArtifact
	workItems    []protocol.TeamWorkItem
	handoffs     map[string]teamHandoffRecord
	interactions []protocol.TeamInteraction
	outbox       []dispatchoutbox.Item
	commits      int
	commitErr    error
}

func newMemTeamHandoffStore() *memTeamHandoffStore {
	return &memTeamHandoffStore{artifacts: map[string]teamHandoffArtifact{}, handoffs: map[string]teamHandoffRecord{}}
}

func (m *memTeamHandoffStore) LoadArtifacts(_ context.Context, ids []string) (map[string]teamHandoffArtifact, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := map[string]teamHandoffArtifact{}
	for _, id := range ids {
		if artifact, ok := m.artifacts[id]; ok {
			out[id] = artifact
		}
	}
	return out, nil
}

func (m *memTeamHandoffStore) RunWorkItemForTeam(_ context.Context, runID, teamID string) (*protocol.TeamWorkItem, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, item := range m.workItems {
		if item.RunID == runID && item.TeamID == teamID && item.State != protocol.TeamWorkStateArchived {
			copied := item
			return &copied, nil
		}
	}
	return nil, nil
}

func (m *memTeamHandoffStore) GetHandoff(_ context.Context, handoffID string) (*teamHandoffRecord, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if record, ok := m.handoffs[handoffID]; ok {
		return &record, nil
	}
	return nil, nil
}

func (m *memTeamHandoffStore) CommitHandoff(_ context.Context, commit teamHandoffCommit) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.commitErr != nil {
		return m.commitErr
	}
	if _, exists := m.handoffs[commit.Record.HandoffID]; exists {
		return errors.New("duplicate handoff id")
	}
	m.commits++
	commit.Record.CreatedAt = time.Now().UTC()
	m.handoffs[commit.Record.HandoffID] = commit.Record
	item := commit.WorkItem
	item.LastEvent = &commit.Event
	m.workItems = append(m.workItems, item)
	commit.Interaction.Timestamp = time.Now().UTC()
	m.interactions = append(m.interactions, commit.Interaction)
	m.outbox = append(m.outbox, commit.Outbox)
	return nil
}

func (m *memTeamHandoffStore) RecordInteraction(_ context.Context, interaction protocol.TeamInteraction) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	interaction.Timestamp = time.Now().UTC()
	m.interactions = append(m.interactions, interaction)
	return nil
}

func (m *memTeamHandoffStore) ListHandoffs(_ context.Context, limit int) ([]teamHandoffRecord, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]teamHandoffRecord, 0, len(m.handoffs))
	for _, record := range m.handoffs {
		if len(out) == limit {
			break
		}
		out = append(out, record)
	}
	return out, nil
}

func (m *memTeamHandoffStore) WorkItem(_ context.Context, teamID, workItemID string) (protocol.TeamWorkItem, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, item := range m.workItems {
		if item.TeamID == teamID && item.WorkItemID == workItemID {
			return item, nil
		}
	}
	return protocol.TeamWorkItem{}, errors.New("work item not found")
}

func (m *memTeamHandoffStore) Interactions(_ context.Context, teamID, workItemID string) ([]protocol.TeamInteraction, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := []protocol.TeamInteraction{}
	for _, interaction := range m.interactions {
		if interaction.TeamID == teamID && interaction.WorkItemID == workItemID {
			out = append(out, interaction)
		}
	}
	return out, nil
}

// projectAcceptance mirrors what the team work signal projection does with a
// correlated "Team accepted work" status signal: the item moves to running.
func (m *memTeamHandoffStore) projectAcceptance(teamID, workItemID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i := range m.workItems {
		if m.workItems[i].TeamID == teamID && m.workItems[i].WorkItemID == workItemID {
			m.workItems[i].State = protocol.TeamWorkStateRunning
			m.workItems[i].UpdatedAt = time.Now().UTC()
		}
	}
}

func (m *memTeamHandoffStore) interactionVerbs(teamID, workItemID string) []string {
	items, _ := m.Interactions(context.Background(), teamID, workItemID)
	verbs := make([]string, 0, len(items))
	for _, item := range items {
		verbs = append(verbs, item.Verb)
	}
	return verbs
}
