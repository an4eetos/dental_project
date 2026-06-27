import type { AdminContentItem, ContentItem } from "../api";
import { fetchContentMediaUrl, listAdminContent, listContent } from "../api";

type PatientContentCache = {
  token: string;
  subscriptionActive: boolean;
  items: ContentItem[];
};

type AdminContentCache = {
  token: string;
  items: AdminContentItem[];
};

let patientContentCache: PatientContentCache | null = null;
let adminContentCache: AdminContentCache | null = null;
const mediaUrlCache = new Map<number, string>();

export function invalidatePatientContentCache(): void {
  patientContentCache = null;
  for (const url of mediaUrlCache.values()) {
    URL.revokeObjectURL(url);
  }
  mediaUrlCache.clear();
}

export function invalidateAdminContentCache(): void {
  adminContentCache = null;
}

export function invalidateAllContentCaches(): void {
  invalidatePatientContentCache();
  invalidateAdminContentCache();
}

export function peekPatientContent(
  token: string,
  subscriptionActive: boolean,
): ContentItem[] | null {
  if (
    patientContentCache &&
    patientContentCache.token === token &&
    patientContentCache.subscriptionActive === subscriptionActive
  ) {
    return patientContentCache.items;
  }
  return null;
}

export async function getPatientContent(
  token: string,
  subscriptionActive: boolean,
  options?: { force?: boolean },
): Promise<ContentItem[]> {
  if (!options?.force) {
    const cached = peekPatientContent(token, subscriptionActive);
    if (cached) {
      return cached;
    }
  }

  const data = await listContent(token);
  patientContentCache = {
    token,
    subscriptionActive,
    items: data.items,
  };
  return data.items;
}

export function peekAdminContent(token: string): AdminContentItem[] | null {
  if (adminContentCache && adminContentCache.token === token) {
    return adminContentCache.items;
  }
  return null;
}

export async function getAdminContentList(
  token: string,
  options?: { force?: boolean },
): Promise<AdminContentItem[]> {
  if (!options?.force) {
    const cached = peekAdminContent(token);
    if (cached) {
      return cached;
    }
  }

  const data = await listAdminContent(token);
  adminContentCache = { token, items: data.items };
  return data.items;
}

export async function getCachedContentMediaUrl(
  token: string,
  mediaId: number,
): Promise<string> {
  const cached = mediaUrlCache.get(mediaId);
  if (cached) {
    return cached;
  }

  const url = await fetchContentMediaUrl(token, mediaId);
  mediaUrlCache.set(mediaId, url);
  return url;
}
