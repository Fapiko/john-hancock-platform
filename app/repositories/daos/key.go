package daos

import "time"

type Key struct {
	ID        string `gorm:"type:uuid;primary_key;"`
	UserID    string
	Name      string
	Data      []byte
	Algorithm string
	Created   time.Time
}
