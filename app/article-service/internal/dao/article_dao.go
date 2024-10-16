package dao

import (
	"time"

	"gorm.io/gorm"
)

type Article struct {
	Id       int `gorm:"primaryKey,autoIncrement"`
	Title    string
	Content  string `gorm:"type:blob"`
	AuthorId int    `gorm:"index:idx_uid_uptime"`
	Ctime    time.Time
	Utime    time.Time `gorm:"index:idx_uid_uptime"`
}

type ArticleDao struct {
	// go get github.com/DATA-DOG/go-sqlmock
	db *gorm.DB
}

func NewArticleDao(db *gorm.DB) *ArticleDao {
	return &ArticleDao{db: db}
}
