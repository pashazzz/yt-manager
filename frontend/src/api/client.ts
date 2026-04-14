import type {
  Show,
  ShowDetail,
  Tag,
  TagInfo,
  TagItems,
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
  // ── Теги ──────────────────────────────────────────────────────────────────
  getTags: () => request<TagInfo[]>('/tags'),
  createTag: (name: string) =>
    request<Tag>('/tags', {
      method: 'POST',
      body: JSON.stringify({ name }),
    }),
  deleteTag: (id: string) => request<void>(`/tags/${id}/delete`, { method: 'POST' }),
  reorderTags: (orderedIds: string[]) =>
    request<void>('/tags/reorder', {
      method: 'POST',
      body: JSON.stringify({ orderedIds }),
    }),
  updateTagSettings: (id: string, useThumb: boolean) =>
    request<void>(`/tags/${id}/settings`, {
      method: 'POST',
      body: JSON.stringify({ useThumb }),
    }),
  getTagItems: (id: string) =>
    request<TagItems>(`/tags/${id}/items`),

  // ── Шоу ──────────────────────────────────────────────────────────────────
  getShows: () => request<Show[]>('/shows'),
  createShow: (playlistUrl: string, tagIds: string[], title?: string) =>
    request<CreateShowResponse>('/shows', {
      method: 'POST',
      body: JSON.stringify({ playlistUrl, tagIds, title }),
    }),
  reorderShows: (tagId: string, orderedIds: string[]) =>
    request<void>('/shows/reorder', {
      method: 'POST',
      body: JSON.stringify({ tagId, orderedIds }),
    }),
  getShow: (id: string) => request<ShowDetail>(`/shows/${id}`),
  deleteShow: (id: string) => request<void>(`/shows/${id}/delete`, { method: 'POST' }),
  updateShowTags: (id: string, tagIds: string[]) =>
    request<{ id: string; tagIds: string[] }>(`/shows/${id}/tags`, {
      method: 'POST',
      body: JSON.stringify({ tagIds }),
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
  updateEpisodeTags: (episodeId: string, tagIds: string[]) =>
    request<void>(`/episodes/${episodeId}/tags`, {
      method: 'POST',
      body: JSON.stringify({ tagIds }),
    }),
  addTagEpisode: (tagId: string, url: string) =>
    request<{ episodes: Episode[] }>(`/tags/${tagId}/episodes`, {
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
  deleteEpisode: (id: string) =>
    request<void>(`/episodes/${id}/delete`, {
      method: 'POST',
    }),
  listEpisodes: () => request<Episode[]>('/episodes'),
}
