import { useState, useEffect } from 'react';
import { Navigate } from 'react-router-dom';

function ProtectedRoute({ children }) {
  const [auth, setAuth] = useState(null);

  useEffect(() => {
    fetch('/api/user', { credentials: 'include' })
      .then((res) => {
        if (res.status === 200) {
          return res.json();
        } else {
          setAuth(false);
          return null;
        }
      })
      .then((data) => {
        if (data) {
          setAuth(true); // Optionally check data for more validation
        }
      })
      .catch(() => setAuth(false));
  }, []);

  if (auth === null) {
    return <div>Loading...</div>;
  }

  return auth ? children : <Navigate to="/login" replace />;
}

export default ProtectedRoute;
