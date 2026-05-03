package server

import (
	"sync"
	"time"
)

type GroupBusSnapshot struct {
	Status          string   `json:"status"`
	PublishedCount  int64    `json:"published_count"`
	LastGroupID     string   `json:"last_group_id,omitempty"`
	LastActorID     string   `json:"last_actor_id,omitempty"`
	LastMessage     string   `json:"last_message,omitempty"`
	LastSubjects    []string `json:"last_subjects,omitempty"`
	LastPublishedAt string   `json:"last_published_at,omitempty"`
	LastError       string   `json:"last_error,omitempty"`
}

// GroupBusMonitor tracks recent group-bus fanout activity for status surfaces.
type GroupBusMonitor struct {
	mu             sync.RWMutex
	publishedCount int64
	lastGroupID    string
	lastActorID    string
	lastMessage    string
	lastSubjects   []string
	lastPublished  time.Time
	lastError      string
}

func NewGroupBusMonitor() *GroupBusMonitor {
	return &GroupBusMonitor{}
}

func (m *GroupBusMonitor) RecordSuccess(groupID, actorID, message string, subjects []string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.publishedCount++
	m.lastGroupID = groupID
	m.lastActorID = actorID
	m.lastMessage = message
	m.lastSubjects = append([]string(nil), subjects...)
	m.lastPublished = time.Now().UTC()
	m.lastError = ""
}

func (m *GroupBusMonitor) RecordFailure(err error) {
	if err == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.lastError = err.Error()
}

func (m *GroupBusMonitor) Snapshot(natsOnline bool) GroupBusSnapshot {
	m.mu.RLock()
	defer m.mu.RUnlock()

	s := GroupBusSnapshot{
		Status:         "offline",
		PublishedCount: m.publishedCount,
		LastGroupID:    m.lastGroupID,
		LastActorID:    m.lastActorID,
		LastMessage:    m.lastMessage,
		LastSubjects:   append([]string(nil), m.lastSubjects...),
		LastError:      m.lastError,
	}
	if !m.lastPublished.IsZero() {
		s.LastPublishedAt = m.lastPublished.Format(time.RFC3339Nano)
	}
	if natsOnline {
		s.Status = "online"
	}
	if natsOnline && m.lastError != "" {
		s.Status = "degraded"
	}
	return s
}
