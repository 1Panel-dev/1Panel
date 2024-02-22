package helper

import (
	"bufio"
	"database/sql"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/backend/global"
	_ "github.com/go-sql-driver/mysql"
)

func init() {}

type dumpOption struct {
	isData bool

	tables      []string
	isAllTable  bool
	isDropTable bool
	writer      io.Writer
}

type DumpOption func(*dumpOption)

func WithDropTable() DumpOption {
	return func(option *dumpOption) {
		option.isDropTable = true
	}
}

func WithData() DumpOption {
	return func(option *dumpOption) {
		option.isData = true
	}
}

func WithWriter(writer io.Writer) DumpOption {
	return func(option *dumpOption) {
		option.writer = writer
	}
}

func Dump(dns string, opts ...DumpOption) error {
	start := time.Now()
	global.LOG.Infof("dump start at %s\n", start.Format("2006-01-02 15:04:05"))
	defer func() {
		end := time.Now()
		global.LOG.Infof("dump end at %s, cost %s\n", end.Format("2006-01-02 15:04:05"), end.Sub(start))
	}()

	var err error

	var o dumpOption

	for _, opt := range opts {
		opt(&o)
	}

	if len(o.tables) == 0 {
		o.isAllTable = true
	}

	if o.writer == nil {
		o.writer = os.Stdout
	}

	buf := bufio.NewWriter(o.writer)
	defer buf.Flush()

	itemFile, lineNumber := "", 0

	itemFile += "-- ----------------------------\n"
	itemFile += "-- MySQL Database Dump\n"
	itemFile += "-- Start Time: " + start.Format("2006-01-02 15:04:05") + "\n"
	itemFile += "-- ----------------------------\n\n\n"

	db, err := sql.Open("mysql", dns)
	if err != nil {
		global.LOG.Errorf("open mysql db failed, err: %v", err)
		return err
	}
	defer db.Close()

	dbName, err := getDBNameFromDNS(dns)
	if err != nil {
		global.LOG.Errorf("get db name from dns failed, err: %v", err)
		return err
	}
	_, err = db.Exec(fmt.Sprintf("USE `%s`", dbName))
	if err != nil {
		global.LOG.Errorf("exec `use %s` failed, err: %v", dbName, err)
		return err
	}

	var tables []string
	if o.isAllTable {
		tmp, err := getAllTables(db)
		if err != nil {
			global.LOG.Errorf("get all tables failed, err: %v", err)
			return err
		}
		tables = tmp
	} else {
		tables = o.tables
	}

	for _, table := range tables {
		if o.isDropTable {
			itemFile += fmt.Sprintf("DROP TABLE IF EXISTS `%s`;\n", table)
		}

		itemFile += "-- ----------------------------\n"
		itemFile += fmt.Sprintf("-- Table structure for %s\n", table)
		itemFile += "-- ----------------------------\n"

		createTableSQL, err := getCreateTableSQL(db, table)
		if err != nil {
			global.LOG.Errorf("get create table sql failed, err: %v", err)
			return err
		}
		itemFile += createTableSQL
		itemFile += ";\n\n\n\n"

		if o.isData {
			itemFile += "-- ----------------------------\n"
			itemFile += fmt.Sprintf("-- Records of %s\n", table)
			itemFile += "-- ----------------------------\n"

			lineRows, err := db.Query(fmt.Sprintf("SELECT * FROM `%s`", table))
			if err != nil {
				global.LOG.Errorf("exec `select * from %s` failed, err: %v", table, err)
				return err
			}
			defer lineRows.Close()

			var columns []string
			columns, err = lineRows.Columns()
			if err != nil {
				global.LOG.Errorf("get columes failed, err: %v", err)
				return err
			}
			columnTypes, err := lineRows.ColumnTypes()
			if err != nil {
				global.LOG.Errorf("get colume types failed, err: %v", err)
				return err
			}
			for lineRows.Next() {
				row := make([]interface{}, len(columns))
				rowPointers := make([]interface{}, len(columns))
				for i := range columns {
					rowPointers[i] = &row[i]
				}
				if err = lineRows.Scan(rowPointers...); err != nil {
					global.LOG.Errorf("scan row data failed, err: %v", err)
					return err
				}
				ssql := loadDataSql(row, columnTypes, table)
				if len(ssql) != 0 {
					itemFile += ssql
					lineNumber++
				}
				if lineNumber > 500 {
					_, _ = buf.WriteString(itemFile)
					itemFile = ""
					lineNumber = 0
				}
			}

			itemFile += "\n\n"
		}
	}

	itemFile += "-- ----------------------------\n"
	itemFile += "-- Dumped by mysqldump\n"
	itemFile += "-- Cost Time: " + time.Since(start).String() + "\n"
	itemFile += "-- ----------------------------\n"

	_, _ = buf.WriteString(itemFile)
	_ = buf.Flush()

	return nil
}

func getCreateTableSQL(db *sql.DB, table string) (string, error) {
	var createTableSQL string
	err := db.QueryRow(fmt.Sprintf("SHOW CREATE TABLE `%s`", table)).Scan(&table, &createTableSQL)
	if err != nil {
		return "", err
	}
	createTableSQL = strings.Replace(createTableSQL, "CREATE TABLE", "CREATE TABLE IF NOT EXISTS", 1)
	return createTableSQL, nil
}

func getAllTables(db *sql.DB) ([]string, error) {
	var tables []string
	rows, err := db.Query("SHOW TABLES")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var table string
		err = rows.Scan(&table)
		if err != nil {
			return nil, err
		}
		tables = append(tables, table)
	}
	return tables, nil
}

func loadDataSql(row []interface{}, columnTypes []*sql.ColumnType, table string) string {
	ssql := "INSERT INTO `" + table + "` VALUES ("
	for i, col := range row {
		if col == nil {
			ssql += "NULL"
		} else {
			Type := columnTypes[i].DatabaseTypeName()
			Type = strings.Replace(Type, "UNSIGNED", "", -1)
			Type = strings.Replace(Type, " ", "", -1)
			switch Type {
			case "TINYINT", "SMALLINT", "MEDIUMINT", "INT", "INTEGER", "BIGINT":
				if bs, ok := col.([]byte); ok {
					ssql += string(bs)
				} else {
					ssql += fmt.Sprintf("%d", col)
				}
			case "FLOAT", "DOUBLE":
				if bs, ok := col.([]byte); ok {
					ssql += string(bs)
				} else {
					ssql += fmt.Sprintf("%f", col)
				}
			case "DECIMAL", "DEC":
				ssql += fmt.Sprintf("%s", col)

			case "DATE":
				t, ok := col.(time.Time)
				if !ok {
					global.LOG.Errorf("the DATE type conversion failed, err value: %v", col)
					return ""
				}
				ssql += fmt.Sprintf("'%s'", t.Format("2006-01-02"))
			case "DATETIME":
				t, ok := col.(time.Time)
				if !ok {
					global.LOG.Errorf("the DATETIME type conversion failed, err value: %v", col)
					return ""
				}
				ssql += fmt.Sprintf("'%s'", t.Format("2006-01-02 15:04:05"))
			case "TIMESTAMP":
				t, ok := col.(time.Time)
				if !ok {
					global.LOG.Errorf("the TIMESTAMP type conversion failed, err value: %v", col)
					return ""
				}
				ssql += fmt.Sprintf("'%s'", t.Format("2006-01-02 15:04:05"))
			case "TIME":
				t, ok := col.([]byte)
				if !ok {
					global.LOG.Errorf("the TIME type conversion failed, err value: %v", col)
					return ""
				}
				ssql += fmt.Sprintf("'%s'", string(t))
			case "YEAR":
				t, ok := col.([]byte)
				if !ok {
					global.LOG.Errorf("the YEAR type conversion failed, err value: %v", col)
					return ""
				}
				ssql += string(t)
			case "CHAR", "VARCHAR", "TINYTEXT", "TEXT", "MEDIUMTEXT", "LONGTEXT":
				ssql += fmt.Sprintf("'%s'", strings.Replace(fmt.Sprintf("%s", col), "'", "''", -1))
			case "BIT", "BINARY", "VARBINARY", "TINYBLOB", "BLOB", "MEDIUMBLOB", "LONGBLOB":
				ssql += fmt.Sprintf("0x%X", col)
			case "ENUM", "SET":
				ssql += fmt.Sprintf("'%s'", col)
			case "BOOL", "BOOLEAN":
				if col.(bool) {
					ssql += "true"
				} else {
					ssql += "false"
				}
			case "JSON":
				ssql += fmt.Sprintf("'%s'", col)
			default:
				global.LOG.Errorf("unsupported colume type: %s", Type)
				return ""
			}
		}
		if i < len(row)-1 {
			ssql += ","
		}
	}
	ssql += ");\n"
	return ssql
}
