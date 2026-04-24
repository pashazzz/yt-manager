import { useEffect, useRef } from 'react'
import type { Episode } from '../../types'
import { usePlayerProgress, type PlayerApi } from './usePlayerProgress'

interface Props {
  episode: Episode
  onProgressSaved: (id: string, currentTime: number, isWatched: boolean) => void
}

// Rutube Player API (postMessage) — события приходят в формате
// { type: 'player:<event>', data: {...} } через window.message.
// Команды отправляются в том же формате в iframe.contentWindow.
interface RutubeMessage {
  type?: string
  data?: {
    time?: number
    duration?: number
    state?: string
  }
}

export default function RutubePlayer({ episode, onProgressSaved }: Props) {
  const iframeRef = useRef<HTMLIFrameElement | null>(null)
  const currentTimeRef = useRef<number>(episode.currentTime)
  const durationRef = useRef<number>(episode.duration)
  const targetSeekRef = useRef<number>(episode.currentTime)
  const seekConfirmedRef = useRef<boolean>(episode.currentTime <= 0)
  const playerApiRef = useRef<PlayerApi | null>({
    getCurrentTime: () => currentTimeRef.current,
    getDuration: () => durationRef.current,
  })

  const { startHeartbeat, stopHeartbeat, saveNow, handleEnd } = usePlayerProgress(
    episode,
    onProgressSaved,
    playerApiRef,
  )

  // При смене эпизода сбрасываем цель seek.
  useEffect(() => {
    targetSeekRef.current = episode.currentTime
    seekConfirmedRef.current = episode.currentTime <= 0
    currentTimeRef.current = episode.currentTime
    durationRef.current = episode.duration
  }, [episode.id, episode.currentTime, episode.duration])

  const postCommand = (type: string, data: Record<string, unknown> = {}) => {
    const win = iframeRef.current?.contentWindow
    if (!win) return
    // Rutube принимает как строку, так и объект; шлём строку для совместимости.
    win.postMessage(JSON.stringify({ type, data }), '*')
  }

  const trySeek = () => {
    const target = targetSeekRef.current
    if (target <= 0 || seekConfirmedRef.current) return
    postCommand('player:setCurrentTime', { time: Math.floor(target) })
  }

  useEffect(() => {
    const handler = (ev: MessageEvent) => {
      // Не проверяем ev.source: при cross-origin iframe сравнение с contentWindow
      // иногда не срабатывает. Фильтруем только по префиксу type === 'player:*',
      // чего достаточно, т.к. формат сообщений специфичен для Rutube.
      let msg: RutubeMessage | null = null
      try {
        msg = typeof ev.data === 'string' ? JSON.parse(ev.data) : (ev.data as RutubeMessage)
      } catch {
        return
      }
      if (!msg || typeof msg.type !== 'string' || !msg.type.startsWith('player:')) return

      const { type, data = {} } = msg

      // player:currentTime приходит периодически во время проигрывания —
      // обновляем ref, чтобы heartbeat/saveNow могли синхронно его прочитать.
      if (typeof data.time === 'number') {
        currentTimeRef.current = data.time
        // Если мы всё ещё в районе 0, а должны быть на target — повторяем seek.
        const target = targetSeekRef.current
        if (!seekConfirmedRef.current && target > 1) {
          if (data.time >= target - 2) {
            // Плеер наконец-то оказался у цели — seek отработал.
            seekConfirmedRef.current = true
          } else if (data.time < 2) {
            // Плеер играет с начала, хотя мы просили target — повторяем seek.
            trySeek()
          }
        }
      }
      if (typeof data.duration === 'number' && data.duration > 0) durationRef.current = data.duration

      switch (type) {
        case 'player:ready':
        case 'player:durationChange': {
          // Плеер готов — пробуем восстановить позицию.
          trySeek()
          break
        }
        case 'player:changeState': {
          const state = data.state
          if (state === 'playing') {
            startHeartbeat()
            // На всякий случай ретраим seek — при первом play плеер уже готов к нему.
            trySeek()
          } else if (state === 'paused' || state === 'stopped') {
            stopHeartbeat()
            void saveNow()
          }
          break
        }
        case 'player:playComplete': {
          // Настоящее окончание видео — помечаем как просмотренное.
          void handleEnd()
          break
        }
      }
    }

    window.addEventListener('message', handler)
    return () => window.removeEventListener('message', handler)
  }, [startHeartbeat, stopHeartbeat, saveNow, handleEnd])

  const src = `https://rutube.ru/play/embed/${encodeURIComponent(episode.videoId)}`

  return (
    <div className="player-wrapper">
      <iframe
        key={episode.id}
        ref={iframeRef}
        src={src}
        width="100%"
        height="100%"
        frameBorder={0}
        allow="clipboard-write; autoplay"
        allowFullScreen
        title={episode.title}
      />
    </div>
  )
}
