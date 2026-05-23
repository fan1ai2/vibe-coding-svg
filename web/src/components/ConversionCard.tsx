import { Link } from 'react-router-dom';
import type { Conversion } from '../api/client';

function statusBadge(s: Conversion['status']) {
  const map = {
    pending: 'bg-yellow-100 text-yellow-800',
    processing: 'bg-blue-100 text-blue-800',
    completed: 'bg-green-100 text-green-800',
    failed: 'bg-red-100 text-red-800',
  };
  return map[s] ?? 'bg-gray-100 text-gray-800';
}

function timeAgo(iso: string): string {
  const diff = Date.now() - new Date(iso).getTime();
  const mins = Math.floor(diff / 60000);
  if (mins < 1) return 'just now';
  if (mins < 60) return `${mins}m ago`;
  const hrs = Math.floor(mins / 60);
  if (hrs < 24) return `${hrs}h ago`;
  return `${Math.floor(hrs / 24)}d ago`;
}

export default function ConversionCard({ conv }: { conv: Conversion }) {
  return (
    <Link
      to={`/workspace/preview/${conv.id}`}
      className="block rounded-xl border border-gray-200 bg-white hover:shadow-md transition-shadow overflow-hidden"
    >
      <div className="aspect-video bg-gray-100 flex items-center justify-center">
        <img
          src={`/api/v1/files/originals/${conv.original_url}`}
          alt=""
          className="w-full h-full object-cover"
          onError={(e) => {
            (e.target as HTMLImageElement).style.display = 'none';
          }}
        />
        {!conv.original_url && (
          <svg className="h-8 w-8 text-gray-300" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5}
              d="M2.25 15.75l5.159-5.159a2.25 2.25 0 013.182 0l5.159 5.159m-1.5-1.5l1.409-1.409a2.25 2.25 0 013.182 0l2.909 2.909M3.75 21h16.5A2.25 2.25 0 0022.5 18.75V5.25A2.25 2.25 0 0020.25 3H3.75A2.25 2.25 0 001.5 5.25v13.5A2.25 2.25 0 003.75 21z" />
          </svg>
        )}
      </div>
      <div className="p-3 flex items-center justify-between">
        <span className="text-xs text-gray-500">{timeAgo(conv.created_at)}</span>
        <span className={`inline-block px-2 py-0.5 rounded-full text-xs font-medium ${statusBadge(conv.status)}`}>
          {conv.status}
        </span>
      </div>
    </Link>
  );
}
