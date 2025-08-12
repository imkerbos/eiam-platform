package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// BaseModel base model
type BaseModel struct {
	ID        string         `json:"id" gorm:"type:varchar(36);primaryKey;not null"`
	CreatedAt time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

// BeforeCreate GORM hook, generate UUID before creation
func (base *BaseModel) BeforeCreate(tx *gorm.DB) error {
	if base.ID == "" {
		base.ID = uuid.New().String()
	}
	return nil
}

// Status status enum
type Status int

const (
	StatusInactive Status = 0 // inactive
	StatusActive   Status = 1 // active
	StatusLocked   Status = 2 // locked
	StatusExpired  Status = 3 // expired
)

func (s Status) String() string {
	switch s {
	case StatusInactive:
		return "inactive"
	case StatusActive:
		return "active"
	case StatusLocked:
		return "locked"
	case StatusExpired:
		return "expired"
	default:
		return "unknown"
	}
}

// Gender gender enum
type Gender int

const (
	GenderUnknown Gender = 0 // unknown
	GenderMale    Gender = 1 // male
	GenderFemale  Gender = 2 // female
)

func (g Gender) String() string {
	switch g {
	case GenderMale:
		return "male"
	case GenderFemale:
		return "female"
	default:
		return "unknown"
	}
}
