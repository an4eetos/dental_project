export type YouTubeVideo = {
  /** YouTube video ID from watch?v=… or youtu.be/… URLs */
  youtubeId: string;
  title: string;
  description?: string;
};

/**
 * Add your YouTube videos here.
 * ID is the part after watch?v= or youtu.be/ (not the full URL).
 */
export const YOUTUBE_VIDEOS: readonly YouTubeVideo[] = [
  {
    youtubeId: "zQZ3SGSwGBI",
    title: "Балалардың ауыз қуысы гигиенасы",
    description: "Рекомендации стоматолога для будущих мам",
  },
  {
    youtubeId: "IFT7drSL35s",
    title: "Гигиена полости рта детей",
    description: "Рекомендации стоматолога для будущих мам",
  },
  {
    youtubeId: "FMU4zgGRbiE",
    title: "Жүктілік кезіндегі ауыз қуысының гигиенасы",
    description: "Рекомендации стоматолога для будущих мам",
  },
  {
    youtubeId: "yKlH5tjZTxI",
    title: "Гигиена полости рта беременных пациентов",
    description: "Рекомендации стоматолога для будущих мам",
  },
];

export function youtubeEmbedUrl(youtubeId: string): string {
  return `https://www.youtube.com/embed/${youtubeId}?rel=0&modestbranding=1&playsinline=1`;
}

export function youtubeWatchUrl(youtubeId: string): string {
  return `https://www.youtube.com/watch?v=${youtubeId}`;
}

export function youtubeThumbnailUrl(youtubeId: string): string {
  return `https://img.youtube.com/vi/${youtubeId}/hqdefault.jpg`;
}
