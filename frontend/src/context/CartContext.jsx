import { createContext, useContext, useEffect, useMemo, useState } from "react";

const STORAGE_KEY = "minishop_cart";

const CartContext = createContext(null);

// Item cart menyimpan snapshot produk (id, name, price, image_url, stock)
// agar tampilan cart tidak perlu fetch ulang. Stok di snapshot hanya untuk
// validasi UX — sumber kebenaran tetap validasi backend saat checkout.
function loadInitialCart() {
  try {
    const raw = localStorage.getItem(STORAGE_KEY);
    const parsed = raw ? JSON.parse(raw) : [];
    return Array.isArray(parsed) ? parsed : [];
  } catch {
    return [];
  }
}

const clampQty = (qty, stock) => Math.max(1, Math.min(qty, stock));

export function CartProvider({ children }) {
  const [items, setItems] = useState(loadInitialCart);

  // Bonus 1.5: persist cart di localStorage agar tidak hilang saat refresh.
  useEffect(() => {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(items));
  }, [items]);

  const addItem = (product, qty = 1) => {
    setItems((prev) => {
      const existing = prev.find((it) => it.id === product.id);
      if (existing) {
        return prev.map((it) =>
          it.id === product.id
            ? { ...it, stock: product.stock, qty: clampQty(it.qty + qty, product.stock) }
            : it
        );
      }
      return [
        ...prev,
        {
          id: product.id,
          name: product.name,
          price: product.price,
          image_url: product.image_url,
          stock: product.stock,
          qty: clampQty(qty, product.stock),
        },
      ];
    });
  };

  const updateQty = (productId, qty) => {
    setItems((prev) =>
      prev.map((it) => (it.id === productId ? { ...it, qty: clampQty(qty, it.stock) } : it))
    );
  };

  // Dipanggil saat backend menolak checkout (422): sinkronkan stok snapshot
  // dengan angka available dari server dan turunkan qty yang melebihi.
  const applyStockErrors = (stockErrors) => {
    setItems((prev) =>
      prev
        .map((it) => {
          const se = stockErrors.find((e) => e.product_id === it.id);
          if (!se) return it;
          return { ...it, stock: se.available, qty: Math.min(it.qty, se.available) };
        })
        .filter((it) => it.qty > 0)
    );
  };

  // Dipanggil saat halaman cart dibuka: menyinkronkan snapshot harga & stok
  // dengan data server terkini. freshById memetakan product id -> {price,
  // stock} (produk masih ada) atau null (produk sudah dihapus). Item
  // dengan stok kini 0 atau produk yang sudah dihapus disingkirkan.
  const applyServerSnapshot = (freshById) => {
    setItems((prev) =>
      prev
        .map((it) => {
          const fresh = freshById[it.id];
          if (fresh === undefined) return it; // tidak berhasil dicek (mis. gagal jaringan)
          if (fresh === null || fresh.stock <= 0) return { ...it, qty: 0 };
          return { ...it, price: fresh.price, stock: fresh.stock, qty: Math.min(it.qty, fresh.stock) };
        })
        .filter((it) => it.qty > 0)
    );
  };

  const removeItem = (productId) =>
    setItems((prev) => prev.filter((it) => it.id !== productId));

  const clearCart = () => setItems([]);

  const value = useMemo(() => {
    const totalItems = items.reduce((sum, it) => sum + it.qty, 0);
    const totalPrice = items.reduce((sum, it) => sum + it.price * it.qty, 0);
    return {
      items,
      totalItems,
      totalPrice,
      addItem,
      updateQty,
      removeItem,
      clearCart,
      applyStockErrors,
      applyServerSnapshot,
    };
  }, [items]);

  return <CartContext.Provider value={value}>{children}</CartContext.Provider>;
}

export function useCart() {
  const ctx = useContext(CartContext);
  if (!ctx) throw new Error("useCart must be used within CartProvider");
  return ctx;
}
