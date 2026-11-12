package helper

import (
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
)

func StreamResponse(c *gin.Context, response *http.Response) {
	defer response.Body.Close()
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("X-Accel-Buffering", "no")
	c.Writer.Flush()
	buffer := make([]byte, 8192)
	for {
		n, err := response.Body.Read(buffer)
		if n > 0 {
			if _, writeErr := c.Writer.Write(buffer[:n]); writeErr != nil {
				return
			}
			c.Writer.Flush()
		}
		if err != nil {
			if err != io.EOF && c.Request.Context().Err() == nil {
				_, _ = c.Writer.WriteString("\ndata: {\"error\": {\"message\": \"Model stream interrupted\"}}\n\n")
				c.Writer.Flush()
			}
			return
		}
	}
}
