import type { ContentBlock, ContentItem, SubscriptionStatus } from "../api";
import { getCachedContentMediaUrl } from "./cache";
import { youtubeEmbedUrl, youtubeThumbnailUrl, youtubeWatchUrl } from "../youtube";

type SubscriptionHandlers = {
  token: string;
  subscription: SubscriptionStatus;
  onSubscriptionChange: () => void;
  openSubscriptionInvoice: (token: string, onChange: () => void) => void;
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

export function renderContentItemCard(
  item: ContentItem,
  handlers: SubscriptionHandlers,
): HTMLElement {
  const card = el("article", item.locked ? "card video-card video-card--locked" : "card content-card");
  card.append(el("h3", "video-title", item.title));
  if (item.description) {
    card.append(el("p", "muted video-desc", item.description));
  }
  if (item.access === "subscription") {
    card.append(el("span", "subscription-badge", item.locked ? "Только для подписчиков" : "Для подписчиков"));
  }

  if (item.locked) {
    const lockedView = el("div", "video-placeholder content-lock-placeholder");
    lockedView.append(
      el("div", "video-lock-overlay"),
    );
    const overlay = lockedView.querySelector(".video-lock-overlay")!;
    overlay.append(el("span", undefined, "🔒"), el("p", undefined, "Эксклюзивный материал для подписчиков"));
    card.append(lockedView);

    const subscribeBtn = document.createElement("button");
    subscribeBtn.type = "button";
    subscribeBtn.className = "button primary subscription-cta";
    subscribeBtn.textContent = `Оформить подписку — ${handlers.subscription.stars_price} ⭐`;
    subscribeBtn.addEventListener("click", () => {
      handlers.openSubscriptionInvoice(handlers.token, handlers.onSubscriptionChange);
    });
    card.append(subscribeBtn);
    return card;
  }

  const blocksHost = el("div", "content-blocks");
  void renderBlocksInto(blocksHost, item.blocks, handlers);
  card.append(blocksHost);
  return card;
}

async function renderBlocksInto(
  host: HTMLElement,
  blocks: ContentBlock[],
  handlers: SubscriptionHandlers,
): Promise<void> {
  for (const block of blocks) {
    host.append(await renderBlock(block, handlers));
  }
}

async function renderBlock(block: ContentBlock, handlers: SubscriptionHandlers): Promise<HTMLElement> {
  switch (block.type) {
    case "text": {
      const wrap = el("div", "content-block content-block--text");
      const html = String(block.data.html ?? "");
      wrap.innerHTML = html;
      return wrap;
    }
    case "youtube": {
      const youtubeId = String(block.data.youtube_id ?? "");
      const wrap = el("div", "content-block content-block--youtube");
      if (!youtubeId) return wrap;

      const playerHost = el("div", "video-player");
      const playBtn = document.createElement("button");
      playBtn.type = "button";
      playBtn.className = "video-placeholder";
      playBtn.setAttribute("aria-label", "Смотреть видео");

      const thumb = document.createElement("img");
      thumb.src = youtubeThumbnailUrl(youtubeId);
      thumb.alt = "";
      thumb.className = "video-thumb";
      thumb.loading = "lazy";
      playBtn.append(thumb, el("span", "video-play", "▶"));

      playBtn.addEventListener("click", () => {
        const iframe = document.createElement("iframe");
        iframe.src = youtubeEmbedUrl(youtubeId);
        iframe.className = "video-embed";
        iframe.setAttribute(
          "allow",
          "accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share",
        );
        iframe.allowFullscreen = true;
        playerHost.replaceChildren(iframe);
      });

      playerHost.append(playBtn);
      wrap.append(playerHost);

      const openExternal = document.createElement("button");
      openExternal.type = "button";
      openExternal.className = "button video-open";
      openExternal.textContent = "Открыть в YouTube";
      openExternal.addEventListener("click", () => {
        const url = youtubeWatchUrl(youtubeId);
        const tg = window.Telegram?.WebApp;
        if (tg?.openLink) tg.openLink(url);
        else window.open(url, "_blank", "noopener,noreferrer");
      });
      wrap.append(openExternal);
      return wrap;
    }
    case "image": {
      const mediaId = Number(block.data.media_id);
      const wrap = el("figure", "content-block content-block--image");
      if (!mediaId) return wrap;
      try {
        const url = await getCachedContentMediaUrl(handlers.token, mediaId);
        const img = document.createElement("img");
        img.src = url;
        img.alt = String(block.data.caption ?? "");
        img.className = "content-media-image";
        img.loading = "lazy";
        wrap.append(img);
        const caption = String(block.data.caption ?? "").trim();
        if (caption) wrap.append(el("figcaption", "muted", caption));
      } catch {
        wrap.append(el("p", "muted", "Не удалось загрузить изображение"));
      }
      return wrap;
    }
    case "video": {
      const mediaId = Number(block.data.media_id);
      const wrap = el("figure", "content-block content-block--video");
      if (!mediaId) return wrap;
      try {
        const url = await getCachedContentMediaUrl(handlers.token, mediaId);
        const video = document.createElement("video");
        video.src = url;
        video.controls = true;
        video.playsInline = true;
        video.className = "content-media-video";
        wrap.append(video);
        const caption = String(block.data.caption ?? "").trim();
        if (caption) wrap.append(el("figcaption", "muted", caption));
      } catch {
        wrap.append(el("p", "muted", "Не удалось загрузить видео"));
      }
      return wrap;
    }
    default:
      return el("div", "content-block");
  }
}
