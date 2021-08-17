package ping

import (
	"github.com/gin-gonic/gin"
)

// @Summary Show the status of the API
// @Description Get the status of the API
// @Accept */*
// @Produce text/plain
// @Success 200 {string} string "OK"
// @Router /ping [get]
// @Tags HealthCheck
func Handler(c *gin.Context) {
	c.String(200, "OK")
}
