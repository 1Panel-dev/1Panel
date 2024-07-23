package middleware

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/backend/app/model"
	"github.com/1Panel-dev/1Panel/backend/app/service"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/1Panel-dev/1Panel/backend/utils/copier"
	"github.com/1Panel-dev/1Panel/cmd/server/docs"
	"github.com/gin-gonic/gin"
)

func OperationLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		if strings.Contains(c.Request.URL.Path, "search") || c.Request.Method == http.MethodGet {
			c.Next()
			return
		}

		source := loadLogInfo(c.Request.URL.Path)
		record := model.OperationLog{
			Source:    source,
			IP:        c.ClientIP(),
			Method:    strings.ToLower(c.Request.Method),
			Path:      strings.ReplaceAll(c.Request.URL.Path, "/api/v1", ""),
			UserAgent: c.Request.UserAgent(),
		}
		var (
			swagger      swaggerJson
			operationDic operationJson
		)
		if err := json.Unmarshal(docs.SwaggerJson, &swagger); err != nil {
			c.Next()
			return
		}
		path, hasPath := swagger.Paths[record.Path]
		if !hasPath {
			c.Next()
			return
		}
		methodMap, isMethodMap := path.(map[string]interface{})
		if !isMethodMap {
			c.Next()
			return
		}
		dataMap, hasPost := methodMap["post"]
		if !hasPost {
			c.Next()
			return
		}
		data, isDataMap := dataMap.(map[string]interface{})
		if !isDataMap {
			c.Next()
			return
		}
		xlog, hasXlog := data["x-panel-log"]
		if !hasXlog {
			c.Next()
			return
		}
		if err := copier.Copy(&operationDic, xlog); err != nil {
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
			for _, funcs := range operationDic.BeforeFunctions {
				for key, value := range formatMap {
					if funcs.InputValue == key {
						var names []string
						if funcs.IsList {
							query := fmt.Sprintf("SELECT %s FROM %s WHERE %s in (?)", funcs.OutputColumn, funcs.DB, funcs.InputColumn)
							_ = global.DB.Raw(query, value).Scan(&names)
						} else {
							query := fmt.Sprintf("SELECT %s FROM %s WHERE %s = ?", funcs.OutputColumn, funcs.DB, funcs.InputColumn)
							_ = global.DB.Raw(query, value).Scan(&names)
						}
						formatMap[funcs.OutputValue] = strings.Join(names, ",")
						break
					}
				}
			}
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
		if c.Request.Header.Get("Content-Encoding") == "gzip" {
			buf := bytes.NewReader(writer.body.Bytes())
			reader, err := gzip.NewReader(buf)
			if err != nil {
				record.Status = constant.StatusFailed
				record.Message = fmt.Sprintf("gzip new reader failed, err: %v", err)
				latency := time.Since(now)
				record.Latency = latency

				if err := service.NewILogService().CreateOperationLog(record); err != nil {
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

		if err := service.NewILogService().CreateOperationLog(record); err != nil {
			global.LOG.Errorf("create operation record failed, err: %v", err)
		}
	}
}

type swaggerJson struct {
	Paths map[string]interface{} `json:"paths"`
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
	path = strings.ReplaceAll(path, "/api/v1", "")
	if !strings.Contains(path, "/") {
		return ""
	}
	pathArrays := strings.Split(path, "/")
	if len(pathArrays) < 2 {
		return ""
	}
	return pathArrays[1]
}
