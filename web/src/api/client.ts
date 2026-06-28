export const API_BASE_URL =
  import.meta.env.VITE_API_BASE_URL?.replace(/\/+$/, "") ||
  "http://127.0.0.1:8080";

export interface ApiResponse<T> {
  ok: boolean;
  message?: string;
  error?: string;
  data?: T;
}

export interface Disk {
  name: string;
  path: string;
  size: number;
}

export interface Partition {
  name: string;
  type: string;
  fit: string;
  start: number;
  size: number;
  status: string;
}

export interface MountedPartition {
  id: string;
  path: string;
  name: string;
  partitionType: string;
  start: number;
  size: number;
}

export interface Session {
  active: boolean;
  mountedId: string;
  user: string;
  uid: number;
  group: string;
  gid: number;
  diskPath: string;
  partitionName: string;
}

export interface FSItem {
  name: string;
  path: string;
  type: "file" | "directory";
  size: number;
  inode: number;
  permissions: string;
  owner: string;
  group: string;
}

export interface DirectoryListing {
  id: string;
  path: string;
  items: FSItem[];
}

export interface FileContent {
  id: string;
  path: string;
  name: string;
  content: string;
  size: number;
}

export interface Metadata extends FSItem {
  id: string;
}

export interface ReportResult {
  name: string;
  path: string;
  url: string;
  contentType: string;
}

export interface CopyResult {
  warnings?: string[];
}

export interface CommandExecution {
  command: string;
  output: string;
  session?: Session;
}

async function request<T>(
  route: string,
  options: RequestInit = {},
): Promise<ApiResponse<T>> {
  try {
    const response = await fetch(`${API_BASE_URL}${route}`, {
      ...options,
      headers: {
        "Content-Type": "application/json",
        ...options.headers,
      },
    });
    const payload = (await response.json()) as ApiResponse<T>;
    if (!response.ok || !payload.ok) {
      throw new Error(payload.error || `Error HTTP ${response.status}`);
    }
    return payload;
  } catch (error) {
    if (error instanceof Error) {
      throw error;
    }
    throw new Error("No fue posible comunicarse con la API");
  }
}

export const api = {
  health: () => request<null>("/api/health"),
  mounts: () => request<MountedPartition[]>("/api/mounts"),
  login: (body: { id: string; user: string; password: string }) =>
    request<Session>("/api/login", {
      method: "POST",
      body: JSON.stringify(body),
    }),
  logout: () => request<null>("/api/logout", { method: "POST" }),
  disks: () => request<Disk[]>("/api/disks"),
  createDisk: (body: {
    path: string;
    size: number;
    unit: string;
    fit: string;
  }) =>
    request<Disk>("/api/disks", {
      method: "POST",
      body: JSON.stringify(body),
    }),
  deleteDisk: (path: string) =>
    request<null>("/api/disks", {
      method: "DELETE",
      body: JSON.stringify({ path }),
    }),
  partitions: (path: string) =>
    request<Partition[]>(
      `/api/partitions?path=${encodeURIComponent(path)}`,
    ),
  createPartition: (body: {
    path: string;
    name: string;
    size: number;
    unit: string;
    type: string;
    fit: string;
  }) =>
    request<null>("/api/partitions", {
      method: "POST",
      body: JSON.stringify(body),
    }),
  resizePartition: (body: {
    path: string;
    name: string;
    add: number;
    unit: string;
  }) =>
    request<null>("/api/partitions/resize", {
      method: "PATCH",
      body: JSON.stringify(body),
    }),
  deletePartition: (body: {
    path: string;
    name: string;
    delete: string;
  }) =>
    request<null>("/api/partitions", {
      method: "DELETE",
      body: JSON.stringify(body),
    }),
  mount: (path: string, name: string) =>
    request<MountedPartition>("/api/mount", {
      method: "POST",
      body: JSON.stringify({ path, name }),
    }),
  unmount: (id: string) =>
    request<null>("/api/unmount", {
      method: "POST",
      body: JSON.stringify({ id }),
    }),
  mkfs: (id: string) =>
    request<null>("/api/mkfs", {
      method: "POST",
      body: JSON.stringify({ id, type: "full" }),
    }),
  listFS: (id: string, path: string) =>
    request<DirectoryListing>(
      `/api/fs/list?id=${encodeURIComponent(id)}&path=${encodeURIComponent(path)}`,
    ),
  readFS: (id: string, path: string) =>
    request<FileContent>(
      `/api/fs/read?id=${encodeURIComponent(id)}&path=${encodeURIComponent(path)}`,
    ),
  statFS: (id: string, path: string) =>
    request<Metadata>(
      `/api/fs/stat?id=${encodeURIComponent(id)}&path=${encodeURIComponent(path)}`,
    ),
  editFile: (path: string, contenido: string) =>
    request<null>("/api/fs/edit", {
      method: "PATCH",
      body: JSON.stringify({ path, contenido }),
    }),
  renamePath: (path: string, name: string) =>
    request<null>("/api/fs/rename", {
      method: "PATCH",
      body: JSON.stringify({ path, name }),
    }),
  removePath: (path: string) =>
    request<null>("/api/fs/remove", {
      method: "DELETE",
      body: JSON.stringify({ path }),
    }),
  copyPath: (path: string, destino: string) =>
    request<CopyResult>("/api/fs/copy", {
      method: "POST",
      body: JSON.stringify({ path, destino }),
    }),
  movePath: (path: string, destino: string) =>
    request<null>("/api/fs/move", {
      method: "PATCH",
      body: JSON.stringify({ path, destino }),
    }),
  executeCommand: (command: string) =>
    request<CommandExecution>("/api/commands/execute", {
      method: "POST",
      body: JSON.stringify({ command }),
    }),
  report: (body: {
    id: string;
    name: string;
    pathFileLs?: string;
    format: string;
  }) =>
    request<ReportResult>("/api/reports", {
      method: "POST",
      body: JSON.stringify(body),
    }),
};
