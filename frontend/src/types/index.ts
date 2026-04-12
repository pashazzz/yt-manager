export interface Show {
  id: string
  title: string
  playlistUrl: string
  ownerId: string
  createdAt: string
}

export interface Episode {
  id: string
  showId: string
  videoId: string
  title: string
  duration: number    // секунды
  currentTime: number // прогресс в секундах
  isWatched: boolean
  orderIndex: number
}

export interface ShowDetail {
  show: Show
  episodes: Episode[]
}

export interface CreateShowResponse {
  show: Show
  episodeCount: number
}

export interface ProgressResponse {
  id: string
  currentTime: number
  isWatched: boolean
}
