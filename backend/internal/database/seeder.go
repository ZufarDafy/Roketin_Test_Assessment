package database

import (
	"fmt"

	"minishop/internal/models"

	"gorm.io/gorm"
)

const (
	catElektronik  = "Elektronik"
	catFashion     = "Fashion"
	catOlahraga    = "Olahraga"
	catRumahTangga = "Rumah Tangga"
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
		{Name: catElektronik},
		{Name: catFashion},
		{Name: catOlahraga},
		{Name: catRumahTangga},
	}
	if err := db.Create(&categories).Error; err != nil {
		return err
	}

	catID := map[string]uint{}
	for _, c := range categories {
		catID[c.Name] = c.ID
	}

	// 50 produk (13 Elektronik, 13 Fashion, 12 Olahraga, 12 Rumah Tangga) —
	// sengaja lebih dari minimum 5-10 di requirement supaya pagination
	// (default page_size=20) langsung terlihat berfungsi tanpa perlu
	// menambah produk manual lewat admin.
	products := []models.Product{
		// Elektronik
		{Name: "Wireless Mouse Logitek M190", Price: 189000, Description: "Mouse nirkabel 2.4GHz, baterai tahan hingga 18 bulan.", Stock: 25, CategoryID: catID[catElektronik]},
		{Name: "Mechanical Keyboard TKL RGB", Price: 549000, Description: "Keyboard mekanik tenkeyless, switch blue, lampu RGB.", Stock: 12, CategoryID: catID[catElektronik]},
		{Name: "Earbuds TWS Bass Boost", Price: 299000, Description: "True wireless earbuds dengan bass kuat dan case charging.", Stock: 30, CategoryID: catID[catElektronik]},
		{Name: "Webcam HD 1080p Autofocus", Price: 249000, Description: "Webcam 1080p dengan autofocus dan mikrofon built-in.", Stock: 20, CategoryID: catID[catElektronik]},
		{Name: "Speaker Bluetooth Portable", Price: 199000, Description: "Speaker bluetooth ringkas, suara jernih, baterai 10 jam.", Stock: 28, CategoryID: catID[catElektronik]},
		{Name: "Power Bank 10000mAh Fast Charging", Price: 189000, Description: "Power bank 10000mAh dengan fast charging dua arah.", Stock: 35, CategoryID: catID[catElektronik]},
		{Name: "Charger USB-C 65W GaN", Price: 259000, Description: "Charger GaN 65W, ringkas, mendukung fast charging laptop & HP.", Stock: 24, CategoryID: catID[catElektronik]},
		{Name: "Mouse Pad Gaming XL", Price: 89000, Description: "Mouse pad gaming ukuran XL, permukaan halus anti slip.", Stock: 45, CategoryID: catID[catElektronik]},
		{Name: "Headset Gaming Surround 7.1", Price: 379000, Description: "Headset gaming surround 7.1 dengan mikrofon noise-cancelling.", Stock: 16, CategoryID: catID[catElektronik]},
		{Name: "Flashdisk 64GB USB 3.0", Price: 95000, Description: "Flashdisk 64GB USB 3.0, transfer cepat, bodi metal.", Stock: 50, CategoryID: catID[catElektronik]},
		{Name: "Kabel HDMI 2.0 2 Meter", Price: 65000, Description: "Kabel HDMI 2.0 mendukung resolusi 4K, panjang 2 meter.", Stock: 40, CategoryID: catID[catElektronik]},
		{Name: "Smartwatch Fitness Tracker", Price: 349000, Description: "Smartwatch dengan pelacak detak jantung dan notifikasi HP.", Stock: 18, CategoryID: catID[catElektronik]},
		{Name: "Router WiFi Dual Band AC1200", Price: 289000, Description: "Router dual band AC1200, jangkauan luas, 4 antena.", Stock: 14, CategoryID: catID[catElektronik]},

		// Fashion
		{Name: "Kaos Polos Cotton Combed 30s", Price: 75000, Description: "Kaos polos bahan cotton combed 30s, adem dan nyaman.", Stock: 50, CategoryID: catID[catFashion]},
		{Name: "Kemeja Flanel Kotak-Kotak", Price: 165000, Description: "Kemeja flanel lengan panjang motif kotak, bahan tebal halus.", Stock: 20, CategoryID: catID[catFashion]},
		{Name: "Celana Chino Slim Fit", Price: 189000, Description: "Celana chino slim fit stretch, nyaman untuk harian.", Stock: 18, CategoryID: catID[catFashion]},
		{Name: "Jaket Hoodie Oversize", Price: 175000, Description: "Jaket hoodie oversize bahan fleece tebal, hangat.", Stock: 22, CategoryID: catID[catFashion]},
		{Name: "Rok Plisket Midi", Price: 145000, Description: "Rok plisket midi, bahan jatuh, cocok untuk kerja maupun santai.", Stock: 19, CategoryID: catID[catFashion]},
		{Name: "Kaos Kaki Motif Polos 3 Pasang", Price: 45000, Description: "Kaos kaki katun 3 pasang, motif polos, elastis.", Stock: 60, CategoryID: catID[catFashion]},
		{Name: "Topi Baseball Cap Adjustable", Price: 85000, Description: "Topi baseball cap adjustable, bahan katun twill.", Stock: 32, CategoryID: catID[catFashion]},
		{Name: "Tas Selempang Kanvas", Price: 135000, Description: "Tas selempang kanvas, muat tablet 10 inci, tali kulit.", Stock: 26, CategoryID: catID[catFashion]},
		{Name: "Dress Casual Wanita", Price: 165000, Description: "Dress casual wanita bahan katun rayon, adem dipakai.", Stock: 21, CategoryID: catID[catFashion]},
		{Name: "Sweater Rajut Unisex", Price: 155000, Description: "Sweater rajut unisex, hangat, cocok untuk cuaca dingin.", Stock: 24, CategoryID: catID[catFashion]},
		{Name: "Ikat Pinggang Kulit Sintetis", Price: 79000, Description: "Ikat pinggang kulit sintetis, gesper metal anti karat.", Stock: 38, CategoryID: catID[catFashion]},
		{Name: "Sarung Tangan Rajut Winter", Price: 55000, Description: "Sarung tangan rajut tebal untuk cuaca dingin.", Stock: 42, CategoryID: catID[catFashion]},
		{Name: "Sandal Slide Casual", Price: 99000, Description: "Sandal slide casual, sol empuk, anti slip.", Stock: 30, CategoryID: catID[catFashion]},

		// Olahraga
		{Name: "Sepatu Lari Ringan Runmax", Price: 425000, Description: "Sepatu lari ringan dengan sol empuk responsif.", Stock: 15, CategoryID: catID[catOlahraga]},
		{Name: "Matras Yoga Anti Slip 6mm", Price: 135000, Description: "Matras yoga tebal 6mm, permukaan anti slip, termasuk strap.", Stock: 22, CategoryID: catID[catOlahraga]},
		{Name: "Dumbbell Set 2x3kg", Price: 175000, Description: "Sepasang dumbbell 3kg dengan lapisan vinyl anti licin.", Stock: 20, CategoryID: catID[catOlahraga]},
		{Name: "Resistance Band Set 5 Level", Price: 89000, Description: "Set resistance band 5 level tahanan, termasuk tas jinjing.", Stock: 34, CategoryID: catID[catOlahraga]},
		{Name: "Skipping Rope Adjustable", Price: 45000, Description: "Tali skipping adjustable dengan bearing ball halus.", Stock: 48, CategoryID: catID[catOlahraga]},
		{Name: "Sarung Tinju Boxing Gloves", Price: 195000, Description: "Sarung tinju bahan kulit sintetis, busa padat pelindung.", Stock: 17, CategoryID: catID[catOlahraga]},
		{Name: "Botol Shaker Protein 700ml", Price: 65000, Description: "Botol shaker 700ml dengan mixer ball, bebas BPA.", Stock: 44, CategoryID: catID[catOlahraga]},
		{Name: "Handuk Olahraga Quick Dry", Price: 55000, Description: "Handuk olahraga microfiber, cepat kering, ringkas.", Stock: 39, CategoryID: catID[catOlahraga]},
		{Name: "Sepeda Statis Mini Portable", Price: 425000, Description: "Sepeda statis mini portable dengan layar penghitung kalori.", Stock: 10, CategoryID: catID[catOlahraga]},
		{Name: "Ankle Weight 2x1kg", Price: 85000, Description: "Pemberat pergelangan kaki 1kg sepasang, strap velcro.", Stock: 28, CategoryID: catID[catOlahraga]},
		{Name: "Foam Roller Muscle Massage", Price: 145000, Description: "Foam roller untuk pemulihan otot pasca olahraga.", Stock: 18, CategoryID: catID[catOlahraga]},
		{Name: "Tas Gym Duffel Bag", Price: 165000, Description: "Tas gym duffel bag dengan kompartemen sepatu terpisah.", Stock: 23, CategoryID: catID[catOlahraga]},

		// Rumah Tangga
		{Name: "Botol Minum Stainless 750ml", Price: 98000, Description: "Botol minum stainless steel, tahan panas & dingin 12 jam.", Stock: 40, CategoryID: catID[catRumahTangga]},
		{Name: "Lampu Meja LED Minimalis", Price: 145000, Description: "Lampu meja LED 3 mode warna, dimmable, port USB.", Stock: 17, CategoryID: catID[catRumahTangga]},
		{Name: "Rak Sepatu Susun 5 Tingkat", Price: 195000, Description: "Rak sepatu susun 5 tingkat, bahan besi, mudah dirakit.", Stock: 15, CategoryID: catID[catRumahTangga]},
		{Name: "Set Container Makanan Kedap Udara", Price: 125000, Description: "Set 5 container makanan kedap udara, bebas BPA.", Stock: 30, CategoryID: catID[catRumahTangga]},
		{Name: "Sapu + Pengki Set", Price: 65000, Description: "Set sapu dan pengki dengan gagang panjang anti karat.", Stock: 36, CategoryID: catID[catRumahTangga]},
		{Name: "Gantungan Baju Lipat Portable", Price: 45000, Description: "Gantungan baju lipat portable, cocok untuk traveling.", Stock: 50, CategoryID: catID[catRumahTangga]},
		{Name: "Termos Air Panas 1.5L", Price: 155000, Description: "Termos air panas 1.5L, mempertahankan suhu hingga 12 jam.", Stock: 25, CategoryID: catID[catRumahTangga]},
		{Name: "Talenan Kayu Anti Bakteri", Price: 75000, Description: "Talenan kayu solid, permukaan anti bakteri, tahan lama.", Stock: 33, CategoryID: catID[catRumahTangga]},
		{Name: "Rak Dinding Melayang Minimalis", Price: 135000, Description: "Rak dinding melayang minimalis, mudah dipasang tanpa bor besar.", Stock: 20, CategoryID: catID[catRumahTangga]},
		{Name: "Keset Kamar Mandi Anti Slip", Price: 55000, Description: "Keset kamar mandi anti slip, cepat menyerap air.", Stock: 42, CategoryID: catID[catRumahTangga]},
		{Name: "Tempat Sampah Injak 15L", Price: 115000, Description: "Tempat sampah injak 15L, tutup senyap, mudah dibersihkan.", Stock: 27, CategoryID: catID[catRumahTangga]},
		{Name: "Set Peralatan Makan Melamin", Price: 145000, Description: "Set peralatan makan melamin 16 pcs, ringan dan tidak mudah pecah.", Stock: 19, CategoryID: catID[catRumahTangga]},
	}
	for i := range products {
		products[i].ImageURL = fmt.Sprintf("https://picsum.photos/seed/minishop-%d/400/300", i+1)
	}
	return db.Create(&products).Error
}
