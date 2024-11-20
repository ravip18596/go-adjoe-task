package main

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/sirupsen/logrus"
	"time"
)

func InitMysql() *sql.DB {
	db, err := sql.Open("mysql", "adjoe:adjoe@tcp(mysql:3306)/adjoe")
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error":    err.Error(),
			"database": "mysql",
		}).Error("Error connecting to database adjoe")
		return nil
	}
	// See "Important settings" section.
	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)
	return db
}

var mysql = InitMysql()
