import { useState, useEffect } from 'react';
import { Navigate } from 'react-router-dom';
import { checkAuthentication } from '../utils/auth';

function AdminProtectedRoute({ children }) {
  const [auth, setAuth] = useState(null);
  const [hasAdminAccess, setHasAdminAccess] = useState(false);

  useEffect(() => {
    const authenticate = async () => {
      const userData = await checkAuthentication();
      if (userData) {
        setAuth(true);
        // Check if user has admin access ("*" or "callowaysutton" groups)
        const groups = userData.groups || [];
        const isAdmin = groups.includes('*') || groups.includes('callowaysutton');
        setHasAdminAccess(isAdmin);
      } else {
        // checkAuthentication will handle clearing cookies and redirecting
        // No need to set auth state as the page will redirect
        return;
      }
    };

    authenticate();
  }, []);

  if (auth === null) {
    return (
      <div className="has-text-centered" style={{ padding: '2rem' }}>
        <div className="loader"></div>
        <p>Loading...</p>
      </div>
    );
  }

  if (!auth) {
    // This shouldn't happen as checkAuthentication redirects, but just in case
    return null;
  }

  if (!hasAdminAccess) {
    return <Navigate to="/dashboard" replace />;
  }

  return children;
}

export default AdminProtectedRoute;
