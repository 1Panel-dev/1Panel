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
	"github.com/1Panel-dev/1Panel/cmd/server/operation"
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
			Method:    c.Request.Method,
			Path:      c.Request.URL.Path,
			UserAgent: c.Request.UserAgent(),
		}
		var (
			operationDics []operationJson
			operationDic  operationJson
		)
		if err := json.Unmarshal(operation.OperationJosn, &operationDics); err != nil {
			c.Next()
			return
		}
		for _, dic := range operationDics {
			if dic.API == record.Path && dic.Method == record.Method {
				operationDic = dic
				break
			}
		}
		if len(operationDic.API) == 0 {
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
					if funcs.Info == key {
						var names []string
						if funcs.IsList {
							if key == "ids" {
								sql := fmt.Sprintf("SELECT %s FROM %s where id in (?);", funcs.Key, funcs.DB)
								fmt.Println(value)
								_ = global.DB.Raw(sql, value).Scan(&names)
							}
						} else {
							_ = global.DB.Raw(fmt.Sprintf("select %s from %s where %s = ?;", funcs.Key, funcs.DB, key), value).Scan(&names)
						}
						formatMap[funcs.Value] = strings.Join(names, ",")
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
	Info   string `json:"info"`
	IsList bool   `json:"isList"`
	DB     string `json:"db"`
	Key    string `json:"key"`
	Value  string `json:"value"`
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
