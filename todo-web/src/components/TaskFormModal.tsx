import { useEffect, useState, type FormEvent } from "react";
import type { Category, Task } from "../lib/api";
import { maxDueDateIso, todayIso } from "../lib/dates";
import { clearDraft, loadDraft, saveDraft } from "../lib/drafts";

const DRAFT_KEY = "task-form";

export type TaskFormValues = {
  title: string;
  description: string;
  dueDate: string;
  categoryId: string;
};

type Props = {
  open: boolean;
  initial?: Task | null;
  categories: Category[];
  onClose: () => void;
  onSubmit: (values: TaskFormValues) => Promise<void>;
};

function emptyValues(): TaskFormValues {
  return { title: "", description: "", dueDate: "", categoryId: "" };
}

function fromTask(t: Task): TaskFormValues {
  return {
    title: t.title ?? "",
    description: t.description ?? "",
    dueDate: t.dueDate ?? "",
    categoryId: t.categoryId ?? "",
  };
}

export function TaskFormModal({ open, initial, categories, onClose, onSubmit }: Props) {
  const [values, setValues] = useState<TaskFormValues>(emptyValues());
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const isEdit = !!initial;
  const draftKey = isEdit ? `${DRAFT_KEY}:${initial!.id}` : `${DRAFT_KEY}:new`;

  useEffect(() => {
    if (!open) return;
    if (initial) {
      const draft = loadDraft<TaskFormValues>(draftKey);
      setValues(draft ?? fromTask(initial));
    } else {
      const draft = loadDraft<TaskFormValues>(draftKey);
      setValues(draft ?? emptyValues());
    }
    setError(null);
  }, [open, initial, draftKey]);

  useEffect(() => {
    if (!open) return;
    saveDraft<TaskFormValues>(draftKey, values);
  }, [open, draftKey, values]);

  useEffect(() => {
    if (!open) return;
    const handler = (e: KeyboardEvent) => {
      if (e.key === "Escape") handleCancel();
    };
    window.addEventListener("keydown", handler);
    return () => window.removeEventListener("keydown", handler);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [open]);

  const handleCancel = () => {
    clearDraft(draftKey);
    onClose();
  };

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    if (submitting) return;
    const trimmedTitle = values.title.trim();
    if (!trimmedTitle) {
      setError("Title is required");
      return;
    }
    if (values.dueDate) {
      if (values.dueDate < todayIso() || values.dueDate > maxDueDateIso()) {
        setError("Due date must be between today and 10 years from now");
        return;
      }
    }
    setSubmitting(true);
    setError(null);
    try {
      await onSubmit({ ...values, title: trimmedTitle });
      clearDraft(draftKey);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to save task");
    } finally {
      setSubmitting(false);
    }
  };

  if (!open) return null;

  return (
    <div className="modal-backdrop" onClick={handleCancel}>
      <form
        className="modal"
        role="dialog"
        aria-modal="true"
        onClick={(e) => e.stopPropagation()}
        onSubmit={handleSubmit}
      >
        <h3>{isEdit ? "Edit task" : "New task"}</h3>
        <label>
          Title
          <input
            type="text"
            value={values.title}
            onChange={(e) => setValues((v) => ({ ...v, title: e.target.value }))}
            required
            maxLength={200}
            autoFocus
          />
        </label>
        <label>
          Description
          <textarea
            value={values.description}
            onChange={(e) => setValues((v) => ({ ...v, description: e.target.value }))}
            rows={3}
            maxLength={2000}
          />
        </label>
        <div className="row">
          <label>
            Due date
            <input
              type="date"
              value={values.dueDate}
              min={todayIso()}
              max={maxDueDateIso()}
              onChange={(e) => setValues((v) => ({ ...v, dueDate: e.target.value }))}
            />
          </label>
          <label>
            Category
            <select
              value={values.categoryId}
              onChange={(e) => setValues((v) => ({ ...v, categoryId: e.target.value }))}
            >
              <option value="">— None —</option>
              {categories.map((c) => (
                <option key={c.id} value={c.id}>
                  {c.name}
                </option>
              ))}
            </select>
          </label>
        </div>
        {error && <div className="form-error">{error}</div>}
        <div className="modal-actions">
          <button type="button" className="btn btn-ghost" onClick={handleCancel}>
            Cancel
          </button>
          <button type="submit" className="btn btn-primary" disabled={submitting}>
            {submitting ? "Saving…" : isEdit ? "Save changes" : "Create task"}
          </button>
        </div>
      </form>
    </div>
  );
}
