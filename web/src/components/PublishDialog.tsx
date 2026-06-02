import { useState } from 'react';

interface PublishDialogProps {
  onClose: () => void;
  onPublish: (name: string, tags: { name: string; type: string }[], theme: string, isPublic: boolean) => void;
  defaultName?: string;
  defaultTags?: string[];
}

export default function PublishDialog({ onClose, onPublish, defaultName = '', defaultTags }: PublishDialogProps) {
  const [name, setName] = useState(defaultName);
  const [tagInput, setTagInput] = useState('');
  const [tags, setTags] = useState<{ name: string; type: string }[]>(
    defaultTags?.map(t => ({ name: t, type: 'usage' })) ?? []
  );
  const [theme, setTheme] = useState('');
  const [isPublic, setIsPublic] = useState(true);

  const addTag = () => {
    const trimmed = tagInput.trim();
    if (trimmed && !tags.some(t => t.name === trimmed)) {
      setTags([...tags, { name: trimmed, type: 'usage' }]);
      setTagInput('');
    }
  };

  const removeTag = (idx: number) => {
    setTags(tags.filter((_, i) => i !== idx));
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!name.trim()) return;
    onPublish(name.trim(), tags, theme.trim(), isPublic);
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40" onClick={onClose}>
      <div className="bg-white rounded-2xl shadow-xl p-6 w-full max-w-md mx-4" onClick={e => e.stopPropagation()}>
        <h3 className="text-lg font-semibold mb-4">发布到图标库</h3>
        <form onSubmit={handleSubmit}>
          <div className="mb-3">
            <label className="block text-xs text-gray-500 mb-1">图标名称</label>
            <input
              type="text"
              value={name}
              onChange={e => setName(e.target.value)}
              className="w-full px-3 py-2 rounded-lg border border-gray-200 text-sm focus:outline-none focus:ring-2 focus:ring-amber-500/20"
              required
            />
          </div>
          <div className="mb-3">
            <label className="block text-xs text-gray-500 mb-1">标签</label>
            <div className="flex gap-1">
              <input
                type="text"
                value={tagInput}
                onChange={e => setTagInput(e.target.value)}
                placeholder="输入标签名后回车"
                className="flex-1 px-3 py-2 rounded-lg border border-gray-200 text-sm focus:outline-none focus:ring-2 focus:ring-amber-500/20"
                onKeyDown={e => { if (e.key === 'Enter') { e.preventDefault(); addTag(); } }}
              />
              <button type="button" onClick={addTag} className="px-3 py-2 text-sm bg-gray-100 rounded-lg hover:bg-gray-200">添加</button>
            </div>
            {tags.length > 0 && (
              <div className="flex flex-wrap gap-1 mt-2">
                {tags.map((t, i) => (
                  <span key={i} className="inline-flex items-center gap-1 px-2 py-0.5 rounded text-xs bg-amber-50 text-amber-700">
                    {t.name}
                    <button type="button" onClick={() => removeTag(i)} className="text-amber-400 hover:text-amber-600">&times;</button>
                  </span>
                ))}
              </div>
            )}
          </div>
          <div className="mb-3">
            <label className="block text-xs text-gray-500 mb-1">主题</label>
            <input
              type="text"
              value={theme}
              onChange={e => setTheme(e.target.value)}
              placeholder="如: business, tech, minimal"
              className="w-full px-3 py-2 rounded-lg border border-gray-200 text-sm focus:outline-none focus:ring-2 focus:ring-amber-500/20"
            />
          </div>
          <div className="mb-4">
            <label className="flex items-center gap-2 text-sm">
              <input type="checkbox" checked={isPublic} onChange={e => setIsPublic(e.target.checked)} />
              公开到图标库
            </label>
          </div>
          <div className="flex gap-2 justify-end">
            <button type="button" onClick={onClose} className="px-4 py-2 text-sm text-gray-600 hover:bg-gray-100 rounded-lg">取消</button>
            <button type="submit" className="px-4 py-2 text-sm bg-amber-500 text-white font-medium rounded-lg hover:bg-amber-600">发布</button>
          </div>
        </form>
      </div>
    </div>
  );
}
