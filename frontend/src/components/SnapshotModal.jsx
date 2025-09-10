import { useState, useEffect } from 'react';

export default function SnapshotModal({ 
  isOpen, 
  onClose, 
  serverId, 
  serverName, 
  onRestore 
}) {
  const [snapshots, setSnapshots] = useState([]);
  const [loading, setLoading] = useState(false);
  const [selectedSnapshot, setSelectedSnapshot] = useState('');
  const [showConfirmation, setShowConfirmation] = useState(false);

  // Fetch snapshots when modal opens
  useEffect(() => {
    if (isOpen) {
      fetchSnapshots();
    }
  }, [isOpen, serverId]);

  const fetchSnapshots = async () => {
    setLoading(true);
    try {
      const response = await fetch(`/api/server/${serverId}/snapshots`, {
        credentials: 'include',
      });
      
      if (response.ok) {
        const data = await response.json();
        setSnapshots(data);
        if (data.length > 0) {
          setSelectedSnapshot(data[0].name);
        }
      } else {
        console.error('Failed to fetch snapshots');
        setSnapshots([]);
      }
    } catch (error) {
      console.error('Error fetching snapshots:', error);
      setSnapshots([]);
    } finally {
      setLoading(false);
    }
  };

  const handleRestore = async () => {
    if (!selectedSnapshot) {
      alert('Please select a snapshot to restore');
      return;
    }

    setLoading(true);
    try {
      const response = await fetch(`/api/server/${serverId}/restore-snapshot`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        credentials: 'include',
        body: JSON.stringify({
          snapshot_name: selectedSnapshot
        })
      });

      const result = await response.json();
      
      if (response.ok) {
        onRestore(result);
        onClose();
      } else {
        alert(`Error: ${result.error || 'Failed to restore snapshot'}`);
      }
    } catch (error) {
      console.error('Error restoring snapshot:', error);
      alert('Failed to restore snapshot');
    } finally {
      setLoading(false);
      setShowConfirmation(false);
    }
  };

  const formatDate = (dateString) => {
    try {
      return new Date(dateString).toLocaleString();
    } catch {
      return dateString;
    }
  };

  const formatSize = (sizeGB) => {
    if (sizeGB < 1) {
      return `${(sizeGB * 1024).toFixed(0)} MB`;
    }
    return `${sizeGB.toFixed(1)} GB`;
  };

  if (!isOpen) return null;

  return (
    <div className="modal is-active">
      <div className="modal-background" onClick={onClose}></div>
      <div className="modal-card">
        <header className="modal-card-head">
          <p className="modal-card-title">Restore Snapshot - {serverName}</p>
          <button className="delete" aria-label="close" onClick={onClose}></button>
        </header>
        
        <section className="modal-card-body">
          {!showConfirmation ? (
            <>
              <div className="notification is-warning">
                <strong>Warning:</strong> Restoring a snapshot will wipe out all current data and restore the server to the specified date. This action cannot be undone.
              </div>

              {loading ? (
                <div className="has-text-centered">
                  <div className="is-loading button is-large is-white"></div>
                  <p>Loading snapshots...</p>
                </div>
              ) : snapshots.length === 0 ? (
                <div className="notification is-info">
                  No snapshots available for this server.
                </div>
              ) : (
                <>
                  <div className="field">
                    <label className="label">Select Snapshot to Restore</label>
                    <div className="control">
                      <div className="select is-fullwidth">
                        <select
                          value={selectedSnapshot}
                          onChange={(e) => setSelectedSnapshot(e.target.value)}
                        >
                          {snapshots.map(snapshot => (
                            <option key={snapshot.id} value={snapshot.name}>
                              {snapshot.name} - {formatDate(snapshot.created_at)} ({formatSize(snapshot.size_gb)})
                            </option>
                          ))}
                        </select>
                      </div>
                    </div>
                  </div>

                  <div className="content">
                    <h6>Snapshot Details:</h6>
                    {snapshots.find(s => s.name === selectedSnapshot) && (
                      <table className="table is-fullwidth">
                        <tbody>
                          <tr>
                            <th>Name</th>
                            <td>{snapshots.find(s => s.name === selectedSnapshot).name}</td>
                          </tr>
                          <tr>
                            <th>Created</th>
                            <td>{formatDate(snapshots.find(s => s.name === selectedSnapshot).created_at)}</td>
                          </tr>
                          <tr>
                            <th>Size</th>
                            <td>{formatSize(snapshots.find(s => s.name === selectedSnapshot).size_gb)}</td>
                          </tr>
                          <tr>
                            <th>Status</th>
                            <td>
                              <span className={`tag ${snapshots.find(s => s.name === selectedSnapshot).status === 'completed' ? 'is-success' : 'is-warning'}`}>
                                {snapshots.find(s => s.name === selectedSnapshot).status}
                              </span>
                            </td>
                          </tr>
                        </tbody>
                      </table>
                    )}
                  </div>
                </>
              )}
            </>
          ) : (
            <div className="content">
              <h4>Confirm Snapshot Restoration</h4>
              <p><strong>Server:</strong> {serverName}</p>
              <p><strong>Snapshot:</strong> {selectedSnapshot}</p>
              
              <div className="notification is-danger">
                <strong>Final Warning:</strong> This will permanently delete all current data and restore the server to the snapshot state. Type "RESTORE" below to proceed.
              </div>
              
              <div className="field">
                <div className="control">
                  <input
                    className="input"
                    type="text"
                    placeholder="Type RESTORE to proceed"
                    id="restoreConfirmInput"
                  />
                </div>
              </div>
            </div>
          )}
        </section>

        <footer className="modal-card-foot">
          {!showConfirmation ? (
            <>
              <button
                className="button is-warning"
                onClick={() => setShowConfirmation(true)}
                disabled={!selectedSnapshot || snapshots.length === 0}
              >
                Continue
              </button>
              <button className="button" onClick={onClose}>Cancel</button>
            </>
          ) : (
            <>
              <button
                className={`button is-danger ${loading ? 'is-loading' : ''}`}
                onClick={() => {
                  const confirmInput = document.getElementById('restoreConfirmInput');
                  if (confirmInput.value === 'RESTORE') {
                    handleRestore();
                  } else {
                    alert('Please type "RESTORE" to proceed');
                  }
                }}
                disabled={loading}
              >
                Restore Snapshot
              </button>
              <button 
                className="button" 
                onClick={() => setShowConfirmation(false)}
                disabled={loading}
              >
                Back
              </button>
            </>
          )}
        </footer>
      </div>
    </div>
  );
}
