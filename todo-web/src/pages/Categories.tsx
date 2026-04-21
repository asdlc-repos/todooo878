import { useCallback, useEffect, useState, type FormEvent } from "react";
import { api, ApiError, type Category } from "../lib/api";
import { useToast } from "../components/Toast";
import { ConfirmDialog } from "../components/ConfirmDialog";
import { clearDraft, loadDraft, saveDraft } from "../lib/drafts";

const DRAFT_KEY = "category-form";

type Draft = { name: string };

export default function CategoriesPage() {
  const toast = useToast();

  const [categories, setCategories] = useState<Category[]>([]);
  const [loading, setLoading] = useState(true);
  const [name, setName] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [pendingDelete, setPendingDelete] = useState<Category | null>(null);

  const load = useCallback(async () => {
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
      await load();
      if (!cancelled) setLoading(false);
    })();
    return () => {
      cancelled = true;
    };
  }, [load]);

  useEffect(() => {
    const draft = loadDraft<Draft>(DRAFT_KEY);
    if (draft?.name) setName(draft.name);
  }, []);

  useEffect(() => {
    saveDraft<Draft>(DRAFT_KEY, { name });
  }, [name]);

  const onCreate = async (e: FormEvent) => {
    e.preventDefault();
    const trimmed = name.trim();
    if (!trimmed) return;
    if (submitting) return;
    setSubmitting(true);
    try {
      const created = await api.createCategory(trimmed);
      setCategories((cs) => [...cs, created]);
      setName("");
      clearDraft(DRAFT_KEY);
      toast.success("Category created");
    } catch (err) {
      const msg = err instanceof ApiError ? err.message : "Failed to create category";
      toast.error(msg);
    } finally {
      setSubmitting(false);
    }
  };

  const confirmDelete = async () => {
    if (!pendingDelete) return;
    const c = pendingDelete;
    setPendingDelete(null);
    try {
      await api.deleteCategory(c.id);
      setCategories((cs) => cs.filter((x) => x.id !== c.id));
      toast.success("Category deleted");
    } catch (err) {
      const msg = err instanceof ApiError ? err.message : "Failed to delete category";
      toast.error(msg);
    }
  };

  return (
    <div className="page">
      <div className="page-header">
        <h1>Categories</h1>
      </div>

      <form className="inline-form" onSubmit={onCreate}>
        <input
          type="text"
          placeholder="Category name"
          value={name}
          onChange={(e) => setName(e.target.value)}
          maxLength={80}
          required
        />
        <button
          type="submit"
          className="btn btn-primary"
          disabled={submitting || !name.trim()}
        >
          {submitting ? "Adding…" : "Add"}
        </button>
      </form>

      {loading ? (
        <div className="muted">Loading…</div>
      ) : categories.length === 0 ? (
        <div className="empty">No categories yet.</div>
      ) : (
        <ul className="category-list">
          {categories.map((c) => (
            <li key={c.id} className="category-item">
              <span>{c.name}</span>
              <button
                type="button"
                className="btn btn-danger-ghost"
                onClick={() => setPendingDelete(c)}
              >
                Delete
              </button>
            </li>
          ))}
        </ul>
      )}

      <ConfirmDialog
        open={!!pendingDelete}
        title="Delete category"
        message={
          pendingDelete
            ? `Delete category "${pendingDelete.name}"? Tasks in this category will lose their label.`
            : ""
        }
        onCancel={() => setPendingDelete(null)}
        onConfirm={confirmDelete}
      />
    </div>
  );
}
