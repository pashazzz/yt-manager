import { useEffect, useState, useCallback } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { api } from '../api/client'
import type { Show, Episode } from '../types'
import VideoPlayer from '../components/VideoPlayer'
import EpisodeList from '../components/EpisodeList'

function fmtDuration(sec: number): string {
  if (!sec) return ''
  const h = Math.floor(sec / 3600)
  const m = Math.floor((sec % 3600) / 60)
  const s = Math.floor(sec % 60)
  if (h > 0) return `${h}:${String(m).padStart(2, '0')}:${String(s).padStart(2, '0')}`
  return `${m}:${String(s).padStart(2, '0')}`
}

export default function ShowPage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()

  const [show, setShow] = useState<Show | null>(null)
  const [episodes, setEpisodes] = useState<Episode[]>([])
  const [currentEpisode, setCurrentEpisode] = useState<Episode | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => {
    if (!id) return
    api.getShow(id)
      .then(detail => {
        setShow(detail.show)
        setEpisodes(detail.episodes)
        // Выбираем первый непросмотренный эпизод, или первый из всех
        const resume = detail.episodes.find(e => !e.isWatched && e.currentTime > 0)
          ?? detail.episodes.find(e => !e.isWatched)
          ?? detail.episodes[0]
        setCurrentEpisode(resume ?? null)
      })
      .catch(err => setError(err instanceof Error ? err.message : 'Ошибка загрузки'))
      .finally(() => setLoading(false))
  }, [id])

  // Обновляем локальный state после heartbeat, чтобы прогресс в сайдбаре обновлялся
  const handleProgressSaved = useCallback(
    (epId: string, currentTime: number, isWatched: boolean) => {
      setEpisodes(prev =>
        prev.map(ep =>
          ep.id === epId ? { ...ep, currentTime, isWatched } : ep,
        ),
      )
    },
    [],
  )

  if (loading) {
    return (
      <div className="page-loader">
        <div className="spinner" />
      </div>
    )
  }

  if (error || !show) {
    return (
      <div className="page-error">
        <span>⚠ {error || 'Шоу не найдено'}</span>
        <button className="btn-primary" onClick={() => navigate('/')}>← На главную</button>
      </div>
    )
  }

  return (
    <div className="show-page">
      {/* ── Header ── */}
      <header className="show-page-header">
        <button className="btn-back" onClick={() => navigate('/')}>
          ← Назад
        </button>
        <h1 className="show-page-title">{show.title}</h1>
      </header>

      <div className="show-page-body">
        {/* ── Player section ── */}
        <section className="player-section">
          {currentEpisode ? (
            <>
              <VideoPlayer
                episode={currentEpisode}
                onProgressSaved={handleProgressSaved}
              />
              <div className="episode-meta">
                <div className="episode-meta-title">{currentEpisode.title}</div>
                <div className="episode-meta-sub">
                  <span>Эпизод {currentEpisode.orderIndex + 1}</span>
                  {currentEpisode.duration > 0 && (
                    <span>{fmtDuration(currentEpisode.duration)}</span>
                  )}
                  {currentEpisode.isWatched && <span style={{ color: 'var(--success)' }}>✓ Просмотрено</span>}
                </div>
              </div>
            </>
          ) : (
            <div className="empty-state">
              <div className="empty-state-icon">📭</div>
              <h3>Нет эпизодов</h3>
              <p>В этом плейлисте не удалось загрузить видео</p>
            </div>
          )}
        </section>

        {/* ── Episode list ── */}
        <EpisodeList
          episodes={episodes}
          currentId={currentEpisode?.id ?? ''}
          onSelect={setCurrentEpisode}
        />
      </div>
    </div>
  )
}
