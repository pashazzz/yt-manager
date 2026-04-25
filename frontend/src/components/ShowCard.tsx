import { useNavigate } from 'react-router-dom'
import type { Show, Episode, Tag } from '../types'
import { thumbForEpisode } from '../utils/thumbnails'

interface Props {
  show: Show
  episodes: Episode[]
  tags: Tag[]
  onDelete: (id: string) => void
  onMove: (showId: string, tagIds: string[]) => void
  listeners?: any
  setNodeRef?: (node: HTMLElement | null) => void
  style?: React.CSSProperties
}

export default function ShowCard({ show, episodes, tags, onDelete, onMove, listeners, setNodeRef, style }: Props) {
  const navigate = useNavigate()
  const firstEp = episodes[0]
  const watched = episodes.filter(e => e.isWatched).length
  const progress = episodes.length > 0 ? (watched / episodes.length) * 100 : 0
  const thumb = firstEp ? thumbForEpisode(firstEp) : ''

  const handleDelete = (e: React.MouseEvent) => {
    e.stopPropagation()
    if (confirm(`Удалить «${show.title}»?`)) onDelete(show.id)
  }

  const handleMove = (e: React.ChangeEvent<HTMLSelectElement>) => {
    e.stopPropagation()
    const targetTagId = e.target.value
    if (!targetTagId) return

    const isAssigned = (show.tagIds || []).includes(targetTagId)
    let newTagIds: string[]
    if (isAssigned) {
      newTagIds = show.tagIds.filter(id => id !== targetTagId)
      // Если это последний тег — можно либо запретить, либо позволить.
      // Обычно лучше оставить хотя бы один, но бэкенд позволяет пустоту.
    } else {
      newTagIds = [...show.tagIds, targetTagId]
    }
    
    onMove(show.id, newTagIds)
    // reset select
    e.target.value = ""
  }

  // Теперь показываем все теги, чтобы можно было как добавлять, так и удалять
  const allTags = tags

  return (
    <div 
      className="show-card" 
      onClick={() => navigate(`/shows/${show.id}`)}
      ref={setNodeRef}
      style={style}
    >
      <div className="show-card-thumb">
        {listeners && (
          <div className="show-card-drag-handle" {...listeners}>
            ☰
          </div>
        )}
        {firstEp && thumb && <img src={thumb} alt={show.title} loading="lazy" />}
        <div className="show-card-overlay">
          <span className="btn-play-overlay">▶ Смотреть</span>
        </div>
      </div>

      {/* Кнопки управления */}
      <div className="show-card-actions">
        {allTags.length > 0 && (
          <div style={{ position: 'relative', display: 'flex' }}>
            <button
              className="btn-card-action"
              title="Управление тегами"
            >
              #
            </button>
            <select
              className="native-dropdown-overlay"
              value=""
              onChange={handleMove}
              onClick={e => e.stopPropagation()}
            >
              <option value="" disabled>Теги...</option>
              {allTags.map(t => {
                const isAssigned = (show.tagIds || []).includes(t.id)
                return (
                  <option key={t.id} value={t.id}>
                    {isAssigned ? '✓ ' : ''}{t.name}
                  </option>
                )
              })}
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
