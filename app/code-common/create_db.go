package main

import (
	"fmt"
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func createDatabase(host string, port int, user, password, dbName string) error {
	// 构建不指定数据库的连接 DSN
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/?charset=utf8mb4&parseTime=True&loc=Local", user, password, host, port)

	// 连接到数据库服务器
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("failed to connect to database server: %w", err)
	}

	// 获取原生 SQL DB 对象
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get raw database connection: %w", err)
	}
	defer sqlDB.Close()

	// 构建创建数据库的 SQL 语句
	createDBSQL := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s` CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;", dbName)

	// 执行创建数据库的 SQL 语句
	if err := db.Exec(createDBSQL).Error; err != nil {
		return fmt.Errorf("failed to create database: %w", err)
	}

	log.Printf("Database `%s` created successfully.", dbName)
	return nil
}

func main() {
	host := "127.0.0.1"
	port := 3306
	user := "root"
	password := "j3391111"
	dbName := "test_db"

	if err := createDatabase(host, port, user, password, dbName); err != nil {
		log.Fatalf("Error: %v", err)
	}
}
