import { useEffect, useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import api from "../api/client";
import { useCart } from "../context/CartContext";
import { formatIDR } from "../utils/format";

export default function CartPage() {
  const { items, totalPrice, updateQty, removeItem, clearCart, applyStockErrors, applyServerSnapshot } =
    useCart();
  const navigate = useNavigate();
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState("");
  const [stockErrors, setStockErrors] = useState([]);
  const [syncNotice, setSyncNotice] = useState(null);

  // Saat halaman cart dibuka: cek ulang harga & stok terkini tiap item ke
  // server, supaya angka yang basi (di localStorage sejak item ditambahkan)
  // segera terlihat/terkoreksi alih-alih baru ketahuan saat checkout gagal.
  // Sengaja hanya sekali saat mount ([]) — dibandingkan terhadap snapshot
  // cart pada render pertama, bukan cart yang terus berubah.
  useEffect(() => {
    let active = true;
    const initialItems = items;
    if (initialItems.length === 0) return undefined;

    (async () => {
      const freshById = {};
      const priceChanges = [];
      const removedNames = [];

      await Promise.all(
        initialItems.map(async (cartItem) => {
          try {
            const res = await api.get(`/products/${cartItem.id}`);
            const fresh = res.data.data;
            if (fresh.stock <= 0) {
              freshById[cartItem.id] = null;
              removedNames.push(cartItem.name);
              return;
            }
            freshById[cartItem.id] = { price: fresh.price, stock: fresh.stock };
            if (fresh.price !== cartItem.price) {
              priceChanges.push({ name: cartItem.name, oldPrice: cartItem.price, newPrice: fresh.price });
            }
          } catch (err) {
            if (err.response?.status === 404) {
              freshById[cartItem.id] = null;
              removedNames.push(cartItem.name);
            }
            // error lain (network dsb.) -> biarkan item apa adanya
          }
        })
      );

      if (!active) return;
      applyServerSnapshot(freshById);
      if (priceChanges.length > 0 || removedNames.length > 0) {
        setSyncNotice({ priceChanges, removedNames });
      }
    })();

    return () => {
      active = false;
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

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

  return (
    <section className="cart">
      <h1>Keranjang</h1>

      {syncNotice && (
        <div className="alert alert--info">
          {syncNotice.priceChanges.length > 0 && (
            <>
              <p>Harga beberapa produk telah diperbarui sejak ditambahkan ke keranjang:</p>
              <ul>
                {syncNotice.priceChanges.map((c) => (
                  <li key={c.name}>
                    {c.name}: {formatIDR(c.oldPrice)} &rarr; {formatIDR(c.newPrice)}
                  </li>
                ))}
              </ul>
            </>
          )}
          {syncNotice.removedNames.length > 0 && (
            <p>
              {syncNotice.removedNames.join(", ")} dihapus dari keranjang karena stok habis atau
              produk sudah tidak tersedia.
            </p>
          )}
          <button className="btn-icon" onClick={() => setSyncNotice(null)} aria-label="Tutup notifikasi">
            &times;
          </button>
        </div>
      )}

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

      {items.length === 0 ? (
        <div className="state">
          <p>Keranjang masih kosong.</p>
          <Link className="btn btn--primary" to="/">
            Lihat Katalog
          </Link>
        </div>
      ) : (
        <>
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

          <div className="cart__summary">
            <div className="cart__total">
              <span>Total</span>
              <strong>{formatIDR(totalPrice)}</strong>
            </div>
            <button className="btn btn--primary" onClick={handleCheckout} disabled={submitting}>
              {submitting ? "Memproses..." : "Checkout"}
            </button>
          </div>
        </>
      )}
    </section>
  );
}
