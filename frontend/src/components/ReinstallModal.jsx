import { useState, useEffect } from 'react';

export default function ReinstallModal({ 
  isOpen, 
  onClose, 
  serverId, 
  serverName, 
  onReinstall 
}) {
  const [osOptions, setOsOptions] = useState([]);
  const [appOptions, setAppOptions] = useState([]);
  const [loading, setLoading] = useState(false);
  const [formData, setFormData] = useState({
    reinstall_type: 'os',
    os_app_id: '',
    authentication: 'password',
    ssh_key: ''
  });
  const [showConfirmation, setShowConfirmation] = useState(false);

  // Fetch OS and App options
  useEffect(() => {
    if (isOpen) {
      fetchOptions();
    }
  }, [isOpen]);

  const fetchOptions = async () => {
    try {
      const [osRes, appRes] = await Promise.all([
        fetch('/api/os_options', { credentials: 'include' }),
        fetch('/api/apps', { credentials: 'include' })
      ]);
      
      const osData = await osRes.json();
      const appData = await appRes.json();
      
      setOsOptions(osData);
      setAppOptions(appData);
      
      // Set default selection
      if (osData.length > 0) {
        setFormData(prev => ({ ...prev, os_app_id: osData[0].id }));
      }
    } catch (error) {
      console.error('Error fetching options:', error);
    }
  };

  const handleSubmit = async () => {
    if (!formData.os_app_id) {
      alert('Please select an OS or App');
      return;
    }

    if (formData.authentication === 'ssh' && !formData.ssh_key.trim()) {
      alert('Please provide an SSH key');
      return;
    }

    setLoading(true);
    try {
      const response = await fetch(`/api/server/${serverId}/reinstall`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        credentials: 'include',
        body: JSON.stringify(formData)
      });

      const result = await response.json();
      
      if (response.ok) {
        onReinstall(result);
        onClose();
      } else {
        alert(`Error: ${result.error || 'Failed to reinstall server'}`);
      }
    } catch (error) {
      console.error('Error reinstalling server:', error);
      alert('Failed to reinstall server');
    } finally {
      setLoading(false);
      setShowConfirmation(false);
    }
  };

  const handleTypeChange = (type) => {
    setFormData(prev => ({
      ...prev,
      reinstall_type: type,
      os_app_id: type === 'os' 
        ? (osOptions.length > 0 ? osOptions[0].id : '')
        : (appOptions.length > 0 ? appOptions[0].id : '')
    }));
  };

  const currentOptions = formData.reinstall_type === 'os' ? osOptions : appOptions;

  if (!isOpen) return null;

  return (
    <div className="modal is-active">
      <div className="modal-background" onClick={onClose}></div>
      <div className="modal-card">
        <header className="modal-card-head">
          <p className="modal-card-title">Reinstall Server - {serverName}</p>
          <button className="delete" aria-label="close" onClick={onClose}></button>
        </header>
        
        <section className="modal-card-body">
          {!showConfirmation ? (
            <>
              <div className="notification is-warning">
                <strong>Warning:</strong> Reinstalling will wipe out all data on the server. This action cannot be undone.
              </div>

              <div className="field">
                <label className="label">Reinstall Type</label>
                <div className="control">
                  <div className="tabs is-boxed">
                    <ul>
                      <li className={formData.reinstall_type === 'os' ? 'is-active' : ''}>
                        <a onClick={() => handleTypeChange('os')}>
                          <span>Operating System</span>
                        </a>
                      </li>
                      <li className={formData.reinstall_type === 'app' ? 'is-active' : ''}>
                        <a onClick={() => handleTypeChange('app')}>
                          <span>Application</span>
                        </a>
                      </li>
                    </ul>
                  </div>
                </div>
              </div>

              <div className="field">
                <label className="label">
                  {formData.reinstall_type === 'os' ? 'Operating System' : 'Application'}
                </label>
                <div className="control">
                  <div className="select is-fullwidth">
                    <select
                      value={formData.os_app_id}
                      onChange={(e) => setFormData(prev => ({ ...prev, os_app_id: parseInt(e.target.value) }))}
                    >
                      <option value="">Select {formData.reinstall_type === 'os' ? 'OS' : 'App'}</option>
                      {currentOptions.map(option => (
                        <option key={option.id} value={option.id}>
                          {formData.reinstall_type === 'os' ? option.operatingsystem : option.app}
                        </option>
                      ))}
                    </select>
                  </div>
                </div>
              </div>

              <div className="field">
                <label className="label">Authentication Method</label>
                <div className="control">
                  <div className="tabs is-boxed">
                    <ul>
                      <li className={formData.authentication === 'password' ? 'is-active' : ''}>
                        <a onClick={() => setFormData(prev => ({ ...prev, authentication: 'password', ssh_key: '' }))}>
                          <span>Auto-generated Password</span>
                        </a>
                      </li>
                      <li className={formData.authentication === 'ssh' ? 'is-active' : ''}>
                        <a onClick={() => setFormData(prev => ({ ...prev, authentication: 'ssh' }))}>
                          <span>SSH Key</span>
                        </a>
                      </li>
                    </ul>
                  </div>
                </div>
              </div>

              {formData.authentication === 'password' && (
                <div className="notification is-info">
                  A random password will be generated automatically. You can retrieve it after reinstallation from the server credentials.
                </div>
              )}

              {formData.authentication === 'ssh' && (
                <div className="field">
                  <label className="label">SSH Public Key</label>
                  <div className="control">
                    <textarea
                      className="textarea"
                      placeholder="ssh-rsa AAAAB3NzaC1yc2E... your-email@example.com"
                      value={formData.ssh_key}
                      onChange={(e) => setFormData(prev => ({ ...prev, ssh_key: e.target.value }))}
                      rows="4"
                    />
                  </div>
                  <p className="help">Paste your SSH public key here. Make sure it's a valid public key.</p>
                </div>
              )}
            </>
          ) : (
            <div className="content">
              <h4>Confirm Reinstallation</h4>
              <p><strong>Server:</strong> {serverName}</p>
              <p><strong>Type:</strong> {formData.reinstall_type === 'os' ? 'Operating System' : 'Application'}</p>
              <p><strong>Selection:</strong> {
                currentOptions.find(opt => opt.id === formData.os_app_id)?.[
                  formData.reinstall_type === 'os' ? 'operatingsystem' : 'app'
                ] || 'Unknown'
              }</p>
              <p><strong>Authentication:</strong> {formData.authentication === 'ssh' ? 'SSH Key' : 'Auto-generated Password'}</p>
              
              <div className="notification is-danger">
                <strong>Final Warning:</strong> This will permanently delete all data on the server. Type "CONFIRM" below to proceed.
              </div>
              
              <div className="field">
                <div className="control">
                  <input
                    className="input"
                    type="text"
                    placeholder="Type CONFIRM to proceed"
                    id="confirmInput"
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
                className="button is-danger"
                onClick={() => setShowConfirmation(true)}
                disabled={!formData.os_app_id || (formData.authentication === 'ssh' && !formData.ssh_key.trim())}
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
                  const confirmInput = document.getElementById('confirmInput');
                  if (confirmInput.value === 'CONFIRM') {
                    handleSubmit();
                  } else {
                    alert('Please type "CONFIRM" to proceed');
                  }
                }}
                disabled={loading}
              >
                Reinstall Server
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
