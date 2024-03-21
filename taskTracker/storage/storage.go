package storage

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"popug_tasktracker/model"
)

func NewDB() *gorm.DB {
	postgresDNS := "host=localhost user=base password=secret dbname=task port=5443 sslmode=disable"
	db, err := gorm.Open(postgres.Open(postgresDNS), &gorm.Config{})
	if err != nil {
		fmt.Println("Failed to connect to postgres: ", err.Error())
		return nil
	}

	_ = db.AutoMigrate(&model.Task{})

	return db
}
