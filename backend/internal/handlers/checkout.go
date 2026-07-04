package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"sort"

	"minishop/internal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type CheckoutHandler struct {
	DB *gorm.DB
}

type checkoutItem struct {
	ProductID uint `json:"product_id" binding:"required"`
	Qty       int  `json:"qty" binding:"required,gt=0"`
}

type checkoutInput struct {
	Items []checkoutItem `json:"items" binding:"required,min=1,dive"`
}

type stockError struct {
	ProductID uint   `json:"product_id"`
	Name      string `json:"name"`
	Requested int    `json:"requested"`
	Available int    `json:"available"`
}

var errStock = errors.New("insufficient stock")
var errNotFound = errors.New("product not found")

// checkoutState menampung hasil antara selama transaction berlangsung.
type checkoutState struct {
	order       models.Order
	stockErrors []stockError
	missingID   uint
}

// Checkout membuat Order dalam satu transaction:
//  1. Setiap produk dikunci dengan SELECT ... FOR UPDATE agar dua checkout
//     bersamaan tidak bisa sama-sama lolos validasi stok (race condition).
//  2. Stok divalidasi ulang di dalam transaction — validasi di frontend
//     hanya untuk UX, sumber kebenaran ada di sini.
//  3. Total dihitung dari harga di database, bukan dari client.
func (h *CheckoutHandler) Checkout(c *gin.Context) {
	var input checkoutInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input: " + err.Error()})
		return
	}

	qtyByProduct, productIDs := normalizeItems(input.Items)

	var state checkoutState
	txErr := h.DB.Transaction(func(tx *gorm.DB) error {
		return state.process(tx, productIDs, qtyByProduct)
	})

	switch {
	case txErr == nil:
		c.JSON(http.StatusCreated, gin.H{"data": state.order})
	case errors.Is(txErr, errStock):
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"error":        "insufficient stock for some items",
			"stock_errors": state.stockErrors,
		})
	case errors.Is(txErr, errNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("product %d not found", state.missingID)})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "checkout failed"})
	}
}

// normalizeItems menggabungkan qty untuk product_id duplikat dan
// mengurutkan id — pemrosesan terurut mencegah deadlock antar dua
// transaksi yang saling menunggu lock produk yang sama.
func normalizeItems(items []checkoutItem) (map[uint]int, []uint) {
	qtyByProduct := map[uint]int{}
	for _, item := range items {
		qtyByProduct[item.ProductID] += item.Qty
	}
	ids := make([]uint, 0, len(qtyByProduct))
	for id := range qtyByProduct {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	return qtyByProduct, ids
}

func (s *checkoutState) process(tx *gorm.DB, productIDs []uint, qtyByProduct map[uint]int) error {
	var items []models.OrderItem
	var total int64

	for _, id := range productIDs {
		qty := qtyByProduct[id]

		product, err := lockProduct(tx, id)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				s.missingID = id
				return errNotFound
			}
			return err
		}

		if qty > product.Stock {
			s.stockErrors = append(s.stockErrors, stockError{
				ProductID: product.ID,
				Name:      product.Name,
				Requested: qty,
				Available: product.Stock,
			})
			continue
		}

		if err := tx.Model(&models.Product{}).Where(byID, product.ID).
			Update("stock", gorm.Expr("stock - ?", qty)).Error; err != nil {
			return err
		}

		subtotal := product.Price * int64(qty)
		total += subtotal
		items = append(items, models.OrderItem{
			ProductID:   product.ID,
			ProductName: product.Name,
			Price:       product.Price,
			Qty:         qty,
			Subtotal:    subtotal,
		})
	}

	if len(s.stockErrors) > 0 {
		return errStock
	}

	s.order = models.Order{Total: total, Items: items}
	return tx.Create(&s.order).Error
}

func lockProduct(tx *gorm.DB, id uint) (models.Product, error) {
	var product models.Product
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&product, byID, id).Error
	return product, err
}
