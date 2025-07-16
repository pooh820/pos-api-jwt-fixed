package db

import (
    "database/sql"
    _ "github.com/go-sql-driver/mysql"
)

var DB *sql.DB

func InitDB() error {
    var err error
    DB, err = sql.Open("mysql", "root:123456@tcp(127.0.0.1:3306)/finery?charset=utf8mb4&parseTime=True&loc=Local")
    if err != nil {
        return err
    }
    return DB.Ping()
}


