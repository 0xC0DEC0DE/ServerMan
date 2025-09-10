import { useEffect, useState, useRef } from 'react';
import { motion } from 'framer-motion';
import DynamicVNCViewer from './VncConsole';
import ReinstallModal from './ReinstallModal';
import SnapshotModal from './SnapshotModal';

export default function ServerItem({ serverId, serverName }) {
  const [details, setDetails] = useState(null);
  const [pingStatus, setPingStatus] = useState(null);
  const [loading, setLoading] = useState(true);

  const [showConsole, setShowConsole] = useState(false);
  const [credentials, setCredentials] = useState(null);
  const [loadingCreds, setLoadingCreds] = useState(false);
  
  // State for management actions
  const [actionLoading, setActionLoading] = useState({});
  const [showReinstallModal, setShowReinstallModal] = useState(false);
  const [showSnapshotModal, setShowSnapshotModal] = useState(false);

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

  // Server management actions
  const handleServerAction = async (action) => {
    const actionName = action.charAt(0).toUpperCase() + action.slice(1);
    
    if (!confirm(`Are you sure you want to ${action} the server?`)) {
      return;
    }

    setActionLoading(prev => ({ ...prev, [action]: true }));
    
    try {
      const response = await fetch(`/api/server/${serverId}/action/${action}`, {
        method: 'POST',
        credentials: 'include',
      });

      const result = await response.json();
      
      if (response.ok) {
        alert(`${actionName} command sent successfully. Check the server status for updates.`);
      } else {
        alert(`Error: ${result.error || `Failed to ${action} server`}`);
      }
    } catch (error) {
      console.error(`Error ${action}ing server:`, error);
      alert(`Failed to ${action} server`);
    } finally {
      setActionLoading(prev => ({ ...prev, [action]: false }));
    }
  };

  const handleReinstall = (result) => {
    alert('Reinstall command sent successfully. The process will take between 5-10 minutes. Check your dashboard for updates.');
  };

  const handlePasswordReset = async () => {
    if (!confirm('Are you sure you want to reset the root password? This will restart the server.')) {
      return;
    }

    setActionLoading(prev => ({ ...prev, resetPassword: true }));
    
    try {
      const response = await fetch(`/api/server/${serverId}/reset-password`, {
        method: 'POST',
        credentials: 'include',
      });

      const result = await response.json();
      
      if (response.ok) {
        alert('Password reset command sent successfully. The server will restart and a new password will be generated. Check the server credentials after restart.');
      } else {
        alert(`Error: ${result.error || 'Failed to reset password'}`);
      }
    } catch (error) {
      console.error('Error resetting password:', error);
      alert('Failed to reset password');
    } finally {
      setActionLoading(prev => ({ ...prev, resetPassword: false }));
    }
  };

  const handleSnapshotRestore = (result) => {
    alert('Snapshot restoration command sent successfully. The process will take some time. Check your dashboard for updates.');
  };

  const handleConsoleToggle = async (action) => {
    const actionText = action === 'enable' ? 'enable' : 'disable';
    if (!confirm(`Are you sure you want to ${actionText} the console? This will force a server restart.`)) {
      return;
    }

    setActionLoading(prev => ({ ...prev, console: true }));
    
    try {
      const response = await fetch(`/api/server/${serverId}/console/${action}`, {
        method: 'POST',
        credentials: 'include',
      });

      const result = await response.json();
      
      if (response.ok) {
        alert(`Console ${actionText} command sent successfully. The server will restart to apply changes.`);
      } else {
        alert(`Error: ${result.error || `Failed to ${actionText} console`}`);
      }
    } catch (error) {
      console.error(`Error ${actionText}ing console:`, error);
      alert(`Failed to ${actionText} console`);
    } finally {
      setActionLoading(prev => ({ ...prev, console: false }));
    }
  };
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
                    <button 
                      className={`button is-danger is-light ${actionLoading.stop ? 'is-loading' : ''}`}
                      onClick={() => handleServerAction('stop')}
                      disabled={actionLoading.stop}
                    >
                      Shutdown
                    </button>
                    <button 
                      className={`button is-warning is-light ${actionLoading.restart ? 'is-loading' : ''}`}
                      onClick={() => handleServerAction('restart')}
                      disabled={actionLoading.restart}
                    >
                      Reboot
                    </button>
                    <div className="dropdown is-hoverable">
                      <div className="dropdown-trigger">
                        <button className="button is-info is-light">
                          <span>More Actions</span>
                          <span className="icon is-small">
                            <span>▼</span>
                          </span>
                        </button>
                      </div>
                      <div className="dropdown-menu">
                        <div className="dropdown-content">
                          <a 
                            className="dropdown-item"
                            onClick={() => setShowReinstallModal(true)}
                          >
                            <span>🔄 Reinstall Server</span>
                          </a>
                          <a 
                            className="dropdown-item"
                            onClick={() => setShowSnapshotModal(true)}
                          >
                            <span>📸 Restore Snapshot</span>
                          </a>
                          <hr className="dropdown-divider" />
                          <a 
                            className={`dropdown-item ${actionLoading.start ? 'has-text-grey' : ''}`}
                            onClick={() => !actionLoading.start && handleServerAction('start')}
                          >
                            <span>▶️ Start Server</span>
                            {actionLoading.start && <span className="is-pulled-right">⏳</span>}
                          </a>
                          <a 
                            className={`dropdown-item ${actionLoading.resetPassword ? 'has-text-grey' : ''}`}
                            onClick={() => !actionLoading.resetPassword && handlePasswordReset()}
                          >
                            <span>🔑 Reset Root Password</span>
                            {actionLoading.resetPassword && <span className="is-pulled-right">⏳</span>}
                          </a>
                          <hr className="dropdown-divider" />
                          <a 
                            className={`dropdown-item ${actionLoading.console ? 'has-text-grey' : ''}`}
                            onClick={() => !actionLoading.console && handleConsoleToggle(details?.vncstatus === 'enabled' ? 'disable' : 'enable')}
                          >
                            <span>🖥️ {details?.vncstatus === 'enabled' ? 'Disable' : 'Enable'} Console</span>
                            {actionLoading.console && <span className="is-pulled-right">⏳</span>}
                          </a>
                        </div>
                      </div>
                    </div>
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

      {/* Reinstall Modal */}
      <ReinstallModal
        isOpen={showReinstallModal}
        onClose={() => setShowReinstallModal(false)}
        serverId={serverId}
        serverName={serverName}
        onReinstall={handleReinstall}
      />

      {/* Snapshot Restore Modal */}
      <SnapshotModal
        isOpen={showSnapshotModal}
        onClose={() => setShowSnapshotModal(false)}
        serverId={serverId}
        serverName={serverName}
        onRestore={handleSnapshotRestore}
      />
    </>
  );
}
