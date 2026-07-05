package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"minishop/internal/models"

	"github.com/gin-gonic/gin"
)

func getProducts(r *gin.Engine, query string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, "/api/products?"+query, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func decodeProducts(t *testing.T, w *httptest.ResponseRecorder) []models.Product {
	t.Helper()
	var resp struct {
		Data []models.Product `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("parse response: %v", err)
	}
	return resp.Data
}

func containsProductID(products []models.Product, id uint) bool {
	for _, p := range products {
		if p.ID == id {
			return true
		}
	}
	return false
}

// Search di bawah 3 karakter harus diperlakukan seperti tidak ada filter
// sama sekali (lihat komentar minSearchLen di product.go). "zz" sengaja
// dipilih karena tidak muncul di nama fixture manapun — bila guard
// berfungsi, kedua produk tetap muncul walau tidak match "zz".
func TestProductListShortSearchIgnoresFilter(t *testing.T) {
	db, r := setup(t)
	target := fixture(t, db, 10, 10000, "target")
	other := fixture(t, db, 10, 10000, "other")

	w := getProducts(r, "search=zz")
	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d: %s", w.Code, w.Body.String())
	}

	products := decodeProducts(t, w)
	if !containsProductID(products, target.ID) || !containsProductID(products, other.ID) {
		t.Errorf("search 2 karakter harus diabaikan (semua produk tetap muncul); got %+v", products)
	}
}

// Search 3 karakter atau lebih harus benar-benar memfilter berdasarkan
// substring nama. "-target" hanya ada di nama fixture target.
func TestProductListLongSearchFilters(t *testing.T) {
	db, r := setup(t)
	target := fixture(t, db, 10, 10000, "target")
	other := fixture(t, db, 10, 10000, "other")

	w := getProducts(r, "search=-target")
	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d: %s", w.Code, w.Body.String())
	}

	products := decodeProducts(t, w)
	if !containsProductID(products, target.ID) {
		t.Errorf("search '-target' harus menemukan produk target %q", target.Name)
	}
	if containsProductID(products, other.ID) {
		t.Errorf("search '-target' tidak boleh menemukan produk lain %q", other.Name)
	}
}
