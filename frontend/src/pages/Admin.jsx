import { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import Navbar from '../components/Navbar';
import { checkAuthentication, clearCookiesAndRedirectHome, redirectToLogin } from '../utils/auth';

export default function Admin() {
  const [users, setUsers] = useState([]);
  const [loading, setLoading] = useState(true);
  const [newUser, setNewUser] = useState({ email: '', groups: [] });
  const [editingUser, setEditingUser] = useState(null);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');

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
    // Verify authentication before loading the page
    const authenticate = async () => {
      const userData = await checkAuthentication();
      if (userData) {
        fetchUsers();
      }
      // If authentication fails, checkAuthentication will handle the redirect
    };

    authenticate();
  }, []);

  const fetchUsers = async () => {
    try {
      const response = await fetch('/api/admin/users', { credentials: 'include' });
      if (handleAuthFailure(response)) {
        return;
      }
      if (response.ok) {
        const data = await response.json();
        setUsers(data.data || []);
      } else {
        setError('Failed to fetch users');
      }
    } catch (err) {
      setError('Error fetching users');
    } finally {
      setLoading(false);
    }
  };

  const handleAddUser = async (e) => {
    e.preventDefault();
    setError('');
    setSuccess('');

    if (!newUser.email || newUser.groups.length === 0) {
      setError('Please provide both email and at least one group');
      return;
    }

    try {
      const response = await fetch('/api/admin/users', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        credentials: 'include',
        body: JSON.stringify(newUser),
      });

      if (handleAuthFailure(response)) {
        return;
      }

      if (response.ok) {
        setSuccess('User added successfully');
        setNewUser({ email: '', groups: [] });
        fetchUsers();
      } else {
        const data = await response.json();
        setError(data.message || 'Failed to add user');
      }
    } catch (err) {
      setError('Error adding user');
    }
  };

  const handleUpdateGroups = async (userEmail, groups) => {
    setError('');
    setSuccess('');

    try {
      const response = await fetch(`/api/admin/users/${encodeURIComponent(userEmail)}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        credentials: 'include',
        body: JSON.stringify({ groups }),
      });

      if (handleAuthFailure(response)) {
        return;
      }

      if (response.ok) {
        setSuccess('User groups updated successfully');
        setEditingUser(null);
        fetchUsers();
      } else {
        const data = await response.json();
        setError(data.message || 'Failed to update user groups');
      }
    } catch (err) {
      setError('Error updating user groups');
    }
  };

  const handleDeleteUser = async (userEmail) => {
    if (!confirm(`Are you sure you want to delete user ${userEmail}?`)) {
      return;
    }

    setError('');
    setSuccess('');

    try {
      const response = await fetch(`/api/admin/users/${encodeURIComponent(userEmail)}`, {
        method: 'DELETE',
        credentials: 'include',
      });

      if (handleAuthFailure(response)) {
        return;
      }

      if (response.ok) {
        setSuccess('User deleted successfully');
        fetchUsers();
      } else {
        const data = await response.json();
        setError(data.message || 'Failed to delete user');
      }
    } catch (err) {
      setError('Error deleting user');
    }
  };

  const handleGroupsChange = (groupString) => {
    const groups = groupString.split(',').map(g => g.trim()).filter(g => g !== '');
    return groups;
  };

  if (loading) {
    return (
      <>
        <section className="section">
          <div className="container">
            <div className="has-text-centered">
              <p>Loading...</p>
            </div>
          </div>
        </section>
      </>
    );
  }

  return (
    <>
      <section className="section">
        <div className="container">
          <div className="level">
            <div className="level-left">
              <div className="level-item">
                <h1 className="title">Admin Panel</h1>
              </div>
            </div>
            <div className="level-right">
              <div className="level-item">
                <Link to="/dashboard" className="button">
                  <span>Back to Dashboard</span>
                </Link>
              </div>
            </div>
          </div>

          <h2 className="subtitle">Manage users and their groups</h2>

          {error && (
            <div className="notification is-danger">
              <button className="delete" onClick={() => setError('')}></button>
              {error}
            </div>
          )}

          {success && (
            <div className="notification is-success">
              <button className="delete" onClick={() => setSuccess('')}></button>
              {success}
            </div>
          )}

          <div className="columns">
            <div className="column is-two-thirds">
              <div className="box">
                <h3 className="title is-5">User Management</h3>
                <table className="table is-fullwidth is-striped">
                  <thead>
                    <tr>
                      <th>Email</th>
                      <th>Groups</th>
                      <th>Actions</th>
                    </tr>
                  </thead>
                  <tbody>
                    {users.map((user) => (
                      <tr key={user.id}>
                        <td>{user.email}</td>
                        <td>
                          {editingUser === user.email ? (
                            <input
                              className="input"
                              type="text"
                              defaultValue={user.groups.join(', ')}
                              onKeyPress={(e) => {
                                if (e.key === 'Enter') {
                                  const groups = handleGroupsChange(e.target.value);
                                  handleUpdateGroups(user.email, groups);
                                }
                              }}
                              onBlur={(e) => {
                                const groups = handleGroupsChange(e.target.value);
                                handleUpdateGroups(user.email, groups);
                              }}
                              autoFocus
                            />
                          ) : (
                            <span className="tag is-multiple">
                              {user.groups.map(group => (
                                <span key={group} className="tag is-info is-light mr-1">{group}</span>
                              ))}
                            </span>
                          )}
                        </td>
                        <td>
                          <div className="buttons">
                            <button
                              className="button is-small is-light is-warning"
                              onClick={() => setEditingUser(user.email)}
                              disabled={editingUser === user.email}
                            >
                              Edit Groups
                            </button>
                            <button
                              className="button is-small is-light is-danger"
                              onClick={() => handleDeleteUser(user.email)}
                            >
                              Delete
                            </button>
                          </div>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>
            
            <div className="column is-one-third">
              <div className="box">
                <h3 className="title is-5">Add New User</h3>
                <form onSubmit={handleAddUser}>
                  <div className="field">
                    <label className="label">Email</label>
                    <div className="control">
                      <input
                        className="input"
                        type="email"
                        placeholder="user@example.com"
                        value={newUser.email}
                        onChange={(e) => setNewUser({ ...newUser, email: e.target.value })}
                        required
                      />
                    </div>
                  </div>
                  
                  <div className="field">
                    <label className="label">Groups</label>
                    <div className="control">
                      <input
                        className="input"
                        type="text"
                        placeholder="*, callowaysutton, group1, group2"
                        onChange={(e) => {
                          const groups = handleGroupsChange(e.target.value);
                          setNewUser({ ...newUser, groups });
                        }}
                      />
                    </div>
                    <p className="help">
                      Enter groups separated by commas. Use "*" for admin access.
                    </p>
                  </div>
                  
                  <div className="field">
                    <div className="control">
                      <button type="submit" className="button is-primary is-fullwidth">
                        Add User
                      </button>
                    </div>
                  </div>
                </form>
              </div>
            </div>
          </div>
        </div>
      </section>
    </>
  );
}
