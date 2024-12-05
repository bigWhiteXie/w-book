package db

import "gorm.io/gorm"

func InitTables(db *gorm.DB) error {
	if err := db.AutoMigrate(&Article{}); err != nil {
		return err
	}
	if err := db.AutoMigrate(&PublishedArticle{}); err != nil {
		return err
	}
	return nil
}
