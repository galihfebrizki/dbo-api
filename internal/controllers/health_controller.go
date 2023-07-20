package controllers

import (
	"net/http"

	"github.com/galihfebrizki/dbo-api/helper"
	"github.com/galihfebrizki/dbo-api/internal/services"

	"github.com/gin-gonic/gin"
)

type HealthController struct {
	HealthService services.IHealthService
}

func NewHealthController(service services.IHealthService) *HealthController {
	return &HealthController{
		HealthService: service,
	}
}

func (h *HealthController) Health(c *gin.Context) {
	ctx := helper.GetGinContext(c)
	c.JSON(http.StatusOK, h.HealthService.HealthCheck(ctx))
}
