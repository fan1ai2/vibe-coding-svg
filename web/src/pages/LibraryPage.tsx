import { useState, useEffect, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import { conversions, svgs, ApiError, type Conversion, type SavedSvg } from '../api/client';
import ConversionCard from '../components/ConversionCard';
import LoadingSpinner from '../components/LoadingSpinner';

const PAGE_SIZE = 20;

type Tab = 'conversions' | 'saved';

export default function LibraryPage() {
  const navigate = useNavigate();
  const [tab, setTab] = useState<Tab>('conversions');
  const [conversionItems, setConversionItems] = useState<Conversion[]>([]);
  const [savedItems, setSavedItems] = useState<SavedSvg[]>([]);
  const [loading, setLoading] = useState(true);
  const [loadingMore, setLoadingMore] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [hasMore, setHasMore] = useState(false);

  useEffect(() => {
    setLoading(true);
    setError(null);
    if (tab === 'conversions') {
      conversions.list(PAGE_SIZE, 0)
        .then(res => {
          setConversionItems(res.data);
          setHasMore(res.data.length === PAGE_SIZE);
        })
        .catch(err => setError(err instanceof ApiError ? err.message : 'Load failed'))
        .finally(() => setLoading(false));
    } else {
      svgs.list(PAGE_SIZE, 0)
        .then(res => {
          setSavedItems(res.data);
          setHasMore(res.data.length === PAGE_SIZE);
        })
        .catch(err => setError(err instanceof ApiError ? err.message : 'Load failed'))
        .finally(() => setLoading(false));
    }
  }, [tab]);

  const loadMore = useCallback(async () => {
    setLoadingMore(true);
    try {
      if (tab === 'conversions') {
        const res = await conversions.list(PAGE_SIZE, conversionItems.length);
        setConversionItems(prev => [...prev, ...res.data]);
        setHasMore(res.data.length === PAGE_SIZE);
      } else {
        const res = await svgs.list(PAGE_SIZE, savedItems.length);
        setSavedItems(prev => [...prev, ...res.data]);
        setHasMore(res.data.length === PAGE_SIZE);
      }
    } catch {
      // ignore
    } finally {
      setLoadingMore(false);
    }
  }, [tab, conversionItems.length, savedItems.length]);

  const handleDeleteSvg = async (id: string) => {
    try {
      await svgs.delete(id);
      setSavedItems(prev => prev.filter(s => s.id !== id));
    } catch {
      // ignore
    }
  };

  if (loading) return <LoadingSpinner label="Loading library..." />;

  const items = tab === 'conversions' ? conversionItems : savedItems;

  return (
    <div className="max-w-6xl mx-auto">
      <div className="flex items-center gap-4 mb-6">
        <h2 className="text-xl font-bold">My Library</h2>
        <div className="flex rounded-lg border border-gray-200 overflow-hidden">
          <button
            onClick={() => setTab('conversions')}
            className={`px-4 py-1.5 text-sm font-medium transition-colors ${tab === 'conversions' ? 'bg-amber-500 text-white' : 'bg-white text-gray-600 hover:bg-gray-50'}`}
          >
            Conversions
          </button>
          <button
            onClick={() => setTab('saved')}
            className={`px-4 py-1.5 text-sm font-medium transition-colors ${tab === 'saved' ? 'bg-amber-500 text-white' : 'bg-white text-gray-600 hover:bg-gray-50'}`}
          >
            Saved SVGs
          </button>
        </div>
      </div>

      {error && (
        <div className="mb-4 rounded-lg bg-red-50 border border-red-200 p-4 text-sm text-red-700">{error}</div>
      )}

      {items.length === 0 ? (
        <div className="text-center py-16 text-gray-400">
          <p className="text-lg">
            {tab === 'conversions' ? 'No conversions yet' : 'No saved SVGs yet'}
          </p>
          <p className="text-sm mt-1">
            {tab === 'conversions'
              ? 'Upload an image to get started.'
              : 'Edit an SVG in the editor and save it.'}
          </p>
        </div>
      ) : (
        <>
          {tab === 'conversions' ? (
            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
              {conversionItems.map(c => (
                <ConversionCard key={c.id} conv={c} />
              ))}
            </div>
          ) : (
            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
              {savedItems.map(s => (
                <div
                  key={s.id}
                  className="rounded-xl border border-gray-200 bg-white hover:shadow-md transition-shadow overflow-hidden"
                >
                  <div
                    className="aspect-video bg-gray-100 flex items-center justify-center cursor-pointer"
                    dangerouslySetInnerHTML={{ __html: s.svg_content ?? '' }}
                    onClick={() => navigate(`/workspace/editor?svg=${s.id}`)}
                  />
                  <div className="p-3 flex items-center justify-between">
                    <span className="text-xs text-gray-500 truncate flex-1 mr-2">{s.name}</span>
                    <div className="flex gap-1">
                      <a
                        href={svgs.downloadUrl(s.id)}
                        className="text-xs text-amber-600 hover:text-amber-700"
                        onClick={e => e.stopPropagation()}
                      >
                        Download
                      </a>
                      <button
                        onClick={(e) => { e.stopPropagation(); handleDeleteSvg(s.id); }}
                        className="text-xs text-red-500 hover:text-red-700"
                      >
                        Delete
                      </button>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          )}

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
