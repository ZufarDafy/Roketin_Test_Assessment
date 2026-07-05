import { NavLink } from "react-router-dom";

export default function AdminNav() {
  return (
    <nav className="admin-tabs">
      <NavLink to="/admin/products">Produk</NavLink>
      <NavLink to="/admin/orders">Order</NavLink>
    </nav>
  );
}
