package ping

import (
	"github.com/gin-gonic/gin"
)

func Handler(c *gin.Context) {
	c.String(200, "OK")
}
