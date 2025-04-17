import React, { useEffect, useState } from 'react';
import UserForm from './components/UserForm';

function App() {
  const [users, setUsers] = useState([]);
  const [editingUser, setEditingUser] = useState(null);

  const fetchUsers = async () => {
    const response = await fetch('http://localhost:8000/users');
    const data = await response.json();
    setUsers(data);
  };

  useEffect(() => {
    fetchUsers();
  }, []);

  const handleCreate = async (userData) => {
    await fetch('http://localhost:8000/users', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(userData)
    });
    fetchUsers();
  };

  const handleUpdate = async (userData) => {
    await fetch(`http://localhost:8000/users/${editingUser.id}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(userData)
    });
    setEditingUser(null);
    fetchUsers();
  };

  const handleDelete = async (id) => {
    await fetch(`http://localhost:8000/users/${id}`, { method: 'DELETE' });
    fetchUsers();
  };

  return (
    <div>
      <h1>User Management</h1>
      
      <h2>{editingUser ? 'Edit User' : 'Add New User'}</h2>
      <UserForm
        user={editingUser}
        onSave={editingUser ? handleUpdate : handleCreate}
        onCancel={() => setEditingUser(null)}
      />

      <h2>User List</h2>
      <ul>
        {users.map(user => (
          <li key={user.id}>
            {user.name} - {user.email}
            <button onClick={() => setEditingUser(user)}>Edit</button>
            <button onClick={() => handleDelete(user.id)}>Delete</button>
          </li>
        ))}
      </ul>
    </div>
  );
}

export default App;
