import { useEffect, useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { api } from '../api/client'
import type { Show, Episode, Section } from '../types'
import ShowCard from '../components/ShowCard'
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
        {items.length === 0 ? (
          <div className="empty-state">
            <div className="empty-state-icon">🎬</div>
            <h3>Раздел пуст</h3>
            <p>Добавь YouTube-плейлист, чтобы смотреть его как сериал</p>
            <button className="btn-primary" onClick={() => setShowModal(true)}>
              Добавить первое шоу
            </button>
          </div>
        ) : (
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
