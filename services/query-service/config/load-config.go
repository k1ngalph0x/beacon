package config

import (
	"os"

	"github.com/joho/godotenv"
)

// type Config struct{
// 	DB *sql.DB
// }

type Config struct {
	DB PostgresConfig
	TOKEN TokenConfig
}

type TokenConfig struct{
	JwtKey string

}

type PostgresConfig struct {
	Host string
	Dbname string
	Username string
	Password string
	Url      string
	Port     string
}

func LoadConfig() (*Config, error) {

	err := godotenv.Load()

	if err!=nil{
		return nil, err
	}

	config := &Config{
		DB: PostgresConfig{
			Host: os.Getenv("DB_HOST"),
			Username: os.Getenv("DB_USERNAME"),
			Password: os.Getenv("DB_PASSWORD"),
			Url: os.Getenv("DB_URL"),
			Port: os.Getenv("DB_PORT"),
			Dbname: os.Getenv("DB_NAME"),
		},

		TOKEN: TokenConfig{
			JwtKey: os.Getenv("JwtKey"),
		},
	}

	return config, nil
}