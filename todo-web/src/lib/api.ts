import { getApiBaseUrl } from "./config";

export class ApiError extends Error {
  status: number;
  constructor(message: string, status: number) {
    super(message);
    this.status = status;
  }
}

function readCsrfCookie(): string | null {
  if (typeof document === "undefined") return null;
  const match = document.cookie.match(/(?:^|;\s*)csrf=([^;]+)/);
  return match ? decodeURIComponent(match[1]) : null;
}

type RequestOptions = {
  method?: string;
  body?: unknown;
  query?: Record<string, string | undefined>;
};

async function request<T>(path: string, opts: RequestOptions = {}): Promise<T> {
  const method = opts.method ?? "GET";
  const url = new URL(
    `${getApiBaseUrl()}${path}`,
    window.location.origin
  );
  if (opts.query) {
    for (const [k, v] of Object.entries(opts.query)) {
      if (v !== undefined && v !== "") url.searchParams.set(k, v);
    }
  }

  const headers: Record<string, string> = {};
  if (opts.body !== undefined) headers["Content-Type"] = "application/json";
  if (method !== "GET" && method !== "HEAD") {
    const csrf = readCsrfCookie();
    if (csrf) headers["X-CSRF-Token"] = csrf;
  }

  const res = await fetch(url.toString(), {
    method,
    credentials: "include",
    headers,
    body: opts.body === undefined ? undefined : JSON.stringify(opts.body),
  });

  if (res.status === 204) return undefined as T;

  const text = await res.text();
  const data = text ? safeParse(text) : null;

  if (!res.ok) {
    const msg =
      (data && typeof data === "object" && "error" in data && typeof (data as { error: unknown }).error === "string"
        ? (data as { error: string }).error
        : null) ?? `Request failed (${res.status})`;
    throw new ApiError(msg, res.status);
  }

  return data as T;
}

function safeParse(text: string): unknown {
  try {
    return JSON.parse(text);
  } catch {
    return null;
  }
}

export type User = { id: string; email: string };
export type Category = { id: string; name: string };
export type Task = {
  id: string;
  title: string;
  description?: string;
  dueDate?: string | null;
  categoryId?: string | null;
  completed: boolean;
  createdAt?: string;
  updatedAt?: string;
};

export type TaskStatus = "all" | "active" | "completed";
export type TaskSort = "due_asc" | "due_desc";

export const api = {
  register: (email: string, password: string) =>
    request<User>("/auth/register", { method: "POST", body: { email, password } }),
  login: (email: string, password: string) =>
    request<User>("/auth/login", { method: "POST", body: { email, password } }),
  logout: () => request<void>("/auth/logout", { method: "POST" }),
  me: () => request<User>("/auth/me"),

  listCategories: () => request<Category[]>("/categories"),
  createCategory: (name: string) =>
    request<Category>("/categories", { method: "POST", body: { name } }),
  deleteCategory: (id: string) =>
    request<void>(`/categories/${encodeURIComponent(id)}`, { method: "DELETE" }),

  listTasks: (params: { category?: string; status?: TaskStatus; sort?: TaskSort }) =>
    request<Task[]>("/tasks", {
      query: {
        category: params.category,
        status: params.status,
        sort: params.sort,
      },
    }),
  createTask: (input: {
    title: string;
    description?: string;
    dueDate?: string | null;
    categoryId?: string | null;
  }) => request<Task>("/tasks", { method: "POST", body: input }),
  updateTask: (
    id: string,
    input: Partial<{
      title: string;
      description: string;
      dueDate: string | null;
      categoryId: string | null;
      completed: boolean;
    }>
  ) => request<Task>(`/tasks/${encodeURIComponent(id)}`, { method: "PUT", body: input }),
  deleteTask: (id: string) =>
    request<void>(`/tasks/${encodeURIComponent(id)}`, { method: "DELETE" }),
};
