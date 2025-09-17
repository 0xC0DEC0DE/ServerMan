import { useEffect, useState } from 'react';
import ServerList from '../components/ServerList';
import { checkAuthentication } from '../utils/auth';

export default function Dashboard() {
  const [isAdmin, setIsAdmin] = useState(false);

  useEffect(() => {
    const authenticate = async () => {
      const userData = await checkAuthentication();
      if (userData && userData.groups && userData.groups.includes('callowaysutton')) {
        setIsAdmin(true);
      }
      // If authentication fails, checkAuthentication will handle the redirect
    };

    authenticate();
  }, []);

  return (
    <section className="section">
      <div className="container">
        <div className="columns">
          <div className="column">
            {/* Header row */}
            <div className="is-flex-tablet is-justify-content-space-between is-align-items-center">
              <h1 className="title mb-0">Dashboard</h1>

              <div className="buttons">
                {isAdmin && (
                  <a href="/admin" className="button is-light">
                    Admin
                  </a>
                )}
                <a
                  href="/logout"
                  className="button is-danger is-light"
                >
                  Log out
                </a>
              </div>
            </div>

            <h2 className="subtitle mt-2">Server management panel</h2>

            <div className="box">
              <ServerList />
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}
