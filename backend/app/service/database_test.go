package service

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/1Panel-dev/1Panel/backend/app/dto"
	_ "github.com/go-sql-driver/mysql"
)

func TestMysql(t *testing.T) {
	connArgs := fmt.Sprintf("%s:%s@tcp(%s:%d)/?charset=utf8", "root", "Calong@2015", "localhost", 2306)
	db, err := sql.Open("mysql", connArgs)
	if err != nil {
		fmt.Println(err)
	}
	defer db.Close()

	rows, err := db.Query("show VARIABLES")
	if err != nil {
		fmt.Println(err)
	}

	variableMap := make(map[string]string)
	for rows.Next() {
		var variableName, variableValue string
		if err := rows.Scan(&variableName, &variableValue); err != nil {
			fmt.Println(err)
		}
		variableMap[variableName] = variableValue
	}
	var info dto.MysqlConf
	arr, err := json.Marshal(variableMap)
	if err != nil {
		fmt.Println(err)
	}
	err = json.Unmarshal(arr, &info)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(info)
}
