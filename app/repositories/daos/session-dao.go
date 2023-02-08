package daos

import "time"

type Session struct {
	ID         string `gorm:"type:uuid;primary_key;"`
	Created    time.Time
	Expiration time.Time
	UserID     string `gorm:"type:uuid;"`
}
