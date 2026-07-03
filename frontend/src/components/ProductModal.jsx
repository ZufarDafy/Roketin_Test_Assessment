import { useEffect } from "react";
import { formatIDR } from "../utils/format";

export default function ProductModal({ product, onClose }) {
  useEffect(() => {
    const onKey = (e) => e.key === "Escape" && onClose();
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [onClose]);

  if (!product) return null;

  const outOfStock = product.stock <= 0;

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
          <button className="btn btn--primary" disabled={outOfStock}>
            {outOfStock ? "Stok habis" : "Tambah ke Keranjang"}
          </button>
        </div>
      </div>
    </div>
  );
}
