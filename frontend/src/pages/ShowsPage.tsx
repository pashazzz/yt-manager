import { useEffect, useState } from 'react'
import { api } from '../api/client'
import type { Show, Episode } from '../types'
import ShowCard from '../components/ShowCard'
import AddShowModal from '../components/AddShowModal'

interface ShowWithEpisodes {
  show: Show
  episodes: Episode[]
}

export default function ShowsPage() {
  const [items, setItems] = useState<ShowWithEpisodes[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [showModal, setShowModal] = useState(false)

  useEffect(() => {
    loadShows()
  }, [])

  async function loadShows() {
    try {
      setLoading(true)
      const shows = await api.getShows()
      // Загружаем эпизоды каждого шоу параллельно
      const details = await Promise.all(
        shows.map(s => api.getShow(s.id).catch(() => ({ show: s, episodes: [] }))),
      )
      setItems(details.map(d => ({ show: d.show, episodes: d.episodes })))
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : 'Ошибка загрузки')
    } finally {
      setLoading(false)
    }
  }

  const handleCreated = async (show: Show) => {
    // Подгружаем эпизоды нового шоу
    const detail = await api.getShow(show.id).catch(() => ({ show, episodes: [] }))
    setItems(prev => [{ show: detail.show, episodes: detail.episodes }, ...prev])
  }

  const handleDelete = async (id: string) => {
    try {
      await api.deleteShow(id)
      setItems(prev => prev.filter(i => i.show.id !== id))
    } catch (err: unknown) {
      alert(err instanceof Error ? err.message : 'Ошибка удаления')
    }
  }

  if (loading) {
    return (
      <div className="page-loader">
        <div className="spinner" />
      </div>
    )
  }

  if (error) {
    return (
      <div className="page-error">
        <span>⚠ {error}</span>
        <button className="btn-primary" onClick={loadShows}>Повторить</button>
      </div>
    )
  }

  return (
    <div className="shows-page">
      <header className="app-header">
        <span className="app-logo">▶ YT Manager</span>
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
            <h3>Пока нет ни одного шоу</h3>
            <p>Добавь YouTube-плейлист, чтобы смотреть его как сериал с отслеживанием прогресса</p>
            <button className="btn-primary" onClick={() => setShowModal(true)}>
              Добавить первое шоу
            </button>
          </div>
        ) : (
          <>
            <h2 className="shows-section-title">Мои шоу</h2>
            <div className="shows-grid">
              {items.map(({ show, episodes }) => (
                <ShowCard
                  key={show.id}
                  show={show}
                  episodes={episodes}
                  onDelete={handleDelete}
                />
              ))}
            </div>
          </>
        )}
      </main>

      {showModal && (
        <AddShowModal onCreated={handleCreated} onClose={() => setShowModal(false)} />
      )}
    </div>
  )
}
