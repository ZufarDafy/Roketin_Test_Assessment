import { BrowserRouter, Link, Route, Routes } from "react-router-dom";
import CatalogPage from "./pages/CatalogPage";

export default function App() {
  return (
    <BrowserRouter>
      <header className="header">
        <div className="container header__inner">
          <Link to="/" className="header__brand">
            🛍️ MiniShop
          </Link>
          <nav className="header__nav">
            <Link to="/">Katalog</Link>
          </nav>
        </div>
      </header>
      <main className="container">
        <Routes>
          <Route path="/" element={<CatalogPage />} />
        </Routes>
      </main>
    </BrowserRouter>
  );
}
