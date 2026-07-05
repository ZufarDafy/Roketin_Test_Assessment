import { useEffect, useState } from "react";

const emptyForm = {
  name: "",
  price: "",
  description: "",
  image_url: "",
  stock: "",
  category_id: "",
};

// Dipakai untuk tambah (product=null) dan edit. Saat edit, field stok
// dikosongkan secara default: backend memperlakukan stok yang tidak
// dikirim sebagai "jangan diubah" (mencegah menimpa stok basi).
export default function ProductFormModal({ product, categories, onSubmit, onClose }) {
  const [form, setForm] = useState(emptyForm);
  const [error, setError] = useState("");
  const [submitting, setSubmitting] = useState(false);

  const isEdit = Boolean(product);

  useEffect(() => {
    if (product) {
      setForm({
        name: product.name,
        price: String(product.price),
        description: product.description ?? "",
        image_url: product.image_url ?? "",
        stock: "",
        category_id: String(product.category_id),
      });
    } else {
      setForm(emptyForm);
    }
    setError("");
  }, [product]);

  const set = (field) => (e) => setForm((f) => ({ ...f, [field]: e.target.value }));

  const handleSubmit = async (e) => {
    e.preventDefault();
    setSubmitting(true);
    setError("");

    const payload = {
      name: form.name.trim(),
      price: Number(form.price),
      description: form.description.trim(),
      image_url: form.image_url.trim(),
      category_id: Number(form.category_id),
    };
    if (form.stock !== "") payload.stock = Number(form.stock);

    try {
      await onSubmit(payload);
      onClose();
    } catch (err) {
      setError(err.response?.data?.error || "Gagal menyimpan produk.");
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal modal--form" role="dialog" aria-modal="true" onClick={(e) => e.stopPropagation()}>
        <button className="modal__close" onClick={onClose} aria-label="Tutup">
          &times;
        </button>
        <form className="form" onSubmit={handleSubmit}>
          <h2>{isEdit ? "Edit Produk" : "Tambah Produk"}</h2>
          {error && <p className="alert alert--error">{error}</p>}

          <label>
            Nama
            <input className="input" required value={form.name} onChange={set("name")} />
          </label>
          <label>
            Harga (Rp)
            <input
              className="input"
              type="number"
              min="0"
              required
              value={form.price}
              onChange={set("price")}
            />
          </label>
          <label>
            Deskripsi
            <textarea className="input" rows="3" value={form.description} onChange={set("description")} />
          </label>
          <label>
            URL Gambar
            <input className="input" value={form.image_url} onChange={set("image_url")} />
          </label>
          <label>
            Stok {isEdit && <small>(kosongkan bila tidak ingin mengubah stok)</small>}
            <input
              className="input"
              type="number"
              min="0"
              required={!isEdit}
              value={form.stock}
              onChange={set("stock")}
              placeholder={isEdit ? `saat ini: ${product.stock}` : ""}
            />
          </label>
          <label>
            Kategori
            <select className="input" required value={form.category_id} onChange={set("category_id")}>
              <option value="">Pilih kategori...</option>
              {categories.map((c) => (
                <option key={c.id} value={c.id}>
                  {c.name}
                </option>
              ))}
            </select>
          </label>

          <button className="btn btn--primary" type="submit" disabled={submitting}>
            {submitting ? "Menyimpan..." : "Simpan"}
          </button>
        </form>
      </div>
    </div>
  );
}
