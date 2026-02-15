package db

import (
	"fmt"

	"github.com/k1ngalph0x/beacon/services/auth-service/config"
	_ "github.com/lib/pq"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func ConnectDB() (*gorm.DB, error){

	config, err := config.LoadConfig()

	if err!=nil{
		return nil, err
	}

	conn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", config.DB.Host, config.DB.Port, config.DB.Username, config.DB.Password, config.DB.Dbname)

	db, err := gorm.Open(postgres.Open(conn), &gorm.Config{})

	if err != nil{
		return nil, err
	}

	fmt.Println("Successfully connected to Database!")

	return db, nil

}