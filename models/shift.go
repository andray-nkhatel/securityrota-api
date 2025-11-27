package models

import "time"

// ShiftType defines day or night shift
type ShiftType string

const (
	ShiftDay   ShiftType = "day"
	ShiftNight ShiftType = "night"
)

// DutyStatus defines officer's duty status
type DutyStatus string

const (
	StatusOnDuty  DutyStatus = "on_duty"
	StatusOffDuty DutyStatus = "off_duty"
)

// Shift represents a duty assignment for an officer on a specific date
type Shift struct {
	ID        uint       `json:"id" gorm:"primaryKey"`
	OfficerID uint       `json:"officer_id" gorm:"not null;index"`
	Officer   Officer    `json:"officer" gorm:"foreignKey:OfficerID"`
	Date      time.Time  `json:"date" gorm:"not null;index"`
	ShiftType ShiftType  `json:"shift_type" gorm:"not null"`
	Status    DutyStatus `json:"status" gorm:"not null"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// WeekRotation tracks which team is on which shift for a given week
type WeekRotation struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	WeekStart    time.Time `json:"week_start" gorm:"uniqueIndex;not null"` // Sunday of the week
	DayShiftTeam int       `json:"day_shift_team" gorm:"not null"`         // Team 1 or 2
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
