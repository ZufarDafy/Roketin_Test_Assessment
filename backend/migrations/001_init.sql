-- Dokumentasi schema MiniShop.
-- Eksekusi aktual dilakukan oleh aplikasi saat start (GORM AutoMigrate +
-- raw SQL di internal/database/database.go), file ini adalah referensi
-- schema yang setara.

CREATE TABLE categories (
    id   BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE
);

-- Harga sebagai BIGINT rupiah utuh (IDR tidak punya sen) — menghindari
-- drift presisi floating point pada kalkulasi total.
-- CHECK constraint menegakkan invariant di level DB sebagai
-- defense-in-depth, tidak hanya mengandalkan disiplin kode aplikasi.
CREATE TABLE products (
    id          BIGSERIAL PRIMARY KEY,
    name        TEXT NOT NULL,
    price       BIGINT NOT NULL CHECK (price >= 0),
    description TEXT,
    image_url   TEXT,
    stock       BIGINT NOT NULL DEFAULT 0 CHECK (stock >= 0),
    category_id BIGINT NOT NULL REFERENCES categories(id),
    created_at  TIMESTAMPTZ,
    updated_at  TIMESTAMPTZ
);

CREATE TABLE orders (
    id         BIGSERIAL PRIMARY KEY,
    total      BIGINT NOT NULL,
    created_at TIMESTAMPTZ
);

-- product_name & price adalah snapshot saat checkout agar riwayat order
-- tidak berubah ketika produk diedit/dihapus.
CREATE TABLE order_items (
    id           BIGSERIAL PRIMARY KEY,
    order_id     BIGINT NOT NULL REFERENCES orders(id),
    product_id   BIGINT,
    product_name TEXT NOT NULL,
    price        BIGINT NOT NULL,
    qty          BIGINT NOT NULL,
    subtotal     BIGINT NOT NULL
);

CREATE INDEX idx_order_items_order_id ON order_items (order_id);

-- Search nama produk memakai ILIKE '%keyword%' (leading wildcard) yang
-- tidak bisa dilayani B-tree; GIN trigram index menyelesaikan ini.
CREATE EXTENSION IF NOT EXISTS pg_trgm;
CREATE INDEX idx_products_name_trgm
    ON products USING GIN (name gin_trgm_ops);

-- Filter kategori (equality) cukup B-tree. Saat search + filter dipakai
-- bersamaan, planner menggabungkan keduanya via BitmapAnd.
CREATE INDEX idx_products_category_id ON products (category_id);
