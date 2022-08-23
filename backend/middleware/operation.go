package middleware

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/app/model"
	"github.com/1Panel-dev/1Panel/app/service"
	"github.com/1Panel-dev/1Panel/global"
	"github.com/gin-gonic/gin"
)

func OperationRecord() gin.HandlerFunc {
	return func(c *gin.Context) {
		var body []byte
		if strings.Contains(c.Request.URL.Path, "search") {
			c.Next()
			return
		}

		if c.Request.Method == http.MethodGet {
			query := c.Request.URL.RawQuery
			query, _ = url.QueryUnescape(query)
			split := strings.Split(query, "&")
			m := make(map[string]string)
			for _, v := range split {
				kv := strings.Split(v, "=")
				if len(kv) == 2 {
					m[kv[0]] = kv[1]
				}
			}
			body, _ = json.Marshal(&m)
		} else {
			var err error
			body, err = ioutil.ReadAll(c.Request.Body)
			if err != nil {
				global.LOG.Errorf("read body from request failed, err: %v", err)
			} else {
				c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(body))
			}
		}
		pathInfo := loadLogInfo(c.Request.URL.Path)

		record := model.OperationLog{
			Group:     pathInfo.group,
			Source:    pathInfo.source,
			Action:    pathInfo.action,
			IP:        c.ClientIP(),
			Method:    c.Request.Method,
			Path:      c.Request.URL.Path,
			UserAgent: c.Request.UserAgent(),
			Body:      string(body),
		}

		writer := responseBodyWriter{
			ResponseWriter: c.Writer,
			body:           &bytes.Buffer{},
		}
		c.Writer = writer
		now := time.Now()

		c.Next()

		latency := time.Since(now)
		record.Latency = latency
		record.Resp = writer.body.String()

		if err := service.NewIOperationService().Create(record); err != nil {
			global.LOG.Errorf("create operation record failed, err: %v", err)
		}
	}
}

type responseBodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (r responseBodyWriter) Write(b []byte) (int, error) {
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}

type pathInfo struct {
	group  string
	source string
	action string
}

func loadLogInfo(path string) pathInfo {
	path = strings.ReplaceAll(path, "/api/v1", "")
	if !strings.Contains(path, "/") {
		return pathInfo{}
	}
	pathArrys := strings.Split(path, "/")
	if len(pathArrys) < 2 {
		return pathInfo{}
	}
	if len(pathArrys) == 2 {
		return pathInfo{group: pathArrys[1]}
	}
	if len(pathArrys) == 3 {
		return pathInfo{group: pathArrys[1], source: pathArrys[2]}
	}
	return pathInfo{group: pathArrys[1], source: pathArrys[2], action: pathArrys[3]}
}
