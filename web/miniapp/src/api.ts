const API_BASE = import.meta.env.VITE_API_BASE_URL ?? "http://localhost:8080";
const DEBUG = import.meta.env.DEV || import.meta.env.VITE_DEBUG === "true";

export function apiBaseUrl(): string {
  return API_BASE;
}

export type UserRole = "patient" | "doctor";

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
};

export type Appointment = {
  id: number;
  preferred_date: string;
  preferred_time: string;
  status: string;
  created_at: string;
};

export type DoctorAppointment = Appointment & {
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
