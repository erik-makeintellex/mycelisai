package searchcap

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

var errSourceNotFound = errors.New("search source not found")

func (s *Service) ListSources() []Source {
	if s == nil {
		return []Source{}
	}
	return cloneSources(s.Status().Sources)
}

func (s *Service) AddSource(input SourceInput) (Source, error) {
	return s.AddSourceWithContext(context.Background(), input)
}

func (s *Service) AddSourceWithContext(ctx context.Context, input SourceInput) (Source, error) {
	if s == nil {
		return Source{}, errors.New("search service unavailable")
	}
	source, err := normalizeSourceInput(input)
	if err != nil {
		return Source{}, err
	}
	source.Managed = true

	s.registryMu.Lock()
	defer s.registryMu.Unlock()
	source.ID = s.uniqueSourceIDLocked(source)
	if s.sourceDB != nil {
		created, err := s.sourceDB.Create(ctx, source)
		if err != nil {
			return Source{}, err
		}
		source = created
	}
	s.registry = append(s.registry, source)
	return source, nil
}

func (s *Service) UpdateSourceWithContext(ctx context.Context, id string, input SourceInput) (Source, error) {
	if s == nil {
		return Source{}, errors.New("search service unavailable")
	}
	id = strings.TrimSpace(id)
	if id == "" {
		return Source{}, errors.New("search source id is required")
	}
	source, err := normalizeSourceInput(input)
	if err != nil {
		return Source{}, err
	}
	source.ID = id
	source.Managed = true

	s.registryMu.Lock()
	defer s.registryMu.Unlock()
	idx := s.registryIndexLocked(id)
	if idx < 0 {
		return Source{}, errSourceNotFound
	}
	if s.sourceDB != nil {
		if err := s.sourceDB.Update(ctx, source); err != nil {
			return Source{}, err
		}
	}
	s.registry[idx] = source
	return source, nil
}

func (s *Service) DeleteSourceWithContext(ctx context.Context, id string) error {
	if s == nil {
		return errors.New("search service unavailable")
	}
	id = strings.TrimSpace(id)
	if id == "" {
		return errors.New("search source id is required")
	}
	s.registryMu.Lock()
	defer s.registryMu.Unlock()
	idx := s.registryIndexLocked(id)
	if idx < 0 {
		return errSourceNotFound
	}
	if s.sourceDB != nil {
		if err := s.sourceDB.Delete(ctx, id); err != nil {
			return err
		}
	}
	s.registry = append(s.registry[:idx], s.registry[idx+1:]...)
	return nil
}

func (s *Service) appendRegisteredSources(base []Source) []Source {
	out := cloneSources(base)
	if s == nil {
		return out
	}
	s.registryMu.RLock()
	defer s.registryMu.RUnlock()
	out = append(out, cloneSources(s.registry)...)
	return out
}

func (s *Service) registryIndexLocked(id string) int {
	for idx, source := range s.registry {
		if source.ID == id {
			return idx
		}
	}
	return -1
}

func (s *Service) uniqueSourceIDLocked(source Source) string {
	base := source.Provider
	if base == "" {
		base = source.Name
	}
	base = slugSourceID(base)
	if base == "" {
		base = "search_source"
	}
	used := map[string]bool{}
	if configuredID := configuredSourceID(s.Provider()); configuredID != "" {
		used[configuredID] = true
	}
	for _, existing := range s.registry {
		used[existing.ID] = true
	}
	if !used[base] {
		return base
	}
	for i := 2; ; i++ {
		candidate := fmt.Sprintf("%s_%d", base, i)
		if !used[candidate] {
			return candidate
		}
	}
}

func configuredSourceID(provider string) string {
	switch provider {
	case ProviderLocalSources:
		return "local_sources"
	case ProviderSearXNG:
		return "searxng"
	case ProviderLocalAPI:
		return "local_api"
	case ProviderBrave:
		return "brave-search"
	default:
		return ""
	}
}
