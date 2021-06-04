package database

import (
	"github.com/Sansui233/proxypool/log"
	"os"

	"github.com/Sansui233/proxypool/config"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func connect() (err error) {
	// localhost url
	dsn := "user=proxypool password=proxypool dbname=proxypool port=5432 sslmode=disable TimeZone=Asia/Shanghai"
	if url := config.Config.DatabaseUrl; url != "" {
		dsn = url
	}
	if url := os.Getenv("DATABASE_URL"); url != "" {
		dsn = url
	}
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err == nil {
		log.Infoln("database: successfully connected to: %s", DB.Name())
	} else {
		DB = nil
		log.Warnln("database connection info: %s \n\t\tUse cache to store proxies", err.Error())
	}
	return
}
