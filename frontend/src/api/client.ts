import type {
  Show,
  ShowDetail,
  Section,
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
  getSections: () => request<Section[]>('/sections'),
  createSection: (name: string) =>
    request<Section>('/sections', { method: 'POST', body: JSON.stringify({ name }) }),
  deleteSection: (id: string) =>
    request<void>(`/sections/${id}`, { method: 'DELETE' }),
  getSectionShows: (id: string) =>
    request<SectionShows>(`/sections/${id}/shows`),

  // ── Шоу ──────────────────────────────────────────────────────────────────
  getShows: () => request<Show[]>('/shows'),
  createShow: (playlistUrl: string, sectionId: string, title?: string) =>
    request<CreateShowResponse>('/shows', {
      method: 'POST',
      body: JSON.stringify({ playlistUrl, sectionId, title }),
    }),
  getShow: (id: string) => request<ShowDetail>(`/shows/${id}`),
  deleteShow: (id: string) => request<void>(`/shows/${id}`, { method: 'DELETE' }),
  moveShow: (id: string, sectionId: string) =>
    request<{ id: string; sectionId: string }>(`/shows/${id}/section`, {
      method: 'PATCH',
      body: JSON.stringify({ sectionId }),
    }),
  updateReverseOrder: (id: string, reverseOrder: boolean) =>
    request<{ id: string; reverseOrder: boolean }>(`/shows/${id}/reverse`, {
      method: 'PATCH',
      body: JSON.stringify({ reverseOrder }),
    }),
  addEpisode: (showId: string, url: string) =>
    request<{ episodes: Episode[] }>(`/shows/${showId}/episodes`, {
      method: 'POST',
      body: JSON.stringify({ url }),
    }),

  // ── Эпизоды ──────────────────────────────────────────────────────────────
  saveProgress: (id: string, currentTime: number, isWatched: boolean) =>
    request<ProgressResponse>(`/episodes/${id}/progress`, {
      method: 'POST',
      body: JSON.stringify({ currentTime, isWatched }),
    }),
}
