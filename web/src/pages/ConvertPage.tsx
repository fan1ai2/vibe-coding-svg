import { useState, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import DropZone from '../components/DropZone';
import LoadingSpinner from '../components/LoadingSpinner';
import { conversions, ApiError } from '../api/client';
import { usePolling } from '../hooks/usePolling';

export default function ConvertPage() {
  const navigate = useNavigate();
  const [uploading, setUploading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [conversionId, setConversionId] = useState<string | null>(null);
  const [status, setStatus] = useState<string | null>(null);

  const pollStatus = useCallback(() => {
    if (!conversionId) return;
    conversions.get(conversionId)
      .then(res => {
        setStatus(res.data.status);
        if (res.data.status === 'completed') {
          navigate(`/workspace/preview/${conversionId}`, { replace: true });
        }
      })
      .catch(() => {});
  }, [conversionId, navigate]);

  usePolling(pollStatus, 1000, status === 'pending' || status === 'processing');

  const handleFile = useCallback(async (file: File) => {
    setError(null);
    setUploading(true);
    try {
      const res = await conversions.upload(file);
      setConversionId(res.data.id);
      setStatus(res.data.status);
    } catch (err) {
      const msg = err instanceof ApiError ? err.message : 'Upload failed';
      setError(msg);
    } finally {
      setUploading(false);
    }
  }, []);

  if (uploading) {
    return <LoadingSpinner label="Uploading..." />;
  }

  if (status === 'pending' || status === 'processing') {
    return (
      <div className="max-w-xl mx-auto">
        <h2 className="text-xl font-bold mb-4">Processing...</h2>
        <LoadingSpinner label={`Status: ${status}`} />
        <p className="text-center text-sm text-gray-500 mt-4">
          Your image is being converted to SVG. This may take a few seconds.
        </p>
      </div>
    );
  }

  return (
    <div className="max-w-xl mx-auto">
      <h2 className="text-xl font-bold mb-6">New Conversion</h2>
      {error && (
        <div className="mb-4 rounded-lg bg-red-50 border border-red-200 p-4 text-sm text-red-700">
          {error}
        </div>
      )}
      <DropZone onFile={handleFile} disabled={uploading} />
    </div>
  );
}
