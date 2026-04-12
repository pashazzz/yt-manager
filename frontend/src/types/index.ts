export interface Show {
  id: string
  title: string
  playlistUrl: string
  ownerId: string
  sectionId: string
  reverseOrder: boolean
  createdAt: string
}

export interface Episode {
  id: string
  showId: string
  videoId: string
  title: string
  duration: number
  currentTime: number
  isWatched: boolean
  orderIndex: number
}

export interface Section {
  id: string
  name: string
  ownerId: string
  isDefault: boolean
  orderIndex: number
  createdAt: string
}

export interface ShowDetail {
  show: Show
  episodes: Episode[]
}

export interface SectionShows {
  section: Section
  shows: Show[]
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
