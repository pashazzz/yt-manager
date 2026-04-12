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
        
        const list = detail.show.reverseOrder ? [...detail.episodes].reverse() : detail.episodes
        const resume = list.find(e => !e.isWatched && e.currentTime > 0)
          ?? list.find(e => !e.isWatched)
          ?? list[0]
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

  const handleToggleWatched = async (ep: Episode) => {
    const newIsWatched = !ep.isWatched
    const newTime = newIsWatched ? ep.duration : 0
    try {
      await api.saveProgress(ep.id, newTime, newIsWatched)
      handleProgressSaved(ep.id, newTime, newIsWatched)
    } catch {
      // Ignored
    }
  }

  const handleToggleReverse = async () => {
    if (!show) return
    try {
      const newRev = !show.reverseOrder
      await api.updateReverseOrder(show.id, newRev)
      setShow(s => s ? { ...s, reverseOrder: newRev } : s)
    } catch (err: unknown) {
      alert(err instanceof Error ? err.message : 'Ошибка')
    }
  }

  const handleAddVideo = async () => {
    const videoUrl = prompt('Введите ссылку на YouTube видео (например https://youtube.com/watch?v=...):')
    if (!videoUrl?.trim() || !show) return
    try {
      const res = await api.addEpisode(show.id, videoUrl.trim())
      setEpisodes(prev => [...prev, ...res.episodes])
    } catch (err: unknown) {
      alert(err instanceof Error ? err.message : 'Ошибка добавления видео')
    }
  }

  const handleReorder = (newEpisodes: Episode[]) => {
    setEpisodes(newEpisodes)
    if (show) {
      api.reorderEpisodes(show.id, newEpisodes.map(e => e.id)).catch(console.error)
    }
  }

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
        {show.playlistUrl !== '' && (
          <button 
            className="btn-ghost" 
            onClick={handleToggleReverse}
            title="Изменить порядок воспроизведения"
            style={{ marginLeft: 'auto', padding: '4px 12px', fontSize: '0.85rem' }}
          >
            {show.reverseOrder ? 'Сначала новые ▼' : 'Сначала старые ▲'}
          </button>
        )}
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
          episodes={show.playlistUrl === '' ? episodes : (show.reverseOrder ? [...episodes].reverse() : episodes)}
          currentId={currentEpisode?.id ?? ''}
          onSelect={setCurrentEpisode}
          onAddVideo={handleAddVideo}
          onToggleWatched={handleToggleWatched}
          isReorderable={show.playlistUrl === ''}
          onReorder={handleReorder}
        />
      </div>
    </div>
  )
}
