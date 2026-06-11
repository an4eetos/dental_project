import {
  ApiError,
  apiBaseUrl,
  authTelegram,
  createAppointment,
  fetchSubmissionPhotoUrl,
  generateSubmissionDraft,
  getAdminStatistics,
  getSubmission,
  listAllAppointments,
  listAnsweredSubmissions,
  listMyAppointments,
  listPendingSubmissions,
  predict,
  respondToSubmission,
  type AdminStatistics,
  type AuthResponse,
  type DoctorAppointment,
  type PhotoSubmission,
  type PredictRequest,
} from "./api";
import {
  PREDICTION_INPUT_FIELDS,
  PREDICTION_OUTPUT_FIELDS,
  type PredictionInputKey,
} from "./prediction-config";
import { clearToken, loadToken, saveToken } from "./storage";
import {
  YOUTUBE_VIDEOS,
  youtubeEmbedUrl,
  youtubeThumbnailUrl,
  youtubeWatchUrl,
  type YouTubeVideo,
} from "./videos-config";

const tg = window.Telegram?.WebApp;

const appRoot = document.getElementById("app");
if (!appRoot) {
  throw new Error("root element #app not found");
}
const app: HTMLElement = appRoot;

type PatientTab = "booking" | "prediction" | "videos";

const PATIENT_TABS: {
  id: PatientTab;
  label: string;
  title: string;
  subtitle: string;
}[] = [
  {
    id: "booking",
    label: "Запись",
    title: "Запись на приём",
    subtitle: "Выберите удобные дату и время",
  },
  {
    id: "prediction",
    label: "Прогноз",
    title: "Прогноз",
    subtitle: "Заполните параметры для прогноза",
  },
  {
    id: "videos",
    label: "Видео",
    title: "Обучающие видео",
    subtitle: "Материалы о здоровье зубов",
  },
];

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

function renderError(message: string, hint?: string): void {
  const box = el("div", "card error");
  box.append(el("p", undefined, message));
  if (hint) {
    box.append(el("p", "muted", hint));
  }
  renderShell(box, "Запись на приём", "Стоматологический AI-ассистент");
}

function patientName(p: {
  first_name: string;
  last_name?: string;
  username?: string;
}): string {
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

function buildBookingContent(token: string): HTMLElement {
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

  listMyAppointments(token)
    .then((data) => {
      listHost.append(renderAppointmentsList(data.appointments));
    })
    .catch(() => {
      listHost.append(el("p", "muted", "Не удалось загрузить список записей."));
    });

  return shell;
}

function buildPredictionContent(token: string): HTMLElement {
  const form = el("form", "card form");
  const inputs = {} as Record<PredictionInputKey, HTMLInputElement>;

  for (const field of PREDICTION_INPUT_FIELDS) {
    const label = el("label", undefined, field.label);
    const input = document.createElement("input");
    input.type = "text";
    input.required = true;
    input.className = "input";
    input.name = field.key;
    inputs[field.key] = input;
    form.append(label, input);
  }

  const submit = document.createElement("button");
  submit.type = "submit";
  submit.className = "button primary";
  submit.textContent = "Получить результат";

  const status = el("p", "status hidden");
  const resultHost = el("div", "result-host hidden");

  form.append(submit, status, resultHost);

  form.addEventListener("submit", async (e) => {
    e.preventDefault();
    status.className = "status";
    status.textContent = "Генерация…";
    resultHost.className = "result-host hidden";
    resultHost.replaceChildren();
    submit.setAttribute("disabled", "true");

    const body = {} as PredictRequest;
    for (const field of PREDICTION_INPUT_FIELDS) {
      body[field.key] = inputs[field.key].value.trim();
    }

    try {
      const res = await predict(token, body);
      status.className = "status hidden";
      resultHost.className = "result-host";
      const resultCard = el("article", "card result");
      resultCard.append(el("h2", "section-title", "Результат"));
      for (const field of PREDICTION_OUTPUT_FIELDS) {
        const row = el("div", "result-row");
        row.append(
          el("p", "result-label", field.label),
          el("p", "result-output", res[field.key]),
        );
        resultCard.append(row);
      }
      resultHost.append(resultCard);
    } catch (err) {
      status.textContent =
        err instanceof ApiError ? err.message : "Не удалось получить прогноз";
    } finally {
      submit.removeAttribute("disabled");
    }
  });

  const stack = el("div", "stack");
  stack.append(form);
  return stack;
}

function buildVideoCard(video: YouTubeVideo): HTMLElement {
  const card = el("article", "card video-card");
  card.append(el("h3", "video-title", video.title));
  if (video.description) {
    card.append(el("p", "muted video-desc", video.description));
  }

  const playerHost = el("div", "video-player");
  const playBtn = document.createElement("button");
  playBtn.type = "button";
  playBtn.className = "video-placeholder";
  playBtn.setAttribute("aria-label", `Смотреть: ${video.title}`);

  const thumb = document.createElement("img");
  thumb.src = youtubeThumbnailUrl(video.youtubeId);
  thumb.alt = "";
  thumb.className = "video-thumb";
  thumb.loading = "lazy";

  const playIcon = el("span", "video-play", "▶");
  playBtn.append(thumb, playIcon);

  playBtn.addEventListener("click", () => {
    const iframe = document.createElement("iframe");
    iframe.src = youtubeEmbedUrl(video.youtubeId);
    iframe.title = video.title;
    iframe.className = "video-embed";
    iframe.setAttribute(
      "allow",
      "accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share",
    );
    iframe.allowFullscreen = true;
    playerHost.replaceChildren(iframe);
  });

  playerHost.append(playBtn);
  card.append(playerHost);

  const openExternal = document.createElement("button");
  openExternal.type = "button";
  openExternal.className = "button video-open";
  openExternal.textContent = "Открыть в YouTube";
  openExternal.addEventListener("click", () => {
    const url = youtubeWatchUrl(video.youtubeId);
    if (tg?.openLink) {
      tg.openLink(url);
    } else {
      window.open(url, "_blank", "noopener,noreferrer");
    }
  });
  card.append(openExternal);

  return card;
}

function buildVideosContent(): HTMLElement {
  const stack = el("div", "stack");

  if (YOUTUBE_VIDEOS.length === 0) {
    stack.append(
      el("p", "muted", "Видео скоро появятся. Добавьте ссылки в videos-config.ts."),
    );
    return stack;
  }

  const list = el("section", "list video-list");
  for (const video of YOUTUBE_VIDEOS) {
    list.append(buildVideoCard(video));
  }
  stack.append(list);
  return stack;
}

function buildPatientTabContent(token: string, tab: PatientTab): HTMLElement {
  switch (tab) {
    case "booking":
      return buildBookingContent(token);
    case "prediction":
      return buildPredictionContent(token);
    case "videos":
      return buildVideosContent();
  }
}

function renderPatientShell(token: string, activeTab: PatientTab): void {
  const tabs = el("nav", "tabs");

  for (const tab of PATIENT_TABS) {
    const btn = document.createElement("button");
    btn.type = "button";
    btn.className = activeTab === tab.id ? "tab active" : "tab";
    btn.textContent = tab.label;
    btn.addEventListener("click", () => renderPatientShell(token, tab.id));
    tabs.append(btn);
  }

  const contentHost = el("div", "tab-content");
  contentHost.append(buildPatientTabContent(token, activeTab));

  const shell = el("div", "stack");
  shell.append(tabs, contentHost);

  const meta = PATIENT_TABS.find((t) => t.id === activeTab) ?? PATIENT_TABS[0];
  renderShell(shell, meta.title, meta.subtitle);
}

type DoctorTab = "appointments" | "pending" | "answered";

const DOCTOR_TABS: { id: DoctorTab; label: string; title: string; subtitle: string }[] = [
  {
    id: "appointments",
    label: "Записи",
    title: "Записи пациентов",
    subtitle: "Все заявки на приём",
  },
  {
    id: "pending",
    label: "Ожидают",
    title: "Фото без ответа",
    subtitle: "Ответьте в течение 48 часов",
  },
  {
    id: "answered",
    label: "Отвеченные",
    title: "Отвеченные фото",
    subtitle: "История ответов врача",
  },
];

function formatDateTime(iso: string): string {
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return iso;
  return d.toLocaleString("ru-RU", {
    day: "2-digit",
    month: "2-digit",
    year: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  });
}

function renderDoctorDashboard(token: string, activeTab: DoctorTab = "pending"): void {
  const tabs = el("nav", "tabs");
  for (const tab of DOCTOR_TABS) {
    const btn = document.createElement("button");
    btn.type = "button";
    btn.className = activeTab === tab.id ? "tab active" : "tab";
    btn.textContent = tab.label;
    btn.addEventListener("click", () => renderDoctorDashboard(token, tab.id));
    tabs.append(btn);
  }

  const contentHost = el("div", "tab-content");
  switch (activeTab) {
    case "appointments":
      contentHost.append(buildDoctorAppointmentsContent(token));
      break;
    case "pending":
      contentHost.append(buildDoctorSubmissionQueue(token, "pending"));
      break;
    case "answered":
      contentHost.append(buildDoctorSubmissionQueue(token, "answered"));
      break;
  }

  const shell = el("div", "stack");
  shell.append(tabs, contentHost);
  const meta = DOCTOR_TABS.find((t) => t.id === activeTab) ?? DOCTOR_TABS[1];
  renderShell(shell, meta.title, meta.subtitle);
}

function buildDoctorAppointmentsContent(token: string): HTMLElement {
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
  void load();
  return shell;
}

function buildDoctorSubmissionQueue(
  token: string,
  queue: "pending" | "answered",
): HTMLElement {
  const listHost = el("div", "list-host");
  const refresh = document.createElement("button");
  refresh.type = "button";
  refresh.className = "button primary";
  refresh.textContent = "Обновить";

  const load = async () => {
    listHost.replaceChildren(el("p", "status", "Загрузка…"));
    try {
      const data =
        queue === "pending"
          ? await listPendingSubmissions(token)
          : await listAnsweredSubmissions(token);
      listHost.replaceChildren(
        renderSubmissionList(token, data.submissions, queue, () => void load()),
      );
    } catch (err) {
      listHost.replaceChildren(
        el(
          "p",
          "muted",
          err instanceof ApiError ? err.message : "Не удалось загрузить фото",
        ),
      );
    }
  };

  refresh.addEventListener("click", () => void load());
  const shell = el("div", "stack");
  shell.append(refresh, listHost);
  void load();
  return shell;
}

function renderSubmissionList(
  token: string,
  items: PhotoSubmission[],
  queue: "pending" | "answered",
  onBack: () => void,
): HTMLElement {
  const list = el("section", "list");
  if (items.length === 0) {
    list.append(
      el(
        "p",
        "muted",
        queue === "pending" ? "Нет фото, ожидающих ответа." : "Отвеченных фото пока нет.",
      ),
    );
    return list;
  }

  for (const item of items) {
    const card = el("article", "card item submission-item");
    card.append(
      el("p", "item-title", patientName(item.patient)),
      el("p", "muted", formatDateTime(item.created_at)),
    );
    if (queue === "answered" && item.responded_at) {
      card.append(el("p", "muted", `Ответ: ${formatDateTime(item.responded_at)}`));
    }
    card.addEventListener("click", () => {
      renderSubmissionDetail(token, item.id, queue, onBack);
    });
    list.append(card);
  }
  return list;
}

function renderSubmissionDetail(
  token: string,
  submissionId: number,
  queue: "pending" | "answered",
  onBack: () => void,
): void {
  const shell = el("div", "stack");
  shell.append(el("p", "status", "Загрузка…"));
  renderShell(shell, "Просмотр фото", "Детали заявки");

  void (async () => {
    try {
      const submission = await getSubmission(token, submissionId);
      const photoUrl = await fetchSubmissionPhotoUrl(token, submissionId);
      shell.replaceChildren(
        buildSubmissionDetailContent(token, submission, photoUrl, queue, onBack),
      );
    } catch (err) {
      shell.replaceChildren(
        el(
          "p",
          "muted",
          err instanceof ApiError ? err.message : "Не удалось загрузить заявку",
        ),
      );
    }
  })();
}

function buildSubmissionDetailContent(
  token: string,
  submission: PhotoSubmission,
  photoUrl: string,
  queue: "pending" | "answered",
  onBack: () => void,
): HTMLElement {
  const shell = el("div", "stack");

  const back = document.createElement("button");
  back.type = "button";
  back.className = "button";
  back.textContent = "← Назад к списку";
  back.addEventListener("click", onBack);
  shell.append(back);

  const card = el("article", "card");
  card.append(
    el("p", "item-title", patientName(submission.patient)),
    el("p", "muted", `Получено: ${formatDateTime(submission.created_at)}`),
  );

  const img = document.createElement("img");
  img.src = photoUrl;
  img.alt = "Фото пациента";
  img.className = "submission-photo";
  card.append(img);
  shell.append(card);

  if (queue === "answered") {
    const answerCard = el("article", "card");
    answerCard.append(el("h2", "section-title", "Ответ врача"));
    answerCard.append(
      el("p", "result-output", submission.doctor_response ?? "—"),
    );
    if (submission.responded_at) {
      answerCard.append(
        el("p", "muted", `Отправлено: ${formatDateTime(submission.responded_at)}`),
      );
    }
    shell.append(answerCard);
    return shell;
  }

  const form = el("form", "card form");
  const label = el("label", undefined, "Ответ пациенту");
  const textarea = document.createElement("textarea");
  textarea.className = "textarea";
  textarea.rows = 10;
  textarea.required = true;
  textarea.placeholder = "Напишите ответ или сгенерируйте черновик ИИ и отредактируйте";

  const draftBtn = document.createElement("button");
  draftBtn.type = "button";
  draftBtn.className = "button";
  draftBtn.textContent = "Сгенерировать черновик ИИ";

  const submit = document.createElement("button");
  submit.type = "submit";
  submit.className = "button primary";
  submit.textContent = "Отправить пациенту";

  const status = el("p", "status hidden");
  form.append(label, textarea, draftBtn, submit, status);

  draftBtn.addEventListener("click", async () => {
    draftBtn.setAttribute("disabled", "true");
    status.className = "status";
    status.textContent = "Генерация черновика…";
    try {
      const res = await generateSubmissionDraft(token, submission.id);
      textarea.value = res.draft_text;
      status.textContent = "Черновик готов. Проверьте и отредактируйте перед отправкой.";
    } catch (err) {
      status.textContent =
        err instanceof ApiError ? err.message : "Не удалось сгенерировать черновик";
    } finally {
      draftBtn.removeAttribute("disabled");
    }
  });

  form.addEventListener("submit", async (e) => {
    e.preventDefault();
    submit.setAttribute("disabled", "true");
    status.className = "status";
    status.textContent = "Отправка…";
    try {
      await respondToSubmission(token, submission.id, textarea.value.trim());
      status.textContent = "Ответ отправлен пациенту.";
      setTimeout(onBack, 800);
    } catch (err) {
      status.textContent =
        err instanceof ApiError ? err.message : "Не удалось отправить ответ";
      submit.removeAttribute("disabled");
    }
  });

  shell.append(form);
  return shell;
}

function renderAdminDashboard(token: string): void {
  const statsHost = el("div", "stats-grid");
  const refresh = document.createElement("button");
  refresh.type = "button";
  refresh.className = "button primary";
  refresh.textContent = "Обновить";

  const load = async () => {
    statsHost.replaceChildren(el("p", "status", "Загрузка…"));
    try {
      const stats = await getAdminStatistics(token);
      statsHost.replaceChildren(renderAdminStats(stats));
    } catch (err) {
      statsHost.replaceChildren(
        el(
          "p",
          "muted",
          err instanceof ApiError ? err.message : "Не удалось загрузить статистику",
        ),
      );
    }
  };

  refresh.addEventListener("click", () => void load());
  const shell = el("div", "stack");
  shell.append(refresh, statsHost);
  renderShell(shell, "Статистика клиники", "Панель администратора");
  void load();
}

function renderAdminStats(stats: AdminStatistics): HTMLElement {
  const grid = el("div", "stats-grid");
  const items: [string, number][] = [
    ["Всего пользователей", stats.total_users],
    ["Пациентов", stats.total_patients],
    ["Врачей", stats.total_doctors],
    ["Администраторов", stats.total_admins],
    ["Всего фото-заявок", stats.total_photo_submissions],
    ["Ожидают ответа", stats.pending_photo_submissions],
    ["Отвечено", stats.answered_photo_submissions],
    ["Всего записей", stats.total_appointments],
    ["Записей: ожидают", stats.pending_appointments],
    ["Записей: подтверждены", stats.confirmed_appointments],
    ["Записей: отменены", stats.cancelled_appointments],
  ];

  for (const [label, value] of items) {
    const card = el("article", "card stat-card");
    card.append(el("p", "stat-label", label), el("p", "stat-value", String(value)));
    grid.append(card);
  }
  return grid;
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

  const apiHint = `API: ${apiBaseUrl()}`;
  console.info("[miniapp] bootstrap", {
    api: apiBaseUrl(),
    hasInitData: Boolean(tg?.initData),
    platform: tg?.platform,
  });

  renderLoading("Авторизация…");

  try {
    const auth = await ensureAuth();
    if (auth.user.role === "admin") {
      renderAdminDashboard(auth.access_token);
      return;
    }
    if (auth.user.role === "doctor") {
      renderDoctorDashboard(auth.access_token, "pending");
      return;
    }
    renderPatientShell(auth.access_token, "booking");
  } catch (err) {
    console.error("[miniapp] auth failed", err);
    if (err instanceof ApiError) {
      clearToken();
      const hint =
        err.status === 0
          ? `${apiHint}. Запрос не дошёл до бэкенда — проверьте туннель и CORS_ALLOW_ORIGINS.`
          : `${apiHint} · HTTP ${err.status}${err.code ? ` · ${err.code}` : ""}`;
      renderError(err.message, hint);
      return;
    }
    renderError(
      err instanceof Error ? err.message : "Ошибка авторизации",
      apiHint,
    );
  }
}

bootstrap();
