import {
  ApiError,
  apiBaseUrl,
  authTelegram,
  createAppointment,
  createSubscriptionInvoice,
  fetchSubmissionMediaUrl,
  generateSubmissionDraft,
  getAdminStatistics,
  getAdminUser,
  listAdminUsers,
  setAdminUserBlocked,
  updateAdminUser,
  getSubmission,
  getSubscriptionStatus,
  listAllAppointments,
  listAnsweredSubmissions,
  listMyAppointments,
  listPendingSubmissions,
  respondAppointment,
  setAppointmentZoomLink,
  suggestAppointmentSlots,
  predict,
  respondToSubmission,
  type AdminStatistics,
  type AdminUser,
  type AdminUserOverview,
  type AppointmentDecision,
  type AuthResponse,
  type DoctorAppointment,
  type PhotoSubmission,
  type PredictRequest,
  type SubscriptionStatus,
} from "./api";
import {
  buildAdminContentRoot,
  type AdminContentState,
} from "./admin/content-ui";
import { renderContentItemCard } from "./content/render-patient";
import {
  getPatientContent,
  invalidatePatientContentCache,
  peekPatientContent,
} from "./content/cache";
import {
  PREDICTION_INPUT_FIELDS,
  PREDICTION_OUTPUT_FIELDS,
  type PredictionInputKey,
} from "./prediction-config";
import { clearToken, loadToken, saveToken } from "./storage";

const tg = window.Telegram?.WebApp;

let subscriptionCache: SubscriptionStatus | null = null;

async function refreshSubscription(token: string): Promise<SubscriptionStatus> {
  const status = await getSubscriptionStatus(token);
  const prevActive = subscriptionCache?.active;
  subscriptionCache = status;
  if (prevActive !== status.active) {
    invalidatePatientContentCache();
  }
  return status;
}

function openSubscriptionInvoice(
  token: string,
  onPaid: () => void,
): void {
  void (async () => {
    try {
      const { invoice_link: invoiceLink } = await createSubscriptionInvoice(token);
      if (tg?.openInvoice) {
        tg.openInvoice(invoiceLink, (status) => {
          if (status === "paid") {
            void refreshSubscription(token).then(onPaid);
          }
        });
        return;
      }
      window.open(invoiceLink, "_blank", "noopener,noreferrer");
    } catch (err) {
      const message =
        err instanceof ApiError ? err.message : "Не удалось открыть оплату";
      if (tg?.showAlert) {
        tg.showAlert(message);
      } else {
        window.alert(message);
      }
    }
  })();
}

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
    case "rejected":
      return "Перенос / отклонена";
    case "cancelled":
      return "Отменена";
    default:
      return status;
  }
}

function visitTypeLabel(visitType?: string): string {
  switch (visitType) {
    case "in_person":
      return "Очный приём";
    case "video":
      return "Видеоконсультация";
    default:
      return "";
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
    const visitLabel =
      a.status === "pending"
        ? visitTypeLabel(a.preferred_visit_type)
        : visitTypeLabel(a.visit_type);
    if (visitLabel) {
      card.append(
        el(
          "p",
          "muted",
          a.status === "pending" ? `Ваш запрос: ${visitLabel}` : visitLabel,
        ),
      );
    }
    if (a.status === "confirmed" && a.visit_type === "video" && !a.zoom_link) {
      card.append(el("p", "muted", "Ссылка на Zoom будет добавлена позже."));
    }
    if (a.status === "confirmed" && a.zoom_link) {
      const link = document.createElement("a");
      link.href = a.zoom_link;
      link.target = "_blank";
      link.rel = "noopener noreferrer";
      link.className = "link";
      link.textContent = "Открыть Zoom";
      card.append(link);
    }
    if (a.status === "rejected" && a.doctor_notes) {
      card.append(el("p", "muted", a.doctor_notes));
    }
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

  const visitTypeLabelEl = el("label", undefined, "Формат приёма");
  const visitTypeSelect = document.createElement("select");
  visitTypeSelect.className = "input";
  visitTypeSelect.required = true;
  visitTypeSelect.append(
    new Option("Выберите формат", ""),
    new Option("Очный приём", "in_person"),
    new Option("Видеоконсультация (Zoom)", "video"),
  );

  const submit = document.createElement("button");
  submit.type = "submit";
  submit.className = "button primary";
  submit.textContent = "Отправить заявку";

  const status = el("p", "status hidden");

  form.append(dateLabel, dateInput, timeLabel, timeInput, visitTypeLabelEl, visitTypeSelect, submit, status);

  const listTitle = el("h2", "section-title", "Мои записи");
  const listHost = el("div", "list-host");

  form.addEventListener("submit", async (e) => {
    e.preventDefault();
    status.className = "status";
    status.textContent = "Отправка…";
    submit.setAttribute("disabled", "true");

    try {
      const preferredVisitType = visitTypeSelect.value as "in_person" | "video" | "";
      if (!preferredVisitType) {
        status.textContent = "Выберите формат приёма.";
        return;
      }
      await createAppointment(
        token,
        dateInput.value,
        timeInput.value,
        preferredVisitType,
      );
      status.textContent = "Заявка отправлена. Администратор свяжется с вами.";
      const data = await listMyAppointments(token);
      listHost.replaceChildren(renderAppointmentsList(data.appointments));
      dateInput.value = "";
      timeInput.value = "";
      visitTypeSelect.value = "";
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

function buildVideosContent(
  token: string,
  subscription: SubscriptionStatus,
  onSubscriptionChange: () => void,
): HTMLElement {
  const stack = el("div", "stack");

  if (!subscription.active) {
    const banner = el(
      "div",
      "subscription-banner",
      `Подписка открывает эксклюзивные видео на ${subscription.duration_days} дн. за ${subscription.stars_price} Telegram Stars.`,
    );
    stack.append(banner);
  } else if (subscription.expires_at) {
    const expires = new Date(subscription.expires_at);
    const label = Number.isNaN(expires.getTime())
      ? "Подписка активна"
      : `Подписка активна до ${expires.toLocaleDateString("ru-RU")}`;
    const banner = el("div", "subscription-banner subscription-banner--active", label);
    stack.append(banner);
  }

  const listHost = el("section", "list video-list");
  stack.append(listHost);
  void loadVideosContent(token, subscription, onSubscriptionChange, listHost);
  return stack;
}

async function loadVideosContent(
  token: string,
  subscription: SubscriptionStatus,
  onSubscriptionChange: () => void,
  listHost: HTMLElement,
): Promise<void> {
  const renderItems = (items: import("./api").ContentItem[]) => {
    listHost.replaceChildren();
    if (items.length === 0) {
      listHost.append(el("p", "muted", "Материалы скоро появятся."));
      return;
    }
    for (const item of items) {
      listHost.append(
        renderContentItemCard(item, {
          token,
          subscription,
          onSubscriptionChange,
          openSubscriptionInvoice,
        }),
      );
    }
  };

  const cached = peekPatientContent(token, subscription.active);
  if (cached) {
    renderItems(cached);
    return;
  }

  listHost.replaceChildren(el("p", "muted", "Загрузка материалов…"));
  try {
    const items = await getPatientContent(token, subscription.active);
    renderItems(items);
  } catch (err) {
    listHost.replaceChildren(
      el(
        "p",
        "muted",
        err instanceof ApiError ? err.message : "Не удалось загрузить материалы",
      ),
    );
  }
}

function buildPatientTabContent(
  token: string,
  tab: PatientTab,
  subscription: SubscriptionStatus,
  onSubscriptionChange: () => void,
): HTMLElement {
  switch (tab) {
    case "booking":
      return buildBookingContent(token);
    case "prediction":
      return buildPredictionContent(token);
    case "videos":
      return buildVideosContent(token, subscription, onSubscriptionChange);
  }
}

function renderPatientShell(token: string, activeTab: PatientTab): void {
  void renderPatientShellAsync(token, activeTab);
}

async function renderPatientShellAsync(token: string, activeTab: PatientTab): Promise<void> {
  let subscription = subscriptionCache;
  if (!subscription) {
    try {
      subscription = await refreshSubscription(token);
    } catch (err) {
      subscription = subscription ?? {
        active: false,
        stars_price: 50,
        duration_days: 30,
      };
      if (activeTab === "videos") {
        console.warn("[miniapp] subscription status unavailable", err);
      }
    }
  }

  const rerender = () => {
    void renderPatientShellAsync(token, activeTab);
  };

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
  contentHost.append(
    buildPatientTabContent(token, activeTab, subscription, rerender),
  );

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
    title: "Заявки без ответа",
    subtitle: "Фото и видео — ответьте в течение 48 часов",
  },
  {
    id: "answered",
    label: "Отвеченные",
    title: "Отвеченные заявки",
    subtitle: "История ответов по фото и видео",
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
      listHost.replaceChildren(
        renderDoctorAppointmentsList(token, data.appointments, load),
      );
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
          err instanceof ApiError ? err.message : "Не удалось загрузить заявки",
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
        queue === "pending"
          ? "Нет заявок (фото или видео), ожидающих ответа."
          : "Отвеченных заявок пока нет.",
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
    card.append(
      el(
        "p",
        "muted",
        item.media_type === "video" ? "Видео" : "Фото",
      ),
    );
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
  renderShell(shell, "Просмотр заявки", "Детали фото или видео");

  void (async () => {
    try {
      const submission = await getSubmission(token, submissionId);
      const mediaUrl = await fetchSubmissionMediaUrl(token, submissionId);
      shell.replaceChildren(
        buildSubmissionDetailContent(token, submission, mediaUrl, queue, onBack),
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
  mediaUrl: string,
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
    el(
      "p",
      "muted",
      submission.media_type === "video" ? "Тип: видео" : "Тип: фото",
    ),
  );

  if (submission.media_type === "video") {
    const video = document.createElement("video");
    video.src = mediaUrl;
    video.controls = true;
    video.playsInline = true;
    video.className = "submission-video";
    card.append(video);
  } else {
    const img = document.createElement("img");
    img.src = mediaUrl;
    img.alt = "Фото пациента";
    img.className = "submission-photo";
    card.append(img);
  }
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

function roleLabel(role: string): string {
  switch (role) {
    case "doctor":
      return "Врач";
    case "admin":
      return "Администратор";
    default:
      return "Пациент";
  }
}

const ADMIN_TABS = [
  { id: "stats" as const, label: "Статистика", title: "Статистика клиники", subtitle: "Панель администратора" },
  { id: "users" as const, label: "Пользователи", title: "Пользователи", subtitle: "Просмотр и редактирование" },
  { id: "content" as const, label: "Контент", title: "Обучающие материалы", subtitle: "Видео, тексты и медиа" },
];

let adminContentState: AdminContentState = { tab: "list" };

function renderAdminDashboard(
  token: string,
  activeTab: "stats" | "users" | "content" = "stats",
): void {
  const tabs = el("nav", "tabs");
  const contentHost = el("div", "tab-content");

  for (const tab of ADMIN_TABS) {
    const btn = document.createElement("button");
    btn.type = "button";
    btn.className = `tab${tab.id === activeTab ? " active" : ""}`;
    btn.textContent = tab.label;
    btn.addEventListener("click", () => {
      if (tab.id !== "content") {
        adminContentState = { tab: "list" };
      }
      renderAdminDashboard(token, tab.id);
    });
    tabs.append(btn);
  }

  switch (activeTab) {
    case "stats":
      contentHost.append(buildAdminStatsContent(token));
      break;
    case "users":
      contentHost.append(buildAdminUsersContent(token));
      break;
    case "content":
      contentHost.append(
        buildAdminContentRoot(token, adminContentState, (next) => {
          adminContentState = next;
          renderAdminDashboard(token, "content");
        }),
      );
      break;
  }

  const shell = el("div", "stack");
  shell.append(tabs, contentHost);
  const meta = ADMIN_TABS.find((t) => t.id === activeTab) ?? ADMIN_TABS[0];
  renderShell(shell, meta.title, meta.subtitle);
}

function buildAdminStatsContent(token: string): HTMLElement {
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
  void load();
  return shell;
}

function buildAdminUsersContent(token: string): HTMLElement {
  const listHost = el("div", "list-host");
  const searchInput = document.createElement("input");
  searchInput.type = "search";
  searchInput.className = "input";
  searchInput.placeholder = "Поиск по имени, username или Telegram ID";

  const roleSelect = document.createElement("select");
  roleSelect.className = "input";
  roleSelect.append(
    new Option("Все роли", ""),
    new Option("Пациенты", "patient"),
    new Option("Врачи", "doctor"),
    new Option("Администраторы", "admin"),
  );

  const refresh = document.createElement("button");
  refresh.type = "button";
  refresh.className = "button primary";
  refresh.textContent = "Обновить";

  const load = async () => {
    listHost.replaceChildren(el("p", "status", "Загрузка…"));
    try {
      const role = roleSelect.value as "" | "patient" | "doctor" | "admin";
      const data = await listAdminUsers(token, {
        search: searchInput.value.trim() || undefined,
        role: role || undefined,
      });
      listHost.replaceChildren(renderAdminUserList(token, data.users, load));
    } catch (err) {
      listHost.replaceChildren(
        el(
          "p",
          "muted",
          err instanceof ApiError ? err.message : "Не удалось загрузить пользователей",
        ),
      );
    }
  };

  searchInput.addEventListener("keydown", (e) => {
    if (e.key === "Enter") void load();
  });
  roleSelect.addEventListener("change", () => void load());
  refresh.addEventListener("click", () => void load());

  const shell = el("div", "stack");
  shell.append(searchInput, roleSelect, refresh, listHost);
  void load();
  return shell;
}

function renderAdminUserList(
  token: string,
  users: AdminUser[],
  onBack: () => void,
): HTMLElement {
  const list = el("section", "list");
  if (users.length === 0) {
    list.append(el("p", "muted", "Пользователи не найдены."));
    return list;
  }

  for (const user of users) {
    const card = el("article", "card item submission-item");
    const title = `${user.first_name}${user.last_name ? ` ${user.last_name}` : ""}`;
    card.append(
      el("p", "item-title", title),
      el("p", "muted", `${roleLabel(user.role)}${user.blocked ? " · заблокирован" : ""}`),
    );
    if (user.username) {
      card.append(el("p", "muted", `@${user.username}`));
    }
    card.addEventListener("click", () => {
      renderAdminUserDetail(token, user.id, onBack);
    });
    list.append(card);
  }
  return list;
}

function renderAdminUserDetail(token: string, userId: number, onBack: () => void): void {
  const shell = el("div", "stack");
  renderShell(shell, "Пользователь", "Карточка пациента / врача");

  const back = document.createElement("button");
  back.type = "button";
  back.className = "button";
  back.textContent = "← Назад к списку";
  back.addEventListener("click", onBack);
  shell.append(back);

  const host = el("div", "list-host");
  host.append(el("p", "status", "Загрузка…"));
  shell.append(host);

  void (async () => {
    try {
      const user = await getAdminUser(token, userId);
      host.replaceChildren(renderAdminUserForm(token, user, onBack));
    } catch (err) {
      host.replaceChildren(
        el(
          "p",
          "muted",
          err instanceof ApiError ? err.message : "Не удалось загрузить пользователя",
        ),
      );
    }
  })();
}

function renderAdminUserForm(
  token: string,
  user: AdminUserOverview,
  onBack: () => void,
): HTMLElement {
  const wrap = el("div", "stack");
  const info = el("article", "card");
  info.append(
    el("p", "muted", `Telegram ID: ${user.telegram_id}`),
    el("p", "muted", `Записей: ${user.appointment_count}`),
    el("p", "muted", `Заявок (фото/видео): ${user.photo_submission_count}`),
    el("p", "muted", `Регистрация: ${formatDateTime(user.created_at)}`),
  );
  wrap.append(info);

  const isAdmin = user.role === "admin";
  const form = el("form", "card form");

  const firstNameLabel = el("label", undefined, "Имя");
  const firstNameInput = document.createElement("input");
  firstNameInput.className = "input";
  firstNameInput.required = true;
  firstNameInput.value = user.first_name;
  firstNameInput.disabled = isAdmin;

  const lastNameLabel = el("label", undefined, "Фамилия");
  const lastNameInput = document.createElement("input");
  lastNameInput.className = "input";
  lastNameInput.value = user.last_name ?? "";
  lastNameInput.disabled = isAdmin;

  const usernameLabel = el("label", undefined, "Username");
  const usernameInput = document.createElement("input");
  usernameInput.className = "input";
  usernameInput.value = user.username ?? "";
  usernameInput.disabled = isAdmin;

  const roleLabelEl = el("label", undefined, "Роль");
  const roleInput = document.createElement("select");
  roleInput.className = "input";
  roleInput.disabled = isAdmin;
  roleInput.append(new Option("Пациент", "patient"), new Option("Врач", "doctor"));
  roleInput.value = user.role === "doctor" ? "doctor" : "patient";

  const save = document.createElement("button");
  save.type = "submit";
  save.className = "button primary";
  save.textContent = "Сохранить";
  save.disabled = isAdmin;

  const blockBtn = document.createElement("button");
  blockBtn.type = "button";
  blockBtn.className = user.blocked ? "button primary" : "button";
  blockBtn.textContent = user.blocked ? "Разблокировать" : "Заблокировать";
  blockBtn.disabled = isAdmin;

  const status = el("p", "status hidden");
  form.append(
    firstNameLabel,
    firstNameInput,
    lastNameLabel,
    lastNameInput,
    usernameLabel,
    usernameInput,
    roleLabelEl,
    roleInput,
    save,
    blockBtn,
    status,
  );

  form.addEventListener("submit", async (e) => {
    e.preventDefault();
    save.setAttribute("disabled", "true");
    status.className = "status";
    status.textContent = "Сохранение…";
    try {
      await updateAdminUser(token, user.id, {
        first_name: firstNameInput.value.trim(),
        last_name: lastNameInput.value.trim(),
        username: usernameInput.value.trim(),
        role: roleInput.value as "patient" | "doctor",
      });
      status.textContent = "Изменения сохранены.";
      setTimeout(onBack, 700);
    } catch (err) {
      status.textContent =
        err instanceof ApiError ? err.message : "Не удалось сохранить изменения";
      save.removeAttribute("disabled");
    }
  });

  blockBtn.addEventListener("click", async () => {
    blockBtn.setAttribute("disabled", "true");
    status.className = "status";
    status.textContent = user.blocked ? "Разблокировка…" : "Блокировка…";
    try {
      await setAdminUserBlocked(token, user.id, !user.blocked);
      status.textContent = user.blocked ? "Пользователь разблокирован." : "Пользователь заблокирован.";
      setTimeout(onBack, 700);
    } catch (err) {
      status.textContent =
        err instanceof ApiError ? err.message : "Не удалось изменить статус блокировки";
      blockBtn.removeAttribute("disabled");
    }
  });

  wrap.append(form);
  return wrap;
}

function renderAdminStats(stats: AdminStatistics): HTMLElement {
  const grid = el("div", "stats-grid");
  const items: [string, number][] = [
    ["Всего пользователей", stats.total_users],
    ["Пациентов", stats.total_patients],
    ["Врачей", stats.total_doctors],
    ["Администраторов", stats.total_admins],
    ["Всего заявок (фото/видео)", stats.total_photo_submissions],
    ["Ожидают ответа (фото/видео)", stats.pending_photo_submissions],
    ["Отвечено (фото/видео)", stats.answered_photo_submissions],
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

function renderDoctorAppointmentsList(
  token: string,
  items: DoctorAppointment[],
  onBack: () => void,
): HTMLElement {
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
    const preferenceLabel = visitTypeLabel(a.preferred_visit_type);
    if (preferenceLabel) {
      card.append(el("p", "muted", `Предпочтение пациента: ${preferenceLabel}`));
    }
    const visitLabel = visitTypeLabel(a.visit_type);
    if (visitLabel && a.status !== "pending") {
      card.append(el("p", "muted", visitLabel));
    }
    if (a.needs_zoom_link) {
      card.append(el("p", "muted", "⚠️ Нужна ссылка на Zoom"));
    }
    if (a.status === "cancelled" || a.status === "rejected") {
      list.append(card);
      continue;
    }
    card.addEventListener("click", () => {
      renderDoctorAppointmentOffer(token, a, onBack);
    });
    list.append(card);
  }
  return list;
}

function renderDoctorAppointmentOffer(
  token: string,
  appointment: DoctorAppointment,
  onBack: () => void,
): void {
  const shell = el("div", "stack");
  renderShell(shell, "Ответ на заявку", patientName(appointment.patient));

  const back = document.createElement("button");
  back.type = "button";
  back.className = "button";
  back.textContent = "← Назад к списку";
  back.addEventListener("click", onBack);
  shell.append(back);

  if (appointment.needs_zoom_link) {
    shell.append(buildZoomLinkForm(token, appointment, onBack));
  }

  if (appointment.status !== "pending") {
    if (!appointment.needs_zoom_link) {
      shell.append(el("p", "muted", "Заявка уже обработана."));
    }
    return;
  }

  const preferenceLabel = visitTypeLabel(appointment.preferred_visit_type);
  if (preferenceLabel) {
    shell.append(el("p", "muted", `Предпочтение пациента: ${preferenceLabel}`));
  }

  const form = el("form", "card form");

  const decisionLabel = el("label", undefined, "Решение врача");
  const decisionSelect = document.createElement("select");
  decisionSelect.className = "input";
  decisionSelect.required = true;
  decisionSelect.append(
    new Option("Выберите вариант", ""),
    new Option("Очный приём", "in_person"),
    new Option("Видеоконсультация (Zoom)", "video"),
    new Option("Перенести / отклонить", "reject"),
  );

  const dateLabel = el("label", undefined, "Дата консультации");
  const dateInput = document.createElement("input");
  dateInput.type = "date";
  dateInput.min = minDate();
  dateInput.max = maxDate();
  dateInput.className = "input";
  dateInput.value = appointment.preferred_date;

  const timeLabel = el("label", undefined, "Время консультации");
  const timeInput = document.createElement("input");
  timeInput.type = "time";
  timeInput.min = "09:00";
  timeInput.max = "20:00";
  timeInput.step = "60";
  timeInput.className = "input";
  timeInput.value = appointment.preferred_time;

  const zoomLabel = el("label", undefined, "Ссылка на Zoom (можно добавить позже)");
  const zoomInput = document.createElement("input");
  zoomInput.type = "url";
  zoomInput.className = "input";
  zoomInput.placeholder = "https://zoom.us/j/...";
  zoomInput.value = appointment.zoom_link ?? "";

  const notesLabel = el("label", undefined, "Комментарий для пациента");
  const notesInput = document.createElement("textarea");
  notesInput.className = "input textarea";
  notesInput.rows = 5;
  notesInput.placeholder = "Укажите доступные даты и время для переноса";

  const generateSlots = document.createElement("button");
  generateSlots.type = "button";
  generateSlots.className = "button";
  generateSlots.textContent = "Сгенерировать доступные слоты";

  const submit = document.createElement("button");
  submit.type = "submit";
  submit.className = "button primary";
  submit.textContent = "Отправить решение пациенту";

  const status = el("p", "status hidden");

  const syncFields = () => {
    const decision = decisionSelect.value as AppointmentDecision | "";
    const isReject = decision === "reject";
    const isVideo = decision === "video";
    const needsSlot = decision === "in_person" || decision === "video";

    dateLabel.classList.toggle("hidden", !needsSlot);
    dateInput.classList.toggle("hidden", !needsSlot);
    dateInput.required = needsSlot;

    timeLabel.classList.toggle("hidden", !needsSlot);
    timeInput.classList.toggle("hidden", !needsSlot);
    timeInput.required = needsSlot;

    zoomLabel.classList.toggle("hidden", !isVideo);
    zoomInput.classList.toggle("hidden", !isVideo);

    notesLabel.textContent = isReject
      ? "Доступные варианты и пояснение для пациента"
      : "Комментарий для пациента (необязательно)";
    notesInput.required = isReject;
    generateSlots.classList.toggle("hidden", !isReject);
  };

  decisionSelect.addEventListener("change", syncFields);
  generateSlots.addEventListener("click", async () => {
    generateSlots.setAttribute("disabled", "true");
    status.className = "status";
    status.textContent = "Генерация…";
    try {
      const data = await suggestAppointmentSlots(token, appointment.id);
      notesInput.value = data.suggested_text;
      status.textContent = "Варианты добавлены — при необходимости отредактируйте.";
    } catch (err) {
      status.textContent =
        err instanceof ApiError ? err.message : "Не удалось сгенерировать слоты";
    } finally {
      generateSlots.removeAttribute("disabled");
    }
  });

  form.append(
    decisionLabel,
    decisionSelect,
    dateLabel,
    dateInput,
    timeLabel,
    timeInput,
    zoomLabel,
    zoomInput,
    notesLabel,
    notesInput,
    generateSlots,
    submit,
    status,
  );
  syncFields();

  form.addEventListener("submit", async (e) => {
    e.preventDefault();
    const decision = decisionSelect.value as AppointmentDecision;
    if (!decision) {
      status.className = "status";
      status.textContent = "Выберите решение.";
      return;
    }

    submit.setAttribute("disabled", "true");
    status.className = "status";
    status.textContent = "Отправка…";
    try {
      await respondAppointment(token, appointment.id, {
        decision,
        preferred_date: dateInput.value,
        preferred_time: timeInput.value,
        zoom_link: zoomInput.value.trim(),
        doctor_notes: notesInput.value.trim(),
      });
      status.textContent = "Пациент получил уведомление.";
      setTimeout(onBack, 900);
    } catch (err) {
      status.textContent =
        err instanceof ApiError ? err.message : "Не удалось отправить решение";
      submit.removeAttribute("disabled");
    }
  });

  shell.append(form);
}

function buildZoomLinkForm(
  token: string,
  appointment: DoctorAppointment,
  onBack: () => void,
): HTMLElement {
  const form = el("form", "card form");
  form.append(el("h2", "section-title", "Добавить ссылку на Zoom"));

  const zoomInput = document.createElement("input");
  zoomInput.type = "url";
  zoomInput.required = true;
  zoomInput.className = "input";
  zoomInput.placeholder = "https://zoom.us/j/...";
  zoomInput.value = appointment.zoom_link ?? "";

  const submit = document.createElement("button");
  submit.type = "submit";
  submit.className = "button primary";
  submit.textContent = "Сохранить и уведомить пациента";

  const status = el("p", "status hidden");
  form.append(zoomInput, submit, status);

  form.addEventListener("submit", async (e) => {
    e.preventDefault();
    submit.setAttribute("disabled", "true");
    status.className = "status";
    status.textContent = "Сохранение…";
    try {
      await setAppointmentZoomLink(token, appointment.id, zoomInput.value.trim());
      status.textContent = "Пациент получил ссылку на Zoom.";
      setTimeout(onBack, 900);
    } catch (err) {
      status.textContent =
        err instanceof ApiError ? err.message : "Не удалось сохранить ссылку";
      submit.removeAttribute("disabled");
    }
  });

  return form;
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
    if (auth.subscription) {
      subscriptionCache = auth.subscription;
    }
    if (auth.user.role === "admin") {
      renderAdminDashboard(auth.access_token);
      return;
    }
    if (auth.user.role === "doctor") {
      renderDoctorDashboard(auth.access_token, "appointments");
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
