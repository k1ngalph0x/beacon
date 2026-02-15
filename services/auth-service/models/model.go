package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct{
	UserId    string    `gorm:"type:uuid;primaryKey" json:"id"`
	Email     string    `gorm:"unique;not null"  json:"email"`
	Password  string    `gorm:"not null" json:"-"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

type RefreshToken struct{
	Id        uint      `gorm:"primaryKey" json:"id"`
	UserId    string    `gorm:"type:uuid;not null;index" json:"user_id"`
	//Email     string    `gorm:"unique;not null"  json:"email"`
	Token     string    `gorm:"unique;not null" json:"token"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	ExpiresAt time.Time `gorm:"not null" json:"expires_at"`
	User      User      `gorm:"foreignKey:UserId;references:UserId;constraint:OnDelete:CASCADE" json:"-"`
}


type Project struct {
	ID        string    `gorm:"type:uuid;primaryKey" json:"id"`
	UserId    string    `gorm:"type:uuid;not null;index" json:"user_id"`
	Name      string    `gorm:"not null" json:"name"`
	PublicKey string    `gorm:"unique;not null" json:"public_key"`
	SecretKey string    `gorm:"unique;not null" json:"-"`
	IsActive  bool      `gorm:"default:true" json:"is_active"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	User      User      `gorm:"foreignKey:UserId;references:UserId;constraint:OnDelete:CASCADE" json:"-"`
}



func(p *Project) BeforeCreate(tx *gorm.DB) error {
	if p.ID == ""{
		p.ID = uuid.New().String()
	}
	return nil
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.UserId == "" {
		u.UserId = uuid.New().String()
	}
	return nil
}