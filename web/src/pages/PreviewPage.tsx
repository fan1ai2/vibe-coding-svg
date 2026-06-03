import { useState, useEffect, useRef } from 'react';
import { useParams, Link, useNavigate } from 'react-router-dom';
import { conversions, icons, ApiError, type Conversion } from '../api/client';
import LoadingSpinner from '../components/LoadingSpinner';
import PublishDialog from '../components/PublishDialog';

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
  const navigate = useNavigate();
  const [conv, setConv] = useState<Conversion | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [svgContent, setSvgContent] = useState<string | null>(null);
  const [originalBlobUrl, setOriginalBlobUrl] = useState<string | null>(null);
  const [showPublish, setShowPublish] = useState(false);
  const [toast, setToast] = useState<string | null>(null);
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

  const handlePublish = async (name: string, tags: { name: string; type: string }[], theme: string, isPublic: boolean) => {
    if (!svgContent) return
    try {
      const res = await icons.create({ name, svg_content: svgContent, is_public: isPublic, tags, theme })
      setShowPublish(false)
      setToast('已发布到图标库')
      setTimeout(() => navigate(`/icons/${res.data.id}`), 1000)
    } catch (err) {
      console.error('Publish failed:', err)
      setToast('发布失败，请重试')
      setTimeout(() => setToast(null), 3000)
    }
  }

  if (loading) return <LoadingSpinner label="Loading preview..." />;
  if (error) return <div className="text-red-600 text-center py-12">{error}</div>;
  if (!conv) return <div className="text-gray-500 text-center py-12">Not found</div>;

  const sizeReduction = conv.file_size_in > 0
    ? Math.round((1 - (conv.file_size_out || 0) / conv.file_size_in) * 100)
    : 0;

  return (
    <div className="max-w-6xl mx-auto">
      {toast && (
        <div className="fixed top-20 right-4 z-50 px-4 py-2 bg-gray-800 text-white text-sm rounded-lg shadow-lg">
          {toast}
        </div>
      )}

      <div className="flex items-center justify-between mb-6">
        <h2 className="text-xl font-bold">Preview</h2>
        <div className="flex gap-3">
          {conv.status === 'completed' && (
            <>
              <button
                onClick={() => {
                  if (svgContent) {
                    sessionStorage.setItem('editor:pendingSvg', svgContent)
                    navigate('/workspace/editor')
                  }
                }}
                disabled={!svgContent}
                className="inline-flex items-center gap-2 rounded-lg bg-amber-500 px-4 py-2 text-sm font-semibold text-white hover:bg-amber-600 transition-colors disabled:opacity-50"
              >
                <svg className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2}
                    d="M15.232 5.232l3.536 3.536m-2.036-5.036a2.5 2.5 0 113.536 3.536L6.5 21.036H3v-3.572L16.732 3.732z" />
                </svg>
                Open in Editor
              </button>
              <button
                onClick={() => {
                  if (svgContent) setShowPublish(true)
                }}
                disabled={!svgContent}
                className="inline-flex items-center gap-2 rounded-lg bg-green-500 px-4 py-2 text-sm font-semibold text-white hover:bg-green-600 transition-colors disabled:opacity-50"
              >
                <svg className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2}
                    d="M8 7H5a2 2 0 00-2 2v9a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-3m-1 4l-3 3m0 0l-3-3m3 3V4" />
                </svg>
                Save to Library
              </button>
              <button
                onClick={async () => {
                  try {
                    const res = await authFetch(conversions.downloadUrl(conv.id));
                    if (!res.ok) throw new Error('Download failed');
                    const blob = await res.blob();
                    const url = URL.createObjectURL(blob);
                    const a = document.createElement('a');
                    a.href = url;
                    a.download = `${conv.id}.svg`;
                    document.body.appendChild(a);
                    a.click();
                    document.body.removeChild(a);
                    URL.revokeObjectURL(url);
                  } catch {
                    // 忽略错误
                  }
                }}
                className="inline-flex items-center gap-2 rounded-lg bg-amber-500 px-4 py-2 text-sm font-medium text-white hover:bg-amber-600 transition-colors"
              >
                <svg className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2}
                    d="M12 10v6m0 0l-3-3m3 3l3-3m2 8H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                </svg>
                Download SVG
              </button>
            </>
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

      {showPublish && (
        <PublishDialog
          onClose={() => setShowPublish(false)}
          onPublish={handlePublish}
          defaultName={`Converted ${conv.id.slice(0, 8)}`}
        />
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
