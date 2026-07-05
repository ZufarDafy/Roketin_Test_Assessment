import { useEffect, useState } from "react";
import api from "../api/client";
import useDebounce from "../hooks/useDebounce";
import ProductCard from "../components/ProductCard";
import ProductModal from "../components/ProductModal";
import Pagination from "../components/Pagination";

// Selaras dengan backend (lihat minSearchLen di product.go): pg_trgm
// membentuk trigram dari pecahan 3 karakter, jadi search < 3 karakter
// tidak bisa memanfaatkan index dan sengaja tidak dikirim ke server.
const MIN_SEARCH_LEN = 3;

export default function CatalogPage() {
  const [products, setProducts] = useState([]);
  const [categories, setCategories] = useState([]);
  const [search, setSearch] = useState("");
  const [categoryId, setCategoryId] = useState("");
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [selected, setSelected] = useState(null);
  const [page, setPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);

  const debouncedSearch = useDebounce(search);
  const trimmedSearch = debouncedSearch.trim();
  // Hint dihitung dari input mentah (bukan debouncedSearch) supaya
  // muncul instan saat mengetik, terlepas dari jeda debounce request.
  const trimmedInput = search.trim();
  const isSearchTooShort = trimmedInput.length > 0 && trimmedInput.length < MIN_SEARCH_LEN;

  useEffect(() => {
    api
      .get("/categories")
      .then((res) => setCategories(res.data.data))
      .catch(() => {});
  }, []);

  // Filter berubah -> selalu mulai lagi dari halaman 1 (halaman lama bisa
  // out-of-range untuk filter yang baru).
  useEffect(() => {
    setPage(1);
  }, [trimmedSearch, categoryId]);

  useEffect(() => {
    // Flag "active" mencegah request lama yang di-abort menimpa state
    // request baru (finally-nya baru jalan setelah effect berikutnya).
    let active = true;
    const controller = new AbortController();
    setLoading(true);
    setError("");

    api
      .get("/products", {
        params: {
          search: trimmedSearch.length >= MIN_SEARCH_LEN ? trimmedSearch : undefined,
          category_id: categoryId || undefined,
          page,
        },
        signal: controller.signal,
      })
      .then((res) => {
        if (!active) return;
        setProducts(res.data.data ?? []);
        setTotalPages(res.data.meta?.total_pages ?? 1);
      })
      .catch(() => {
        if (!active) return;
        setError("Gagal memuat produk. Pastikan backend berjalan.");
      })
      .finally(() => {
        if (active) setLoading(false);
      });

    return () => {
      active = false;
      controller.abort();
    };
  }, [trimmedSearch, categoryId, page]);

  return (
    <>
      <div className="toolbar">
        <div className="toolbar__search">
          <input
            type="search"
            className="input"
            placeholder="Cari produk..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
          />
          {isSearchTooShort && (
            <p className="search-hint">Ketik minimal 3 huruf untuk mencari</p>
          )}
        </div>
        <select
          className="input"
          value={categoryId}
          onChange={(e) => setCategoryId(e.target.value)}
        >
          <option value="">Semua kategori</option>
          {categories.map((c) => (
            <option key={c.id} value={c.id}>
              {c.name}
            </option>
          ))}
        </select>
      </div>

      {error && <p className="state state--error">{error}</p>}
      {loading && !error && <p className="state">Memuat produk...</p>}
      {!loading && !error && products.length === 0 && (
        <p className="state">Tidak ada produk yang cocok.</p>
      )}

      <div className="product-grid">
        {products.map((p) => (
          <ProductCard key={p.id} product={p} onSelect={setSelected} />
        ))}
      </div>

      <Pagination page={page} totalPages={totalPages} onChange={setPage} />

      <ProductModal product={selected} onClose={() => setSelected(null)} />
    </>
  );
}
