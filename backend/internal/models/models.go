package models

import "time"

type Category struct {
	ID   uint   `json:"id" gorm:"primaryKey"`
	Name string `json:"name" gorm:"uniqueIndex;not null"`
}

// Harga disimpan sebagai int64 rupiah (IDR tidak punya sen) —
// menghindari drift presisi floating point pada kalkulasi total.
type Product struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"not null"`
	Price       int64     `json:"price" gorm:"not null;check:price >= 0"`
	Description string    `json:"description"`
	ImageURL    string    `json:"image_url"`
	Stock       int       `json:"stock" gorm:"not null;default:0;check:stock >= 0"`
	CategoryID  uint      `json:"category_id" gorm:"not null"`
	Category    *Category `json:"category,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Order struct {
	ID        uint        `json:"id" gorm:"primaryKey"`
	Total     int64       `json:"total" gorm:"not null"`
	Items     []OrderItem `json:"items"`
	CreatedAt time.Time   `json:"created_at"`
}

// OrderItem menyimpan snapshot nama & harga produk saat checkout,
// sehingga riwayat order tidak berubah ketika produk diedit/dihapus.
type OrderItem struct {
	ID          uint   `json:"id" gorm:"primaryKey"`
	OrderID     uint   `json:"order_id" gorm:"not null;index"`
	ProductID   uint   `json:"product_id"`
	ProductName string `json:"product_name" gorm:"not null"`
	Price       int64  `json:"price" gorm:"not null"`
	Qty         int    `json:"qty" gorm:"not null"`
	Subtotal    int64  `json:"subtotal" gorm:"not null"`
}
