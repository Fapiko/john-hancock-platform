package daos

import "time"

type Certificate struct {
	ID      string `gorm:"type:uuid;primary_key;"`
	UserID  string
	Name    string
	Data    []byte
	Type    string
	Created time.Time
}
