package repositories

import (
	"errors"

	"gorm.io/gorm"
)

func convertNotFound(err error) error {
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrNoRecord
	}

	return err
}
