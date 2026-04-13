import { useNavigate } from 'react-router-dom'
import type { Show, Episode, Section } from '../types'

interface Props {
  show: Show
  episodes: Episode[]
  sections: Section[]
  onDelete: (id: string) => void
  onMove: (showId: string, sectionId: string) => void
}

function thumbUrl(videoId: string | undefined) {
  return videoId ? `https://img.youtube.com/vi/${videoId}/mqdefault.jpg` : ''
}

export default function ShowCard({ show, episodes, sections, onDelete, onMove }: Props) {
  const navigate = useNavigate()
  const firstEp = episodes[0]
  const watched = episodes.filter(e => e.isWatched).length
  const progress = episodes.length > 0 ? (watched / episodes.length) * 100 : 0

  const handleDelete = (e: React.MouseEvent) => {
    e.stopPropagation()
    if (confirm(`Удалить «${show.title}»?`)) onDelete(show.id)
  }

  const handleMove = (e: React.ChangeEvent<HTMLSelectElement>) => {
    e.stopPropagation()
    const targetSection = e.target.value
    if (targetSection) onMove(show.id, targetSection)
    // reset select
    e.target.value = ""
  }

  const otherSections = sections.filter(s => s.id !== show.sectionId)

  return (
    <div className="show-card" onClick={() => navigate(`/shows/${show.id}`)}>
      <div className="show-card-thumb">
        {firstEp && <img src={thumbUrl(firstEp.videoId)} alt={show.title} loading="lazy" />}
        <div className="show-card-overlay">
          <span className="btn-play-overlay">▶ Смотреть</span>
        </div>
      </div>

      {/* Кнопки управления */}
      <div className="show-card-actions">
        {otherSections.length > 0 && (
          <div style={{ position: 'relative', display: 'flex' }}>
            <button
              className="btn-card-action"
              title="Переместить в раздел"
            >
              ⋯
            </button>
            <select
              className="native-dropdown-overlay"
              value=""
              onChange={handleMove}
              onClick={e => e.stopPropagation()}
            >
              <option value="" disabled>Переместить в</option>
              {otherSections.map(s => (
                <option key={s.id} value={s.id}>{s.name}</option>
              ))}
            </select>
          </div>
        )}
        <button className="btn-card-action btn-card-delete" onClick={handleDelete} title="Удалить">
          ✕
        </button>
      </div>

      <div className="show-card-info">
        <div className="show-card-title">{show.title}</div>
        <div className="show-card-meta">
          <span>{episodes.length} эп.</span>
          {watched > 0 && <span className="badge-watched">{watched} просм.</span>}
        </div>
        {progress > 0 && (
          <div className="show-card-progress">
            <div className="show-card-progress-bar" style={{ width: `${progress}%` }} />
          </div>
        )}
      </div>
    </div>
  )
}
