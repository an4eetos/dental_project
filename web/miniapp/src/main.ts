import {
  ApiError,
  authTelegram,
  createAppointment,
  listAllAppointments,
  listMyAppointments,
  type AuthResponse,
  type DoctorAppointment,
} from "./api";
import { clearToken, loadToken, saveToken } from "./storage";

const tg = window.Telegram?.WebApp;

const appRoot = document.getElementById("app");
if (!appRoot) {
  throw new Error("root element #app not found");
}
const app: HTMLElement = appRoot;

function el(tag: string, className?: string, text?: string): HTMLElement {
  const node = document.createElement(tag);
  if (className) node.className = className;
  if (text) node.textContent = text;
  return node;
}

function minDate(): string {
  const d = new Date();
  const y = d.getFullYear();
  const m = String(d.getMonth() + 1).padStart(2, "0");
  const day = String(d.getDate()).padStart(2, "0");
  return `${y}-${m}-${day}`;
}

function maxDate(): string {
  const d = new Date();
  d.setDate(d.getDate() + 90);
  const y = d.getFullYear();
  const m = String(d.getMonth() + 1).padStart(2, "0");
  const day = String(d.getDate()).padStart(2, "0");
  return `${y}-${m}-${day}`;
}

function statusLabel(status: string): string {
  switch (status) {
    case "pending":
      return "Ожидает подтверждения";
    case "confirmed":
      return "Подтверждена";
    case "cancelled":
      return "Отменена";
    default:
      return status;
  }
}

function applyTheme(): void {
  const scheme = tg?.colorScheme ?? "light";
  document.body.dataset.theme = scheme;
  const bg = tg?.themeParams.bg_color;
  if (bg) document.body.style.background = bg;
}

function renderShell(content: HTMLElement, title: string, subtitle: string): void {
  app.replaceChildren();
  const header = el("header", "header");
  header.append(el("h1", "title", title), el("p", "subtitle", subtitle));
  app.append(header, content);
}

function renderLoading(message: string): void {
  renderShell(
    el("p", "status", message),
    "Запись на приём",
    "Стоматологический AI-ассистент",
  );
}

function renderError(message: string): void {
  const box = el("div", "card error");
  box.append(el("p", undefined, message));
  renderShell(box, "Запись на приём", "Стоматологический AI-ассистент");
}

function patientName(p: DoctorAppointment["patient"]): string {
  const parts = [p.first_name, p.last_name].filter(Boolean);
  const base = parts.join(" ").trim();
  return p.username ? `${base} (@${p.username})` : base;
}

function renderAppointmentsList(
  items: Awaited<ReturnType<typeof listMyAppointments>>["appointments"],
): HTMLElement {
  const list = el("section", "list");
  if (items.length === 0) {
    list.append(el("p", "muted", "У вас пока нет записей."));
    return list;
  }
  for (const a of items) {
    const card = el("article", "card item");
    card.append(
      el("p", "item-title", `${a.preferred_date} в ${a.preferred_time}`),
      el("p", "muted", statusLabel(a.status)),
    );
    list.append(card);
  }
  return list;
}

function renderBookingForm(token: string): void {
  const form = el("form", "card form");
  const dateLabel = el("label", undefined, "Предпочтительная дата");
  const dateInput = document.createElement("input");
  dateInput.type = "date";
  dateInput.required = true;
  dateInput.min = minDate();
  dateInput.max = maxDate();
  dateInput.className = "input";

  const timeLabel = el("label", undefined, "Предпочтительное время");
  const timeInput = document.createElement("input");
  timeInput.type = "time";
  timeInput.required = true;
  timeInput.min = "09:00";
  timeInput.max = "20:00";
  timeInput.step = "60";
  timeInput.className = "input";

  const submit = document.createElement("button");
  submit.type = "submit";
  submit.className = "button primary";
  submit.textContent = "Отправить заявку";

  const status = el("p", "status hidden");

  form.append(dateLabel, dateInput, timeLabel, timeInput, submit, status);

  const listTitle = el("h2", "section-title", "Мои записи");
  const listHost = el("div", "list-host");

  form.addEventListener("submit", async (e) => {
    e.preventDefault();
    status.className = "status";
    status.textContent = "Отправка…";
    submit.setAttribute("disabled", "true");

    try {
      await createAppointment(token, dateInput.value, timeInput.value);
      status.textContent = "Заявка отправлена. Администратор свяжется с вами.";
      const data = await listMyAppointments(token);
      listHost.replaceChildren(renderAppointmentsList(data.appointments));
      dateInput.value = "";
      timeInput.value = "";
    } catch (err) {
      status.textContent =
        err instanceof ApiError ? err.message : "Не удалось отправить заявку";
    } finally {
      submit.removeAttribute("disabled");
    }
  });

  const shell = el("div", "stack");
  shell.append(form, listTitle, listHost);
  renderShell(shell, "Запись на приём", "Выберите удобные дату и время");

  listMyAppointments(token)
    .then((data) => {
      listHost.append(renderAppointmentsList(data.appointments));
    })
    .catch(() => {
      listHost.append(el("p", "muted", "Не удалось загрузить список записей."));
    });
}

function renderDoctorDashboard(token: string): void {
  const listHost = el("div", "list-host");
  const refresh = document.createElement("button");
  refresh.type = "button";
  refresh.className = "button primary";
  refresh.textContent = "Обновить";

  const load = async () => {
    listHost.replaceChildren(el("p", "status", "Загрузка…"));
    try {
      const data = await listAllAppointments(token);
      listHost.replaceChildren(renderDoctorAppointmentsList(data.appointments));
    } catch (err) {
      listHost.replaceChildren(
        el(
          "p",
          "muted",
          err instanceof ApiError ? err.message : "Не удалось загрузить записи",
        ),
      );
    }
  };

  refresh.addEventListener("click", () => void load());

  const shell = el("div", "stack");
  shell.append(refresh, listHost);
  renderShell(shell, "Записи пациентов", "Все заявки на приём");

  void load();
}

function renderDoctorAppointmentsList(items: DoctorAppointment[]): HTMLElement {
  const list = el("section", "list");
  if (items.length === 0) {
    list.append(el("p", "muted", "Записей пока нет."));
    return list;
  }
  for (const a of items) {
    const card = el("article", "card item");
    card.append(
      el("p", "item-title", `${a.preferred_date} в ${a.preferred_time}`),
      el("p", undefined, patientName(a.patient)),
      el("p", "muted", statusLabel(a.status)),
    );
    list.append(card);
  }
  return list;
}

async function ensureAuth(): Promise<AuthResponse> {
  if (tg?.initData) {
    const res = await authTelegram(tg.initData);
    saveToken(res.access_token);
    return res;
  }

  const cached = loadToken();
  if (cached) {
    return {
      access_token: cached,
      expires_at: "",
      user: { id: 0, telegram_id: 0, role: "patient", first_name: "" },
    };
  }

  throw new Error(
    "Откройте приложение из Telegram. Для локальной разработки нужен реальный initData.",
  );
}

async function bootstrap(): Promise<void> {
  applyTheme();
  tg?.ready();
  tg?.expand();

  renderLoading("Авторизация…");

  try {
    const auth = await ensureAuth();
    if (auth.user.role === "doctor") {
      renderDoctorDashboard(auth.access_token);
      return;
    }
    renderBookingForm(auth.access_token);
  } catch (err) {
    if (err instanceof ApiError) {
      clearToken();
      renderError(err.message);
      return;
    }
    renderError(err instanceof Error ? err.message : "Ошибка авторизации");
  }
}

bootstrap();
