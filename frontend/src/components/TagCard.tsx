import { useNavigate } from 'react-router-dom'
import type { TagInfo } from '../types'

// Набор градиентов для карточек — назначается по индексу
const GRADIENTS = [
  'linear-gradient(135deg, #6c47ff, #ff4778)',
  'linear-gradient(135deg, #00e5a0, #6c47ff)',
  'linear-gradient(135deg, #ff4778, #ffb347)',
  'linear-gradient(135deg, #47c8ff, #6c47ff)',
  'linear-gradient(135deg, #ffb347, #ff4778)',
  'linear-gradient(135deg, #6c47ff, #00e5a0)',
]

interface Props {
  tag: TagInfo
  index: number
  onDelete: (id: string) => void
  onToggleThumb: (id: string, useThumb: boolean) => void
}

export default function TagCard({ tag, index, onDelete, onToggleThumb }: Props) {
  const navigate = useNavigate()
  const gradient = GRADIENTS[index % GRADIENTS.length]

  const handleDelete = (e: React.MouseEvent) => {
    e.stopPropagation()
    if (confirm(`Удалить тег «${tag.name}»?\nШоу и видео останутся, но тег будет снят.`)) {
      onDelete(tag.id)
    }
  }

  const handleToggle = (e: React.MouseEvent) => {
    e.stopPropagation()
    onToggleThumb(tag.id, !tag.useThumb)
  }

  const thumbUrl = tag.firstVideoId ? `https://img.youtube.com/vi/${tag.firstVideoId}/mqdefault.jpg` : null
  const showThumb = tag.useThumb && thumbUrl

  const style: React.CSSProperties = {
    background: showThumb 
      ? `linear-gradient(to bottom, rgba(9,9,15,0.7), rgba(9,9,15,0.9)), url(${thumbUrl})`
      : gradient,
    backgroundSize: 'cover',
    backgroundPosition: 'center',
  }

  return (
    <div
      className="tag-card"
      style={style}
      onClick={() => navigate(`/tags/${tag.id}`)}
    >
      <div className="tag-card-top-actions">
        <button 
          className="btn-delete-tag" 
          onClick={handleToggle} 
          title={tag.useThumb ? "Переключить на градиент" : "Переключить на обложку"}
          style={{ marginRight: '8px', opacity: tag.useThumb ? 1 : 0.4 }}
        >
          {tag.useThumb ? '🖼' : '🎨'}
        </button>
        {!tag.isDefault && (
          <button className="btn-delete-tag" onClick={handleDelete} title="Удалить тег">
            ✕
          </button>
        )}
      </div>

      {tag.isDefault && <div className="tag-default-badge">Default</div>}

      <div className="tag-card-name">{tag.name}</div>
      <div className="tag-card-count">
        {tag.showCount} {tag.showCount === 1 ? 'шоу' : 'шоу'}
        {' · '}
        {tag.episodeCount} {tag.episodeCount === 1 ? 'видео' : 'видео'}
      </div>
    </div>
  )
}
