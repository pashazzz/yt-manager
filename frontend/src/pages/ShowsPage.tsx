import { useEffect, useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { api } from '../api/client'
import type { Show, Episode, Section } from '../types'
import ShowCard from '../components/ShowCard'
import EpisodeList from '../components/EpisodeList'
import AddShowModal from '../components/AddShowModal'

interface ShowWithEpisodes {
  show: Show
  episodes: Episode[]
}

export default function ShowsPage() {
  const { sectionId } = useParams<{ sectionId: string }>()
  const navigate = useNavigate()

  const [section, setSection] = useState<Section | null>(null)
  const [items, setItems] = useState<ShowWithEpisodes[]>([])
  const [sections, setSections] = useState<Section[]>([]) // все разделы (для модалки и move)
  const [singlesShow, setSinglesShow] = useState<Show | null>(null)
  const [singlesEpisodes, setSinglesEpisodes] = useState<Episode[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [showModal, setShowModal] = useState(false)

  useEffect(() => {
    if (!sectionId) return
    load()
  }, [sectionId])

  async function load() {
    try {
      setLoading(true)
      const [sectionData, allSections] = await Promise.all([
        api.getSectionShows(sectionId!),
        api.getSections(),
      ])
      setSection(sectionData.section)
      setSections(allSections)
      setSinglesShow(sectionData.singlesShow ?? null)
      setSinglesEpisodes(sectionData.singlesEpisodes ?? [])

      // Загружаем эпизоды для каждого шоу (для прогресса на карточках)
      const withEpisodes = await Promise.all(
        sectionData.shows.map(async show => {
          const detail = await api.getShow(show.id).catch(() => ({ show, episodes: [] }))
          return { show: detail.show, episodes: detail.episodes }
        }),
      )
      setItems(withEpisodes)
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : 'Ошибка загрузки')
    } finally {
      setLoading(false)
    }
  }

  const handleCreated = async (show: Show) => {
    const detail = await api.getShow(show.id).catch(() => ({ show, episodes: [] }))
    // Показываем шоу только если оно в текущем разделе
    if (show.sectionId === sectionId) {
      setItems(prev => [{ show: detail.show, episodes: detail.episodes }, ...prev])
    }
  }

  const handleDelete = async (id: string) => {
    try {
      await api.deleteShow(id)
      setItems(prev => prev.filter(i => i.show.id !== id))
    } catch (err: unknown) {
      alert(err instanceof Error ? err.message : 'Ошибка удаления')
    }
  }

  const handleMove = async (showId: string, targetSectionId: string) => {
    try {
      await api.moveShow(showId, targetSectionId)
      // Убираем шоу из текущего раздела
      setItems(prev => prev.filter(i => i.show.id !== showId))
    } catch (err: unknown) {
      alert(err instanceof Error ? err.message : 'Ошибка перемещения')
    }
  }

  const handleAddSingleVideo = async () => {
    const videoUrl = prompt('Введите ссылку на YouTube видео:')
    if (!videoUrl?.trim() || !sectionId) return
    try {
      const res = await api.addSectionEpisode(sectionId, videoUrl.trim())
      setSinglesEpisodes(prev => [...prev, ...res.episodes])
      if (!singlesShow) load() // перезагрузка, если скрытое шоу только что создано
    } catch (err: unknown) {
      alert(err instanceof Error ? err.message : 'Ошибка')
    }
  }

  const handleReorderSingles = (newEpisodes: Episode[]) => {
    setSinglesEpisodes(newEpisodes)
    if (singlesShow) {
      api.reorderEpisodes(singlesShow.id, newEpisodes.map(e => e.id)).catch(console.error)
    }
  }

  const handleToggleWatchedSingle = async (ep: Episode) => {
    const newIsWatched = !ep.isWatched
    const newTime = newIsWatched ? ep.duration : 0
    try {
      await api.saveProgress(ep.id, newTime, newIsWatched)
      setSinglesEpisodes(prev => prev.map(e => e.id === ep.id ? { ...e, isWatched: newIsWatched, currentTime: newTime } : e))
    } catch {
      // ignore
    }
  }

  const defaultSection = sections.find(s => s.isDefault)

  if (loading) return <div className="page-loader"><div className="spinner" /></div>
  if (error) return (
    <div className="page-error">
      <span>⚠ {error}</span>
      <button className="btn-primary" onClick={() => navigate('/')}>← На главную</button>
    </div>
  )

  return (
    <div className="shows-page">
      <header className="app-header">
        <button className="btn-back-inline" onClick={() => navigate('/')}>← Разделы</button>
        <span className="app-logo" style={{ fontSize: '1rem' }}>{section?.name}</span>
        <button className="btn-add" onClick={() => setShowModal(true)}>
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5">
            <path d="M12 5v14M5 12h14" strokeLinecap="round" />
          </svg>
          Добавить шоу
        </button>
      </header>

      <main className="shows-content">
        {items.length === 0 && singlesEpisodes.length === 0 ? (
          <div className="empty-state">
            <div className="empty-state-icon">🎬</div>
            <h3>Раздел пуст</h3>
            <p>Добавь YouTube-плейлист или одиночные видео</p>
            <div style={{ display: 'flex', gap: '12px', marginTop: '8px' }}>
              <button className="btn-primary" onClick={() => setShowModal(true)}>
                Добавить шоу
              </button>
              <button className="btn-ghost" onClick={handleAddSingleVideo} style={{ border: '1px solid rgba(255,255,255,0.1)' }}>
                Добавить видео
              </button>
            </div>
          </div>
        ) : (
          <>
            {items.length > 0 && (
              <>
                <h2 className="shows-section-title">
                  Шоу · {items.length}
                </h2>
                <div className="shows-grid">
                  {items.map(({ show, episodes }) => (
                    <ShowCard
                      key={show.id}
                      show={show}
                      episodes={episodes}
                      sections={sections}
                      onDelete={handleDelete}
                      onMove={handleMove}
                    />
                  ))}
                </div>
              </>
            )}

            <div style={{ marginTop: items.length > 0 ? '48px' : '0' }}>
              <h2 className="shows-section-title">Отдельные видео</h2>
              <EpisodeList
                episodes={singlesEpisodes}
                currentId=""
                onSelect={(ep) => singlesShow && navigate(`/shows/${singlesShow.id}?episode=${ep.id}`)}
                onToggleWatched={handleToggleWatchedSingle}
                onAddVideo={handleAddSingleVideo}
                isReorderable={true}
                onReorder={handleReorderSingles}
                variant="inline"
              />
            </div>
          </>
        )}
      </main>

      {showModal && (
        <AddShowModal
          sections={sections}
          defaultSectionId={sectionId ?? defaultSection?.id ?? ''}
          onCreated={handleCreated}
          onClose={() => setShowModal(false)}
        />
      )}
    </div>
  )
}
