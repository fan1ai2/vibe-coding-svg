import { useState, useEffect } from 'react';
import { tags as tagsApi, type IconTag } from '../api/client';

interface SearchBarProps {
  onSearch: (query: string, selectedTags: string[]) => void;
  initialQuery?: string;
}

export default function SearchBar({ onSearch, initialQuery = '' }: SearchBarProps) {
  const [query, setQuery] = useState(initialQuery);
  const [tagList, setTagList] = useState<IconTag[]>([]);
  const [selectedTags, setSelectedTags] = useState<string[]>([]);

  useEffect(() => {
    tagsApi.list(30, 'popular').then(res => setTagList(res.data)).catch(() => {});
  }, []);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    onSearch(query, selectedTags);
  };

  const toggleTag = (slug: string) => {
    setSelectedTags(prev => {
      const next = prev.includes(slug) ? prev.filter(t => t !== slug) : [...prev, slug];
      onSearch(query, next);
      return next;
    });
  };

  return (
    <form onSubmit={handleSubmit} className="mb-6">
      <input
        type="text"
        value={query}
        onChange={e => setQuery(e.target.value)}
        placeholder="搜索图标..."
        className="w-full px-4 py-2.5 rounded-lg border border-gray-200 bg-white text-sm focus:outline-none focus:ring-2 focus:ring-amber-500/20 focus:border-amber-400"
      />
      {tagList.length > 0 && (
        <div className="flex flex-wrap gap-1.5 mt-3">
          {tagList.map(tag => (
            <button
              key={tag.slug}
              type="button"
              onClick={() => toggleTag(tag.slug)}
              className={`px-2.5 py-1 rounded-full text-xs font-medium transition-colors ${
                selectedTags.includes(tag.slug)
                  ? 'bg-amber-500 text-white'
                  : 'bg-gray-100 text-gray-600 hover:bg-gray-200'
              }`}
            >
              {tag.name}
            </button>
          ))}
        </div>
      )}
    </form>
  );
}
