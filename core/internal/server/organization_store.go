package server

import (
	"sort"
	"strings"
	"sync"
)

type OrganizationStore struct {
	mu    sync.RWMutex
	items map[string]OrganizationHomePayload
}

func NewOrganizationStore() *OrganizationStore {
	return &OrganizationStore{items: make(map[string]OrganizationHomePayload)}
}

func (s *OrganizationStore) List() []OrganizationSummary {
	s.mu.RLock()
	defer s.mu.RUnlock()

	summaries := make([]OrganizationSummary, 0, len(s.items))
	for _, item := range s.items {
		summaries = append(summaries, item.OrganizationSummary)
	}
	sort.Slice(summaries, func(i, j int) bool {
		return summaries[i].Name < summaries[j].Name
	})
	return summaries
}

func (s *OrganizationStore) Save(home OrganizationHomePayload) OrganizationHomePayload {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items[home.ID] = home
	return home
}

func (s *OrganizationStore) Get(id string) (OrganizationHomePayload, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	item, ok := s.items[id]
	return item, ok
}

func (s *OrganizationStore) Update(id string, update func(OrganizationHomePayload) OrganizationHomePayload) (OrganizationHomePayload, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	item, ok := s.items[id]
	if !ok {
		return OrganizationHomePayload{}, false
	}

	item = update(item)
	s.items[id] = item
	return item, true
}

func (s *AdminServer) templateBundlesPath() string {
	if strings.TrimSpace(s.TemplateBundlesPath) != "" {
		return s.TemplateBundlesPath
	}
	return "config/templates"
}

func (s *AdminServer) organizationStore() *OrganizationStore {
	if s.Organizations == nil {
		s.Organizations = NewOrganizationStore()
	}
	return s.Organizations
}
