import { useState, useEffect } from 'react';
import { Navigate } from 'react-router-dom';

function AdminProtectedRoute({ children }) {
  const [auth, setAuth] = useState(null);
  const [hasAdminAccess, setHasAdminAccess] = useState(false);

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
          setAuth(true);
          // Check if user has admin access ("*" or "callowaysutton" groups)
          const groups = data.groups;
          const isAdmin = groups.includes('*') || groups.includes('callowaysutton');
          setHasAdminAccess(isAdmin);
        }
      })
      .catch(() => {
        setAuth(false);
        setHasAdminAccess(false);
      });
  }, []);

  if (auth === null) {
    return <div className="has-text-centered"><p>Loading...</p></div>;
  }

  if (!auth) {
    return <Navigate to="/login" replace />;
  }

  if (!hasAdminAccess) {
    return <Navigate to="/dashboard" replace />;
  }

  return children;
}

export default AdminProtectedRoute;
