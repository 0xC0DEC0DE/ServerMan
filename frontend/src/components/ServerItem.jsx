import { useEffect, useState, useRef } from 'react';
import { motion } from 'framer-motion';
import DynamicVNCViewer from './VncConsole';

export default function ServerItem({ serverId, serverName }) {
  const [details, setDetails] = useState(null);
  const [pingStatus, setPingStatus] = useState(null);
  const [loading, setLoading] = useState(true);

  const [showConsole, setShowConsole] = useState(false);
  const [credentials, setCredentials] = useState<{ password: string } | null>(null);
  const [loadingCreds, setLoadingCreds] = useState(false);

  // Fetch server details
  useEffect(() => {
    async function fetchDetails() {
      try {
        const res = await fetch(`/api/server/${serverId}`, {
          credentials: 'include',
        });
        const data = await res.json();
        setDetails(data);
      } catch (err) {
        console.error('Error loading server', err);
      } finally {
        setLoading(false);
      }
    }
    fetchDetails();
  }, [serverId]);

  // Ping server IP
  useEffect(() => {
    const pingServer = async () => {
      if (details && details.ip) {
        const res = await fetch(`/api/ping/${details.ip}`);
        const data = await res.json();
        setPingStatus(data.status);
      }
    };
    pingServer();
  }, [details]);

  // Ping server domain
  const [serverNameStatus, setServerNameStatus] = useState(null);
  useEffect(() => {
    const pingServerName = async () => {
      if (serverName) {
        const res = await fetch(`/api/ping/${serverName}`);
        const data = await res.json();
        setServerNameStatus(data.status);
      }
    };
    pingServerName();
  }, [serverName]);

  // Fetch VNC credentials
  const fetchCredentials = async () => {
    setLoadingCreds(true);
    try {
      const res = await fetch(`/api/server/${serverId}/credentials`, {
        credentials: 'include',
      });
      const data = await res.json();
      const password = data.vnc_password;
      if (!password) {
        throw new Error('No VNC password returned');
      }
      setCredentials({ "password": password }); // expecting { password: "..." }
      setShowConsole(true);
    } catch (err) {
      console.error('Error fetching credentials', err);
    } finally {
      setLoadingCreds(false);
    }
  };

  return (
    <>
      <motion.div
        whileHover={{ scale: 1.01, boxShadow: '0px 5px 15px rgba(0,0,0,0.2)' }}
        transition={{ type: 'spring', stiffness: 300 }}
        className="box mb-3"
      >
        {loading ? (
          // Skeleton loading state
          <div>
            <p className="has-background-light mb-2" style={{ height: '1.5rem', width: '40%' }}></p>
            <p className="has-background-light mb-2" style={{ height: '1rem', width: '60%' }}></p>
            <p className="has-background-light" style={{ height: '1rem', width: '30%' }}></p>
          </div>
        ) : (
          details && (
            <div className="content">
              <div>
                <div className="is-flex-tablet is-align-texts-center">
                  <strong className="mt-2 mb-0-tablet">{serverName}</strong>

                  <span className="ml-auto buttons are-small">
                    <button className="button is-danger is-light">Shutdown</button>
                    <button className="button is-warning is-light">Reboot</button>
                    <button
                      className={`button is-light ${loadingCreds ? 'is-loading' : ''}`}
                      onClick={fetchCredentials}
                    >
                      View Console
                    </button>
                  </span>
                </div>
                <hr />
                {pingStatus === null ? (
                  <span className="tag is-light is-loading">Checking {details.ip}...</span>
                ) : pingStatus === 'up' ? (
                  <span className="tag is-success">{details.ip}</span>
                ) : (
                  <span className="tag is-danger">{details.ip}</span>
                )}{' '}
                |{' '}
                {serverNameStatus === null ? (
                  <span className="tag is-light is-loading">Checking {serverName}...</span>
                ) : serverNameStatus === 'up' ? (
                  <span className="tag is-success">{serverName}</span>
                ) : (
                  <span className="tag is-danger">{serverName}</span>
                )}
              </div>

              <div className="content">
                <table className="table is-fullwidth is-narrow">
                  <tbody>
                    <tr>
                      <th>OS</th>
                      <td>{details.operatingsystem}</td>
                    </tr>
                    <tr>
                      <th>CPU</th>
                      <td>{details.cpu} Cores</td>
                    </tr>
                    <tr>
                      <th>Memory</th>
                      <td>{details.mem} GB</td>
                    </tr>
                    <tr>
                      <th>Disk</th>
                      <td>{details.disk} GB</td>
                    </tr>
                    <tr>
                      <th>VNC</th>
                      <td>{details.vncstatus}</td>
                    </tr>
                    <tr>
                      <th>Snapshots</th>
                      <td>{details.dailysnapshots}</td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </div>
          )
        )}
      </motion.div>

      {/* Console Modal */}
      {showConsole && credentials && (
        <div className="modal is-active">
          <div className="modal-background" onClick={() => setShowConsole(false)}></div>
          <div className="modal-card" style={{ width: '90%', height: '90%' }}>
            <header className="modal-card-head">
              <p className="modal-card-title">Console - {serverName}</p>
              <button
                className="delete"
                aria-label="close"
                onClick={() => setShowConsole(false)}
              ></button>
            </header>
            <section className="modal-card-body" style={{ height: '100%', display: 'flex', flexDirection: 'column' }}>
              <DynamicVNCViewer
                url={`ws?token=${serverName}`}
                password={credentials.password}
              />
            </section>
          </div>
        </div>
      )}
    </>
  );
}
