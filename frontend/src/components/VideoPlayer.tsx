import { useState, useRef, useEffect, useCallback } from 'react'
import YouTube, { YouTubeEvent } from 'react-youtube'
import type { Episode } from '../types'
import { api } from '../api/client'

interface Props {
  episode: Episode
  onProgressSaved: (id: string, currentTime: number, isWatched: boolean) => void
}

const HEARTBEAT_INTERVAL = 10_000 // 10 секунд
const WATCHED_THRESHOLD = 0.95     // 95% = просмотрено

// Определяем мобильное устройство один раз
const isMobile = /iPhone|iPad|iPod|Android/i.test(navigator.userAgent)

export default function VideoPlayer({ episode, onProgressSaved }: Props) {
  const playerRef = useRef<YouTubeEvent['target'] | null>(null)
  const heartbeatRef = useRef<ReturnType<typeof setInterval> | null>(null)
  const episodeRef = useRef(episode) // актуальный episode без пересоздания интервала
  const [showMobileOverlay, setShowMobileOverlay] = useState(isMobile)

  // Всегда держим свежий episode в ref
  useEffect(() => { episodeRef.current = episode }, [episode])

  // Отправляем прогресс на бэкенд
  const saveProgress = useCallback(
    async (currentTime: number, isWatched: boolean) => {
      try {
        await api.saveProgress(episodeRef.current.id, currentTime, isWatched)
        onProgressSaved(episodeRef.current.id, currentTime, isWatched)
      } catch {
        // Молча игнорируем сетевые ошибки — данные сохранятся при следующем heartbeat
      }
    },
    [onProgressSaved],
  )

  // Запуск heartbeat
  const startHeartbeat = useCallback(() => {
    if (heartbeatRef.current) clearInterval(heartbeatRef.current)
    heartbeatRef.current = setInterval(async () => {
      const player = playerRef.current
      if (!player) return
      try {
        const currentTime: number = player.getCurrentTime()
        const duration: number = player.getDuration()
        const isWatched = duration > 0 && currentTime / duration >= WATCHED_THRESHOLD
        await saveProgress(currentTime, isWatched)
      } catch {
        // Плеер может не быть готов
      }
    }, HEARTBEAT_INTERVAL)
  }, [saveProgress])

  // Остановка heartbeat
  const stopHeartbeat = useCallback(() => {
    if (heartbeatRef.current) {
      clearInterval(heartbeatRef.current)
      heartbeatRef.current = null
    }
  }, [])

  // Финальное сохранение при уходе со страницы или смене эпизода
  useEffect(() => {
    return () => {
      stopHeartbeat()
      const player = playerRef.current
      if (!player) return
      try {
        const currentTime: number = player.getCurrentTime()
        const duration: number = player.getDuration()
        const isWatched = duration > 0 && currentTime / duration >= WATCHED_THRESHOLD
        // fire-and-forget: используем sendBeacon если возможно
        const body = JSON.stringify({
          currentTime,
          isWatched,
        })
        navigator.sendBeacon?.(
          `/api/v1/episodes/${episodeRef.current.id}/progress`,
          new Blob([body], { type: 'application/json' }),
        )
      } catch {
        //
      }
    }
  }, [stopHeartbeat])

  // Когда плеер готов — перемотать на сохранённую позицию
  const handleReady = (e: YouTubeEvent) => {
    playerRef.current = e.target
    if (episode.currentTime > 5) {
      e.target.seekTo(episode.currentTime, true)
    }
    // На десктопе сразу запускаем heartbeat
    if (!isMobile) {
      startHeartbeat()
    }
  }

  const handlePlay = () => {
    setShowMobileOverlay(false)
    startHeartbeat()
  }

  const handlePause = async () => {
    stopHeartbeat()
    const player = playerRef.current
    if (!player) return
    const currentTime: number = player.getCurrentTime()
    const duration: number = player.getDuration()
    await saveProgress(currentTime, duration > 0 && currentTime / duration >= WATCHED_THRESHOLD)
  }

  const handleEnd = async () => {
    stopHeartbeat()
    await saveProgress(episode.duration, true)
  }

  const handleMobileStart = () => {
    setShowMobileOverlay(false)
    playerRef.current?.playVideo()
    startHeartbeat()
  }

  const opts = {
    width: '100%',
    height: '100%',
    playerVars: {
      autoplay: isMobile ? 0 : 1,
      playsinline: 1,      // iOS: не открывать в системном плеере
      rel: 0,              // Не показывать рекомендации по окончании
      modestbranding: 1,
    },
  }

  return (
    <div className="player-wrapper">
      {showMobileOverlay && (
        <div className="player-overlay-mobile" onClick={handleMobileStart}>
          <button className="btn-start-watching">
            <span className="play-icon">▶</span>
            Начать просмотр
          </button>
        </div>
      )}

      <YouTube
        key={episode.id}         // Перемонтировать при смене эпизода
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
