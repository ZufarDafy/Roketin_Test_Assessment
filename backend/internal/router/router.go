package router

import (
	"time"

	"minishop/internal/handlers"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const productByID = "/products/:id"

func New(db *gorm.DB, allowedOrigins []string) *gin.Engine {
	r := gin.Default()
	r.SetTrustedProxies(nil)

	r.Use(cors.New(cors.Config{
		AllowOrigins:     allowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept"},
		MaxAge:           12 * time.Hour,
		AllowCredentials: false,
	}))

	product := &handlers.ProductHandler{DB: db}
	category := &handlers.CategoryHandler{DB: db}
	checkout := &handlers.CheckoutHandler{DB: db}
	order := &handlers.OrderHandler{DB: db}

	api := r.Group("/api")
	{
		api.GET("/products", product.List)
		api.GET(productByID, product.Get)
		api.POST("/products", product.Create)
		api.PUT(productByID, product.Update)
		api.DELETE(productByID, product.Delete)

		api.GET("/categories", category.List)

		api.POST("/checkout", checkout.Checkout)

		api.GET("/orders", order.List)
		api.GET("/orders/:id", order.Get)
	}

	return r
}
