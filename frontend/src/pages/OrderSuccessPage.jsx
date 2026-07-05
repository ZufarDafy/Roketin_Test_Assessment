import { useEffect, useState } from "react";
import { Link, useLocation, useParams } from "react-router-dom";
import api from "../api/client";
import { formatIDR } from "../utils/format";

export default function OrderSuccessPage() {
  const { id } = useParams();
  const location = useLocation();
  const [order, setOrder] = useState(location.state?.order ?? null);
  const [error, setError] = useState("");

  // Fallback fetch bila halaman dibuka langsung/di-refresh (state hilang).
  useEffect(() => {
    if (order) return;
    api
      .get(`/orders/${id}`)
      .then((res) => setOrder(res.data.data))
      .catch(() => setError("Order tidak ditemukan."));
  }, [id, order]);

  if (error) return <p className="state state--error">{error}</p>;
  if (!order) return <p className="state">Memuat ringkasan order...</p>;

  return (
    <section className="order-success">
      <div className="order-success__icon">✅</div>
      <h1>Order Berhasil!</h1>
      <p className="order-success__meta">
        Order <strong>#{order.id}</strong> &middot;{" "}
        {new Date(order.created_at).toLocaleString("id-ID", {
          dateStyle: "long",
          timeStyle: "short",
        })}
      </p>

      <div className="order-summary">
        <table className="table">
          <thead>
            <tr>
              <th>Produk</th>
              <th>Harga</th>
              <th>Qty</th>
              <th>Subtotal</th>
            </tr>
          </thead>
          <tbody>
            {order.items.map((it) => (
              <tr key={it.id}>
                <td>{it.product_name}</td>
                <td>{formatIDR(it.price)}</td>
                <td>{it.qty}</td>
                <td>{formatIDR(it.subtotal)}</td>
              </tr>
            ))}
          </tbody>
          <tfoot>
            <tr>
              <td colSpan="3">Total</td>
              <td>
                <strong>{formatIDR(order.total)}</strong>
              </td>
            </tr>
          </tfoot>
        </table>
      </div>

      <Link className="btn btn--primary" to="/">
        Kembali Belanja
      </Link>
    </section>
  );
}
