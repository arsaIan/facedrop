package models

import (
	"gorm.io/gorm"
)

type UserRole string

const (
	RoleHost  UserRole = "host"
	RoleGuest UserRole = "guest"
	RoleAdmin UserRole = "admin"
)

type User struct {
	gorm.Model
	Email     string   `gorm:"uniqueIndex;not null" json:"email"`
	Password  string   `gorm:"not null" json:"password"`
	FirstName string   `json:"first_name"`
	LastName  string   `json:"last_name"`
	Role      UserRole `gorm:"type:varchar(10);not null" json:"role"`
}

type UserFace struct {
	gorm.Model
	UserID uint `gorm:"not null" json:"user_id"`
	User   User `gorm:"foreignKey:UserID" json:"user"`
	FaceURL string `gorm:"not null" json:"face_id"`
}

