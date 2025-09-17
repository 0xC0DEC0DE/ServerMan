import { useEffect, useState } from 'react';
import ServerItem from './ServerItem';
import { clearCookiesAndRedirectHome, redirectToLogin } from '../utils/auth';

export default function ServerList() {
  const [servers, setServers] = useState([]);
  const [loading, setLoading] = useState(true);

  // Helper function to handle auth failures in API calls
  const handleAuthFailure = (response) => {
    if (response.status === 401 || response.status === 403) {
      // Check if user has cookies to decide where to redirect
      const hasAuthCookies = document.cookie.includes('auth-session');
      if (hasAuthCookies) {
        clearCookiesAndRedirectHome();
      } else {
        redirectToLogin();
      }
      return true;
    }
    return false;
  };

  useEffect(() => {
    async function fetchServers() {
      try {
        const res = await fetch('/api/servers', { credentials: 'include' });
        
        if (handleAuthFailure(res)) {
          return;
        }

        if (!res.ok) {
          throw new Error('Failed to fetch servers');
        }

        const data = await res.json();

        // Remove any servers with a "domainstatus" other than "Active" or an empty "domain"
        const filtered = data.filter(
          (s) =>
            (s.domainstatus === 'Active' || s.domainstatus === 'Pending') &&
            s.domain &&
            s.domain.trim() !== '',
        );

        // Sort alphabetically by domain
        filtered.sort((a, b) => a.domain.localeCompare(b.domain));
        setServers(filtered);
      } catch (err) {
        console.error('Error fetching servers', err);
      } finally {
        setLoading(false);
      }
    }
    fetchServers();
  }, []);

  if (loading) {
    return (
      <div>
        <h2 className="title is-4 mb-4">Servers</h2>
        {/* Skeleton rows */}
        {[1, 2, 3, 4].map((n) => (
          <div key={n} className="box mb-3">
            <p
              className="has-background-light mb-2"
              style={{ height: '1.5rem', width: '50%' }}
            ></p>
            <p
              className="has-background-light mb-2"
              style={{ height: '1rem', width: '70%' }}
            ></p>
          </div>
        ))}
      </div>
    );
  }

  return (
    <div>
      <h2 className="title is-4 mb-4">Servers</h2>
      {servers.length === 0 ? (
        <div className="notification has-text-centered">
          No servers found, please wait for an administrator to add servers to
          your account.
        </div>
      ) : (
        servers.map((server) => (
          <ServerItem
            key={server.id}
            serverId={server.id}
            serverName={server.domain}
          />
        ))
      )}
    </div>
  );
}
