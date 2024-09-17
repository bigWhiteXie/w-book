package model

import (
	"time"
)

type User struct {
	Id       int    `gorm:"primaryKey,autoIncrement"`
	Email    string `gorm:"unique"`
	Password string
	Ctime    time.Time
	Utime    time.Time
}
