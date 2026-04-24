import { useCallback, useEffect, useRef } from 'react'
import type { MutableRefObject } from 'react'
import { api } from '../../api/client'
import type { Episode } from '../../types'

// PlayerApi — минимальный интерфейс, который должна предоставлять
// каждая конкретная реализация плеера (YouTube, Rutube, …). Методы
// должны возвращать "последнее известное" значение синхронно.
export interface PlayerApi {
  getCurrentTime(): number
  getDuration(): number
}

const HEARTBEAT_INTERVAL = 10_000
const WATCHED_THRESHOLD = 0.95

// usePlayerProgress инкапсулирует общую логику отслеживания прогресса
// (heartbeat, сохранение при паузе/конце, sendBeacon на unmount) независимо
// от конкретного плеера. Плеер даёт хуку ref на свой PlayerApi и вызывает
// возвращаемые колбэки в соответствующих событиях.
export function usePlayerProgress(
  episode: Episode,
  onProgressSaved: (id: string, currentTime: number, isWatched: boolean) => void,
  playerApiRef: MutableRefObject<PlayerApi | null>,
) {
  const heartbeatRef = useRef<ReturnType<typeof setInterval> | null>(null)
  const episodeRef = useRef(episode)

  useEffect(() => {
    episodeRef.current = episode
  }, [episode])

  const saveProgress = useCallback(
    async (currentTime: number, isWatched: boolean) => {
      try {
        await api.saveProgress(episodeRef.current.id, currentTime, isWatched)
        onProgressSaved(episodeRef.current.id, currentTime, isWatched)
      } catch {
        // Молча игнорируем сетевые ошибки — повторим в следующий heartbeat
      }
    },
    [onProgressSaved],
  )

  const snapshot = useCallback((): { currentTime: number; isWatched: boolean } | null => {
    const api = playerApiRef.current
    if (!api) return null
    try {
      const currentTime = api.getCurrentTime()
      const duration = api.getDuration()
      const isWatched = duration > 0 && currentTime / duration >= WATCHED_THRESHOLD
      return { currentTime, isWatched }
    } catch {
      return null
    }
  }, [playerApiRef])

  const stopHeartbeat = useCallback(() => {
    if (heartbeatRef.current) {
      clearInterval(heartbeatRef.current)
      heartbeatRef.current = null
    }
  }, [])

  const startHeartbeat = useCallback(() => {
    stopHeartbeat()
    heartbeatRef.current = setInterval(() => {
      const snap = snapshot()
      if (!snap) return
      void saveProgress(snap.currentTime, snap.isWatched)
    }, HEARTBEAT_INTERVAL)
  }, [saveProgress, snapshot, stopHeartbeat])

  const saveNow = useCallback(async () => {
    const snap = snapshot()
    if (!snap) return
    await saveProgress(snap.currentTime, snap.isWatched)
  }, [saveProgress, snapshot])

  const handleEnd = useCallback(async () => {
    stopHeartbeat()
    await saveProgress(episodeRef.current.duration, true)
  }, [saveProgress, stopHeartbeat])

  // Финальное сохранение при размонтировании / смене эпизода.
  useEffect(() => {
    return () => {
      stopHeartbeat()
      const snap = snapshot()
      if (!snap) return
      try {
        const body = JSON.stringify({
          currentTime: snap.currentTime,
          isWatched: snap.isWatched,
        })
        navigator.sendBeacon?.(
          `/api/v1/episodes/${episodeRef.current.id}/progress`,
          new Blob([body], { type: 'application/json' }),
        )
      } catch {
        //
      }
    }
  }, [snapshot, stopHeartbeat])

  return { startHeartbeat, stopHeartbeat, saveNow, handleEnd }
}

// isMobile — грубое определение мобильного устройства, чтобы корректно
// откладывать запуск heartbeat до первого явного воспроизведения.
export const isMobile = /iPhone|iPad|iPod|Android/i.test(
  typeof navigator !== 'undefined' ? navigator.userAgent : '',
)
