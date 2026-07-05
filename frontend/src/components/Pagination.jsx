export default function Pagination({ page, totalPages, onChange }) {
  if (totalPages <= 1) return null;

  return (
    <nav className="pagination" aria-label="Navigasi halaman">
      <button
        className="btn btn--small"
        onClick={() => onChange(page - 1)}
        disabled={page <= 1}
      >
        &larr; Sebelumnya
      </button>
      <span className="pagination__info">
        Halaman {page} dari {totalPages}
      </span>
      <button
        className="btn btn--small"
        onClick={() => onChange(page + 1)}
        disabled={page >= totalPages}
      >
        Berikutnya &rarr;
      </button>
    </nav>
  );
}
