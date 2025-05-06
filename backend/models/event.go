package models

import (
	"time"

	"gorm.io/gorm"
)
func (p *Photo) AfterFind(tx *gorm.DB) (err error) {
	// Create a new Photo with only URL and EventID
	stripped := Photo{
		URL:     p.URL,
		EventID: p.EventID,
	}
	// Copy the stripped version back to the original
	*p = stripped
	return nil
}
type EventStatus string

const (
	EventStatusCreated EventStatus = "created"
	EventStatusInactive EventStatus = "inactive"
	EventStatusReady   EventStatus = "ready"
)

type EventSubscriberStatus string

const (
	Subscribed EventSubscriberStatus = "subscribed"
	Pending EventSubscriberStatus = "pending"
	Delivered EventSubscriberStatus = "delivered"
	Failed EventSubscriberStatus = "failed"
)

type Event struct {
	gorm.Model
	Title       string    `gorm:"not null" json:"title"`
	Description string    `json:"description"`
	CreatedAt   time.Time `gorm:"not null" json:"created_at"`
	EndDate     time.Time `json:"end_date"`
	CreatedBy   uint      `gorm:"not null" json:"created_by"`
	Creator     User      `gorm:"foreignKey:CreatedBy" json:"creator"`
	Subscribers []User    `gorm:"many2many:event_subscribers;" json:"subscribers,omitempty"`
	Photos      []Photo   `gorm:"foreignKey:EventID" json:"photos,omitempty"`
	QRCode      string    `json:"qr_code"`
	Status      EventStatus `gorm:"default:created" json:"status"`
}

type Photo struct {
	gorm.Model
	URL         string    `gorm:"not null" json:"url"`
	EventID     uint      `gorm:"not null" json:"event_id"`
	Event       Event     `gorm:"foreignKey:EventID" json:"event"`
	UploadedBy  uint      `gorm:"not null" json:"uploaded_by"`
	Uploader    User      `gorm:"foreignKey:UploadedBy" json:"uploader"`
	UploadedAt  time.Time `gorm:"not null" json:"uploaded_at"`
	Description string    `json:"description"`
	IsPublic    bool      `gorm:"default:true" json:"is_public"`
}

type EventSubscriber struct {
	EventID      uint      `gorm:"primaryKey" json:"event_id"`
	UserID       uint      `gorm:"primaryKey" json:"user_id"`
	Status       string    `gorm:"default:subscribed" json:"status"`
	User         User      `gorm:"foreignKey:UserID" json:"user"`
	SubscribedAt time.Time `gorm:"not null" json:"subscribed_at"`
} 

type EventMatch struct {
	gorm.Model
	EventID uint `gorm:"not null" json:"event_id"`
	UserID  uint `gorm:"not null" json:"user_id"`
	PhotoURL string `gorm:"not null" json:"photo_url"`
}
