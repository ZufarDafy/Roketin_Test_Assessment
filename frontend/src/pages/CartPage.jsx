import { useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import api from "../api/client";
import { useCart } from "../context/CartContext";
import { formatIDR } from "../utils/format";

export default function CartPage() {
  const { items, totalPrice, updateQty, removeItem, clearCart, applyStockErrors } = useCart();
  const navigate = useNavigate();
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState("");
  const [stockErrors, setStockErrors] = useState([]);

  const handleCheckout = async () => {
    setSubmitting(true);
    setError("");
    setStockErrors([]);
    try {
      const res = await api.post("/checkout", {
        items: items.map((it) => ({ product_id: it.id, qty: it.qty })),
      });
      const order = res.data.data;
      clearCart();
      navigate(`/order-success/${order.id}`, { state: { order } });
    } catch (err) {
      const data = err.response?.data;
      if (err.response?.status === 422 && data?.stock_errors) {
        setStockErrors(data.stock_errors);
        applyStockErrors(data.stock_errors);
        setError("Sebagian item melebihi stok tersedia. Jumlah di keranjang sudah disesuaikan — silakan periksa lalu checkout lagi.");
      } else {
        setError(data?.error || "Checkout gagal. Coba lagi.");
      }
    } finally {
      setSubmitting(false);
    }
  };

  if (items.length === 0 && !error) {
    return (
      <div className="state">
        <p>Keranjang masih kosong.</p>
        <Link className="btn btn--primary" to="/">
          Lihat Katalog
        </Link>
      </div>
    );
  }

  return (
    <section className="cart">
      <h1>Keranjang</h1>

      {error && (
        <div className="alert alert--error">
          <p>{error}</p>
          {stockErrors.length > 0 && (
            <ul>
              {stockErrors.map((se) => (
                <li key={se.product_id}>
                  {se.name}: diminta {se.requested}, tersedia {se.available}
                </li>
              ))}
            </ul>
          )}
        </div>
      )}

      <div className="cart__list">
        {items.map((it) => (
          <div className="cart-item" key={it.id}>
            <img src={it.image_url} alt={it.name} />
            <div className="cart-item__info">
              <strong>{it.name}</strong>
              <span className="cart-item__price">{formatIDR(it.price)}</span>
            </div>
            <div className="qty-stepper">
              <button onClick={() => updateQty(it.id, it.qty - 1)} aria-label="Kurangi">
                &minus;
              </button>
              <input
                type="number"
                min="1"
                max={it.stock}
                value={it.qty}
                onChange={(e) => updateQty(it.id, Number(e.target.value) || 1)}
                aria-label={`Jumlah ${it.name}`}
              />
              <button
                onClick={() => updateQty(it.id, it.qty + 1)}
                disabled={it.qty >= it.stock}
                aria-label="Tambah"
              >
                +
              </button>
            </div>
            <strong className="cart-item__subtotal">{formatIDR(it.price * it.qty)}</strong>
            <button
              className="btn-icon"
              onClick={() => removeItem(it.id)}
              aria-label={`Hapus ${it.name}`}
            >
              🗑
            </button>
          </div>
        ))}
      </div>

      {items.length > 0 && (
        <div className="cart__summary">
          <div className="cart__total">
            <span>Total</span>
            <strong>{formatIDR(totalPrice)}</strong>
          </div>
          <button className="btn btn--primary" onClick={handleCheckout} disabled={submitting}>
            {submitting ? "Memproses..." : "Checkout"}
          </button>
        </div>
      )}
    </section>
  );
}
