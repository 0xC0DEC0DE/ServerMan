import { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { checkAuthentication } from '../utils/auth';

export default function Navbar() {
  const [isAdmin, setIsAdmin] = useState(false);
  const [userData, setUserData] = useState(null);

  useEffect(() => {
    const authenticate = async () => {
      const result = await checkAuthentication();
      if (result) {
        setUserData(result);
        if (result.groups && result.groups.includes('callowaysutton')) {
          setIsAdmin(true);
        }
      }
      // If authentication fails, checkAuthentication will handle the redirect
    };

    authenticate();
  }, []);

  return (
    <nav className="navbar is-dark" role="navigation" aria-label="main navigation">
      <div className="container">
        <div className="navbar-brand">
          <Link className="navbar-item" to="/dashboard">
            <span className="title is-4 has-text-white">CCS Management</span>
          </Link>
        </div>

        <div className="navbar-menu">
          <div className="navbar-end">
            <div className="navbar-item">
              <div className="buttons">
                <Link to="/dashboard" className="button is-dark is-primary">

                  <span>Back to Dashboard</span>
                </Link>
              </div>
            </div>
          </div>
        </div>
      </div>
    </nav>
  );
}