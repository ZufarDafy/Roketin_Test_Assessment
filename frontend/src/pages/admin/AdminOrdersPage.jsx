import { useEffect, useState } from "react";
import api from "../../api/client";
import AdminNav from "../../components/AdminNav";
import Pagination from "../../components/Pagination";
import { formatIDR } from "../../utils/format";

export default function AdminOrdersPage() {
  const [orders, setOrders] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [openId, setOpenId] = useState(null);
  const [page, setPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);

  useEffect(() => {
    setLoading(true);
    api
      .get("/orders", { params: { page } })
      .then((res) => {
        setOrders(res.data.data ?? []);
        setTotalPages(res.data.meta?.total_pages ?? 1);
      })
      .catch(() => setError("Gagal memuat orders."))
      .finally(() => setLoading(false));
  }, [page]);

  if (loading) return <p className="state">Memuat...</p>;
  if (error) return <p className="state state--error">{error}</p>;
  if (orders.length === 0) return <p className="state">Belum ada order.</p>;

  return (
    <section>
      <AdminNav />
      <div className="page-head">
        <h1>Daftar Order</h1>
      </div>
      <div className="table-wrap">
        <table className="table">
          <thead>
            <tr>
              <th>ID</th>
              <th>Tanggal</th>
              <th>Jumlah Item</th>
              <th>Total</th>
              <th></th>
            </tr>
          </thead>
          <tbody>
            {orders.map((o) => (
              <OrderRow
                key={o.id}
                order={o}
                open={openId === o.id}
                onToggle={() => setOpenId(openId === o.id ? null : o.id)}
              />
            ))}
          </tbody>
        </table>
      </div>

      <Pagination page={page} totalPages={totalPages} onChange={setPage} />
    </section>
  );
}

function OrderRow({ order, open, onToggle }) {
  const itemCount = order.items.reduce((sum, it) => sum + it.qty, 0);

  return (
    <>
      <tr>
        <td>#{order.id}</td>
        <td>
          {new Date(order.created_at).toLocaleString("id-ID", {
            dateStyle: "medium",
            timeStyle: "short",
          })}
        </td>
        <td>{itemCount}</td>
        <td>{formatIDR(order.total)}</td>
        <td>
          <button className="btn btn--small" onClick={onToggle}>
            {open ? "Tutup" : "Detail"}
          </button>
        </td>
      </tr>
      {open && (
        <tr className="order-detail-row">
          <td colSpan="5">
            <table className="table table--nested">
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
            </table>
          </td>
        </tr>
      )}
    </>
  );
}
