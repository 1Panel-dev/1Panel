package service

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"testing"
	"time"

	_ "gorm.io/driver/mysql"
)

func TestMysql(t *testing.T) {
	connArgs := fmt.Sprintf("%s:%s@tcp(%s:%d)/?charset=utf8", "root", "Calong@2015", "localhost", 2306)
	db, err := sql.Open("mysql", connArgs)
	if err != nil {
		fmt.Println(err)
	}
	defer db.Close()
	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()
	res, err := db.ExecContext(ctx, "CREATE DATABASE IF NOT EXISTS "+"songli")
	if err != nil {
		log.Printf("Error %s when creating DB\n", err)
		return
	}

	// res, err := db.Exec("SHOW DATABASES;")
	// if err != nil {
	// 	fmt.Println(err)
	// }
	fmt.Println(res)
}
