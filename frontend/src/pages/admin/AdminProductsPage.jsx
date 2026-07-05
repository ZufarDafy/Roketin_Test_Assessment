import { useCallback, useEffect, useState } from "react";
import api from "../../api/client";
import AdminNav from "../../components/AdminNav";
import ProductFormModal from "../../components/ProductFormModal";
import { formatIDR } from "../../utils/format";

export default function AdminProductsPage() {
  const [products, setProducts] = useState([]);
  const [categories, setCategories] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  // modal: null = tertutup, "new" = tambah, object = edit produk tsb
  const [modal, setModal] = useState(null);

  const fetchProducts = useCallback(() => {
    setLoading(true);
    api
      .get("/products")
      .then((res) => setProducts(res.data.data ?? []))
      .catch(() => setError("Gagal memuat produk."))
      .finally(() => setLoading(false));
  }, []);

  useEffect(() => {
    fetchProducts();
    api
      .get("/categories")
      .then((res) => setCategories(res.data.data))
      .catch(() => {});
  }, [fetchProducts]);

  const handleSubmit = async (payload) => {
    if (modal === "new") {
      await api.post("/products", payload);
    } else {
      await api.put(`/products/${modal.id}`, payload);
    }
    fetchProducts();
  };

  const handleDelete = async (product) => {
    if (!window.confirm(`Hapus produk "${product.name}"?`)) return;
    try {
      await api.delete(`/products/${product.id}`);
      fetchProducts();
    } catch {
      setError("Gagal menghapus produk.");
    }
  };

  return (
    <section>
      <AdminNav />
      <div className="page-head">
        <h1>Kelola Produk</h1>
        <button className="btn btn--primary" onClick={() => setModal("new")}>
          + Tambah Produk
        </button>
      </div>

      {error && <p className="alert alert--error">{error}</p>}
      {loading ? (
        <p className="state">Memuat...</p>
      ) : (
        <div className="table-wrap">
          <table className="table">
            <thead>
              <tr>
                <th>ID</th>
                <th>Nama</th>
                <th>Kategori</th>
                <th>Harga</th>
                <th>Stok</th>
                <th>Aksi</th>
              </tr>
            </thead>
            <tbody>
              {products.map((p) => (
                <tr key={p.id}>
                  <td>{p.id}</td>
                  <td>{p.name}</td>
                  <td>{p.category?.name}</td>
                  <td>{formatIDR(p.price)}</td>
                  <td>{p.stock}</td>
                  <td className="table__actions">
                    <button className="btn btn--small" onClick={() => setModal(p)}>
                      Edit
                    </button>
                    <button className="btn btn--small btn--danger" onClick={() => handleDelete(p)}>
                      Hapus
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {modal !== null && (
        <ProductFormModal
          product={modal === "new" ? null : modal}
          categories={categories}
          onSubmit={handleSubmit}
          onClose={() => setModal(null)}
        />
      )}
    </section>
  );
}
