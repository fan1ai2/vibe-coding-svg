import { useState, useEffect } from 'react';
import { conversions, ApiError, type Conversion } from '../api/client';
import ConversionCard from '../components/ConversionCard';
import LoadingSpinner from '../components/LoadingSpinner';

export default function LibraryPage() {
  const [items, setItems] = useState<Conversion[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    conversions.list(50, 0)
      .then(res => setItems(res.data))
      .catch(err => setError(err instanceof ApiError ? err.message : 'Load failed'))
      .finally(() => setLoading(false));
  }, []);

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
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
          {items.map(c => (
            <ConversionCard key={c.id} conv={c} />
          ))}
        </div>
      )}
    </div>
  );
}
