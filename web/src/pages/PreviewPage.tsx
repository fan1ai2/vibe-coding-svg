import { useState, useEffect, useRef } from 'react';
import { useParams, Link } from 'react-router-dom';
import { conversions, ApiError, type Conversion } from '../api/client';
import LoadingSpinner from '../components/LoadingSpinner';

function authFetch(url: string, init?: RequestInit): Promise<Response> {
  const token = localStorage.getItem('token');
  const headers: Record<string, string> = {
    ...(init?.headers as Record<string, string> ?? {}),
  };
  if (token) headers['Authorization'] = `Bearer ${token}`;
  return fetch(url, { ...init, headers });
}

export default function PreviewPage() {
  const { id } = useParams<{ id: string }>();
  const [conv, setConv] = useState<Conversion | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [svgContent, setSvgContent] = useState<string | null>(null);
  const [originalBlobUrl, setOriginalBlobUrl] = useState<string | null>(null);
  const revokedRef = useRef<string | null>(null);

  useEffect(() => {
    if (!id) return;
    conversions.get(id)
      .then(res => {
        setConv(res.data);
        const promises: Promise<void>[] = [];
        if (res.data.status === 'completed' && res.data.svg_url) {
          promises.push(
            authFetch(`/api/v1/files/results/${res.data.svg_url}`)
              .then(r => r.text())
              .then(svg => setSvgContent(svg))
          );
          promises.push(
            authFetch(`/api/v1/files/originals/${res.data.original_url}`)
              .then(r => r.blob())
              .then(blob => {
                const url = URL.createObjectURL(blob);
                if (revokedRef.current) URL.revokeObjectURL(revokedRef.current);
                revokedRef.current = url;
                setOriginalBlobUrl(url);
              })
          );
        }
        return Promise.all(promises);
      })
      .catch(err => {
        setError(err instanceof ApiError ? err.message : 'Failed to load');
      })
      .finally(() => setLoading(false));

    return () => {
      if (revokedRef.current) URL.revokeObjectURL(revokedRef.current);
    };
  }, [id]);

  if (loading) return <LoadingSpinner label="Loading preview..." />;
  if (error) return <div className="text-red-600 text-center py-12">{error}</div>;
  if (!conv) return <div className="text-gray-500 text-center py-12">Not found</div>;

  const sizeReduction = conv.file_size_in > 0
    ? Math.round((1 - (conv.file_size_out || 0) / conv.file_size_in) * 100)
    : 0;

  return (
    <div className="max-w-5xl mx-auto">
      <div className="flex items-center justify-between mb-6">
        <h2 className="text-xl font-bold">Preview</h2>
        <div className="flex gap-3">
          {conv.status === 'completed' && (
            <a
              href={conversions.downloadUrl(conv.id)}
              className="inline-flex items-center gap-2 rounded-lg bg-indigo-600 px-4 py-2 text-sm font-semibold text-white hover:bg-indigo-500 transition-colors"
            >
              <svg className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2}
                  d="M12 10v6m0 0l-3-3m3 3l3-3m2 8H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
              </svg>
              Download SVG
            </a>
          )}
          <Link
            to="/workspace/library"
            className="inline-flex items-center gap-2 rounded-lg border border-gray-300 px-4 py-2 text-sm font-semibold text-gray-700 hover:bg-gray-50 transition-colors"
          >
            Back to Library
          </Link>
        </div>
      </div>

      {conv.status === 'failed' && (
        <div className="mb-6 rounded-lg bg-red-50 border border-red-200 p-4 text-sm text-red-700">
          Conversion failed: {conv.error_message || 'Unknown error'}
        </div>
      )}

      {/* Comparison view */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        <div className="rounded-xl border border-gray-200 bg-white p-4">
          <h3 className="text-xs font-semibold text-gray-400 uppercase mb-3">Original ({conv.format_in})</h3>
          <div className="aspect-square flex items-center justify-center bg-gray-50 rounded-lg overflow-hidden">
            <img
              src={originalBlobUrl ?? ''}
              alt="Original"
              className="max-w-full max-h-full object-contain"
            />
          </div>
        </div>

        <div className="rounded-xl border border-gray-200 bg-white p-4">
          <h3 className="text-xs font-semibold text-gray-400 uppercase mb-3">SVG Result</h3>
          <div className="aspect-square flex items-center justify-center bg-gray-50 rounded-lg overflow-hidden">
            {svgContent ? (
              <div
                className="max-w-full max-h-full"
                dangerouslySetInnerHTML={{ __html: svgContent }}
              />
            ) : conv.status === 'pending' || conv.status === 'processing' ? (
              <LoadingSpinner label="Processing..." />
            ) : (
              <span className="text-gray-400 text-sm">Not available</span>
            )}
          </div>
        </div>
      </div>

      {/* Metadata */}
      {conv.status === 'completed' && (
        <div className="mt-6 grid grid-cols-2 sm:grid-cols-4 gap-4">
          <MetaItem label="Input Size" value={formatBytes(conv.file_size_in)} />
          <MetaItem label="Output Size" value={formatBytes(conv.file_size_out)} />
          <MetaItem label="Reduction" value={`${sizeReduction}%`} />
          <MetaItem label="Paths" value={String(conv.path_count)} />
          <MetaItem label="Colors" value={String(conv.color_count)} />
          <MetaItem label="Format" value={conv.format_in.toUpperCase()} />
        </div>
      )}
    </div>
  );
}

function MetaItem({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-lg border border-gray-200 bg-white p-3">
      <div className="text-xs text-gray-400">{label}</div>
      <div className="text-sm font-semibold text-gray-900">{value}</div>
    </div>
  );
}

function formatBytes(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
}
