import { useState, useEffect } from 'react';
import { checkAuthentication } from '../utils/auth';

function ProtectedRoute({ children }) {
  const [auth, setAuth] = useState(null);
  const [userData, setUserData] = useState(null);

  useEffect(() => {
    const authenticate = async () => {
      const result = await checkAuthentication();
      if (result) {
        setAuth(true);
        setUserData(result);
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

  return auth ? children : null;
}

export default ProtectedRoute;
