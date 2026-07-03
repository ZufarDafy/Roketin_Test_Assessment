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

// Checkout membuat Order dalam satu transaction:
//  1. Setiap produk dikunci dengan SELECT ... FOR UPDATE agar dua checkout
//     bersamaan tidak bisa sama-sama lolos validasi stok (race condition).
//  2. Stok divalidasi ulang di dalam transaction — validasi di frontend
//     hanya untuk UX, sumber kebenaran ada di sini.
//  3. Total dihitung dari harga di database, bukan dari client.
//
// Produk diproses berurutan berdasarkan product_id agar dua transaksi yang
// saling menunggu lock tidak deadlock.
func (h *CheckoutHandler) Checkout(c *gin.Context) {
	var input checkoutInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input: " + err.Error()})
		return
	}

	// Gabungkan qty untuk product_id duplikat, lalu urutkan.
	qtyByProduct := map[uint]int{}
	for _, item := range input.Items {
		qtyByProduct[item.ProductID] += item.Qty
	}
	productIDs := make([]uint, 0, len(qtyByProduct))
	for id := range qtyByProduct {
		productIDs = append(productIDs, id)
	}
	sort.Slice(productIDs, func(i, j int) bool { return productIDs[i] < productIDs[j] })

	var order models.Order
	var stockErrors []stockError
	var missingID uint

	txErr := h.DB.Transaction(func(tx *gorm.DB) error {
		var items []models.OrderItem
		var total float64

		for _, id := range productIDs {
			qty := qtyByProduct[id]

			var product models.Product
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
				First(&product, "id = ?", id).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					missingID = id
					return errNotFound
				}
				return err
			}

			if qty > product.Stock {
				stockErrors = append(stockErrors, stockError{
					ProductID: product.ID,
					Name:      product.Name,
					Requested: qty,
					Available: product.Stock,
				})
				continue
			}

			if err := tx.Model(&models.Product{}).Where("id = ?", product.ID).
				Update("stock", gorm.Expr("stock - ?", qty)).Error; err != nil {
				return err
			}

			subtotal := product.Price * float64(qty)
			total += subtotal
			items = append(items, models.OrderItem{
				ProductID:   product.ID,
				ProductName: product.Name,
				Price:       product.Price,
				Qty:         qty,
				Subtotal:    subtotal,
			})
		}

		if len(stockErrors) > 0 {
			return errStock
		}

		order = models.Order{Total: total, Items: items}
		return tx.Create(&order).Error
	})

	switch {
	case txErr == nil:
		c.JSON(http.StatusCreated, gin.H{"data": order})
	case errors.Is(txErr, errStock):
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"error":       "insufficient stock for some items",
			"stock_errors": stockErrors,
		})
	case errors.Is(txErr, errNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("product %d not found", missingID)})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "checkout failed"})
	}
}
