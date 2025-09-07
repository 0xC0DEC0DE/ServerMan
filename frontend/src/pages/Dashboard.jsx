import { useEffect, useState } from 'react';
import ServerList from '../components/ServerList';

export default function Dashboard() {
  const [isAdmin, setIsAdmin] = useState(false);

  useEffect(() => {
    const fetchUser = async () => {
      try {
        const res = await fetch('/api/user');
        if (!res.ok) return;
        const user = await res.json();

        if (user.groups && user.groups.includes('callowaysutton')) {
          setIsAdmin(true);
        }
      } catch (err) {
        console.error('Failed to fetch user:', err);
      }
    };

    fetchUser();
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
