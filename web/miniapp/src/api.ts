const API_BASE = import.meta.env.VITE_API_BASE_URL ?? "http://localhost:8080";

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

async function request<T>(
  path: string,
  options: RequestInit & { token?: string } = {},
): Promise<T> {
  const headers: Record<string, string> = {
    "Content-Type": "application/json",
    ...(options.headers as Record<string, string>),
  };
  if (options.token) {
    headers.Authorization = `Bearer ${options.token}`;
  }

  const res = await fetch(`${API_BASE}${path}`, {
    ...options,
    headers,
  });

  const body = await res.json().catch(() => ({}));
  if (!res.ok) {
    const msg =
      typeof body.message === "string" ? body.message : "Ошибка запроса";
    throw new ApiError(msg, res.status, body.code);
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
