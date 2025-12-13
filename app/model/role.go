package model

import (
	"time"

	"github.com/google/uuid"
)

type Role struct {
	ID          uuid.UUID    `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Name        string       `json:"name" gorm:"type:varchar(50);unique;not null"`
	Description string       `json:"description" gorm:"type:text"`
	Permissions []Permission `json:"permissions,omitempty" gorm:"many2many:role_permissions;"`
	CreatedAt   time.Time    `json:"created_at" gorm:"default:now()"`
}

func (Role) TableName() string {
	return "roles"
}

type Permission struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Name        string    `json:"name" gorm:"type:varchar(100);unique;not null"`
	Resource    string    `json:"resource" gorm:"type:varchar(50);not null"`
	Action      string    `json:"action" gorm:"type:varchar(50);not null"`
	Description string    `json:"description" gorm:"type:text"`
}

func (Permission) TableName() string {
	return "permissions"
}

type RolePermission struct {
	RoleID       uuid.UUID `gorm:"type:uuid;primaryKey"`
	PermissionID uuid.UUID `gorm:"type:uuid;primaryKey"`
}

func (RolePermission) TableName() string {
	return "role_permissions"
}
