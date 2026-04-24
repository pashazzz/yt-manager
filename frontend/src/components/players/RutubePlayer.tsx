import { useEffect, useRef } from 'react'
import type { Episode } from '../../types'
import { usePlayerProgress, isMobile, type PlayerApi } from './usePlayerProgress'

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
  const seekedRef = useRef(false)
  const playerApiRef = useRef<PlayerApi | null>({
    getCurrentTime: () => currentTimeRef.current,
    getDuration: () => durationRef.current,
  })

  const { startHeartbeat, stopHeartbeat, saveNow, handleEnd } = usePlayerProgress(
    episode,
    onProgressSaved,
    playerApiRef,
  )

  const postCommand = (type: string, data: Record<string, unknown> = {}) => {
    const win = iframeRef.current?.contentWindow
    if (!win) return
    // Rutube принимает как строку, так и объект; шлём строку для совместимости.
    win.postMessage(JSON.stringify({ type, data }), '*')
  }

  useEffect(() => {
    const handler = (ev: MessageEvent) => {
      // Игнорируем сообщения не от Rutube iframe
      if (!iframeRef.current || ev.source !== iframeRef.current.contentWindow) return

      let msg: RutubeMessage | null = null
      try {
        msg = typeof ev.data === 'string' ? JSON.parse(ev.data) : (ev.data as RutubeMessage)
      } catch {
        return
      }
      if (!msg || typeof msg.type !== 'string' || !msg.type.startsWith('player:')) return

      const { type, data = {} } = msg

      if (typeof data.time === 'number') currentTimeRef.current = data.time
      if (typeof data.duration === 'number' && data.duration > 0) durationRef.current = data.duration

      switch (type) {
        case 'player:ready':
        case 'player:init': {
          // Rutube сообщает длительность в отдельном событии; пытаемся перемотать
          // как только она появится.
          if (!seekedRef.current && episode.currentTime > 0) {
            postCommand('player:setCurrentTime', { time: Math.floor(episode.currentTime) })
            seekedRef.current = true
          }
          if (!isMobile) startHeartbeat()
          break
        }
        case 'player:durationChange': {
          if (!seekedRef.current && episode.currentTime > 0) {
            postCommand('player:setCurrentTime', { time: Math.floor(episode.currentTime) })
            seekedRef.current = true
          }
          break
        }
        case 'player:playing':
        case 'player:play': {
          startHeartbeat()
          break
        }
        case 'player:paused':
        case 'player:pause': {
          stopHeartbeat()
          void saveNow()
          break
        }
        case 'player:stopped':
        case 'player:ended': {
          void handleEnd()
          break
        }
        case 'player:changeState': {
          const state = data.state
          if (state === 'playing') startHeartbeat()
          else if (state === 'paused') {
            stopHeartbeat()
            void saveNow()
          } else if (state === 'stopped' || state === 'ended') {
            void handleEnd()
          }
          break
        }
      }
    }

    window.addEventListener('message', handler)
    return () => window.removeEventListener('message', handler)
  }, [episode.currentTime, startHeartbeat, stopHeartbeat, saveNow, handleEnd])

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
