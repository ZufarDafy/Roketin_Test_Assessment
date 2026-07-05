# MiniShop — Product Catalog & Cart

Aplikasi katalog produk + keranjang belanja sederhana (test case Fullstack Engineer Roketin). Client bisa melihat katalog, mencari & memfilter produk, menambah ke keranjang, dan checkout. Admin bisa mengelola produk (CRUD) dan melihat daftar order.

## Arsitektur Sistem
```mermaid
%%{init: {'theme': 'dark'}}%%
graph TD
    Client([Browser / Pengguna]) <-->|Interaksi UI & Tampilan| Frontend[Frontend: React + Vite]
    Frontend <-->|REST API Request & Response| Backend[Backend: Go + Gin]
    Backend <-->|SQL Query & Data Result| DB[(Database: PostgreSQL 16)]
    
    classDef clientNode fill:#111111,stroke:#555555,stroke-width:2px,color:#FFFFFF;
    classDef frontendNode fill:#1A1A1A,stroke:#666666,stroke-width:2px,color:#FFFFFF;
    classDef backendNode fill:#242424,stroke:#777777,stroke-width:2px,color:#FFFFFF;
    classDef dbNode fill:#2E2E2E,stroke:#888888,stroke-width:2px,color:#FFFFFF;
    
    class Client clientNode;
    class Frontend frontendNode;
    class Backend backendNode;
    class DB dbNode;
```

## Tech Stack & Alasan

| Layer | Teknologi | Alasan |
|---|---|---|
| Backend | Go + Gin + GORM | REST API ringkas, ekosistem matang; GORM menyediakan migration & transaction API yang rapi |
| Database | PostgreSQL 16 (Docker) | Mendukung `pg_trgm` untuk mempercepat search substring; transaksi & row-locking solid untuk checkout |
| Frontend | React + Vite | Setup cepat, Hot Module Replacement, ekosistem luas |

### Catatan Desain selama Develop

- **Uang sebagai `int64` rupiah utuh** — harga/subtotal/total disimpan dan dihitung sebagai integer (`BIGINT`), menghindari kemungkinan drift presisi floating point, ini sangat kritikal jika tidak diperhatikan karena akan membuat kesalahan kalkulasi.
- **Validasi Stock** — Validasi sudah diterapkan pada code aplikasi Frontend maupun Backend, dan juga diterapkan pada Database dengan menambahkan `CHECK (stock >= 0)` dan `CHECK (price >= 0)` sebagai defense-in-depth; stok negatif ditolak database sendiri, tidak hanya oleh validasi aplikasi.
- **Stok opsional pada PUT produk** — bila field `stock` tidak dikirim, stok tidak disentuh. Sebelumnya AI Tools yang saya gunakan (Claude) membuat logika PUT produk dengan mengharuskan menambahkan stock sebagai parameter, namun secara logika saya pribadi jika perubahan produk dilakukan mengharuskan menambahkan stock sebagai parameter ada terdapat kemungkinan stock yang dimasukkan sudah berubah, sehingga Ini mencegah lost update: form edit yang menyimpan stok basi tidak menimpa hasil pengurangan stok dari checkout yang terjadi di sela-selanya. Update juga berjalan dalam transaction dengan row lock.
- **Search & indexing** — saya meminta AI untuk menggunakan fitur trigram pada Postgres untuk fitur pencarian pada aplikasi, Sepengetahuan saya trigram ini dapat mengoptimalkan speed query database karena melakukan indexing pada kolom tabel (walaupun untuk saat ini mungkin speed tidak akan terlihat signifikan karena jumlah data yang sedikit, pada kasus yang saya temukan speed terasa signifikan di 50ribu data), misal nama produk akan diindexing dengna cara dipecah menjadi 3 huruf, sehingga database tidak perlu melakukan full sequential read, search nama memakai `ILIKE '%keyword%'` (leading wildcard) yang menggunakan (`pg_trgm`). Filter kategori memakai B-tree biasa; saat keduanya dipakai bersamaan planner menggabungkan lewat BitmapAnd. Dengan 10 produk seed planner tetap memilih seq scan (lebih murah di tabel kecil) — index ini keputusan desain untuk skala data nyata, terverifikasi terpakai via `EXPLAIN` dengan `enable_seqscan=off`.
- **Optimasi Trigram** — Saya juga sudah menambahkan penyesuaian trigram ini pada frontend dan backend, dimana jika ingin melakukan search produk minimal input adalah 3 huruf, sehingga request ke backend dan query database akan berkurang.
- **Checkout transaksional** — satu transaction dengan `SELECT ... FOR UPDATE` per produk (diurutkan by id untuk mencegah deadlock), validasi stok ulang di server, stok dikurangi atomik, total dihitung dari harga database (bukan kiriman client). Tercakup integration test, termasuk kasus dua checkout konkuren memperebutkan stok yang sama: satu 201, satu 422, stok berakhir 0 tanpa oversell.
- **Snapshot harga** — `order_items` menyimpan nama & harga produk saat checkout sehingga riwayat order tidak berubah saat produk diedit/dihapus.

## Menjalankan Lokal

Prasyarat: Go 1.22+, Node 18+, Docker.

```bash
# 1. Database
docker compose up -d

# 2. Backend (port 8080) — migrasi + seeder otomatis saat start
cd backend
cp .env.example .env   # Windows (cmd): copy .env.example .env
go run ./cmd/server

# 3. Frontend (port 5173)
cd frontend
npm install
npm run dev
```

Konfigurasi backend **wajib** via env — tidak ada nilai default di kode: copy `backend/.env.example` ke `backend/.env` sebelum menjalankan server (nilai contohnya sudah cocok dengan docker-compose). Frontend opsional via `frontend/.env.example`.

Seeder otomatis mengisi **4 kategori + 10 produk dummy** saat tabel products kosong.

### Menjalankan test

Integration test checkout butuh PostgreSQL jalan dan `TEST_DB_DSN` — dibaca dari `backend/.env` (sudah ada di `.env.example`) atau dari env shell/CI (yang menang bila keduanya diset). Tanpa `TEST_DB_DSN`, test di-skip dengan pesan jelas.

```bash
cd backend
go test ./... -v
```

Mencakup: checkout sukses (penggabungan qty duplikat, total server-side, stok berkurang), stok tidak cukup (422 + detail per item, stok utuh), produk tidak ditemukan (404), **race test** dua checkout konkuren tanpa oversell, search minimal 3 karakter, dan pagination (default, penjelajahan seluruh halaman tanpa duplikat, page_size dipangkas ke maksimum, page tidak valid jatuh ke halaman 1).

## Endpoint API

Base URL: `http://localhost:8080/api`

| Method | Path | Deskripsi |
|---|---|---|
| GET | `/products?search=&category_id=&page=&page_size=` | List produk (paginated), search nama, filter kategori |
| GET | `/products/:id` | Detail produk |
| POST | `/products` | Tambah produk |
| PUT | `/products/:id` | Edit produk (`stock` opsional — tidak dikirim = tidak diubah) |
| DELETE | `/products/:id` | Hapus produk |
| GET | `/categories` | List kategori |
| POST | `/checkout` | Buat order, kurangi stok |
| GET | `/orders?page=&page_size=` | List order (paginated) |
| GET | `/orders/:id` | Detail order |

### Pagination

`GET /products` dan `GET /orders` menerima query `page` (default 1) dan `page_size` (default 20, maksimum 100 — nilai di luar batas dipangkas, nilai tidak valid jatuh ke default). Response menyertakan objek `meta`:

```json
{
  "data": [ ... ],
  "meta": { "page": 1, "page_size": 20, "total": 47, "total_pages": 3 }
}
```

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

- Belum ada autentikasi — halaman admin (`/admin/products`, `/admin/orders`) terbuka tanpa login (di luar scope requirement).
- Kategori hanya bisa dibaca (`GET /categories`) — tidak ada CRUD kategori, seeder yang mengisi.
- Stok pada item cart adalah snapshot saat produk ditambahkan, bisa basi bila stok berubah di server; backend tetap sumber kebenaran saat checkout — bila ditolak (422 karena stok kurang), cart otomatis disinkronkan dengan stok terbaru dari server.
