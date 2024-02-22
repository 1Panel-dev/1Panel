package helper

import (
	"bufio"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/backend/global"
)

type sourceOption struct {
	dryRun      bool
	mergeInsert int
	debug       bool
}
type SourceOption func(*sourceOption)

func WithMergeInsert(size int) SourceOption {
	return func(o *sourceOption) {
		o.mergeInsert = size
	}
}

type dbWrapper struct {
	DB     *sql.DB
	debug  bool
	dryRun bool
}

func newDBWrapper(db *sql.DB, dryRun, debug bool) *dbWrapper {

	return &dbWrapper{
		DB:     db,
		dryRun: dryRun,
		debug:  debug,
	}
}

func (db *dbWrapper) Exec(query string, args ...interface{}) (sql.Result, error) {
	if db.debug {
		global.LOG.Debugf("query %s", query)
	}

	if db.dryRun {
		return nil, nil
	}
	return db.DB.Exec(query, args...)
}

func Source(dns string, reader io.Reader, opts ...SourceOption) error {
	start := time.Now()
	global.LOG.Infof("source start at %s", start.Format("2006-01-02 15:04:05"))
	defer func() {
		end := time.Now()
		global.LOG.Infof("source end at %s, cost %s", end.Format("2006-01-02 15:04:05"), end.Sub(start))
	}()

	var err error
	var db *sql.DB
	var o sourceOption
	for _, opt := range opts {
		opt(&o)
	}

	dbName, err := getDBNameFromDNS(dns)
	if err != nil {
		global.LOG.Errorf("get db name from dns failed, err: %v", err)
		return err
	}

	db, err = sql.Open("mysql", dns)
	if err != nil {
		global.LOG.Errorf("open mysql db failed, err: %v", err)
		return err
	}
	defer db.Close()

	dbWrapper := newDBWrapper(db, o.dryRun, o.debug)

	_, err = dbWrapper.Exec(fmt.Sprintf("USE `%s`;", dbName))
	if err != nil {
		global.LOG.Errorf("exec `use %s` failed, err: %v", dbName, err)
		return err
	}

	db.SetConnMaxLifetime(3600)

	r := bufio.NewReader(reader)
	_, err = dbWrapper.Exec("SET autocommit=0;")
	if err != nil {
		global.LOG.Errorf("exec `set autocommit=0` failed, err: %v", err)
		return err
	}

	for {
		line, err := readLine(r)
		if err != nil {
			if err == io.EOF {
				break
			}
			global.LOG.Errorf("read sql failed, err: %v", err)
			return err
		}

		ssql, err := trim(line)
		if err != nil {
			global.LOG.Errorf("trim sql failed, err: %v", err)
			return err
		}

		afterInsertSql := ""
		if o.mergeInsert > 1 && strings.HasPrefix(ssql, "INSERT INTO") {
			var insertSQLs []string
			insertSQLs = append(insertSQLs, ssql)
			for i := 0; i < o.mergeInsert-1; i++ {
				line, err := readLine(r)
				if err != nil {
					if err == io.EOF {
						break
					}
					return err
				}
				ssql2, err := trim(line)
				if err != nil {
					global.LOG.Errorf("trim merge insert sql failed, err: %v", err)
					return err
				}
				if strings.HasPrefix(ssql2, "INSERT INTO") {
					insertSQLs = append(insertSQLs, ssql2)
					continue
				}
				afterInsertSql = ssql2
				break
			}
			ssql, err = mergeInsert(insertSQLs)
			if err != nil {
				global.LOG.Errorf("do merge insert failed, err: %v", err)
				return err
			}
		}

		_, err = dbWrapper.Exec(ssql)
		if err != nil {
			global.LOG.Errorf("exec sql failed, err: %v", err)
			return err
		}
		if len(afterInsertSql) != 0 {
			_, err = dbWrapper.Exec(afterInsertSql)
			if err != nil {
				global.LOG.Errorf("exec sql failed, err: %v", err)
				return err
			}
		}
	}

	_, err = dbWrapper.Exec("COMMIT;")
	if err != nil {
		global.LOG.Errorf("exec `commit` failed, err: %v", err)
		return err
	}

	_, err = dbWrapper.Exec("SET autocommit=1;")
	if err != nil {
		global.LOG.Errorf("exec `autocommit=1` failed, err: %v", err)
		return err
	}

	return nil
}

func mergeInsert(insertSQLs []string) (string, error) {
	if len(insertSQLs) == 0 {
		return "", errors.New("no input provided")
	}
	builder := strings.Builder{}
	sql1 := insertSQLs[0]
	sql1 = strings.TrimSuffix(sql1, ";")
	builder.WriteString(sql1)
	for i, insertSQL := range insertSQLs[1:] {
		if i < len(insertSQLs)-1 {
			builder.WriteString(",")
		}

		valuesIdx := strings.Index(insertSQL, "VALUES")
		if valuesIdx == -1 {
			return "", errors.New("invalid SQL: missing VALUES keyword")
		}
		sqln := insertSQL[valuesIdx:]
		sqln = strings.TrimPrefix(sqln, "VALUES")
		sqln = strings.TrimSuffix(sqln, ";")
		builder.WriteString(sqln)

	}
	builder.WriteString(";")

	return builder.String(), nil
}

func trim(s string) (string, error) {
	s = strings.TrimLeft(s, "\n")
	s = strings.TrimSpace(s)
	return s, nil
}

func getDBNameFromDNS(dns string) (string, error) {
	ss1 := strings.Split(dns, "/")
	if len(ss1) == 2 {
		ss2 := strings.Split(ss1[1], "?")
		if len(ss2) == 2 {
			return ss2[0], nil
		}
	}

	return "", fmt.Errorf("dns error: %s", dns)
}

func readLine(r *bufio.Reader) (string, error) {
	lineItem, err := r.ReadString('\n')
	if err != nil {
		if err == io.EOF {
			return lineItem, err
		}
		global.LOG.Errorf("read merge insert sql failed, err: %v", err)
		return "", err
	}
	if strings.HasSuffix(lineItem, ";\n") {
		return lineItem, nil
	}
	lineAppend, err := readLine(r)
	if err != nil {
		if err == io.EOF {
			return lineItem, err
		}
		global.LOG.Errorf("read merge insert sql failed, err: %v", err)
		return "", err
	}

	return lineItem + lineAppend, nil
}
