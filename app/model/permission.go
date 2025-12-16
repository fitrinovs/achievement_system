package model

import (
	"time"
	"github.com/google/uuid"
)

type Permission struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Name        string    `json:"name" gorm:"type:varchar(100);unique;not null"`
	Resource    string    `json:"resource" gorm:"type:varchar(50);not null"`
	Action      string    `json:"action" gorm:"type:varchar(50);not null"`
	Description string    `json:"description" gorm:"type:text"`
	
	CreatedAt   time.Time    `json:"created_at" gorm:"default:now()"`
	UpdatedAt   time.Time    `json:"updated_at" gorm:"default:now()"`
}

func (Permission) TableName() string {
	return "permissions"
}

// RolePermission adalah tabel pivot many-to-many
type RolePermission struct {
	RoleID       uuid.UUID `gorm:"type:uuid;primaryKey"`
	PermissionID uuid.UUID `gorm:"type:uuid;primaryKey"`
}

func (RolePermission) TableName() string {
	return "role_permissions"
}