package model

import (
	"database/sql"
	"time"
)

type User struct {
	Id       int            `gorm:"primaryKey,autoIncrement"`
	Email    sql.NullString `gorm:"unique"`
	Phone    sql.NullString `gorm:"unique"`
	Password string
	Ctime    time.Time
	Utime    time.Time
}
