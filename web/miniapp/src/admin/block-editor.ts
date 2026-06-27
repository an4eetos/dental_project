import type { ContentBlock } from "../api";
import { uploadAdminContentMedia } from "../api";

export type BlockEditor = {
  root: HTMLElement;
  getBlocks: () => ContentBlock[];
  setBlocks: (blocks: ContentBlock[]) => void;
};

type BlockEditorOptions = {
  token: string;
  contentItemId?: number;
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

export function createBlockEditor(options: BlockEditorOptions): BlockEditor {
  const blocks: ContentBlock[] = [];
  const listHost = el("div", "block-editor-list");

  const syncFromDOM = () => {
    blocks.length = 0;
    for (const row of listHost.querySelectorAll<HTMLElement>("[data-block-type]")) {
      const type = row.dataset.blockType as ContentBlock["type"];
      switch (type) {
        case "text": {
          const html = row.querySelector<HTMLTextAreaElement>("textarea")?.value.trim() ?? "";
          if (html) blocks.push({ type: "text", data: { html: `<p>${escapeHtml(html).replace(/\n/g, "<br>")}</p>` } });
          break;
        }
        case "youtube": {
          const youtubeId = row.querySelector<HTMLInputElement>('input[name="youtube"]')?.value.trim() ?? "";
          if (youtubeId) blocks.push({ type: "youtube", data: { youtube_id: youtubeId } });
          break;
        }
        case "image":
        case "video": {
          const mediaId = Number(row.dataset.mediaId ?? "0");
          const caption = row.querySelector<HTMLInputElement>('input[name="caption"]')?.value.trim() ?? "";
          if (mediaId > 0) blocks.push({ type, data: { media_id: mediaId, caption } });
          break;
        }
      }
    }
  };

  const renderRow = (block: ContentBlock, index: number) => {
    const row = el("div", "card block-editor-row");
    row.dataset.blockType = block.type;

    const header = el("div", "block-editor-row-header");
    header.append(el("strong", undefined, blockLabel(block.type)));
    const actions = el("div", "block-editor-row-actions");

    const upBtn = document.createElement("button");
    upBtn.type = "button";
    upBtn.className = "button";
    upBtn.textContent = "↑";
    upBtn.disabled = index === 0;
    upBtn.addEventListener("click", () => {
      syncFromDOM();
      if (index > 0) {
        [blocks[index - 1], blocks[index]] = [blocks[index], blocks[index - 1]];
        paint(blocks);
      }
    });

    const downBtn = document.createElement("button");
    downBtn.type = "button";
    downBtn.className = "button";
    downBtn.textContent = "↓";
    downBtn.addEventListener("click", () => {
      syncFromDOM();
      if (index < blocks.length - 1) {
        [blocks[index + 1], blocks[index]] = [blocks[index], blocks[index + 1]];
        paint(blocks);
      }
    });

    const delBtn = document.createElement("button");
    delBtn.type = "button";
    delBtn.className = "button";
    delBtn.textContent = "Удалить";
    delBtn.addEventListener("click", () => {
      syncFromDOM();
      blocks.splice(index, 1);
      paint(blocks);
    });

    actions.append(upBtn, downBtn, delBtn);
    header.append(actions);
    row.append(header);

    const body = el("div", "block-editor-row-body");
    if (block.type === "text") {
      const area = document.createElement("textarea");
      area.className = "input block-editor-text";
      area.rows = 4;
      area.placeholder = "Текст материала…";
      const html = String(block.data.html ?? "");
      area.value = stripHtml(html);
      body.append(area);
    } else if (block.type === "youtube") {
      const input = document.createElement("input");
      input.type = "text";
      input.name = "youtube";
      input.className = "input";
      input.placeholder = "YouTube ID или ссылка";
      input.value = String(block.data.youtube_id ?? "");
      body.append(input);
    } else if (block.type === "image" || block.type === "video") {
      row.dataset.mediaId = String(block.data.media_id ?? 0);
      const mediaId = Number(block.data.media_id ?? 0);
      if (mediaId > 0) {
        body.append(el("p", "muted", `Загружено: media #${mediaId}`));
      }
      const fileInput = document.createElement("input");
      fileInput.type = "file";
      fileInput.accept = block.type === "image" ? "image/*" : "video/*";
      fileInput.className = "input";
      const status = el("p", "muted");
      fileInput.addEventListener("change", () => {
        const file = fileInput.files?.[0];
        if (!file) return;
        status.textContent = "Загрузка…";
        void uploadAdminContentMedia(options.token, file, options.contentItemId)
          .then((res) => {
            row.dataset.mediaId = String(res.media_id);
            status.textContent = `Загружено: media #${res.media_id}`;
          })
          .catch((err) => {
            status.textContent = err instanceof Error ? err.message : "Ошибка загрузки";
          });
      });
      const caption = document.createElement("input");
      caption.type = "text";
      caption.name = "caption";
      caption.className = "input";
      caption.placeholder = "Подпись (необязательно)";
      caption.value = String(block.data.caption ?? "");
      body.append(fileInput, status, caption);
    }
    row.append(body);
    return row;
  };

  const paint = (next: ContentBlock[]) => {
    blocks.length = 0;
    blocks.push(...next);
    listHost.replaceChildren();
    next.forEach((block, index) => listHost.append(renderRow(block, index)));
  };

  const toolbar = el("div", "block-editor-toolbar");
  const addBtn = (label: string, type: ContentBlock["type"]) => {
    const btn = document.createElement("button");
    btn.type = "button";
    btn.className = "button";
    btn.textContent = label;
    btn.addEventListener("click", () => {
      syncFromDOM();
      blocks.push(emptyBlock(type));
      paint(blocks);
    });
    toolbar.append(btn);
  };
  addBtn("+ Текст", "text");
  addBtn("+ YouTube", "youtube");
  addBtn("+ Фото", "image");
  addBtn("+ Видео", "video");

  const root = el("div", "block-editor");
  root.append(toolbar, listHost);

  return {
    root,
    getBlocks: () => {
      syncFromDOM();
      return [...blocks];
    },
    setBlocks: (next) => paint(next.length ? next : [emptyBlock("text")]),
  };
}

function emptyBlock(type: ContentBlock["type"]): ContentBlock {
  switch (type) {
    case "text":
      return { type, data: { html: "" } };
    case "youtube":
      return { type, data: { youtube_id: "" } };
    case "image":
    case "video":
      return { type, data: { media_id: 0, caption: "" } };
  }
}

function blockLabel(type: ContentBlock["type"]): string {
  switch (type) {
    case "text":
      return "Текст";
    case "youtube":
      return "YouTube";
    case "image":
      return "Фото";
    case "video":
      return "Видео";
  }
}

function stripHtml(html: string): string {
  return html.replace(/<br\s*\/?>/gi, "\n").replace(/<[^>]+>/g, "").trim();
}

function escapeHtml(value: string): string {
  return value
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/"/g, "&quot;");
}
