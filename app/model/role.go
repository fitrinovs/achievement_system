// File: app/model/role.go
package model

import (
	"time"

	"github.com/google/uuid"
	// Import model Permission jika diletakkan di package model yang sama
)

type Role struct {
	ID          uuid.UUID    `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Name        string       `json:"name" gorm:"type:varchar(50);unique;not null"`
	Description string       `json:"description" gorm:"type:text"`
	
	// Gunakan nama package jika Permission berada di package berbeda
	Permissions []Permission `json:"permissions,omitempty" gorm:"many2many:role_permissions;"`
	
	CreatedAt   time.Time    `json:"created_at" gorm:"default:now()"`
}

func (Role) TableName() string {
	return "roles"
}