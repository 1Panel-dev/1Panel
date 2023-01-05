package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
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

		group := loadLogInfo(c.Request.URL.Path)
		record := model.OperationLog{
			Group:     group,
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
			body, err := ioutil.ReadAll(c.Request.Body)
			if err == nil {
				c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(body))
			}
			bodyMap := make(map[string]interface{})
			_ = json.Unmarshal(body, &bodyMap)
			for _, key := range operationDic.BodyKeys {
				if _, ok := bodyMap[key]; ok {
					formatMap[key] = bodyMap[key]
				}
			}
		}
		if len(operationDic.BeforeFuntions) != 0 {
			for _, funcs := range operationDic.BeforeFuntions {
				for key, value := range formatMap {
					if funcs.InputValue == key {
						var names []string
						if funcs.IsList {
							sql := fmt.Sprintf("SELECT %s FROM %s where %s in (?);", funcs.OutputColume, funcs.DB, funcs.InputColume)
							fmt.Println(value)
							_ = global.DB.Raw(sql, value).Scan(&names)
						} else {
							_ = global.DB.Raw(fmt.Sprintf("select %s from %s where %s = ?;", funcs.OutputColume, funcs.DB, funcs.InputColume), value).Scan(&names)
						}
						formatMap[funcs.OutputValue] = strings.Join(names, ",")
						break
					}
				}
			}
		}
		for key, value := range formatMap {
			if strings.Contains(operationDic.FormatEN, "["+key+"]") {
				if arrys, ok := value.([]string); ok {
					operationDic.FormatZH = strings.ReplaceAll(operationDic.FormatZH, "["+key+"]", fmt.Sprintf("[%v]", strings.Join(arrys, ",")))
					operationDic.FormatEN = strings.ReplaceAll(operationDic.FormatEN, "["+key+"]", fmt.Sprintf("[%v]", strings.Join(arrys, ",")))
				} else {
					operationDic.FormatZH = strings.ReplaceAll(operationDic.FormatZH, "["+key+"]", fmt.Sprintf("[%v]", value))
					operationDic.FormatEN = strings.ReplaceAll(operationDic.FormatEN, "["+key+"]", fmt.Sprintf("[%v]", value))
				}
			}
		}
		record.DetailEN = operationDic.FormatEN
		record.DetailZH = operationDic.FormatZH

		writer := responseBodyWriter{
			ResponseWriter: c.Writer,
			body:           &bytes.Buffer{},
		}
		c.Writer = writer
		now := time.Now()

		c.Next()

		var res response
		_ = json.Unmarshal(writer.body.Bytes(), &res)
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
	API            string         `json:"api"`
	Method         string         `json:"method"`
	BodyKeys       []string       `json:"bodyKeys"`
	ParamKeys      []string       `json:"paramKeys"`
	BeforeFuntions []functionInfo `json:"beforeFuntions"`
	FormatZH       string         `json:"formatZH"`
	FormatEN       string         `json:"formatEN"`
}
type functionInfo struct {
	InputColume  string `json:"input_colume"`
	InputValue   string `json:"input_value"`
	IsList       bool   `json:"isList"`
	DB           string `json:"db"`
	OutputColume string `json:"output_colume"`
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
	pathArrys := strings.Split(path, "/")
	if len(pathArrys) < 2 {
		return ""
	}
	return pathArrys[1]
}
