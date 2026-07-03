# MiniShop — Product Catalog & Cart

Aplikasi katalog produk + keranjang belanja sederhana (test case Fullstack Engineer Roketin). Pengunjung bisa melihat katalog, mencari & memfilter produk, menambah ke keranjang, dan checkout. Admin bisa mengelola produk (CRUD) dan melihat daftar order.

> **Status:** Hari 1 — backend lengkap + frontend katalog. Cart, checkout UI, dan halaman admin menyusul di Hari 2.

## Tech Stack & Alasan

| Layer | Teknologi | Alasan |
|---|---|---|
| Backend | Go + Gin + GORM | REST API ringkas, ekosistem matang; GORM menyediakan migration & transaction API yang rapi |
| Database | PostgreSQL 16 (Docker) | Mendukung `pg_trgm` untuk mempercepat search substring; transaksi & row-locking solid untuk checkout |
| Frontend | React + Vite | Setup cepat, HMR, ekosistem luas |

### Catatan desain

- **Search & indexing** — search nama memakai `ILIKE '%keyword%'` (leading wildcard) yang tidak bisa dilayani B-tree, jadi dibuatkan **GIN trigram index** (`pg_trgm`). Filter kategori memakai B-tree biasa; saat keduanya dipakai bersamaan planner menggabungkan lewat BitmapAnd. Dengan 10 produk seed planner tetap memilih seq scan (lebih murah di tabel kecil) — index ini keputusan desain untuk skala data nyata, terverifikasi terpakai via `EXPLAIN` dengan `enable_seqscan=off`.
- **Checkout transaksional** — satu transaction dengan `SELECT ... FOR UPDATE` per produk (diurutkan by id untuk mencegah deadlock), validasi stok ulang di server, stok dikurangi atomik, total dihitung dari harga database (bukan kiriman client). Teruji dengan dua checkout konkuren memperebutkan stok yang sama: satu 201, satu 422, stok berakhir 0 tanpa oversell.
- **Snapshot harga** — `order_items` menyimpan nama & harga produk saat checkout sehingga riwayat order tidak berubah saat produk diedit/dihapus.

## Menjalankan Lokal

Prasyarat: Go 1.22+, Node 18+, Docker.

```bash
# 1. Database
docker compose up -d

# 2. Backend (port 8080) — migrasi + seeder otomatis saat start
cd backend
go run ./cmd/server

# 3. Frontend (port 5173)
cd frontend
npm install
npm run dev
```

Konfigurasi opsional via env (lihat `backend/.env.example` dan `frontend/.env.example`); default sudah cocok dengan docker-compose.

Seeder otomatis mengisi **4 kategori + 10 produk dummy** saat tabel products kosong.

## Endpoint API

Base URL: `http://localhost:8080/api`

| Method | Path | Deskripsi |
|---|---|---|
| GET | `/products?search=&category_id=` | List produk, search nama, filter kategori |
| GET | `/products/:id` | Detail produk |
| POST | `/products` | Tambah produk |
| PUT | `/products/:id` | Edit produk |
| DELETE | `/products/:id` | Hapus produk |
| GET | `/categories` | List kategori |
| POST | `/checkout` | Buat order, kurangi stok |
| GET | `/orders` | List order |
| GET | `/orders/:id` | Detail order |

### Contoh request/response

**POST `/products`**

```json
{ "name": "Produk Baru", "price": 10000, "description": "deskripsi", "image_url": "https://...", "stock": 5, "category_id": 2 }
```

→ `201` `{ "data": { "id": 11, "name": "Produk Baru", ... } }`

Validasi: `name` & `category_id` wajib, `price ≥ 0`, `stock ≥ 0`, `category_id` harus merujuk kategori yang ada → `400` jika gagal.

**POST `/checkout`**

```json
{ "items": [ { "product_id": 1, "qty": 2 }, { "product_id": 4, "qty": 3 } ] }
```

→ `201`

```json
{ "data": { "id": 1, "total": 603000, "items": [ { "product_id": 1, "product_name": "...", "price": 189000, "qty": 2, "subtotal": 378000 }, ... ] } }
```

→ `422` jika stok tidak cukup:

```json
{ "error": "insufficient stock for some items", "stock_errors": [ { "product_id": 2, "name": "...", "requested": 999, "available": 12 } ] }
```

→ `404` jika produk tidak ditemukan, `400` jika payload tidak valid (mis. `qty ≤ 0`).

## Known Limitations

- Cart, checkout flow UI, dan halaman admin belum ada di frontend (rencana Hari 2).
- Belum ada autentikasi — halaman admin terbuka (di luar scope requirement).
- Belum ada pagination pada list produk/order.
- Kategori hanya bisa dibaca (tidak ada CRUD kategori) — seeder yang mengisi.
