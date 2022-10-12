package handlers

import (
	"github.com/gin-gonic/gin"

	mylog "github.com/alexveli/diploma/pkg/log"
)

func newResponse(c *gin.Context, statusCode int, message interface{}) {
	mylog.SugarLogger.Infof("sending message %s with status %d", message, statusCode)
	c.Writer.WriteHeader(statusCode)
	var contents []byte
	switch message := message.(type) {
	case string:
		contents = []byte(message)
	case []byte:
		contents = message
	}
	_, err := c.Writer.Write(contents)
	if err != nil {
		mylog.SugarLogger.Errorf("cannot write response, %v", err)
	}
	c.Abort()
	mylog.SugarLogger.Infof("Response is %v", c)
}
