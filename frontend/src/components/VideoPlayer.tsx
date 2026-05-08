import type { ComponentType } from 'react'
import type { Episode, Provider } from '../types'
import YouTubePlayer from './players/YouTubePlayer'
import RutubePlayer from './players/RutubePlayer'

interface Props {
  episode: Episode
  onProgressSaved: (id: string, currentTime: number, isWatched: boolean) => void
}

type PlayerComponent = ComponentType<Props>

// Реестр плееров: ключ — каноническое имя провайдера (см. internal/providers/*.go).
// При добавлении нового провайдера достаточно расширить эту карту.
const PLAYERS: Partial<Record<Provider, PlayerComponent>> = {
  youtube: YouTubePlayer,
  rutube: RutubePlayer,
}

const FALLBACK: PlayerComponent = YouTubePlayer

// VideoPlayer — диспетчер, выбирающий конкретную реализацию плеера
// по `episode.provider`. Старые записи без поля provider считаются YouTube.
export default function VideoPlayer({ episode, onProgressSaved }: Props) {
  const provider = (episode.provider ?? 'youtube') as Provider
  const Player = PLAYERS[provider] ?? FALLBACK
  return <Player episode={episode} onProgressSaved={onProgressSaved} />
}
