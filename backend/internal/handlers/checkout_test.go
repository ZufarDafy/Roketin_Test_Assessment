package handlers_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"

	"minishop/internal/database"
	"minishop/internal/models"
	"minishop/internal/router"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/gorm"
)

// init memuat backend/.env sebagai sumber TEST_DB_DSN.
func init() {
	if _, file, _, ok := runtime.Caller(0); ok {
		_ = godotenv.Load(filepath.Join(filepath.Dir(file), "..", "..", ".env"))
	}
}

// Integration test — butuh PostgreSQL berjalan (docker compose up -d)
// dan TEST_DB_DSN diset (via shell atau backend/.env). Test tidak
// memakai config.Load() agar tidak terikat konfigurasi server.
func setup(t *testing.T) (*gorm.DB, *gin.Engine) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	dsn := os.Getenv("TEST_DB_DSN")
	if dsn == "" {
		t.Skip("TEST_DB_DSN is not set; skipping integration test")
	}

	db, err := database.Connect(dsn)
	if err != nil {
		t.Skipf("database not available, skipping integration test: %v", err)
	}
	if err := database.Migrate(db); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db, router.New(db, []string{"http://localhost:5173"})
}

// fixture membuat kategori + produk uji dan mendaftarkan pembersihannya.
func fixture(t *testing.T, db *gorm.DB, stock int, price int64) models.Product {
	t.Helper()

	category := models.Category{Name: fmt.Sprintf("TEST-cat-%s", t.Name())}
	if err := db.Create(&category).Error; err != nil {
		t.Fatalf("create category: %v", err)
	}
	product := models.Product{
		Name:       fmt.Sprintf("TEST-product-%s", t.Name()),
		Price:      price,
		Stock:      stock,
		CategoryID: category.ID,
	}
	if err := db.Create(&product).Error; err != nil {
		t.Fatalf("create product: %v", err)
	}

	t.Cleanup(func() {
		var orderIDs []uint
		db.Model(&models.OrderItem{}).Where("product_id = ?", product.ID).
			Distinct().Pluck("order_id", &orderIDs)
		if len(orderIDs) > 0 {
			db.Delete(&models.OrderItem{}, "order_id IN ?", orderIDs)
			db.Delete(&models.Order{}, "id IN ?", orderIDs)
		}
		db.Delete(&models.Product{}, "id = ?", product.ID)
		db.Delete(&models.Category{}, "id = ?", category.ID)
	})
	return product
}

func postCheckout(r *gin.Engine, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/api/checkout", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func currentStock(t *testing.T, db *gorm.DB, id uint) int {
	t.Helper()
	var product models.Product
	if err := db.First(&product, "id = ?", id).Error; err != nil {
		t.Fatalf("fetch product: %v", err)
	}
	return product.Stock
}

func TestCheckoutSuccess(t *testing.T) {
	db, r := setup(t)
	product := fixture(t, db, 10, 189000)

	// Dua entri product_id sama harus digabung (qty 2+1 = 3).
	body := fmt.Sprintf(
		`{"items":[{"product_id":%d,"qty":2},{"product_id":%d,"qty":1}]}`,
		product.ID, product.ID)
	w := postCheckout(r, body)

	if w.Code != http.StatusCreated {
		t.Fatalf("want 201, got %d: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Data models.Order `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("parse response: %v", err)
	}
	if want := int64(3 * 189000); resp.Data.Total != want {
		t.Errorf("total: want %d, got %d", want, resp.Data.Total)
	}
	if len(resp.Data.Items) != 1 || resp.Data.Items[0].Qty != 3 {
		t.Errorf("want 1 item with qty 3, got %+v", resp.Data.Items)
	}
	if got := currentStock(t, db, product.ID); got != 7 {
		t.Errorf("stock: want 7, got %d", got)
	}
}

func TestCheckoutInsufficientStock(t *testing.T) {
	db, r := setup(t)
	product := fixture(t, db, 5, 10000)

	w := postCheckout(r, fmt.Sprintf(`{"items":[{"product_id":%d,"qty":6}]}`, product.ID))

	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("want 422, got %d: %s", w.Code, w.Body.String())
	}

	var resp struct {
		StockErrors []struct {
			ProductID uint `json:"product_id"`
			Requested int  `json:"requested"`
			Available int  `json:"available"`
		} `json:"stock_errors"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("parse response: %v", err)
	}
	if len(resp.StockErrors) != 1 {
		t.Fatalf("want 1 stock error, got %+v", resp.StockErrors)
	}
	se := resp.StockErrors[0]
	if se.ProductID != product.ID || se.Requested != 6 || se.Available != 5 {
		t.Errorf("stock error mismatch: %+v", se)
	}
	if got := currentStock(t, db, product.ID); got != 5 {
		t.Errorf("stock must be untouched: want 5, got %d", got)
	}
}

func TestCheckoutProductNotFound(t *testing.T) {
	_, r := setup(t)

	w := postCheckout(r, `{"items":[{"product_id":999999999,"qty":1}]}`)
	if w.Code != http.StatusNotFound {
		t.Fatalf("want 404, got %d: %s", w.Code, w.Body.String())
	}
}

// Dua checkout konkuren memperebutkan sisa stok yang sama: tepat satu
// harus berhasil dan stok berakhir 0 — membuktikan SELECT FOR UPDATE
// mencegah oversell.
func TestCheckoutConcurrentNoOversell(t *testing.T) {
	db, r := setup(t)
	product := fixture(t, db, 12, 10000)
	body := fmt.Sprintf(`{"items":[{"product_id":%d,"qty":12}]}`, product.ID)

	codes := make([]int, 2)
	var wg sync.WaitGroup
	for i := range codes {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			codes[i] = postCheckout(r, body).Code
		}(i)
	}
	wg.Wait()

	success, rejected := 0, 0
	for _, code := range codes {
		switch code {
		case http.StatusCreated:
			success++
		case http.StatusUnprocessableEntity:
			rejected++
		}
	}
	if success != 1 || rejected != 1 {
		t.Errorf("want exactly one 201 and one 422, got %v", codes)
	}
	if got := currentStock(t, db, product.ID); got != 0 {
		t.Errorf("stock: want 0, got %d", got)
	}
}
