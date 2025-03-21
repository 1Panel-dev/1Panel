package middleware

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path"
	"reflect"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/core/app/model"
	"github.com/1Panel-dev/1Panel/core/app/repo"
	"github.com/1Panel-dev/1Panel/core/cmd/server/docs"
	"github.com/1Panel-dev/1Panel/core/constant"
	"github.com/1Panel-dev/1Panel/core/global"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func OperationLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		if strings.Contains(c.Request.URL.Path, "search") || c.Request.Method == http.MethodGet {
			c.Next()
			return
		}

		source := loadLogInfo(c.Request.URL.Path)
		pathItem := strings.TrimPrefix(c.Request.URL.Path, "/api/v2")
		pathItem = strings.TrimPrefix(pathItem, "/api/v2/core")
		currentNode := c.Request.Header.Get("CurrentNode")
		record := &model.OperationLog{
			Source:    source,
			Node:      currentNode,
			IP:        c.ClientIP(),
			Method:    strings.ToLower(c.Request.Method),
			Path:      pathItem,
			UserAgent: c.Request.UserAgent(),
		}
		swagger := make(map[string]operationJson)
		if err := json.Unmarshal(docs.XLogJson, &swagger); err != nil {
			c.Next()
			return
		}
		operationDic, hasPath := swagger[record.Path]
		if !hasPath {
			c.Next()
			return
		}
		if len(operationDic.FormatZH) == 0 {
			c.Next()
			return
		}

		formatMap := make(map[string]interface{})
		if len(operationDic.BodyKeys) != 0 {
			body, err := io.ReadAll(c.Request.Body)
			if err == nil {
				c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
			}
			bodyMap := make(map[string]interface{})
			_ = json.Unmarshal(body, &bodyMap)
			for _, key := range operationDic.BodyKeys {
				if _, ok := bodyMap[key]; ok {
					formatMap[key] = bodyMap[key]
				}
			}
		}
		if len(operationDic.BeforeFunctions) != 0 {
			dbItem, err := newDB(record.Path)
			if err != nil {
				c.Next()
				return
			}
			for _, funcs := range operationDic.BeforeFunctions {
				for key, value := range formatMap {
					if funcs.InputValue == key {
						var names []string
						if funcs.IsList {
							sql := fmt.Sprintf("SELECT %s FROM %s where %s in (?);", funcs.OutputColumn, funcs.DB, funcs.InputColumn)
							_ = dbItem.Raw(sql, value).Scan(&names)
						} else {
							_ = dbItem.Raw(fmt.Sprintf("select %s from %s where %s = ?;", funcs.OutputColumn, funcs.DB, funcs.InputColumn), value).Scan(&names)
						}
						formatMap[funcs.OutputValue] = strings.Join(names, ",")
						break
					}
				}
			}
			closeDB(dbItem)
		}
		for key, value := range formatMap {
			if strings.Contains(operationDic.FormatEN, "["+key+"]") {
				t := reflect.TypeOf(value)
				if t.Kind() != reflect.Array && t.Kind() != reflect.Slice {
					operationDic.FormatZH = strings.ReplaceAll(operationDic.FormatZH, "["+key+"]", fmt.Sprintf("[%v]", value))
					operationDic.FormatEN = strings.ReplaceAll(operationDic.FormatEN, "["+key+"]", fmt.Sprintf("[%v]", value))
				} else {
					val := reflect.ValueOf(value)
					length := val.Len()

					var elements []string
					for i := 0; i < length; i++ {
						element := val.Index(i).Interface().(string)
						elements = append(elements, element)
					}
					operationDic.FormatZH = strings.ReplaceAll(operationDic.FormatZH, "["+key+"]", fmt.Sprintf("[%v]", strings.Join(elements, ",")))
					operationDic.FormatEN = strings.ReplaceAll(operationDic.FormatEN, "["+key+"]", fmt.Sprintf("[%v]", strings.Join(elements, ",")))
				}
			}
		}
		record.DetailEN = strings.ReplaceAll(operationDic.FormatEN, "[]", "")
		record.DetailZH = strings.ReplaceAll(operationDic.FormatZH, "[]", "")

		writer := responseBodyWriter{
			ResponseWriter: c.Writer,
			body:           &bytes.Buffer{},
		}
		c.Writer = writer
		now := time.Now()

		c.Next()

		datas := writer.body.Bytes()
		logRepo := repo.NewILogRepo()
		if c.Request.Header.Get("Content-Encoding") == "gzip" {
			buf := bytes.NewReader(writer.body.Bytes())
			reader, err := gzip.NewReader(buf)
			if err != nil {
				record.Status = constant.StatusFailed
				record.Message = fmt.Sprintf("gzip new reader failed, err: %v", err)
				latency := time.Since(now)
				record.Latency = latency

				if err := logRepo.CreateOperationLog(record); err != nil {
					global.LOG.Errorf("create operation record failed, err: %v", err)
				}
				return
			}
			defer reader.Close()
			datas, _ = io.ReadAll(reader)
		}
		var res response
		_ = json.Unmarshal(datas, &res)
		if res.Code == 200 {
			record.Status = constant.StatusSuccess
		} else {
			record.Status = constant.StatusFailed
			record.Message = res.Message
		}

		latency := time.Since(now)
		record.Latency = latency

		if err := logRepo.CreateOperationLog(record); err != nil {
			global.LOG.Errorf("create operation record failed, err: %v", err)
		}
	}
}

type operationJson struct {
	API             string         `json:"api"`
	Method          string         `json:"method"`
	BodyKeys        []string       `json:"bodyKeys"`
	ParamKeys       []string       `json:"paramKeys"`
	BeforeFunctions []functionInfo `json:"beforeFunctions"`
	FormatZH        string         `json:"formatZH"`
	FormatEN        string         `json:"formatEN"`
}
type functionInfo struct {
	InputColumn  string `json:"input_column"`
	InputValue   string `json:"input_value"`
	IsList       bool   `json:"isList"`
	DB           string `json:"db"`
	OutputColumn string `json:"output_column"`
	OutputValue  string `json:"output_value"`
}

type response struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type responseBodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (r responseBodyWriter) Write(b []byte) (int, error) {
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}

func loadLogInfo(path string) string {
	path = replaceStr(path, "/api/v2", "/core", "/xpack")
	if !strings.Contains(path, "/") {
		return ""
	}
	pathArrays := strings.Split(path, "/")
	if len(pathArrays) < 2 {
		return ""
	}
	return pathArrays[1]
}

func newDB(pathItem string) (*gorm.DB, error) {
	dbFile := ""
	switch {
	case strings.HasPrefix(pathItem, "/core"):
		dbFile = path.Join(global.CONF.Base.InstallDir, "1panel/db/core.db")
	case strings.HasPrefix(pathItem, "/xpack"):
		dbFile = path.Join(global.CONF.Base.InstallDir, "1panel/db/xpack.db")
	default:
		dbFile = path.Join(global.CONF.Base.InstallDir, "1panel/db/agent.db")
	}

	db, _ := gorm.Open(sqlite.Open(dbFile), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetConnMaxIdleTime(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)
	return db, nil
}

func closeDB(db *gorm.DB) {
	sqlDB, err := db.DB()
	if err != nil {
		return
	}
	_ = sqlDB.Close()
}

func replaceStr(val string, rep ...string) string {
	for _, item := range rep {
		val = strings.ReplaceAll(val, item, "")
	}
	return val
}
