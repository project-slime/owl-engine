package database

import (
	"fmt"
	"owl-engine/pkg/config"
	"owl-engine/pkg/xlogs"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

var DB *gorm.DB

func Setup(conf *config.MySQLOptions) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		conf.Username, conf.Password, conf.Host, conf.Port, conf.DBName)

	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
		SkipDefaultTransaction: true, // 禁用默认事务
	})

	if err != nil {
		xlogs.Fatalf("database client initialize fail, %s", err.Error())
	}

	db, err := DB.DB()
	if err != nil {
		xlogs.Fatalf("get database client instance error, %s", err.Error())
	}

	db.SetMaxIdleConns(conf.MaxIdleConnections)
	db.SetMaxOpenConns(conf.MaxOpenConnections)
	db.SetConnMaxLifetime(conf.MaxConnectionLifeTime)

	xlogs.Info("database client initialize success")
}
