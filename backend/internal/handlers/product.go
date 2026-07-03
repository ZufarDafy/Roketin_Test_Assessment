package handlers

import (
	"errors"
	"net/http"

	"minishop/internal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ProductHandler struct {
	DB *gorm.DB
}

type productInput struct {
	Name        string  `json:"name" binding:"required"`
	Price       float64 `json:"price" binding:"required,gte=0"`
	Description string  `json:"description"`
	ImageURL    string  `json:"image_url"`
	Stock       *int    `json:"stock" binding:"required,gte=0"`
	CategoryID  uint    `json:"category_id" binding:"required"`
}

// List menangani search nama (ILIKE, didukung GIN trigram index) dan
// filter kategori (B-tree index) sekaligus.
func (h *ProductHandler) List(c *gin.Context) {
	q := h.DB.Preload("Category").Order("id")

	if search := c.Query("search"); search != "" {
		q = q.Where("name ILIKE ?", "%"+search+"%")
	}
	if categoryID := c.Query("category_id"); categoryID != "" {
		q = q.Where("category_id = ?", categoryID)
	}

	var products []models.Product
	if err := q.Find(&products).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch products"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": products})
}

func (h *ProductHandler) Get(c *gin.Context) {
	var product models.Product
	if err := h.DB.Preload("Category").First(&product, "id = ?", c.Param("id")).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch product"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": product})
}

func (h *ProductHandler) Create(c *gin.Context) {
	input, ok := h.bindAndValidate(c)
	if !ok {
		return
	}

	product := models.Product{
		Name:        input.Name,
		Price:       input.Price,
		Description: input.Description,
		ImageURL:    input.ImageURL,
		Stock:       *input.Stock,
		CategoryID:  input.CategoryID,
	}
	if err := h.DB.Create(&product).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create product"})
		return
	}
	h.DB.Preload("Category").First(&product, product.ID)
	c.JSON(http.StatusCreated, gin.H{"data": product})
}

func (h *ProductHandler) Update(c *gin.Context) {
	var product models.Product
	if err := h.DB.First(&product, "id = ?", c.Param("id")).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch product"})
		return
	}

	input, ok := h.bindAndValidate(c)
	if !ok {
		return
	}

	product.Name = input.Name
	product.Price = input.Price
	product.Description = input.Description
	product.ImageURL = input.ImageURL
	product.Stock = *input.Stock
	product.CategoryID = input.CategoryID
	if err := h.DB.Save(&product).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update product"})
		return
	}
	h.DB.Preload("Category").First(&product, product.ID)
	c.JSON(http.StatusOK, gin.H{"data": product})
}

func (h *ProductHandler) Delete(c *gin.Context) {
	res := h.DB.Delete(&models.Product{}, "id = ?", c.Param("id"))
	if res.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete product"})
		return
	}
	if res.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "product deleted"})
}

// bindAndValidate memvalidasi payload dan memastikan category_id merujuk
// kategori yang ada.
func (h *ProductHandler) bindAndValidate(c *gin.Context) (productInput, bool) {
	var input productInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input: " + err.Error()})
		return input, false
	}

	var count int64
	if err := h.DB.Model(&models.Category{}).Where("id = ?", input.CategoryID).Count(&count).Error; err != nil || count == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "category_id does not refer to an existing category"})
		return input, false
	}
	return input, true
}
