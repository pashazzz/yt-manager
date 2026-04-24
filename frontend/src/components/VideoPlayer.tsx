import type { Episode } from '../types'
import YouTubePlayer from './players/YouTubePlayer'
import RutubePlayer from './players/RutubePlayer'

interface Props {
  episode: Episode
  onProgressSaved: (id: string, currentTime: number, isWatched: boolean) => void
}

// VideoPlayer — диспетчер, выбирающий конкретную реализацию плеера
// по `episode.provider`. Старые записи без поля provider считаются YouTube.
export default function VideoPlayer({ episode, onProgressSaved }: Props) {
  const provider = episode.provider ?? 'youtube'

  if (provider === 'rutube') {
    return <RutubePlayer episode={episode} onProgressSaved={onProgressSaved} />
  }
  return <YouTubePlayer episode={episode} onProgressSaved={onProgressSaved} />
}
