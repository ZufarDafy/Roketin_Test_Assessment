package handlers

import (
	"errors"
	"net/http"

	"minishop/internal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type OrderHandler struct {
	DB *gorm.DB
}

func (h *OrderHandler) List(c *gin.Context) {
	var total int64
	if err := h.DB.Model(&models.Order{}).Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch orders"})
		return
	}
	page, pageSize, offset := parsePagination(c)

	var orders []models.Order
	if err := h.DB.Preload("Items").Order("id DESC").Limit(pageSize).Offset(offset).Find(&orders).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch orders"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": orders, "meta": buildPaginationMeta(page, pageSize, total)})
}

func (h *OrderHandler) Get(c *gin.Context) {
	var order models.Order
	if err := h.DB.Preload("Items").First(&order, "id = ?", c.Param("id")).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch order"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": order})
}
