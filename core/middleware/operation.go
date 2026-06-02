package middleware

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/url"
	"path"
	"reflect"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/core/app/model"
	"github.com/1Panel-dev/1Panel/core/app/repo"
	"github.com/1Panel-dev/1Panel/core/cmd/server/docs"
	"github.com/1Panel-dev/1Panel/core/constant"
	"github.com/1Panel-dev/1Panel/core/global"
	psessionUtils "github.com/1Panel-dev/1Panel/core/init/session/psession"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

const (
	headerNeedOperationResolve = "X-Need-Op-Resolve"
	headerOperationResolved    = "X-Op-Resolved"
)

func OperationLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Request.Header.Del(headerNeedOperationResolve)

		if strings.Contains(c.Request.URL.Path, "search") || c.Request.Method == http.MethodGet {
			c.Next()
			return
		}

		source := loadLogInfo(c.Request.URL.Path)
		pathItem := strings.TrimPrefix(c.Request.URL.Path, "/api/v2")
		pathItem = strings.TrimPrefix(pathItem, "/api/v2/core")
		currentNodeItem := c.Request.Header.Get("CurrentNode")
		currentNode, _ := url.QueryUnescape(currentNodeItem)
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
			if fullPath := normalizeOperationPath(c.FullPath()); fullPath != "" {
				operationDic, hasPath = swagger[fullPath]
			}
		}
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
			if strings.Contains(c.Request.Header.Get("Content-Type"), "multipart/form-data") {
				bodyMap, _ = parseMultipart(body, c.Request.Header.Get("Content-Type"))
			} else {
				decoder := json.NewDecoder(bytes.NewReader(body))
				decoder.UseNumber()
				_ = decoder.Decode(&bodyMap)
			}
			for _, key := range operationDic.BodyKeys {
				if _, ok := bodyMap[key]; ok {
					formatMap[key] = bodyMap[key]
				}
			}
		}
		needAgentResolve := len(operationDic.BeforeFunctions) != 0 && len(currentNode) != 0 && currentNode != "local" && !strings.HasPrefix(record.Path, "/core")
		allowCoreFallback := strings.HasPrefix(record.Path, "/core/xpack") || !ShouldProxyToAgent(c.Request.URL.Path) || len(currentNode) == 0 || currentNode == "local"
		if needAgentResolve {
			c.Request.Header.Set(headerNeedOperationResolve, "1")
			defer func() {
				c.Request.Header.Del(headerNeedOperationResolve)
			}()
		}

		writer := responseBodyWriter{
			ResponseWriter: c.Writer,
			body:           &bytes.Buffer{},
			captureBody:    !strings.Contains(strings.ToLower(c.Request.URL.Path), "download"),
		}
		c.Writer = &writer
		now := time.Now()

		c.Next()

		record.User = LoadOperationUser(c)

		if len(operationDic.BeforeFunctions) != 0 {
			if needAgentResolve {
				mergeResolvedData(writer.resolvedHeader, formatMap)
			}

			if allowCoreFallback && !hasAllResolvedData(formatMap, operationDic.BeforeFunctions) {
				dbItem, err := newDB(record.Path)
				if err == nil {
					resolveByDB(dbItem, formatMap, operationDic.BeforeFunctions)
					closeDB(dbItem)
				}
			}
		}
		fillOperationDetail(&operationDic, formatMap)
		record.DetailEN = strings.ReplaceAll(operationDic.FormatEN, "[]", "")
		record.DetailZH = strings.ReplaceAll(operationDic.FormatZH, "[]", "")

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
		contentType := strings.ToLower(c.Writer.Header().Get("Content-Type"))
		isJSONResponse := strings.Contains(contentType, "application/json")
		if isJSONResponse {
			_ = json.Unmarshal(datas, &res)
			if res.Code == 200 {
				record.Status = constant.StatusSuccess
			} else {
				record.Status = constant.StatusFailed
				record.Message = res.Message
			}
		} else {
			statusCode := c.Writer.Status()
			if statusCode >= 200 && statusCode < 400 {
				record.Status = constant.StatusSuccess
			} else {
				record.Status = constant.StatusFailed
				record.Message = http.StatusText(statusCode)
			}
		}

		latency := time.Since(now)
		record.Latency = latency

		if err := logRepo.CreateOperationLog(record); err != nil {
			global.LOG.Errorf("create operation record failed, err: %v", err)
		}
	}
}

func LoadOperationUser(c *gin.Context) string {
	sessionUser, ok := c.Get(psessionUtils.GinContextSessionUserKey)
	if ok {
		psession, ok := sessionUser.(psessionUtils.SessionUser)
		if ok {
			return psession.Name
		}
	}
	apiUsername, ok := c.Get("API_AUTH_USERNAME")
	if !ok {
		return ""
	}
	username, ok := apiUsername.(string)
	if !ok {
		return ""
	}
	return username
}

func fillOperationDetail(operationDic *operationJson, formatMap map[string]interface{}) {
	for key, value := range formatMap {
		if !strings.Contains(operationDic.FormatEN, "["+key+"]") {
			continue
		}
		t := reflect.TypeOf(value)
		if t == nil || (t.Kind() != reflect.Array && t.Kind() != reflect.Slice) {
			operationDic.FormatZH = strings.ReplaceAll(operationDic.FormatZH, "["+key+"]", fmt.Sprintf("[%v]", value))
			operationDic.FormatEN = strings.ReplaceAll(operationDic.FormatEN, "["+key+"]", fmt.Sprintf("[%v]", value))
			continue
		}

		val := reflect.ValueOf(value)
		length := val.Len()
		elements := make([]string, 0, length)
		for i := 0; i < length; i++ {
			elements = append(elements, fmt.Sprintf("%v", val.Index(i).Interface()))
		}
		replaced := fmt.Sprintf("[%v]", strings.Join(elements, ","))
		operationDic.FormatZH = strings.ReplaceAll(operationDic.FormatZH, "["+key+"]", replaced)
		operationDic.FormatEN = strings.ReplaceAll(operationDic.FormatEN, "["+key+"]", replaced)
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
	body           *bytes.Buffer
	resolvedHeader string
	captureBody    bool
}

func (r *responseBodyWriter) sanitizeResolvedHeader() {
	if len(r.resolvedHeader) == 0 {
		r.resolvedHeader = r.ResponseWriter.Header().Get(headerOperationResolved)
	}
	r.ResponseWriter.Header().Del(headerOperationResolved)
}

func (r *responseBodyWriter) WriteHeader(code int) {
	r.sanitizeResolvedHeader()
	r.ResponseWriter.WriteHeader(code)
}

func (r *responseBodyWriter) WriteHeaderNow() {
	r.sanitizeResolvedHeader()
	r.ResponseWriter.WriteHeaderNow()
}

func (r *responseBodyWriter) Write(b []byte) (int, error) {
	r.sanitizeResolvedHeader()
	if r.captureBody {
		r.body.Write(b)
	}
	return r.ResponseWriter.Write(b)
}

func loadLogInfo(path string) string {
	path = replaceStr(path, "/api/v2", "/core", "/xpack", "/enterprise")
	if !strings.Contains(path, "/") {
		return ""
	}
	pathArrays := strings.Split(path, "/")
	if len(pathArrays) < 2 {
		return ""
	}
	return pathArrays[1]
}

func normalizeOperationPath(reqPath string) string {
	pathItem := strings.TrimPrefix(reqPath, "/api/v2")
	pathItem = strings.TrimPrefix(pathItem, "/api/v2/core")
	return pathItem
}

func newDB(pathItem string) (*gorm.DB, error) {
	dbFile := ""
	switch {
	case strings.HasPrefix(pathItem, "/core/xpack") || strings.HasPrefix(pathItem, "/xpack"):
		dbFile = path.Join(global.CONF.Base.InstallDir, "1panel/db/xpack.db")
	case strings.HasPrefix(pathItem, "/core"):
		dbFile = path.Join(global.CONF.Base.InstallDir, "1panel/db/core.db")
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
	sqlDB.SetMaxOpenConns(4)
	sqlDB.SetMaxIdleConns(1)
	sqlDB.SetConnMaxIdleTime(15 * time.Minute)
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

func resolveByDB(dbItem *gorm.DB, values map[string]interface{}, beforeFunctions []functionInfo) {
	for _, funcs := range beforeFunctions {
		for key, value := range values {
			if funcs.InputValue != key {
				continue
			}
			var names []string
			if funcs.IsList {
				sql := fmt.Sprintf("SELECT %s FROM %s where %s in (?);", funcs.OutputColumn, funcs.DB, funcs.InputColumn)
				_ = dbItem.Raw(sql, value).Scan(&names)
			} else {
				sql := fmt.Sprintf("select %s from %s where %s = ?;", funcs.OutputColumn, funcs.DB, funcs.InputColumn)
				_ = dbItem.Raw(sql, value).Scan(&names)
			}
			values[funcs.OutputValue] = strings.Join(names, ",")
			break
		}
	}
}

func replaceStr(val string, rep ...string) string {
	for _, item := range rep {
		val = strings.ReplaceAll(val, item, "")
	}
	return val
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

func mergeResolvedData(headerVal string, values map[string]interface{}) {
	if len(headerVal) == 0 {
		return
	}
	data, err := base64.RawURLEncoding.DecodeString(headerVal)
	if err != nil {
		return
	}
	resolved := make(map[string]string)
	if err := json.Unmarshal(data, &resolved); err != nil {
		return
	}
	for key, value := range resolved {
		values[key] = value
	}
}

func hasAllResolvedData(values map[string]interface{}, beforeFunctions []functionInfo) bool {
	for _, item := range beforeFunctions {
		if _, ok := values[item.OutputValue]; ok {
			continue
		}
		return false
	}
	return true
}
