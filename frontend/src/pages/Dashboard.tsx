import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useAuth } from '../context/AuthContext';
import api from '../lib/api';
import type { Project } from '../types';
import ThemeToggle from '../components/ThemeToggle';

function NewProjectModal({ onClose }: { onClose: () => void }) {
  const qc = useQueryClient();
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [error, setError] = useState('');

  const mutation = useMutation({
    mutationFn: () => api.post('/projects', { name, description: description || undefined }),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ['projects'] }); onClose(); },
    onError: (err: any) => setError(err.response?.data?.error || 'Failed to create project'),
  });

  return (
    <div className="modal-overlay" onClick={e => e.target === e.currentTarget && onClose()}>
      <div className="modal">
        <div className="modal-header">
          <h3>New Project</h3>
          <button className="btn btn-ghost btn-sm" onClick={onClose}>✕</button>
        </div>
        {error && <div className="error-banner">{error}</div>}
        <form onSubmit={e => { e.preventDefault(); mutation.mutate(); }}>
          <div className="form-group">
            <label>Project Name *</label>
            <input
              className="form-control"
              placeholder="My awesome project"
              value={name}
              onChange={e => setName(e.target.value)}
              required autoFocus
            />
          </div>
          <div className="form-group">
            <label>Description</label>
            <textarea
              className="form-control"
              placeholder="What's this project about?"
              value={description}
              onChange={e => setDescription(e.target.value)}
            />
          </div>
          <div className="modal-footer">
            <button type="button" className="btn btn-secondary" onClick={onClose}>Cancel</button>
            <button type="submit" className="btn btn-primary" disabled={mutation.isPending}>
              {mutation.isPending ? 'Creating…' : 'Create Project'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}

export default function Dashboard() {
  const { user, logout } = useAuth();
  const navigate = useNavigate();
  const [showModal, setShowModal] = useState(false);

  const { data, isLoading, isError } = useQuery({
    queryKey: ['projects'],
    queryFn: () => api.get<{ projects: Project[] }>('/projects').then(r => r.data.projects),
  });

  const projects = data ?? [];

  return (
    <>
      <nav className="navbar">
        <span className="navbar-brand">⚡ TaskFlow</span>
        <div className="navbar-right">
          <ThemeToggle />
          <span className="text-muted text-sm">Hi, {user?.name?.split(' ')[0]}</span>
          <button className="btn btn-ghost btn-sm" onClick={logout}>Sign out</button>
        </div>
      </nav>

      <div className="main-content">
        <div className="projects-header">
          <h2>My Projects</h2>
          <button className="btn btn-primary" onClick={() => setShowModal(true)}>+ New Project</button>
        </div>

        {isLoading && (
          <div className="spinner-wrap"><div className="spinner" /></div>
        )}

        {isError && (
          <div className="error-banner">Failed to load projects. Please refresh.</div>
        )}

        {!isLoading && !isError && projects.length === 0 && (
          <div className="empty-state">
            <div className="empty-icon">📋</div>
            <h3>No projects yet</h3>
            <p>Create your first project to start managing tasks</p>
            <button className="btn btn-primary" onClick={() => setShowModal(true)}>+ Create Project</button>
          </div>
        )}

        {!isLoading && projects.length > 0 && (
          <div className="projects-grid">
            {projects.map(p => (
              <div key={p.id} className="card project-card" onClick={() => navigate(`/projects/${p.id}`)}>
                <h3>{p.name}</h3>
                <p>{p.description || 'No description provided.'}</p>
                <div className="project-card-footer">
                  <span className="project-card-date">
                    {new Date(p.created_at).toLocaleDateString('en-IN', { day: 'numeric', month: 'short', year: 'numeric' })}
                  </span>
                  <span className="btn btn-ghost btn-sm">Open →</span>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>

      {showModal && <NewProjectModal onClose={() => setShowModal(false)} />}
    </>
  );
}
