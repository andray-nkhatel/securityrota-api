package models

import "time"

// OfficerRole defines the role of an officer
type OfficerRole string

const (
	RoleSergeant OfficerRole = "sergeant"
	RoleFemale   OfficerRole = "female"
	RoleRegular  OfficerRole = "regular"
)

// Officer represents a security officer
type Officer struct {
	ID        uint        `json:"id" gorm:"primaryKey"`
	Name      string      `json:"name" gorm:"uniqueIndex;not null"`
	Role      OfficerRole `json:"role" gorm:"not null;default:'regular'"`
	Team      int         `json:"team" gorm:"not null"` // 1 or 2 for rotation teams
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
}
