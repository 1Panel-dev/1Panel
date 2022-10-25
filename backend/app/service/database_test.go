package service

import (
	"database/sql"
	"fmt"
	"reflect"
	"testing"

	"github.com/1Panel-dev/1Panel/backend/app/dto"
	_ "github.com/go-sql-driver/mysql"
)

func TestMysql(t *testing.T) {
	connArgs := fmt.Sprintf("%s:%s@tcp(%s:%d)/?charset=utf8", "root", "Calong@2015", "172.16.10.143", 3306)
	db, err := sql.Open("mysql", connArgs)
	if err != nil {
		fmt.Println(err)
	}
	defer db.Close()

	rows, err := db.Query("show variables")
	if err != nil {
		fmt.Println(err)
	}
	variableMap := make(map[string]int)

	for rows.Next() {
		var variableName string
		var variableValue int
		if err := rows.Scan(&variableName, &variableValue); err != nil {
			fmt.Println(err)
		}
		variableMap[variableName] = variableValue
	}
	for k, v := range variableMap {
		fmt.Println(k, v)
	}
}

func TestMs(t *testing.T) {
	db, err := newDatabaseClient()
	if err != nil {
		fmt.Println(err)
	}
	defer db.Close()

	variables := dto.MysqlVariablesUpdate{
		Version:              "5.7.39",
		KeyBufferSize:        8388608,
		QueryCacheSize:       1048576,
		TmpTableSize:         16777216,
		InnodbBufferPoolSize: 134217728,
		InnodbLogBufferSize:  16777216,
		SortBufferSize:       262144,
		ReadBufferSize:       131072,

		ReadRndBufferSize: 262144,
		JoinBufferSize:    262144,
		ThreadStack:       262144,
		BinlogCachSize:    32768,
		ThreadCacheSize:   9,
		TableOpenCache:    2000,
		MaxConnections:    150,
	}

	v := reflect.ValueOf(variables)
	typeOfS := v.Type()
	for i := 0; i < v.NumField(); i++ {
		fmt.Printf("SET GLOBAL %s=%d \n", typeOfS.Field(i).Name, v.Field(i).Interface())
	}
}
