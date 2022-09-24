package handlers

import (
	"github.com/gin-gonic/gin"

	mylog "github.com/alexveli/diploma/pkg/log"
)

func newResponse(c *gin.Context, statusCode int, message any) {
	mylog.SugarLogger.Infof("sending message %s with status %d", message, statusCode)
	c.Writer.WriteHeader(statusCode)
	var contents []byte
	switch message.(type) {
	case string:
		contents = []byte(message.(string))
	case []byte:
		contents = message.([]byte)
	}
	c.Writer.Write(contents)
	c.Abort()
	mylog.SugarLogger.Infof("Response is %v", c)
}
