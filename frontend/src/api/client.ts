import type { Show, ShowDetail, CreateShowResponse, ProgressResponse } from '../types'

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
  /** Список шоу текущего профиля */
  getShows: () => request<Show[]>('/shows'),

  /** Добавить плейлист по URL */
  createShow: (playlistUrl: string) =>
    request<CreateShowResponse>('/shows', {
      method: 'POST',
      body: JSON.stringify({ playlistUrl }),
    }),

  /** Шоу + все его эпизоды */
  getShow: (id: string) => request<ShowDetail>(`/shows/${id}`),

  /** Удалить шоу */
  deleteShow: (id: string) => request<void>(`/shows/${id}`, { method: 'DELETE' }),

  /** Сохранить прогресс просмотра */
  saveProgress: (id: string, currentTime: number, isWatched: boolean) =>
    request<ProgressResponse>(`/episodes/${id}/progress`, {
      method: 'POST',
      body: JSON.stringify({ currentTime, isWatched }),
    }),
}
