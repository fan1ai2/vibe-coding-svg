import { Routes, Route, Navigate } from 'react-router-dom';
import { AuthProvider, useAuth } from './context/AuthContext';
import Layout from './components/Layout';
import LoadingSpinner from './components/LoadingSpinner';
import LandingPage from './pages/LandingPage';
import CallbackPage from './pages/CallbackPage';
import ConvertPage from './pages/ConvertPage';
import PreviewPage from './pages/PreviewPage';
import LibraryPage from './pages/LibraryPage';

function ProtectedRoute({ children }: { children: React.ReactNode }) {
  const { token, loading } = useAuth();
  if (loading) return <LoadingSpinner label="Checking authentication..." />;
  if (!token) return <Navigate to="/" replace />;
  return <>{children}</>;
}

export default function App() {
  return (
    <AuthProvider>
      <Routes>
        <Route path="/" element={<LandingPage />} />
        <Route path="/callback" element={<CallbackPage />} />
        <Route
          path="/workspace"
          element={
            <ProtectedRoute>
              <Layout />
            </ProtectedRoute>
          }
        >
          <Route index element={<Navigate to="convert" replace />} />
          <Route path="convert" element={<ConvertPage />} />
          <Route path="preview/:id" element={<PreviewPage />} />
          <Route path="library" element={<LibraryPage />} />
        </Route>
        <Route path="*" element={<div className="p-8 text-center text-gray-500">404 — Page not found</div>} />
      </Routes>
    </AuthProvider>
  );
}
