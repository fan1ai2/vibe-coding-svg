import { useEffect } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import LoadingSpinner from '../components/LoadingSpinner';

export default function CallbackPage() {
  const [params] = useSearchParams();
  const navigate = useNavigate();

  useEffect(() => {
    const jwt = params.get('token');
    if (jwt) {
      localStorage.setItem('token', jwt);
      navigate('/workspace', { replace: true });
    } else {
      navigate('/', { replace: true });
    }
  }, [params, navigate]);

  return <LoadingSpinner label="Signing you in..." />;
}
