package middleware

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/1Panel-dev/1Panel/agent/cmd/server/docs"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/re"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

const (
	headerNeedOperationResolve = "X-Need-Op-Resolve"
	headerOperationResolved    = "X-Op-Resolved"
)

var (
	logMetaOnce    sync.Once
	logMetaData    map[string]operationMeta
	logMetaLoadErr error
)

type operationMeta struct {
	BodyKeys        []string       `json:"bodyKeys"`
	BeforeFunctions []functionInfo `json:"beforeFunctions"`
}

type functionInfo struct {
	InputColumn  string `json:"input_column"`
	InputValue   string `json:"input_value"`
	IsList       bool   `json:"isList"`
	DB           string `json:"db"`
	OutputColumn string `json:"output_column"`
	OutputValue  string `json:"output_value"`
}

func OperationResolveMeta() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetHeader(headerNeedOperationResolve) != "1" {
			c.Next()
			return
		}

		metaMap, err := loadOperationMeta()
		if err != nil {
			c.Next()
			return
		}
		reqPath := strings.TrimPrefix(c.Request.URL.Path, "/api/v2")
		meta, ok := metaMap[reqPath]
		if !ok || len(meta.BeforeFunctions) == 0 {
			c.Next()
			return
		}

		values := make(map[string]interface{})
		if len(meta.BodyKeys) > 0 {
			body, err := io.ReadAll(c.Request.Body)
			if err == nil {
				c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
				bodyMap := make(map[string]interface{})
				if strings.Contains(c.Request.Header.Get("Content-Type"), "multipart/form-data") {
					bodyMap, _ = parseMultipart(body, c.Request.Header.Get("Content-Type"))
				} else {
					decoder := json.NewDecoder(bytes.NewReader(body))
					decoder.UseNumber()
					_ = decoder.Decode(&bodyMap)
				}
				for _, key := range meta.BodyKeys {
					if value, ok := bodyMap[key]; ok {
						values[key] = value
					}
				}
			}
		}

		resolved, err := resolveOperationValues(reqPath, values, meta.BeforeFunctions)
		if err == nil && len(resolved) > 0 {
			if data, err := json.Marshal(resolved); err == nil {
				c.Header(headerOperationResolved, base64.RawURLEncoding.EncodeToString(data))
			}
		}
		c.Next()
	}
}

func loadOperationMeta() (map[string]operationMeta, error) {
	logMetaOnce.Do(func() {
		logMetaData = make(map[string]operationMeta)
		logMetaLoadErr = json.Unmarshal(docs.XLogJson, &logMetaData)
	})
	return logMetaData, logMetaLoadErr
}

func parseMultipart(formData []byte, contentType string) (map[string]interface{}, error) {
	d, params, err := mime.ParseMediaType(contentType)
	if err != nil || d != "multipart/form-data" {
		return nil, http.ErrNotMultipart
	}
	boundary, ok := params["boundary"]
	if !ok {
		return nil, http.ErrMissingBoundary
	}
	reader := multipart.NewReader(bytes.NewReader(formData), boundary)
	ret := make(map[string]interface{})

	f, err := reader.ReadForm(32 << 20)
	if err != nil {
		return nil, err
	}
	for k, v := range f.Value {
		if len(v) > 0 {
			ret[k] = v[0]
		}
	}
	for k, v := range f.File {
		if len(v) > 0 {
			ret[k] = v[0].Filename
		}
	}
	return ret, nil
}

func resolveOperationValues(pathItem string, values map[string]interface{}, beforeFunctions []functionInfo) (map[string]string, error) {
	dbItem, err := newResolveDB(pathItem)
	if err != nil {
		return nil, err
	}
	defer closeResolveDB(dbItem)

	resolved := make(map[string]string)
	for _, funcs := range beforeFunctions {
		if !isSafeIdentifier(funcs.DB) || !isSafeIdentifier(funcs.InputColumn) || !isSafeIdentifier(funcs.OutputColumn) {
			continue
		}
		for key, value := range values {
			if funcs.InputValue != key {
				continue
			}
			var names []string
			if funcs.IsList {
				sql := fmt.Sprintf("SELECT %s FROM %s where %s in (?);", funcs.OutputColumn, funcs.DB, funcs.InputColumn)
				_ = dbItem.Raw(sql, value).Scan(&names)
			} else {
				sql := fmt.Sprintf("SELECT %s FROM %s where %s = ?;", funcs.OutputColumn, funcs.DB, funcs.InputColumn)
				_ = dbItem.Raw(sql, value).Scan(&names)
			}
			outputValue := strings.Join(names, ",")
			resolved[funcs.OutputValue] = outputValue
			values[funcs.OutputValue] = outputValue
			break
		}
	}
	return resolved, nil
}

func newResolveDB(pathItem string) (*gorm.DB, error) {
	dbFile := ""
	switch {
	case strings.HasPrefix(pathItem, "/core"):
		dbFile = path.Join(global.CONF.Base.InstallDir, "1panel/db/core.db")
	case strings.HasPrefix(pathItem, "/xpack"):
		dbFile = path.Join(global.CONF.Base.InstallDir, "1panel/db/xpack.db")
	default:
		dbFile = path.Join(global.CONF.Base.InstallDir, "1panel/db/agent.db")
	}

	db, err := gorm.Open(sqlite.Open(dbFile), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		return nil, err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(4)
	sqlDB.SetMaxIdleConns(1)
	sqlDB.SetConnMaxIdleTime(15 * time.Minute)
	sqlDB.SetConnMaxLifetime(time.Hour)
	return db, nil
}

func closeResolveDB(db *gorm.DB) {
	sqlDB, err := db.DB()
	if err != nil {
		return
	}
	_ = sqlDB.Close()
}

func isSafeIdentifier(val string) bool {
	return re.GetRegex(re.SQLIdentifierPattern).MatchString(val)
}
