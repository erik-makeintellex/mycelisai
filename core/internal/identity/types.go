package identity

import (
	"time"

	"github.com/google/uuid"
)

// Mission represents the overarching goal of a Swarm instance.
type Mission struct {
	ID        uuid.UUID `json:"id" db:"id"`
	OwnerID   uuid.UUID `json:"owner_id" db:"owner_id"`
	Name      string    `json:"name" db:"name"`
	Directive string    `json:"directive" db:"directive"` // The "Why"
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// Team represents a node in the command hierarchy.
type Team struct {
	ID        uuid.UUID  `json:"id" db:"id"`
	MissionID uuid.UUID  `json:"mission_id" db:"mission_id"`         // Link to Root
	ParentID  *uuid.UUID `json:"parent_id,omitempty" db:"parent_id"` // Pointer allows NULL (Root)
	Name      string     `json:"name" db:"name"`
	Path      string     `json:"path" db:"path"` // Helper for frontend breadcrumbs

	// Virtual Field for Recursive JSON response
	Children  []Team    `json:"children,omitempty" db:"-"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}
