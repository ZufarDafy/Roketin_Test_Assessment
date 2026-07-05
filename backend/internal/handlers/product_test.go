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

type paginationMetaJSON struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

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

func decodeProductsWithMeta(t *testing.T, w *httptest.ResponseRecorder) ([]models.Product, paginationMetaJSON) {
	t.Helper()
	var resp struct {
		Data []models.Product    `json:"data"`
		Meta paginationMetaJSON `json:"meta"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("parse response: %v", err)
	}
	return resp.Data, resp.Meta
}

func containsProductID(products []models.Product, id uint) bool {
	for _, p := range products {
		if p.ID == id {
			return true
		}
	}
	return false
}

// fixtureCategory membuat kategori uji dengan nama unik dan mendaftarkan
// pembersihannya. Dipakai saat satu test butuh beberapa produk dalam satu
// kategori yang sama (mis. untuk menguji pagination) — fixture() di
// checkout_test.go selalu membuat kategori baru per produk sehingga tidak
// cocok untuk skenario ini.
func fixtureCategory(t *testing.T, db *gorm.DB, label string) models.Category {
	t.Helper()

	category := models.Category{Name: fmt.Sprintf("TEST-cat-%s-%s", t.Name(), label)}
	if err := db.Create(&category).Error; err != nil {
		t.Fatalf("create category: %v", err)
	}
	t.Cleanup(func() {
		db.Delete(&models.Category{}, "id = ?", category.ID)
	})
	return category
}

// fixtureProduct membuat produk uji di kategori tertentu dan mendaftarkan
// pembersihannya. t.Cleanup berjalan LIFO, jadi dipanggil setelah
// fixtureCategory agar produk terhapus lebih dulu (FK ke categories).
func fixtureProduct(t *testing.T, db *gorm.DB, categoryID uint, name string, stock int, price int64) models.Product {
	t.Helper()

	product := models.Product{Name: name, Price: price, Stock: stock, CategoryID: categoryID}
	if err := db.Create(&product).Error; err != nil {
		t.Fatalf("create product: %v", err)
	}
	t.Cleanup(func() {
		db.Delete(&models.Product{}, "id = ?", product.ID)
	})
	return product
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

// Tanpa page/page_size, harus dapat halaman pertama dengan page_size
// default dan meta.total menghitung seluruh baris yang match filter.
func TestProductListPaginationDefaults(t *testing.T) {
	db, r := setup(t)
	category := fixtureCategory(t, db, "defaults")
	for i := 0; i < 3; i++ {
		fixtureProduct(t, db, category.ID, fmt.Sprintf("TEST-product-defaults-%d", i), 10, 10000)
	}

	w := getProducts(r, fmt.Sprintf("category_id=%d", category.ID))
	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d: %s", w.Code, w.Body.String())
	}

	products, meta := decodeProductsWithMeta(t, w)
	if len(products) != 3 {
		t.Errorf("want 3 products, got %d", len(products))
	}
	if meta.Page != 1 || meta.PageSize != 20 || meta.Total != 3 || meta.TotalPages != 1 {
		t.Errorf("meta tidak sesuai default: %+v", meta)
	}
}

// Menjelajah 5 produk dengan page_size=2 harus menghasilkan 3 halaman
// (2+2+1) tanpa duplikat maupun yang terlewat.
func TestProductListPaginationWalksAllPages(t *testing.T) {
	db, r := setup(t)
	category := fixtureCategory(t, db, "walk")
	want := map[uint]bool{}
	for i := 0; i < 5; i++ {
		p := fixtureProduct(t, db, category.ID, fmt.Sprintf("TEST-product-walk-%d", i), 10, 10000)
		want[p.ID] = true
	}

	seen := map[uint]bool{}
	expectedCounts := []int{2, 2, 1}
	for i, page := range []int{1, 2, 3} {
		w := getProducts(r, fmt.Sprintf("category_id=%d&page=%d&page_size=2", category.ID, page))
		if w.Code != http.StatusOK {
			t.Fatalf("page %d: want 200, got %d: %s", page, w.Code, w.Body.String())
		}
		products, meta := decodeProductsWithMeta(t, w)
		if len(products) != expectedCounts[i] {
			t.Errorf("page %d: want %d item, got %d", page, expectedCounts[i], len(products))
		}
		if meta.Total != 5 || meta.TotalPages != 3 || meta.Page != page || meta.PageSize != 2 {
			t.Errorf("page %d: meta tidak sesuai: %+v", page, meta)
		}
		for _, p := range products {
			if seen[p.ID] {
				t.Errorf("produk %d muncul di lebih dari satu halaman", p.ID)
			}
			seen[p.ID] = true
		}
	}
	for id := range want {
		if !seen[id] {
			t.Errorf("produk %d tidak pernah muncul di halaman manapun", id)
		}
	}
}

// page_size di atas batas maksimum harus dipangkas ke maxPageSize (100).
func TestProductListPaginationCapsPageSize(t *testing.T) {
	_, r := setup(t)

	w := getProducts(r, "page_size=1000")
	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d: %s", w.Code, w.Body.String())
	}
	_, meta := decodeProductsWithMeta(t, w)
	if meta.PageSize != 100 {
		t.Errorf("page_size harus dipangkas ke 100, got %d", meta.PageSize)
	}
}

// page tidak valid (0, negatif, atau bukan angka) harus jatuh ke halaman 1.
func TestProductListPaginationInvalidPageDefaultsToFirst(t *testing.T) {
	_, r := setup(t)

	for _, query := range []string{"page=0", "page=-5", "page=abc"} {
		w := getProducts(r, query)
		if w.Code != http.StatusOK {
			t.Fatalf("query %q: want 200, got %d: %s", query, w.Code, w.Body.String())
		}
		_, meta := decodeProductsWithMeta(t, w)
		if meta.Page != 1 {
			t.Errorf("query %q: page harus default ke 1, got %d", query, meta.Page)
		}
	}
}
