package dao

import (
	"codexie.com/w-book-user/internal/model"
	"gorm.io/gorm"
)

func InitTables(db *gorm.DB) error {
	if err := db.AutoMigrate(&model.User{}); err != nil {
		return err
	}

	return nil
}
