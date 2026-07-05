import { useEffect, useState } from "react";
import api from "../api/client";
import useDebounce from "../hooks/useDebounce";
import ProductCard from "../components/ProductCard";
import ProductModal from "../components/ProductModal";

export default function CatalogPage() {
  const [products, setProducts] = useState([]);
  const [categories, setCategories] = useState([]);
  const [search, setSearch] = useState("");
  const [categoryId, setCategoryId] = useState("");
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [selected, setSelected] = useState(null);

  const debouncedSearch = useDebounce(search);

  useEffect(() => {
    api
      .get("/categories")
      .then((res) => setCategories(res.data.data))
      .catch(() => {});
  }, []);

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
          search: debouncedSearch || undefined,
          category_id: categoryId || undefined,
        },
        signal: controller.signal,
      })
      .then((res) => {
        if (!active) return;
        setProducts(res.data.data ?? []);
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
  }, [debouncedSearch, categoryId]);

  return (
    <>
      <div className="toolbar">
        <input
          type="search"
          className="input"
          placeholder="Cari produk..."
          value={search}
          onChange={(e) => setSearch(e.target.value)}
        />
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

      <ProductModal product={selected} onClose={() => setSelected(null)} />
    </>
  );
}
