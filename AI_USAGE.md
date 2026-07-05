# AI Usage

## Tools

- **Claude Code** (Anthropic) dengan model Claude Fable 5 — dipakai sebagai pair programmer: menyusun rencana, menulis boilerplate, dan mengeksekusi smoke test. **Keputusan arsitektur (pemilihan stack, strategi indexing trigram, desain checkout) ditentukan dan direview oleh saya.**

## Contoh Prompt untuk Bagian Krusial

### 1. Strategi indexing untuk search & filter

> "untuk database kita perlu garis bawahi bahwa functional requirement terdapat search dan filter, bagaimana dan indexingnya, lalu saya menyarankan menggunakan trigram pada postgres supaya lebih mempercepat query untuk searchnya"

Hasil: search `ILIKE '%keyword%'` dilayani GIN index dengan `gin_trgm_ops` (B-tree tidak bisa dipakai karena leading wildcard), filter kategori memakai B-tree, dan keduanya digabung planner via BitmapAnd. Diverifikasi dengan `EXPLAIN` bahwa planner memakai `idx_products_name_trgm`.

### 2. Logika checkout & pengurangan stok

Prompt (inti dari rencana yang disepakati sebelum implementasi):

> "Checkout harus dalam satu transaction GORM: lock tiap produk dengan SELECT ... FOR UPDATE, validasi ulang stok di dalam transaction (qty > stok → rollback dan 422 dengan pesan per item), kurangi stok, buat order + order_items dengan snapshot nama/harga, dan total dihitung di server dari harga database — bukan dari client."

Hasil: `backend/internal/handlers/checkout.go` — termasuk penggabungan qty untuk product_id duplikat dan pemrosesan terurut by id untuk mencegah deadlock. Diverifikasi dengan tes race: dua checkout konkuren memperebutkan stok yang sama → satu 201, satu 422, stok berakhir 0 (tidak oversell).
