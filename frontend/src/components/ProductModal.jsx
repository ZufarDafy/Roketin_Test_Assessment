import { useEffect, useState } from "react";
import { useCart } from "../context/CartContext";
import { formatIDR } from "../utils/format";

export default function ProductModal({ product, onClose }) {
  const { addItem } = useCart();
  const [qty, setQty] = useState(1);
  const [added, setAdded] = useState(false);

  useEffect(() => {
    const onKey = (e) => e.key === "Escape" && onClose();
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [onClose]);

  useEffect(() => {
    setQty(1);
    setAdded(false);
  }, [product]);

  if (!product) return null;

  const outOfStock = product.stock <= 0;

  const handleAdd = () => {
    addItem(product, qty);
    setAdded(true);
    setTimeout(onClose, 600);
  };

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div
        className="modal"
        role="dialog"
        aria-modal="true"
        aria-label={product.name}
        onClick={(e) => e.stopPropagation()}
      >
        <button className="modal__close" onClick={onClose} aria-label="Tutup">
          &times;
        </button>
        <img className="modal__image" src={product.image_url} alt={product.name} />
        <div className="modal__body">
          <span className="product-card__category">{product.category?.name}</span>
          <h2>{product.name}</h2>
          <p className="modal__desc">{product.description}</p>
          <div className="modal__meta">
            <strong className="modal__price">{formatIDR(product.price)}</strong>
            <span className="product-card__stock">Stok: {product.stock}</span>
          </div>
          {!outOfStock && (
            <div className="qty-row">
              <label htmlFor="modal-qty">Jumlah</label>
              <div className="qty-stepper">
                <button onClick={() => setQty((q) => Math.max(1, q - 1))} aria-label="Kurangi">
                  &minus;
                </button>
                <input
                  id="modal-qty"
                  type="number"
                  min="1"
                  max={product.stock}
                  value={qty}
                  onChange={(e) =>
                    setQty(Math.max(1, Math.min(product.stock, Number(e.target.value) || 1)))
                  }
                />
                <button
                  onClick={() => setQty((q) => Math.min(product.stock, q + 1))}
                  aria-label="Tambah"
                >
                  +
                </button>
              </div>
            </div>
          )}
          <button className="btn btn--primary" disabled={outOfStock || added} onClick={handleAdd}>
            {outOfStock ? "Stok habis" : added ? "Ditambahkan ✓" : "Tambah ke Keranjang"}
          </button>
        </div>
      </div>
    </div>
  );
}
