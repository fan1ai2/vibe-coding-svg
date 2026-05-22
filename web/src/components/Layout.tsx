import { NavLink, Outlet, useLocation } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';

export default function Layout() {
  const { userId, logout } = useAuth();
  const location = useLocation();

  const linkClass = (path: string) =>
    `block px-4 py-2 rounded-lg text-sm font-medium transition-colors ${
      location.pathname === path
        ? 'bg-indigo-100 text-indigo-700'
        : 'text-gray-600 hover:bg-gray-100 hover:text-gray-900'
    }`;

  return (
    <div className="flex min-h-screen">
      {/* Sidebar */}
      <aside className="w-64 bg-white border-r border-gray-200 flex flex-col">
        <div className="p-6 border-b border-gray-100">
          <h1 className="text-lg font-bold text-gray-900">SVG Converter</h1>
          <p className="text-xs text-gray-500 mt-1 truncate">{userId}</p>
        </div>

        <nav className="flex-1 p-4 space-y-1">
          <NavLink to="/workspace/convert" className={linkClass('/workspace/convert')}>
            + New Conversion
          </NavLink>
          <NavLink to="/workspace/library" className={linkClass('/workspace/library')}>
            My Library
          </NavLink>
        </nav>

        <div className="p-4 border-t border-gray-100">
          <button
            onClick={logout}
            className="w-full text-left px-4 py-2 rounded-lg text-sm text-gray-500 hover:bg-gray-100 transition-colors"
          >
            Log out
          </button>
        </div>
      </aside>

      {/* Main content */}
      <main className="flex-1 p-8">
        <Outlet />
      </main>
    </div>
  );
}
