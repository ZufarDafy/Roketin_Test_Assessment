import { BrowserRouter, NavLink, Route, Routes } from "react-router-dom";
import { CartProvider, useCart } from "./context/CartContext";
import CatalogPage from "./pages/CatalogPage";
import CartPage from "./pages/CartPage";
import OrderSuccessPage from "./pages/OrderSuccessPage";
import AdminProductsPage from "./pages/admin/AdminProductsPage";
import AdminOrdersPage from "./pages/admin/AdminOrdersPage";

function Header() {
  const { totalItems } = useCart();

  return (
    <header className="header">
      <div className="container header__inner">
        <NavLink to="/" className="header__brand">
          🛍️ MiniShop
        </NavLink>
        <nav className="header__nav">
          <NavLink to="/" end>
            Katalog
          </NavLink>
          <NavLink to="/cart">
            Keranjang
            {totalItems > 0 && <span className="cart-badge">{totalItems}</span>}
          </NavLink>
          <NavLink to="/admin/products">Admin</NavLink>
        </nav>
      </div>
    </header>
  );
}

export default function App() {
  return (
    <CartProvider>
      <BrowserRouter>
        <Header />
        <main className="container">
          <Routes>
            <Route path="/" element={<CatalogPage />} />
            <Route path="/cart" element={<CartPage />} />
            <Route path="/order-success/:id" element={<OrderSuccessPage />} />
            <Route path="/admin/products" element={<AdminProductsPage />} />
            <Route path="/admin/orders" element={<AdminOrdersPage />} />
          </Routes>
        </main>
      </BrowserRouter>
    </CartProvider>
  );
}
