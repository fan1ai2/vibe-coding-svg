import { Routes, Route, Navigate } from 'react-router-dom';
import { AuthProvider, useAuth } from './context/AuthContext';
import ErrorBoundary from './components/ErrorBoundary';
import Navbar from './components/Navbar';
import Footer from './components/Footer';
import WorkspaceShell from './components/WorkspaceShell';
import LoadingSpinner from './components/LoadingSpinner';
import LandingPage from './pages/LandingPage';
import CallbackPage from './pages/CallbackPage';
import ConvertPage from './pages/ConvertPage';
import PreviewPage from './pages/PreviewPage';
import LibraryPage from './pages/LibraryPage';
import EditorPage from './pages/EditorPage';

function ProtectedRoute({ children }: { children: React.ReactNode }) {
  const { token, loading } = useAuth();
  if (loading) return <LoadingSpinner label="正在检查认证状态..." />;
  if (!token) return <Navigate to="/" replace />;
  return <>{children}</>;
}

export default function App() {
  return (
    <AuthProvider>
      <Routes>
        {/* 公开页面：首页 */}
        <Route path="/" element={<><Navbar /><LandingPage /><Footer /></>} />

        {/* 公开页面：OAuth 回调 */}
        <Route path="/callback" element={<CallbackPage />} />

        {/* 工作区（需登录） */}
        <Route
          path="/workspace"
          element={
            <ProtectedRoute>
              <ErrorBoundary>
                <WorkspaceShell />
              </ErrorBoundary>
            </ProtectedRoute>
          }
        >
          <Route index element={<Navigate to="convert" replace />} />
          <Route path="convert" element={<ConvertPage />} />
          <Route path="preview/:id" element={<PreviewPage />} />
          <Route path="library" element={<LibraryPage />} />
          <Route path="editor" element={<EditorPage />} />
        </Route>

        {/* 404 */}
        <Route path="*" element={<div className="p-8 text-center text-gray-500">404 — 页面未找到</div>} />
      </Routes>
    </AuthProvider>
  );
}
