import {
  createAdminContent,
  deleteAdminContent,
  getAdminContent,
  listAdminContent,
  reorderAdminContent,
  updateAdminContent,
  type AdminContentItem,
  type SaveContentPayload,
} from "../api";
import { createBlockEditor } from "./block-editor";

type AdminContentTab = "list" | "edit";

export type AdminContentState = {
  tab: AdminContentTab;
  editingId?: number;
};

function el<K extends keyof HTMLElementTagNameMap>(
  tag: K,
  className?: string,
  text?: string,
): HTMLElementTagNameMap[K] {
  const node = document.createElement(tag);
  if (className) node.className = className;
  if (text != null) node.textContent = text;
  return node;
}

export function buildAdminContentRoot(
  token: string,
  state: AdminContentState,
  onStateChange: (next: AdminContentState) => void,
): HTMLElement {
  if (state.tab === "edit") {
    return buildAdminContentEditor(token, state.editingId, () => onStateChange({ tab: "list" }));
  }
  return buildAdminContentList(token, onStateChange);
}

function buildAdminContentList(
  token: string,
  onStateChange: (next: AdminContentState) => void,
): HTMLElement {
  const wrap = el("div", "stack");
  const listHost = el("div", "list-host");
  const errorHost = el("p", "error hidden");

  const createBtn = document.createElement("button");
  createBtn.type = "button";
  createBtn.className = "button primary";
  createBtn.textContent = "+ Создать материал";
  createBtn.addEventListener("click", () => onStateChange({ tab: "edit" }));

  const refreshBtn = document.createElement("button");
  refreshBtn.type = "button";
  refreshBtn.className = "button";
  refreshBtn.textContent = "Обновить";

  const toolbar = el("div", "toolbar-row");
  toolbar.append(createBtn, refreshBtn);
  wrap.append(toolbar, errorHost, listHost);

  const load = async () => {
    errorHost.classList.add("hidden");
    listHost.replaceChildren(el("p", "muted", "Загрузка…"));
    try {
      const data = await listAdminContent(token);
      renderList(data.items);
    } catch (err) {
      errorHost.textContent = err instanceof Error ? err.message : "Ошибка загрузки";
      errorHost.classList.remove("hidden");
      listHost.replaceChildren();
    }
  };

  const renderList = (items: AdminContentItem[]) => {
    listHost.replaceChildren();
    if (items.length === 0) {
      listHost.append(el("p", "muted", "Материалов пока нет. Создайте первый."));
      return;
    }

    const list = el("section", "list");
    items.forEach((item, index) => {
      const card = el("article", "card admin-content-card");
      const titleRow = el("div", "admin-content-card-title");
      titleRow.append(el("h3", undefined, item.title));
      const badges = el("div", "admin-content-badges");
      badges.append(el("span", "badge", item.access === "subscription" ? "Подписка" : "Публичный"));
      badges.append(el("span", "badge", item.published ? "Опубликован" : "Черновик"));
      titleRow.append(badges);
      card.append(titleRow);
      if (item.description) card.append(el("p", "muted", item.description));

      const actions = el("div", "admin-content-actions");
      const editBtn = document.createElement("button");
      editBtn.type = "button";
      editBtn.className = "button";
      editBtn.textContent = "Редактировать";
      editBtn.addEventListener("click", () => onStateChange({ tab: "edit", editingId: item.id }));

      const upBtn = document.createElement("button");
      upBtn.type = "button";
      upBtn.className = "button";
      upBtn.textContent = "↑";
      upBtn.disabled = index === 0;
      upBtn.addEventListener("click", () => {
        void reorder(token, items, index, index - 1, load);
      });

      const downBtn = document.createElement("button");
      downBtn.type = "button";
      downBtn.className = "button";
      downBtn.textContent = "↓";
      downBtn.disabled = index === items.length - 1;
      downBtn.addEventListener("click", () => {
        void reorder(token, items, index, index + 1, load);
      });

      const deleteBtn = document.createElement("button");
      deleteBtn.type = "button";
      deleteBtn.className = "button";
      deleteBtn.textContent = "Удалить";
      deleteBtn.addEventListener("click", () => {
        if (!window.confirm(`Удалить «${item.title}»?`)) return;
        void deleteAdminContent(token, item.id).then(load).catch((err) => {
          errorHost.textContent = err instanceof Error ? err.message : "Ошибка удаления";
          errorHost.classList.remove("hidden");
        });
      });

      actions.append(editBtn, upBtn, downBtn, deleteBtn);
      card.append(actions);
      list.append(card);
    });
    listHost.append(list);
  };

  refreshBtn.addEventListener("click", () => void load());
  void load();
  return wrap;
}

async function reorder(
  token: string,
  items: AdminContentItem[],
  from: number,
  to: number,
  reload: () => void,
): Promise<void> {
  const ids = items.map((item) => item.id);
  [ids[from], ids[to]] = [ids[to], ids[from]];
  await reorderAdminContent(token, ids);
  reload();
}

function buildAdminContentEditor(
  token: string,
  editingId: number | undefined,
  onBack: () => void,
): HTMLElement {
  const wrap = el("div", "stack");
  const errorHost = el("p", "error hidden");
  const form = el("div", "stack admin-content-form");

  const titleInput = document.createElement("input");
  titleInput.type = "text";
  titleInput.className = "input";
  titleInput.placeholder = "Заголовок";
  titleInput.required = true;

  const descInput = document.createElement("input");
  descInput.type = "text";
  descInput.className = "input";
  descInput.placeholder = "Краткое описание (необязательно)";

  const accessSelect = document.createElement("select");
  accessSelect.className = "input";
  accessSelect.innerHTML = `
    <option value="public">Публичный</option>
    <option value="subscription">Только для подписчиков</option>
  `;

  const publishedLabel = el("label", "checkbox-row");
  const publishedInput = document.createElement("input");
  publishedInput.type = "checkbox";
  publishedLabel.append(publishedInput, document.createTextNode(" Опубликован"));

  const editor = createBlockEditor({ token, contentItemId: editingId });

  form.append(
    el("label", undefined, "Заголовок"),
    titleInput,
    el("label", undefined, "Описание"),
    descInput,
    el("label", undefined, "Доступ"),
    accessSelect,
    publishedLabel,
    el("h3", undefined, "Блоки контента"),
    editor.root,
  );

  const actions = el("div", "toolbar-row");
  const backBtn = document.createElement("button");
  backBtn.type = "button";
  backBtn.className = "button";
  backBtn.textContent = "Назад";
  backBtn.addEventListener("click", onBack);

  const saveBtn = document.createElement("button");
  saveBtn.type = "button";
  saveBtn.className = "button primary";
  saveBtn.textContent = editingId ? "Сохранить" : "Создать";

  actions.append(backBtn, saveBtn);
  wrap.append(errorHost, form, actions);

  const loadExisting = async () => {
    if (!editingId) {
      editor.setBlocks([{ type: "text", data: { html: "" } }]);
      return;
    }
    try {
      const item = await getAdminContent(token, editingId);
      titleInput.value = item.title;
      descInput.value = item.description ?? "";
      accessSelect.value = item.access;
      publishedInput.checked = item.published;
      editor.setBlocks(item.blocks.length ? item.blocks : [{ type: "text", data: { html: "" } }]);
    } catch (err) {
      errorHost.textContent = err instanceof Error ? err.message : "Ошибка загрузки";
      errorHost.classList.remove("hidden");
    }
  };

  saveBtn.addEventListener("click", () => {
    const payload: SaveContentPayload = {
      title: titleInput.value.trim(),
      description: descInput.value.trim(),
      access: accessSelect.value as SaveContentPayload["access"],
      published: publishedInput.checked,
      blocks: editor.getBlocks(),
    };
    if (!payload.title) {
      errorHost.textContent = "Укажите заголовок";
      errorHost.classList.remove("hidden");
      return;
    }
    if (payload.blocks.length === 0) {
      errorHost.textContent = "Добавьте хотя бы один блок контента";
      errorHost.classList.remove("hidden");
      return;
    }
    saveBtn.disabled = true;
    errorHost.classList.add("hidden");
    const savePromise = editingId
      ? updateAdminContent(token, editingId, payload)
      : createAdminContent(token, payload);
    void savePromise
      .then(() => onBack())
      .catch((err) => {
        errorHost.textContent = err instanceof Error ? err.message : "Ошибка сохранения";
        errorHost.classList.remove("hidden");
      })
      .finally(() => {
        saveBtn.disabled = false;
      });
  });

  void loadExisting();
  return wrap;
}
