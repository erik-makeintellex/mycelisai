package inputs

import (
	"context"
	"fmt"
	"strings"
	"sync"
)

type Service struct {
	mu      sync.RWMutex
	sources map[string]Source
	store   *Store
}

func NewService() *Service {
	return &Service{sources: map[string]Source{}}
}

func (s *Service) UseStore(ctx context.Context, store *Store) error {
	if s == nil {
		return ErrUnavailable
	}
	s.mu.Lock()
	s.store = store
	s.mu.Unlock()
	if store == nil {
		return nil
	}
	sources, err := store.List(ctx)
	if err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	loaded := map[string]Source{}
	for _, source := range sources {
		if owner := subjectOwner(loaded, source.AllowedIngressSubject, source.ID); owner != "" {
			return fmt.Errorf("%w: %s is already owned by %s", ErrSubjectInUse, source.AllowedIngressSubject, owner)
		}
		loaded[source.ID] = source
	}
	s.sources = loaded
	return nil
}

func (s *Service) List() []Source {
	if s == nil {
		return nil
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]Source, 0, len(s.sources))
	for _, source := range s.sources {
		result = append(result, source)
	}
	return result
}

func (s *Service) Get(ctx context.Context, id string) (Source, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return Source{}, ErrNotFound
	}
	if s != nil && s.store != nil {
		return s.store.Get(ctx, id)
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	source, ok := s.sources[id]
	if !ok {
		return Source{}, ErrNotFound
	}
	return source, nil
}

func (s *Service) Add(ctx context.Context, input SourceInput) (Source, error) {
	source, err := NormalizeSourceInput(input)
	if err != nil {
		return Source{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if owner := subjectOwner(s.sources, source.AllowedIngressSubject, source.ID); owner != "" {
		return Source{}, fmt.Errorf("%w: %s is already owned by %s", ErrSubjectInUse, source.AllowedIngressSubject, owner)
	}
	if s.store != nil {
		source, err = s.store.Create(ctx, source)
		if err != nil {
			return Source{}, err
		}
	}
	s.sources[source.ID] = source
	return source, nil
}

func (s *Service) Update(ctx context.Context, id string, input SourceInput) (Source, error) {
	input.ID = strings.TrimSpace(id)
	source, err := NormalizeSourceInput(input)
	if err != nil {
		return Source{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if owner := subjectOwner(s.sources, source.AllowedIngressSubject, source.ID); owner != "" {
		return Source{}, fmt.Errorf("%w: %s is already owned by %s", ErrSubjectInUse, source.AllowedIngressSubject, owner)
	}
	if s.store != nil {
		source, err = s.store.Update(ctx, source)
		if err != nil {
			return Source{}, err
		}
	} else {
		_, exists := s.sources[source.ID]
		if !exists {
			return Source{}, ErrNotFound
		}
	}
	s.sources[source.ID] = source
	return source, nil
}

func subjectOwner(sources map[string]Source, subject, excludingID string) string {
	for id, source := range sources {
		if id != excludingID && source.AllowedIngressSubject == subject {
			return id
		}
	}
	return ""
}

func (s *Service) Delete(ctx context.Context, id string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return ErrNotFound
	}
	if s != nil && s.store != nil {
		if err := s.store.Delete(ctx, id); err != nil {
			return err
		}
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.sources, id)
	return nil
}

func (s *Service) Buffer(ctx context.Context, sourceID, mode, channelKey string, limit int) (BufferView, error) {
	if s == nil || s.store == nil {
		source, err := s.Get(ctx, sourceID)
		if err != nil {
			return BufferView{}, err
		}
		if mode == "" {
			mode = source.BufferMode
		}
		return BufferView{Mode: mode, Source: source}, nil
	}
	return s.store.Buffer(ctx, sourceID, mode, channelKey, limit)
}
