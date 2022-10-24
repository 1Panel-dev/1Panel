package service

import (
	"database/sql"
	"fmt"
	"testing"

	_ "github.com/go-sql-driver/mysql"
)

func TestMysql(t *testing.T) {
	connArgs := fmt.Sprintf("%s:%s@tcp(%s:%d)/?charset=utf8", "root", "Calong@2015", "172.16.10.143", 3306)
	db, err := sql.Open("mysql", connArgs)
	if err != nil {
		fmt.Println(err)
	}
	defer db.Close()

	rows, err := db.Query("show master status")
	if err != nil {
		fmt.Println(err)
	}

	masterRows := make([]string, 5)
	for rows.Next() {
		if err := rows.Scan(&masterRows[0], &masterRows[1], &masterRows[2], &masterRows[3], &masterRows[4]); err != nil {
			fmt.Println(err)
		}
	}
	for k, v := range masterRows {
		fmt.Println(k, v)
	}
}
