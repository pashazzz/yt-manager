import type {
  Show,
  ShowDetail,
  Section,
  SectionInfo,
  SectionShows,
  Episode,
  CreateShowResponse,
  ProgressResponse,
} from '../types'

const BASE = '/api/v1'

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(BASE + path, {
    headers: { 'Content-Type': 'application/json', ...init?.headers },
    ...init,
  })
  if (!res.ok) {
    const body = await res.json().catch(() => ({}))
    throw new Error(body.error ?? `HTTP ${res.status}`)
  }
  if (res.status === 204) return undefined as T
  return res.json()
}

export const api = {
  // ── Разделы ───────────────────────────────────────────────────────────────
  getSections: () => request<SectionInfo[]>('/sections'),
  createSection: (name: string) =>
    request<Section>('/sections', {
      method: 'POST',
      body: JSON.stringify({ name }),
    }),
  deleteSection: (id: string) => request<void>(`/sections/${id}/delete`, { method: 'POST' }),
  reorderSections: (orderedIds: string[]) =>
    request<void>('/sections/reorder', {
      method: 'POST',
      body: JSON.stringify({ orderedIds }),
    }),
  updateSectionSettings: (id: string, useThumb: boolean) =>
    request<void>(`/sections/${id}/settings`, {
      method: 'POST',
      body: JSON.stringify({ useThumb }),
    }),
  getSectionShows: (id: string) =>
    request<SectionShows>(`/sections/${id}/shows`),

  // ── Шоу ──────────────────────────────────────────────────────────────────
  getShows: () => request<Show[]>('/shows'),
  createShow: (playlistUrl: string, sectionId: string, title?: string) =>
    request<CreateShowResponse>('/shows', {
      method: 'POST',
      body: JSON.stringify({ playlistUrl, sectionId, title }),
    }),
  reorderShows: (sectionId: string, orderedIds: string[]) =>
    request<void>('/shows/reorder', {
      method: 'POST',
      body: JSON.stringify({ sectionId, orderedIds }),
    }),
  getShow: (id: string) => request<ShowDetail>(`/shows/${id}`),
  deleteShow: (id: string) => request<void>(`/shows/${id}/delete`, { method: 'POST' }),
  moveShow: (id: string, sectionId: string) =>
    request<{ id: string; sectionId: string }>(`/shows/${id}/section`, {
      method: 'POST',
      body: JSON.stringify({ sectionId }),
    }),
  updateReverseOrder: (id: string, reverseOrder: boolean) =>
    request<{ id: string; reverseOrder: boolean }>(`/shows/${id}/reverse`, {
      method: 'POST',
      body: JSON.stringify({ reverseOrder }),
    }),
  addEpisode: (showId: string, url: string) =>
    request<{ episodes: Episode[] }>(`/shows/${showId}/episodes`, {
      method: 'POST',
      body: JSON.stringify({ url }),
    }),
  moveEpisode: (episodeId: string, sectionId: string) =>
    request<void>(`/episodes/${episodeId}/move`, {
      method: 'POST',
      body: JSON.stringify({ sectionId }),
    }),
  addSectionEpisode: (sectionId: string, url: string) =>
    request<{ episodes: Episode[] }>(`/sections/${sectionId}/episodes`, {
      method: 'POST',
      body: JSON.stringify({ url }),
    }),
  reorderEpisodes: (showId: string, orderedIds: string[]) =>
    request<void>(`/shows/${showId}/episodes/reorder`, {
      method: 'POST',
      body: JSON.stringify({ orderedIds }),
    }),

  // ── Эпизоды ──────────────────────────────────────────────────────────────
  saveProgress: (id: string, currentTime: number, isWatched: boolean) =>
    request<ProgressResponse>(`/episodes/${id}/progress`, {
      method: 'POST',
      body: JSON.stringify({ currentTime, isWatched }),
    }),
}
