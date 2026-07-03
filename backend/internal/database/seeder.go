package database

import (
	"fmt"

	"minishop/internal/models"

	"gorm.io/gorm"
)

// Seed mengisi kategori & produk dummy. Idempotent: tidak menambah data
// jika tabel products sudah terisi.
func Seed(db *gorm.DB) error {
	var count int64
	if err := db.Model(&models.Product{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	categories := []models.Category{
		{Name: "Elektronik"},
		{Name: "Fashion"},
		{Name: "Olahraga"},
		{Name: "Rumah Tangga"},
	}
	if err := db.Create(&categories).Error; err != nil {
		return err
	}

	catID := map[string]uint{}
	for _, c := range categories {
		catID[c.Name] = c.ID
	}

	products := []models.Product{
		{Name: "Wireless Mouse Logitek M190", Price: 189000, Description: "Mouse nirkabel 2.4GHz, baterai tahan hingga 18 bulan.", Stock: 25, CategoryID: catID["Elektronik"]},
		{Name: "Mechanical Keyboard TKL RGB", Price: 549000, Description: "Keyboard mekanik tenkeyless, switch blue, lampu RGB.", Stock: 12, CategoryID: catID["Elektronik"]},
		{Name: "Earbuds TWS Bass Boost", Price: 299000, Description: "True wireless earbuds dengan bass kuat dan case charging.", Stock: 30, CategoryID: catID["Elektronik"]},
		{Name: "Kaos Polos Cotton Combed 30s", Price: 75000, Description: "Kaos polos bahan cotton combed 30s, adem dan nyaman.", Stock: 50, CategoryID: catID["Fashion"]},
		{Name: "Kemeja Flanel Kotak-Kotak", Price: 165000, Description: "Kemeja flanel lengan panjang motif kotak, bahan tebal halus.", Stock: 20, CategoryID: catID["Fashion"]},
		{Name: "Celana Chino Slim Fit", Price: 189000, Description: "Celana chino slim fit stretch, nyaman untuk harian.", Stock: 18, CategoryID: catID["Fashion"]},
		{Name: "Sepatu Lari Ringan Runmax", Price: 425000, Description: "Sepatu lari ringan dengan sol empuk responsif.", Stock: 15, CategoryID: catID["Olahraga"]},
		{Name: "Matras Yoga Anti Slip 6mm", Price: 135000, Description: "Matras yoga tebal 6mm, permukaan anti slip, termasuk strap.", Stock: 22, CategoryID: catID["Olahraga"]},
		{Name: "Botol Minum Stainless 750ml", Price: 98000, Description: "Botol minum stainless steel, tahan panas & dingin 12 jam.", Stock: 40, CategoryID: catID["Rumah Tangga"]},
		{Name: "Lampu Meja LED Minimalis", Price: 145000, Description: "Lampu meja LED 3 mode warna, dimmable, port USB.", Stock: 17, CategoryID: catID["Rumah Tangga"]},
	}
	for i := range products {
		products[i].ImageURL = fmt.Sprintf("https://picsum.photos/seed/minishop-%d/400/300", i+1)
	}
	return db.Create(&products).Error
}
