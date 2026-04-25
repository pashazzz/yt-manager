import type { Provider } from '../types'

// thumbForEpisode возвращает URL превью эпизода. Стратегия:
//   - если есть сохранённый thumbnailUrl (приходит из yt-dlp при добавлении) — используем его;
//   - иначе для YouTube строим URL по videoId — это работает для всех старых эпизодов;
//   - иначе пустая строка (превью неизвестно).
export function thumbForEpisode(ep: {
  provider?: Provider
  videoId: string
  thumbnailUrl?: string
}): string {
  if (ep.thumbnailUrl) return ep.thumbnailUrl
  const provider = ep.provider || 'youtube'
  if (provider === 'youtube' && ep.videoId) {
    return `https://img.youtube.com/vi/${ep.videoId}/mqdefault.jpg`
  }
  return ''
}

// thumbForVideo — версия для случаев, когда у нас есть только videoId+provider+url
// (например, статистика тега из бэка), без полного объекта Episode.
export function thumbForVideo(opts: {
  provider?: Provider
  videoId?: string
  thumbnailUrl?: string
}): string {
  if (opts.thumbnailUrl) return opts.thumbnailUrl
  const provider = opts.provider || 'youtube'
  if (provider === 'youtube' && opts.videoId) {
    return `https://img.youtube.com/vi/${opts.videoId}/mqdefault.jpg`
  }
  return ''
}
