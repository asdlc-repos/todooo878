import { useCallback, useEffect, useMemo, useState } from "react";
import {
  api,
  ApiError,
  type Category,
  type Task,
  type TaskSort,
  type TaskStatus,
} from "../lib/api";
import { useToast } from "../components/Toast";
import { ConfirmDialog } from "../components/ConfirmDialog";
import { TaskFormModal, type TaskFormValues } from "../components/TaskFormModal";
import { formatDate, isOverdue } from "../lib/dates";

export default function TasksPage() {
  const toast = useToast();

  const [tasks, setTasks] = useState<Task[]>([]);
  const [categories, setCategories] = useState<Category[]>([]);
  const [loading, setLoading] = useState(true);

  const [statusFilter, setStatusFilter] = useState<TaskStatus>("all");
  const [categoryFilter, setCategoryFilter] = useState<string>("");
  const [sort, setSort] = useState<TaskSort>("due_asc");

  const [formOpen, setFormOpen] = useState(false);
  const [editing, setEditing] = useState<Task | null>(null);
  const [pendingDelete, setPendingDelete] = useState<Task | null>(null);

  const loadTasks = useCallback(async () => {
    try {
      const list = await api.listTasks({
        status: statusFilter,
        category: categoryFilter || undefined,
        sort,
      });
      setTasks(list);
    } catch (err) {
      const msg = err instanceof ApiError ? err.message : "Failed to load tasks";
      toast.error(msg);
    }
  }, [statusFilter, categoryFilter, sort, toast]);

  const loadCategories = useCallback(async () => {
    try {
      const list = await api.listCategories();
      setCategories(list);
    } catch (err) {
      const msg = err instanceof ApiError ? err.message : "Failed to load categories";
      toast.error(msg);
    }
  }, [toast]);

  useEffect(() => {
    let cancelled = false;
    (async () => {
      setLoading(true);
      await Promise.all([loadTasks(), loadCategories()]);
      if (!cancelled) setLoading(false);
    })();
    return () => {
      cancelled = true;
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  useEffect(() => {
    loadTasks();
  }, [loadTasks]);

  const categoriesById = useMemo(() => {
    const m = new Map<string, Category>();
    for (const c of categories) m.set(c.id, c);
    return m;
  }, [categories]);

  const openCreate = () => {
    setEditing(null);
    setFormOpen(true);
  };

  const openEdit = (t: Task) => {
    setEditing(t);
    setFormOpen(true);
  };

  const closeForm = () => {
    setFormOpen(false);
    setEditing(null);
  };

  const submitForm = async (values: TaskFormValues) => {
    const payload = {
      title: values.title,
      description: values.description || undefined,
      dueDate: values.dueDate ? values.dueDate : null,
      categoryId: values.categoryId ? values.categoryId : null,
    };
    try {
      if (editing) {
        await api.updateTask(editing.id, payload);
        toast.success("Task updated");
      } else {
        await api.createTask({
          title: payload.title,
          description: payload.description,
          dueDate: payload.dueDate ?? undefined,
          categoryId: payload.categoryId,
        });
        toast.success("Task created");
      }
      closeForm();
      await loadTasks();
    } catch (err) {
      const msg = err instanceof ApiError ? err.message : "Failed to save task";
      toast.error(msg);
      throw err;
    }
  };

  const toggleComplete = async (t: Task) => {
    try {
      await api.updateTask(t.id, { completed: !t.completed });
      setTasks((ts) =>
        ts.map((x) => (x.id === t.id ? { ...x, completed: !t.completed } : x))
      );
    } catch (err) {
      const msg = err instanceof ApiError ? err.message : "Failed to update task";
      toast.error(msg);
    }
  };

  const confirmDelete = async () => {
    if (!pendingDelete) return;
    const t = pendingDelete;
    setPendingDelete(null);
    try {
      await api.deleteTask(t.id);
      toast.success("Task deleted");
      setTasks((ts) => ts.filter((x) => x.id !== t.id));
    } catch (err) {
      const msg = err instanceof ApiError ? err.message : "Failed to delete task";
      toast.error(msg);
    }
  };

  return (
    <div className="page">
      <div className="page-header">
        <h1>Tasks</h1>
        <button className="btn btn-primary" onClick={openCreate} type="button">
          New task
        </button>
      </div>

      <div className="filters">
        <label>
          Status
          <select
            value={statusFilter}
            onChange={(e) => setStatusFilter(e.target.value as TaskStatus)}
          >
            <option value="all">All</option>
            <option value="active">Active</option>
            <option value="completed">Completed</option>
          </select>
        </label>
        <label>
          Category
          <select
            value={categoryFilter}
            onChange={(e) => setCategoryFilter(e.target.value)}
          >
            <option value="">All categories</option>
            {categories.map((c) => (
              <option key={c.id} value={c.id}>
                {c.name}
              </option>
            ))}
          </select>
        </label>
        <label>
          Sort
          <select value={sort} onChange={(e) => setSort(e.target.value as TaskSort)}>
            <option value="due_asc">Due date ↑</option>
            <option value="due_desc">Due date ↓</option>
          </select>
        </label>
      </div>

      {loading ? (
        <div className="muted">Loading…</div>
      ) : tasks.length === 0 ? (
        <div className="empty">No tasks yet. Create your first one.</div>
      ) : (
        <ul className="task-list">
          {tasks.map((t) => {
            const overdue = !t.completed && isOverdue(t.dueDate);
            const cat = t.categoryId ? categoriesById.get(t.categoryId) : null;
            return (
              <li
                key={t.id}
                className={`task-item${t.completed ? " done" : ""}${
                  overdue ? " overdue" : ""
                }`}
              >
                <label className="task-check">
                  <input
                    type="checkbox"
                    checked={t.completed}
                    onChange={() => toggleComplete(t)}
                  />
                </label>
                <div className="task-body" onClick={() => openEdit(t)}>
                  <div className="task-title">{t.title}</div>
                  {t.description && <div className="task-desc">{t.description}</div>}
                  <div className="task-meta">
                    {t.dueDate && (
                      <span className={`chip${overdue ? " chip-overdue" : ""}`}>
                        Due {formatDate(t.dueDate)}
                      </span>
                    )}
                    {cat && <span className="chip">{cat.name}</span>}
                  </div>
                </div>
                <div className="task-actions">
                  <button
                    type="button"
                    className="btn btn-ghost"
                    onClick={() => openEdit(t)}
                  >
                    Edit
                  </button>
                  <button
                    type="button"
                    className="btn btn-danger-ghost"
                    onClick={() => setPendingDelete(t)}
                  >
                    Delete
                  </button>
                </div>
              </li>
            );
          })}
        </ul>
      )}

      <TaskFormModal
        open={formOpen}
        initial={editing}
        categories={categories}
        onClose={closeForm}
        onSubmit={submitForm}
      />

      <ConfirmDialog
        open={!!pendingDelete}
        title="Delete task"
        message={
          pendingDelete
            ? `Delete "${pendingDelete.title}"? This cannot be undone.`
            : ""
        }
        onCancel={() => setPendingDelete(null)}
        onConfirm={confirmDelete}
      />
    </div>
  );
}
