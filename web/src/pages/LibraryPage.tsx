import { useState, useEffect, useCallback } from 'react';
import { conversions, ApiError, type Conversion } from '../api/client';
import ConversionCard from '../components/ConversionCard';
import LoadingSpinner from '../components/LoadingSpinner';

const PAGE_SIZE = 20;

export default function LibraryPage() {
  const [items, setItems] = useState<Conversion[]>([]);
  const [loading, setLoading] = useState(true);
  const [loadingMore, setLoadingMore] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [hasMore, setHasMore] = useState(false);

  useEffect(() => {
    conversions.list(PAGE_SIZE, 0)
      .then(res => {
        setItems(res.data);
        setHasMore(res.data.length === PAGE_SIZE);
      })
      .catch(err => setError(err instanceof ApiError ? err.message : 'Load failed'))
      .finally(() => setLoading(false));
  }, []);

  const loadMore = useCallback(async () => {
    setLoadingMore(true);
    try {
      const res = await conversions.list(PAGE_SIZE, items.length);
      setItems(prev => [...prev, ...res.data]);
      setHasMore(res.data.length === PAGE_SIZE);
    } catch {
      // ignore load-more errors
    } finally {
      setLoadingMore(false);
    }
  }, [items.length]);

  if (loading) return <LoadingSpinner label="Loading library..." />;

  return (
    <div className="max-w-5xl mx-auto">
      <h2 className="text-xl font-bold mb-6">My Library</h2>
      {error && (
        <div className="mb-4 rounded-lg bg-red-50 border border-red-200 p-4 text-sm text-red-700">{error}</div>
      )}
      {items.length === 0 ? (
        <div className="text-center py-16 text-gray-400">
          <p className="text-lg">No conversions yet</p>
          <p className="text-sm mt-1">Upload an image to get started.</p>
        </div>
      ) : (
        <>
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
            {items.map(c => (
              <ConversionCard key={c.id} conv={c} />
            ))}
          </div>
          {hasMore && (
            <div className="mt-6 text-center">
              <button
                onClick={loadMore}
                disabled={loadingMore}
                className="inline-flex items-center gap-2 rounded-lg border border-gray-300 px-6 py-3 text-sm font-semibold text-gray-700 hover:bg-gray-50 disabled:opacity-50 transition-colors"
              >
                {loadingMore ? (
                  <>
                    <svg className="h-4 w-4 animate-spin" fill="none" viewBox="0 0 24 24">
                      <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
                      <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
                    </svg>
                    Loading...
                  </>
                ) : (
                  'Load more'
                )}
              </button>
            </div>
          )}
        </>
      )}
    </div>
  );
}
