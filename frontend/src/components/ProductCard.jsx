import { formatIDR } from "../utils/format";

export default function ProductCard({ product, onSelect }) {
  const outOfStock = product.stock <= 0;

  return (
    <article className="product-card" onClick={() => onSelect(product)}>
      <div className="product-card__image-wrap">
        <img src={product.image_url} alt={product.name} loading="lazy" />
        {outOfStock && <span className="badge badge--danger">Stok habis</span>}
      </div>
      <div className="product-card__body">
        <span className="product-card__category">{product.category?.name}</span>
        <h3 className="product-card__name">{product.name}</h3>
        <p className="product-card__desc">{product.description}</p>
        <div className="product-card__footer">
          <strong>{formatIDR(product.price)}</strong>
          <span className="product-card__stock">Stok: {product.stock}</span>
        </div>
      </div>
    </article>
  );
}
