package handlers

import (
	"errors"
	"net/http"
	"strings"

	"minishop/internal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ProductHandler struct {
	DB *gorm.DB
}

type createProductInput struct {
	Name        string `json:"name" binding:"required"`
	Price       *int64 `json:"price" binding:"required,gte=0"`
	Description string `json:"description"`
	ImageURL    string `json:"image_url"`
	Stock       *int   `json:"stock" binding:"required,gte=0"`
	CategoryID  uint   `json:"category_id" binding:"required"`
}

// Stock opsional pada update: bila tidak dikirim, stok tidak disentuh.
// Ini mencegah lost update — form edit yang menyimpan stok basi akan
// menimpa hasil pengurangan stok dari checkout yang terjadi di sela-selanya (critical).
type updateProductInput struct {
	Name        string `json:"name" binding:"required"`
	Price       *int64 `json:"price" binding:"required,gte=0"`
	Description string `json:"description"`
	ImageURL    string `json:"image_url"`
	Stock       *int   `json:"stock" binding:"omitempty,gte=0"`
	CategoryID  uint   `json:"category_id" binding:"required"`
}

// escapeLike meng-escape wildcard ILIKE agar '%' dan '_' dari user
// diperlakukan sebagai karakter biasa, bukan pola.
var escapeLike = strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)

const (
	byID           = "id = ?"
	msgNotFound    = "product not found"
	msgFetchFailed = "failed to fetch product"
)

// List menangani search nama (ILIKE, didukung GIN trigram index) dan
// filter kategori (B-tree index) sekaligus.
func (h *ProductHandler) List(c *gin.Context) {
	q := h.DB.Preload("Category").Order("id")

	if search := c.Query("search"); search != "" {
		q = q.Where("name ILIKE ?", "%"+escapeLike.Replace(search)+"%")
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
	if err := h.DB.Preload("Category").First(&product, byID, c.Param("id")).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": msgNotFound})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": msgFetchFailed})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": product})
}

func (h *ProductHandler) Create(c *gin.Context) {
	var input createProductInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input: " + err.Error()})
		return
	}
	if !h.categoryExists(c, input.CategoryID) {
		return
	}

	product := models.Product{
		Name:        input.Name,
		Price:       *input.Price,
		Description: input.Description,
		ImageURL:    input.ImageURL,
		Stock:       *input.Stock,
		CategoryID:  input.CategoryID,
	}
	if err := h.DB.Create(&product).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create product"})
		return
	}
	h.respondWithProduct(c, http.StatusCreated, product.ID)
}

func (h *ProductHandler) Update(c *gin.Context) {
	var input updateProductInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input: " + err.Error()})
		return
	}
	if !h.categoryExists(c, input.CategoryID) {
		return
	}

	product.Name = input.Name
	product.Price = input.Price
	product.Description = input.Description
	product.ImageURL = input.ImageURL
	product.Stock = *input.Stock
	product.CategoryID = input.CategoryID
	h.respondWithProduct(c, http.StatusOK, c.Param("id"))
}

func (h *ProductHandler) Delete(c *gin.Context) {
	res := h.DB.Delete(&models.Product{}, byID, c.Param("id"))
	if res.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete product"})
		return
	}
	if res.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": msgNotFound})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "product deleted"})
}

func (h *ProductHandler) categoryExists(c *gin.Context, categoryID uint) bool {
	var count int64
	if err := h.DB.Model(&models.Category{}).Where(byID, categoryID).Count(&count).Error; err != nil || count == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "category_id does not refer to an existing category"})
		return false
	}
	return true
}

func (h *ProductHandler) respondWithProduct(c *gin.Context, status int, id any) {
	var product models.Product
	if err := h.DB.Preload("Category").First(&product, byID, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": msgFetchFailed})
		return
	}
	c.JSON(status, gin.H{"data": product})
}
