package dao

import (
	"context"
	"fmt"
	"testing"

	"codexie.com/w-book-common/ioc"
	"codexie.com/w-book-common/sql"
	"codexie.com/w-book-user/internal/config"
	"codexie.com/w-book-user/internal/model"

	"github.com/zeromicro/go-zero/core/conf"
	"gorm.io/gorm"
)

func initDb() *gorm.DB {
	var c config.Config
	conf.MustLoad("/usr/local/go_project/w-book/app/user-service/etc/user.yaml", &c)
	return ioc.InitGormDB(c.MySQLConf)
}

func TestFindOne(t *testing.T) {
	userDao := NewUserDao(initDb())

	one, err := userDao.FindOne(context.Background(), &model.User{Email: sql.StringToNullString("2607219580@qq.com")})
	fmt.Printf("user %v, err:%s", one, err)
	one2, err2 := userDao.FindOne(context.Background(), &model.User{Email: sql.StringToNullString("2107219580@qq.com")})
	fmt.Printf("user %v, err:%s", one2, err2)
}
