import { useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { DndContext, useDraggable, useDroppable } from '@dnd-kit/core';
import type { DragEndEvent } from '@dnd-kit/core';
import api from '../lib/api';
import type { Project, Task } from '../types';
import ThemeToggle from '../components/ThemeToggle';

const STATUS_COLUMNS: { key: Task['status']; label: string; colClass: string }[] = [
  { key: 'todo',        label: 'To Do',       colClass: 'col-todo' },
  { key: 'in_progress', label: 'In Progress',  colClass: 'col-progress' },
  { key: 'done',        label: 'Done',         colClass: 'col-done' },
];

const PRIORITY_NEXT: Record<Task['status'], Task['status'] | null> = {
  todo:        'in_progress',
  in_progress: 'done',
  done:        null,
};

function AddTaskModal({ projectId, onClose }: { projectId: string; onClose: () => void }) {
  const qc = useQueryClient();
  const [title, setTitle]       = useState('');
  const [description, setDesc]  = useState('');
  const [priority, setPriority] = useState<Task['priority']>('medium');
  const [assignee, setAssignee] = useState('');
  const [dueDate, setDueDate]   = useState('');
  const [error, setError]       = useState('');

  const mutation = useMutation({
    mutationFn: () => api.post(`/projects/${projectId}/tasks`, {
      title,
      description: description || undefined,
      priority,
      assignee_id: assignee || undefined,
      due_date:    dueDate   || undefined,
    }),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ['tasks', projectId] }); onClose(); },
    onError: (err: any) => setError(err.response?.data?.error || 'Failed to create task'),
  });

  return (
    <div className="modal-overlay" onClick={e => e.target === e.currentTarget && onClose()}>
      <div className="modal">
        <div className="modal-header">
          <h3>Add Task</h3>
          <button className="btn btn-ghost btn-sm" onClick={onClose}>✕</button>
        </div>
        {error && <div className="error-banner">{error}</div>}
        <form onSubmit={e => { e.preventDefault(); mutation.mutate(); }}>
          <div className="form-group">
            <label>Title *</label>
            <input className="form-control" placeholder="Task title" value={title} onChange={e => setTitle(e.target.value)} required autoFocus />
          </div>
          <div className="form-group">
            <label>Description</label>
            <textarea className="form-control" placeholder="Optional details…" value={description} onChange={e => setDesc(e.target.value)} />
          </div>
          <div className="form-group">
            <label>Priority</label>
            <select className="form-control" value={priority} onChange={e => setPriority(e.target.value as Task['priority'])}>
              <option value="low">Low</option>
              <option value="medium">Medium</option>
              <option value="high">High</option>
            </select>
          </div>
          <div className="form-group">
            <label>Assignee (name or ID)</label>
            <input className="form-control" placeholder="Optional" value={assignee} onChange={e => setAssignee(e.target.value)} />
          </div>
          <div className="form-group">
            <label>Due Date</label>
            <input type="date" className="form-control" value={dueDate} onChange={e => setDueDate(e.target.value)} />
          </div>
          <div className="modal-footer">
            <button type="button" className="btn btn-secondary" onClick={onClose}>Cancel</button>
            <button type="submit" className="btn btn-primary" disabled={mutation.isPending}>
              {mutation.isPending ? 'Adding…' : 'Add Task'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}

function DroppableColumn({ id, colClass, label, children, count }: { id: string; colClass: string; label: string; children: React.ReactNode; count: number }) {
  const { isOver, setNodeRef } = useDroppable({ id });
  return (
    <div 
      ref={setNodeRef} 
      className={`board-column ${colClass}`} 
      style={{ border: isOver ? '2px dashed var(--primary)' : undefined }}
    >
      <div className="column-header">
        <div className="column-title">
          <span className="column-dot" />
          {label}
        </div>
        <span className="column-count">{count}</span>
      </div>
      {children}
    </div>
  );
}

function TaskCard({ task, onMove }: { task: Task; onMove: (id: string, st: Task['status'] | 'delete') => void }) {
  const { attributes, listeners, setNodeRef, transform, isDragging } = useDraggable({
    id: task.id,
    data: { status: task.status }
  });

  const nextStatus = PRIORITY_NEXT[task.status];

  const style: React.CSSProperties | undefined = transform ? {
    transform: `translate3d(${transform.x}px, ${transform.y}px, 0)`,
    zIndex: 100,
    opacity: isDragging ? 0.8 : 1,
    boxShadow: isDragging ? 'var(--shadow-lg)' : undefined,
  } : undefined;

  return (
    <div ref={setNodeRef} style={style} {...attributes} {...listeners} className="task-card">
      <div className="task-title">{task.title}</div>
      {task.description && <div className="task-desc">{task.description}</div>}
      <div className="task-footer">
        <span className={`badge badge-${task.priority}`} style={{ textTransform: 'capitalize' }}>{task.priority}</span>
        {task.due_date && (
          <span className="text-sm text-muted">
            📅 {new Date(task.due_date).toLocaleDateString('en-IN', { day: 'numeric', month: 'short' })}
          </span>
        )}
        {nextStatus && (
          <button
            className="status-btn"
            onPointerDown={e => e.stopPropagation()} /* prevent dragging when clicking button */
            onClick={(e) => { e.stopPropagation(); onMove(task.id, nextStatus); }}
            title={`Move to ${nextStatus.replace('_', ' ')}`}
          >
            {nextStatus === 'in_progress' ? '▶ Start' : '✓ Done'}
          </button>
        )}
        <button
          className="btn btn-ghost btn-sm"
          style={{ marginLeft: 'auto', color: 'var(--danger)', padding: '0.2rem 0.4rem', cursor: 'pointer', pointerEvents: 'auto' }}
          onPointerDown={e => e.stopPropagation()}
          onClick={(e) => { e.stopPropagation(); onMove(task.id, 'delete'); }}
          title="Delete task"
        >
          🗑
        </button>
      </div>
    </div>
  );
}

export default function ProjectBoard() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const qc = useQueryClient();
  const [showAddTask, setShowAddTask] = useState(false);
  const [taskToDelete, setTaskToDelete] = useState<Task | null>(null);

  const { data: project } = useQuery({
    queryKey: ['project', id],
    queryFn: () => api.get<{ project: Project }>(`/projects/${id}`).then(r => r.data.project),
    enabled: !!id,
  });

  const { data: tasks, isLoading, isError } = useQuery({
    queryKey: ['tasks', id],
    queryFn: () => api.get<{ tasks: Task[] }>(`/projects/${id}/tasks`).then(r => r.data.tasks),
    enabled: !!id,
  });

  const moveTaskMutation = useMutation({
    mutationFn: ({ taskId, newStatus }: { taskId: string; newStatus: Task['status'] }) => 
      api.patch(`/tasks/${taskId}`, { status: newStatus }),
    onMutate: async ({ taskId, newStatus }) => {
      await qc.cancelQueries({ queryKey: ['tasks', id] });
      const prev = qc.getQueryData<Task[]>(['tasks', id]);
      qc.setQueryData<Task[]>(['tasks', id], old =>
        old?.map(t => t.id === taskId ? { ...t, status: newStatus } : t)
      );
      return { prev };
    },
    onError: (_err, _vars, ctx) => qc.setQueryData(['tasks', id], ctx?.prev),
    onSettled: () => qc.invalidateQueries({ queryKey: ['tasks', id] }),
  });

  const handleDragEnd = (event: DragEndEvent) => {
    const { active, over } = event;
    if (!over) return;
    const taskId = active.id as string;
    const newStatus = over.id as Task['status'];
    const currentStatus = active.data.current?.status;
    if (currentStatus && newStatus !== currentStatus) {
      moveTaskMutation.mutate({ taskId, newStatus });
    }
  };

  const handleMoveStatus = (taskId: string, newStatus: Task['status'] | 'delete') => {
    if (newStatus === 'delete') {
      const t = allTasks.find(x => x.id === taskId);
      if (t) setTaskToDelete(t);
      return;
    }
    moveTaskMutation.mutate({ taskId, newStatus });
  };

  const deleteMutation = useMutation({
    mutationFn: (taskId: string) => api.delete(`/tasks/${taskId}`),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ['tasks', id] }); setTaskToDelete(null); },
  });

  const allTasks = tasks ?? [];

  return (
    <>
      <nav className="navbar">
        <div style={{ display: 'flex', alignItems: 'center', gap: '1rem' }}>
          <button className="btn btn-ghost btn-sm" onClick={() => navigate('/')}>← Back</button>
          <span className="navbar-brand">⚡ TaskFlow</span>
        </div>
        <div style={{ display: 'flex', alignItems: 'center', gap: '1rem' }}>
          <ThemeToggle />
          <button className="btn btn-primary btn-sm" onClick={() => setShowAddTask(true)}>+ Add Task</button>
        </div>
      </nav>

      <div className="main-content">
        <div className="board-header">
          <div>
            <h2>{project?.name || ''}</h2>
            {project?.description && (
              <p className="text-muted text-sm" style={{ marginTop: '.25rem' }}>{project.description}</p>
            )}
          </div>
        </div>

        {isLoading && <div className="spinner-wrap"><div className="spinner" /></div>}
        {isError  && <div className="error-banner">Failed to load tasks. Please refresh.</div>}

        {!isLoading && !isError && (
          <DndContext onDragEnd={handleDragEnd}>
            <div className="board-columns">
              {STATUS_COLUMNS.map(col => {
                const colTasks = allTasks.filter(t => t.status === col.key);
                return (
                  <DroppableColumn key={col.key} id={col.key} colClass={col.colClass} label={col.label} count={colTasks.length}>
                    {colTasks.length === 0 ? (
                      <div className="empty-state" style={{ padding: '2rem .5rem' }}>
                        <div className="empty-icon" style={{ fontSize: '1.5rem' }}>✦</div>
                        <p style={{ fontSize: '.8rem' }}>No tasks here</p>
                      </div>
                    ) : (
                      colTasks.map(task => (
                        <TaskCard key={task.id} task={task} onMove={handleMoveStatus} />
                      ))
                    )}
                  </DroppableColumn>
                );
              })}
            </div>
          </DndContext>
        )}
      </div>

      {showAddTask && id && <AddTaskModal projectId={id} onClose={() => setShowAddTask(false)} />}
      
      {taskToDelete && (
        <div className="modal-overlay" onClick={e => e.target === e.currentTarget && setTaskToDelete(null)}>
          <div className="modal" style={{ maxWidth: '400px' }}>
            <div className="modal-header">
              <h3>Delete Task</h3>
              <button className="btn btn-ghost btn-sm" onClick={() => setTaskToDelete(null)}>✕</button>
            </div>
            <p>Are you sure you want to delete the task <strong>"{taskToDelete.title}"</strong>?</p>
            <p className="text-sm text-dim mt-1">This action cannot be undone.</p>
            <div className="modal-footer mt-2">
              <button className="btn btn-secondary" onClick={() => setTaskToDelete(null)}>Cancel</button>
              <button 
                className="btn btn-danger" 
                onClick={() => deleteMutation.mutate(taskToDelete.id)}
                disabled={deleteMutation.isPending}
              >
                {deleteMutation.isPending ? 'Deleting...' : 'Delete'}
              </button>
            </div>
          </div>
        </div>
      )}
    </>
  );
}
