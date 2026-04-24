import { useRef } from 'react'
import YouTube, { YouTubeEvent } from 'react-youtube'
import type { Episode } from '../../types'
import { usePlayerProgress, isMobile, type PlayerApi } from './usePlayerProgress'

interface Props {
  episode: Episode
  onProgressSaved: (id: string, currentTime: number, isWatched: boolean) => void
}

export default function YouTubePlayer({ episode, onProgressSaved }: Props) {
  const playerApiRef = useRef<PlayerApi | null>(null)
  const ytPlayerRef = useRef<YouTubeEvent['target'] | null>(null)

  const { startHeartbeat, stopHeartbeat, saveNow, handleEnd } = usePlayerProgress(
    episode,
    onProgressSaved,
    playerApiRef,
  )

  const handleReady = (e: YouTubeEvent) => {
    ytPlayerRef.current = e.target
    playerApiRef.current = {
      getCurrentTime: () => e.target.getCurrentTime(),
      getDuration: () => e.target.getDuration(),
    }
    if (!isMobile) startHeartbeat()
  }

  const handlePlay = () => startHeartbeat()

  const handlePause = async () => {
    stopHeartbeat()
    await saveNow()
  }

  const opts = {
    width: '100%',
    height: '100%',
    playerVars: {
      autoplay: 0,
      start: Math.floor(episode.currentTime),
      playsinline: 1,
      rel: 0,
      modestbranding: 1,
    },
  }

  return (
    <div className="player-wrapper">
      <YouTube
        key={episode.id}
        videoId={episode.videoId}
        opts={opts}
        onReady={handleReady}
        onPlay={handlePlay}
        onPause={handlePause}
        onEnd={handleEnd}
      />
    </div>
  )
}
