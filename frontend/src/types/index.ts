export interface Show {
  id: string
  title: string
  playlistUrl: string
  ownerId: string
  tagIds: string[]
  reverseOrder: boolean
  isSingles: boolean
  orderIndex: number
  createdAt: string
}

export type Provider = 'youtube' | 'rutube' | (string & {})

export interface Episode {
  id: string
  showId: string
  provider?: Provider
  videoId: string
  title: string
  duration: number
  currentTime: number
  isWatched: boolean
  orderIndex: number
  tagIds?: string[]
  thumbnailUrl?: string
}

export interface Tag {
  id: string
  name: string
  ownerId: string
  isDefault: boolean
  orderIndex: number
  useThumb: boolean
  createdAt: string
}

export interface TagInfo extends Tag {
  showCount: number
  episodeCount: number
  firstVideoId: string
  firstProvider?: Provider
  firstThumbnailUrl?: string
}

export interface ShowDetail {
  show: Show
  episodes: Episode[]
}

export interface TagItems {
  tag: Tag
  shows: Show[]
  singlesShow?: Show
  singlesEpisodes?: Episode[]
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
