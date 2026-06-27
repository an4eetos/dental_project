const DEBUG = import.meta.env.DEV || import.meta.env.VITE_DEBUG === "true";

/** Ensures absolute URL; without https:// fetch treats host as a path on the Mini App origin → HTTP 405. */
function normalizeApiBase(raw: string | undefined): string {
  const fallback = "http://localhost:8080";
  let base = (raw ?? fallback).trim();
  if (!base) {
    base = fallback;
  }
  if (!/^https?:\/\//i.test(base)) {
    base = `https://${base}`;
  }
  return base.replace(/\/+$/, "");
}

const API_BASE = normalizeApiBase(import.meta.env.VITE_API_BASE_URL);

export function apiBaseUrl(): string {
  return API_BASE;
}

export type UserRole = "patient" | "doctor" | "admin";

export type User = {
  id: number;
  telegram_id: number;
  role: UserRole;
  username?: string;
  first_name: string;
  last_name?: string;
  avatar_url?: string;
};

export type AuthResponse = {
  access_token: string;
  expires_at: string;
  user: User;
  subscription?: SubscriptionStatus;
};

export type SubscriptionStatus = {
  active: boolean;
  expires_at?: string;
  stars_price: number;
  duration_days: number;
};

export type CreateInvoiceResponse = {
  invoice_link: string;
};

export type Appointment = {
  id: number;
  preferred_date: string;
  preferred_time: string;
  status: string;
  visit_type?: "in_person" | "video";
  zoom_link?: string;
  doctor_notes?: string;
  created_at: string;
};

export type DoctorAppointment = Appointment & {
  needs_zoom_link?: boolean;
  patient: {
    id: number;
    telegram_id: number;
    username?: string;
    first_name: string;
    last_name?: string;
  };
};

export class ApiError extends Error {
  constructor(
    message: string,
    readonly status: number,
    readonly code?: string,
  ) {
    super(message);
  }
}

function fallbackMessage(status: number, rawText: string): string {
  if (status === 0) {
    return "Нет ответа от сервера (сеть, CORS или неверный URL API)";
  }
  if (status === 405) {
    return `Метод не разрешён (405). Укажите VITE_API_BASE_URL с https:// (сейчас: ${API_BASE})`;
  }
  if (status === 404) {
    return `API не найден (404). Проверьте VITE_API_BASE_URL: ${API_BASE}`;
  }
  if (status >= 502) {
    return `Сервер недоступен (${status}). Туннель или бэкенд не запущен?`;
  }
  if (rawText && rawText.trim().startsWith("<")) {
    return `Ответ не JSON (похоже на HTML). URL API: ${API_BASE}`;
  }
  return `Ошибка запроса (HTTP ${status})`;
}

async function request<T>(
  path: string,
  options: RequestInit & { token?: string } = {},
): Promise<T> {
  const url = `${API_BASE}${path}`;
  const headers: Record<string, string> = {
    "Content-Type": "application/json",
    ...(options.headers as Record<string, string>),
  };
  if (options.token) {
    headers.Authorization = `Bearer ${options.token}`;
  }

  let res: Response;
  try {
    res = await fetch(url, {
      ...options,
      headers,
    });
  } catch (err) {
    const detail = err instanceof Error ? err.message : String(err);
    if (DEBUG) {
      console.error("[api] fetch failed", { url, method: options.method ?? "GET", detail });
    }
    throw new ApiError(
      `Не удалось связаться с API: ${detail}. URL: ${API_BASE}`,
      0,
      "network_error",
    );
  }

  const rawText = await res.text();
  let body: Record<string, unknown> = {};
  if (rawText) {
    try {
      body = JSON.parse(rawText) as Record<string, unknown>;
    } catch {
      body = {};
    }
  }

  if (!res.ok) {
    const msg =
      typeof body.message === "string"
        ? body.message
        : fallbackMessage(res.status, rawText);
    const code = typeof body.code === "string" ? body.code : undefined;
    if (DEBUG) {
      console.error("[api] error response", {
        url,
        status: res.status,
        code,
        message: msg,
        body: rawText.slice(0, 500),
      });
    }
    throw new ApiError(msg, res.status, code);
  }

  if (DEBUG) {
    console.debug("[api] ok", { url, status: res.status });
  }
  return body as T;
}

export async function authTelegram(initData: string): Promise<AuthResponse> {
  return request<AuthResponse>("/auth/telegram", {
    method: "POST",
    body: JSON.stringify({ init_data: initData }),
  });
}

export async function createAppointment(
  token: string,
  preferredDate: string,
  preferredTime: string,
): Promise<Appointment> {
  return request<Appointment>("/appointments", {
    method: "POST",
    token,
    body: JSON.stringify({
      preferred_date: preferredDate,
      preferred_time: preferredTime,
    }),
  });
}

export async function listMyAppointments(
  token: string,
): Promise<{ appointments: Appointment[] }> {
  return request<{ appointments: Appointment[] }>("/appointments/me", {
    method: "GET",
    token,
  });
}

export async function listAllAppointments(
  token: string,
): Promise<{ appointments: DoctorAppointment[] }> {
  return request<{ appointments: DoctorAppointment[] }>("/appointments", {
    method: "GET",
    token,
  });
}

export type AppointmentDecision = "in_person" | "video" | "reject";

export async function respondAppointment(
  token: string,
  appointmentId: number,
  body: {
    decision: AppointmentDecision;
    preferred_date?: string;
    preferred_time?: string;
    zoom_link?: string;
    doctor_notes?: string;
  },
): Promise<DoctorAppointment> {
  return request<DoctorAppointment>(`/appointments/${appointmentId}/respond`, {
    method: "PATCH",
    token,
    body: JSON.stringify(body),
  });
}

export async function setAppointmentZoomLink(
  token: string,
  appointmentId: number,
  zoomLink: string,
): Promise<DoctorAppointment> {
  return request<DoctorAppointment>(`/appointments/${appointmentId}/zoom-link`, {
    method: "PATCH",
    token,
    body: JSON.stringify({ zoom_link: zoomLink }),
  });
}

export async function suggestAppointmentSlots(
  token: string,
  appointmentId: number,
): Promise<{ suggested_text: string }> {
  return request<{ suggested_text: string }>(
    `/appointments/${appointmentId}/suggest-slots`,
    {
      method: "POST",
      token,
    },
  );
}

export type PredictRequest = {
  age: string;
  pregnancy_weeks: string;
  kpu_index: string;
  brushing_per_day: string;
  dentist_visit_during_pregnancy: string;
  parent_caries: string;
  saliva_ph: string;
};

export type PredictResponse = {
  child_caries_probability: string;
  risk_group: string;
  action: string;
  recommendations: string;
};

export async function predict(
  token: string,
  body: PredictRequest,
): Promise<PredictResponse> {
  return request<PredictResponse>("/predict", {
    method: "POST",
    token,
    body: JSON.stringify(body),
  });
}

export type PhotoSubmissionPatient = {
  id: number;
  telegram_id: number;
  username?: string;
  first_name: string;
  last_name?: string;
};

export type PhotoSubmission = {
  id: number;
  status: "pending" | "answered";
  created_at: string;
  responded_at?: string;
  doctor_response?: string;
  ai_draft?: {
    visible_issues: string[];
    confidence: string;
    recommendations: string[];
  };
  patient: PhotoSubmissionPatient;
};

export async function listPendingSubmissions(
  token: string,
): Promise<{ submissions: PhotoSubmission[] }> {
  return request<{ submissions: PhotoSubmission[] }>("/submissions/pending", {
    method: "GET",
    token,
  });
}

export async function listAnsweredSubmissions(
  token: string,
): Promise<{ submissions: PhotoSubmission[] }> {
  return request<{ submissions: PhotoSubmission[] }>("/submissions/answered", {
    method: "GET",
    token,
  });
}

export async function getSubmission(
  token: string,
  id: number,
): Promise<PhotoSubmission> {
  return request<PhotoSubmission>(`/submissions/${id}`, {
    method: "GET",
    token,
  });
}

export async function generateSubmissionDraft(
  token: string,
  id: number,
): Promise<{ submission: PhotoSubmission; draft_text: string }> {
  return request<{ submission: PhotoSubmission; draft_text: string }>(
    `/submissions/${id}/draft`,
    { method: "POST", token },
  );
}

export async function respondToSubmission(
  token: string,
  id: number,
  response: string,
): Promise<PhotoSubmission> {
  return request<PhotoSubmission>(`/submissions/${id}/respond`, {
    method: "POST",
    token,
    body: JSON.stringify({ response }),
  });
}

export async function fetchSubmissionPhotoUrl(
  token: string,
  id: number,
): Promise<string> {
  const url = `${API_BASE}/submissions/${id}/photo`;
  let res: Response;
  try {
    res = await fetch(url, {
      headers: { Authorization: `Bearer ${token}` },
    });
  } catch (err) {
    const detail = err instanceof Error ? err.message : String(err);
    throw new ApiError(`Не удалось загрузить фото: ${detail}`, 0, "network_error");
  }
  if (!res.ok) {
    throw new ApiError("Не удалось загрузить фото", res.status);
  }
  const blob = await res.blob();
  return URL.createObjectURL(blob);
}

export type AdminStatistics = {
  total_users: number;
  total_patients: number;
  total_doctors: number;
  total_admins: number;
  total_photo_submissions: number;
  pending_photo_submissions: number;
  answered_photo_submissions: number;
  total_appointments: number;
  pending_appointments: number;
  confirmed_appointments: number;
  cancelled_appointments: number;
};

export async function getAdminStatistics(token: string): Promise<AdminStatistics> {
  return request<AdminStatistics>("/admin/statistics", {
    method: "GET",
    token,
  });
}

export type AdminUser = {
  id: number;
  telegram_id: number;
  role: UserRole;
  username?: string;
  first_name: string;
  last_name?: string;
  avatar_url?: string;
  blocked: boolean;
  created_at: string;
  updated_at: string;
};

export type AdminUserOverview = AdminUser & {
  appointment_count: number;
  photo_submission_count: number;
};

export type AdminUserListResponse = {
  users: AdminUser[];
};

export async function listAdminUsers(
  token: string,
  params: { search?: string; role?: UserRole; limit?: number; offset?: number } = {},
): Promise<AdminUserListResponse> {
  const query = new URLSearchParams();
  if (params.search) query.set("search", params.search);
  if (params.role) query.set("role", params.role);
  if (params.limit != null) query.set("limit", String(params.limit));
  if (params.offset != null) query.set("offset", String(params.offset));
  const qs = query.toString();
  return request<AdminUserListResponse>(`/admin/users${qs ? `?${qs}` : ""}`, {
    method: "GET",
    token,
  });
}

export async function getAdminUser(token: string, userId: number): Promise<AdminUserOverview> {
  return request<AdminUserOverview>(`/admin/users/${userId}`, {
    method: "GET",
    token,
  });
}

export async function updateAdminUser(
  token: string,
  userId: number,
  body: {
    first_name?: string;
    last_name?: string;
    username?: string;
    role?: UserRole;
  },
): Promise<AdminUser> {
  return request<AdminUser>(`/admin/users/${userId}`, {
    method: "PATCH",
    token,
    body: JSON.stringify(body),
  });
}

export async function setAdminUserBlocked(
  token: string,
  userId: number,
  blocked: boolean,
): Promise<AdminUser> {
  return request<AdminUser>(`/admin/users/${userId}/block`, {
    method: "PATCH",
    token,
    body: JSON.stringify({ blocked }),
  });
}

export async function getSubscriptionStatus(token: string): Promise<SubscriptionStatus> {
  return request<SubscriptionStatus>("/subscription/me", {
    method: "GET",
    token,
  });
}

export async function createSubscriptionInvoice(
  token: string,
): Promise<CreateInvoiceResponse> {
  return request<CreateInvoiceResponse>("/subscription/invoice", {
    method: "POST",
    token,
    body: JSON.stringify({}),
  });
}

export type ContentBlock = {
  type: "text" | "youtube" | "image" | "video";
  data: Record<string, unknown>;
};

export type ContentItem = {
  id: number;
  title: string;
  description?: string;
  access: "public" | "subscription";
  locked: boolean;
  blocks: ContentBlock[];
  sort_order?: number;
};

export type ContentListResponse = {
  items: ContentItem[];
};

export type AdminContentItem = {
  id: number;
  title: string;
  description?: string;
  access: "public" | "subscription";
  published: boolean;
  sort_order: number;
  blocks: ContentBlock[];
};

export type AdminContentListResponse = {
  items: AdminContentItem[];
};

export type SaveContentPayload = {
  title: string;
  description?: string;
  access: "public" | "subscription";
  published: boolean;
  blocks: ContentBlock[];
};

export async function listContent(token: string): Promise<ContentListResponse> {
  return request<ContentListResponse>("/content", {
    method: "GET",
    token,
  });
}

export async function fetchContentMediaUrl(token: string, mediaId: number): Promise<string> {
  const url = `${API_BASE}/content/media/${mediaId}`;
  let res: Response;
  try {
    res = await fetch(url, {
      headers: { Authorization: `Bearer ${token}` },
    });
  } catch (err) {
    const detail = err instanceof Error ? err.message : String(err);
    throw new ApiError(`Не удалось загрузить медиа: ${detail}`, 0, "network_error");
  }
  if (!res.ok) {
    throw new ApiError("Не удалось загрузить медиа", res.status);
  }
  const blob = await res.blob();
  return URL.createObjectURL(blob);
}

export async function listAdminContent(token: string): Promise<AdminContentListResponse> {
  return request<AdminContentListResponse>("/admin/content", {
    method: "GET",
    token,
  });
}

export async function getAdminContent(token: string, id: number): Promise<AdminContentItem> {
  return request<AdminContentItem>(`/admin/content/${id}`, {
    method: "GET",
    token,
  });
}

export async function createAdminContent(
  token: string,
  body: SaveContentPayload,
): Promise<AdminContentItem> {
  return request<AdminContentItem>("/admin/content", {
    method: "POST",
    token,
    body: JSON.stringify(body),
  });
}

export async function updateAdminContent(
  token: string,
  id: number,
  body: SaveContentPayload,
): Promise<AdminContentItem> {
  return request<AdminContentItem>(`/admin/content/${id}`, {
    method: "PUT",
    token,
    body: JSON.stringify(body),
  });
}

export async function deleteAdminContent(token: string, id: number): Promise<void> {
  await request<void>(`/admin/content/${id}`, {
    method: "DELETE",
    token,
  });
}

export async function reorderAdminContent(token: string, ids: number[]): Promise<void> {
  await request<void>("/admin/content/reorder", {
    method: "PATCH",
    token,
    body: JSON.stringify({ ids }),
  });
}

export async function uploadAdminContentMedia(
  token: string,
  file: File,
  contentItemId?: number,
): Promise<{ media_id: number }> {
  const form = new FormData();
  form.append("file", file);
  if (contentItemId != null) {
    form.append("content_item_id", String(contentItemId));
  }

  const url = `${API_BASE}/admin/content/media`;
  let res: Response;
  try {
    res = await fetch(url, {
      method: "POST",
      headers: { Authorization: `Bearer ${token}` },
      body: form,
    });
  } catch (err) {
    const detail = err instanceof Error ? err.message : String(err);
    throw new ApiError(`Не удалось загрузить файл: ${detail}`, 0, "network_error");
  }

  const text = await res.text();
  if (!res.ok) {
    let message = fallbackMessage(res.status, text);
    try {
      const parsed = JSON.parse(text) as { message?: string };
      if (parsed.message) message = parsed.message;
    } catch {
      // keep fallback
    }
    throw new ApiError(message, res.status);
  }
  return JSON.parse(text) as { media_id: number };
}
