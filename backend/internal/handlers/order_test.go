package handlers_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"minishop/internal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Order List dipakai bersama test lain (tiap checkout sukses membuat
// order baru), jadi test ini memverifikasi keberadaan & tidak-ada-
// duplikat lewat containment saat menjelajah halaman — bukan total
// count eksak seperti test pagination produk yang bisa diisolasi lewat
// category_id.
func TestOrderListPaginationWalksWithoutDuplicates(t *testing.T) {
	db, r := setup(t)

	createdIDs := createTestOrders(t, db, r, 3)
	seen := walkAllOrderPages(t, r)

	for _, id := range createdIDs {
		if seen[id] != 1 {
			t.Errorf("order %d muncul %d kali saat menjelajah halaman (want tepat 1)", id, seen[id])
		}
	}
}

// createTestOrders membuat n order lewat checkout (masing-masing 1 produk
// baru) dan mengembalikan ID order yang terbentuk.
func createTestOrders(t *testing.T, db *gorm.DB, r *gin.Engine, n int) []uint {
	t.Helper()

	ids := make([]uint, 0, n)
	for i := 0; i < n; i++ {
		product := fixture(t, db, 10, 10000, fmt.Sprintf("order-%d", i))
		body := fmt.Sprintf(`{"items":[{"product_id":%d,"qty":1}]}`, product.ID)
		w := postCheckout(r, body)
		if w.Code != http.StatusCreated {
			t.Fatalf("checkout %d: want 201, got %d: %s", i, w.Code, w.Body.String())
		}

		var resp struct {
			Data models.Order `json:"data"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("parse checkout response: %v", err)
		}
		ids = append(ids, resp.Data.ID)
	}
	return ids
}

// walkAllOrderPages menjelajahi /orders dengan page_size=1 dari halaman 1
// sampai total_pages, mengembalikan berapa kali tiap order ID muncul.
func walkAllOrderPages(t *testing.T, r *gin.Engine) map[uint]int {
	t.Helper()

	firstPage := getOrders(r, "page=1&page_size=1")
	if firstPage.Code != http.StatusOK {
		t.Fatalf("want 200, got %d: %s", firstPage.Code, firstPage.Body.String())
	}
	_, meta := decodeOrdersWithMeta(t, firstPage)
	if meta.PageSize != 1 || meta.Page != 1 {
		t.Fatalf("meta halaman pertama tidak sesuai: %+v", meta)
	}
	if meta.TotalPages < 3 {
		t.Fatalf("total_pages (%d) harus >= 3 order yang baru dibuat", meta.TotalPages)
	}

	seen := map[uint]int{}
	for page := 1; page <= meta.TotalPages; page++ {
		w := getOrders(r, fmt.Sprintf("page=%d&page_size=1", page))
		if w.Code != http.StatusOK {
			t.Fatalf("page %d: want 200, got %d: %s", page, w.Code, w.Body.String())
		}
		orders, _ := decodeOrdersWithMeta(t, w)
		if len(orders) != 1 {
			t.Fatalf("page %d: want 1 order (page_size=1), got %d", page, len(orders))
		}
		seen[orders[0].ID]++
	}
	return seen
}

func getOrders(r *gin.Engine, query string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, "/api/orders?"+query, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func decodeOrdersWithMeta(t *testing.T, w *httptest.ResponseRecorder) ([]models.Order, paginationMetaJSON) {
	t.Helper()
	var resp struct {
		Data []models.Order     `json:"data"`
		Meta paginationMetaJSON `json:"meta"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("parse response: %v", err)
	}
	return resp.Data, resp.Meta
}
