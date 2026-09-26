package server

import (
	"context"
	"sort"
	"sync"
	"testing"
)

// memoryOrganizationRepository is test-only. Production has no in-memory
// organization store; tests inject this explicitly.
type memoryOrganizationRepository struct {
	mu    sync.Mutex
	items map[string]OrganizationHomePayload
	err   error
}

func newMemoryOrganizationStore() *OrganizationStore {
	return NewOrganizationStoreWithRepository(&memoryOrganizationRepository{items: map[string]OrganizationHomePayload{}})
}

func (m *memoryOrganizationRepository) List(context.Context) ([]OrganizationSummary, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.err != nil {
		return nil, m.err
	}
	summaries := make([]OrganizationSummary, 0, len(m.items))
	for _, item := range m.items {
		summaries = append(summaries, item.OrganizationSummary)
	}
	sort.Slice(summaries, func(i, j int) bool {
		if summaries[i].Name != summaries[j].Name {
			return summaries[i].Name < summaries[j].Name
		}
		return summaries[i].ID < summaries[j].ID
	})
	return summaries, nil
}

func (m *memoryOrganizationRepository) Insert(_ context.Context, home OrganizationHomePayload) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.err != nil {
		return m.err
	}
	if _, exists := m.items[home.ID]; exists {
		return ErrOrganizationConflict
	}
	m.items[home.ID] = home
	return nil
}

func (m *memoryOrganizationRepository) Get(_ context.Context, id string) (OrganizationHomePayload, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.err != nil {
		return OrganizationHomePayload{}, m.err
	}
	item, ok := m.items[id]
	if !ok {
		return OrganizationHomePayload{}, ErrOrganizationNotFound
	}
	return item, nil
}

func (m *memoryOrganizationRepository) Update(_ context.Context, id string, update func(OrganizationHomePayload) (OrganizationHomePayload, error)) (OrganizationHomePayload, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.err != nil {
		return OrganizationHomePayload{}, m.err
	}
	item, ok := m.items[id]
	if !ok {
		return OrganizationHomePayload{}, ErrOrganizationNotFound
	}
	next, err := update(item)
	if err != nil {
		return OrganizationHomePayload{}, err
	}
	item = preserveOrganizationIdentity(item, next)
	m.items[id] = item
	return item, nil
}

func (m *memoryOrganizationRepository) Delete(_ context.Context, id string) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.err != nil {
		return false, m.err
	}
	_, ok := m.items[id]
	delete(m.items, id)
	return ok, nil
}

// withMemoryOrganizations injects the explicit test repository.
func withMemoryOrganizations(s *AdminServer) {
	s.Organizations = newMemoryOrganizationStore()
}

// seedOrganization stores home through the store and fails the test on error.
func seedOrganization(t *testing.T, s *AdminServer, home OrganizationHomePayload) OrganizationHomePayload {
	t.Helper()
	saved, err := s.organizationStore().Save(context.Background(), home)
	if err != nil {
		t.Fatalf("seed organization %q: %v", home.ID, err)
	}
	return saved
}

func mustGetOrganization(t *testing.T, s *AdminServer, id string) OrganizationHomePayload {
	t.Helper()
	home, err := s.organizationStore().Get(context.Background(), id)
	if err != nil {
		t.Fatalf("get organization %q: %v", id, err)
	}
	return home
}
