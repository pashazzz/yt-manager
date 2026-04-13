import { useNavigate } from 'react-router-dom'
import type { SectionInfo } from '../types'

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
  section: SectionInfo
  index: number
  onDelete: (id: string) => void
  onToggleThumb: (id: string, useThumb: boolean) => void
}

export default function SectionCard({ section, index, onDelete, onToggleThumb }: Props) {
  const navigate = useNavigate()
  const gradient = GRADIENTS[index % GRADIENTS.length]

  const handleDelete = (e: React.MouseEvent) => {
    e.stopPropagation()
    if (confirm(`Удалить раздел «${section.name}»?\nШоу переедут в раздел Default.`)) {
      onDelete(section.id)
    }
  }

  const handleToggle = (e: React.MouseEvent) => {
    e.stopPropagation()
    onToggleThumb(section.id, !section.useThumb)
  }

  const thumbUrl = section.firstVideoId ? `https://img.youtube.com/vi/${section.firstVideoId}/mqdefault.jpg` : null
  const showThumb = section.useThumb && thumbUrl

  const style: React.CSSProperties = {
    background: showThumb 
      ? `linear-gradient(to bottom, rgba(9,9,15,0.7), rgba(9,9,15,0.9)), url(${thumbUrl})`
      : gradient,
    backgroundSize: 'cover',
    backgroundPosition: 'center',
  }

  return (
    <div
      className="section-card"
      style={style}
      onClick={() => navigate(`/sections/${section.id}`)}
    >
      <div className="section-card-top-actions">
        <button 
          className="btn-delete-section" 
          onClick={handleToggle} 
          title={section.useThumb ? "Переключить на градиент" : "Переключить на обложку"}
          style={{ marginRight: '8px', opacity: section.useThumb ? 1 : 0.4 }}
        >
          {section.useThumb ? '🖼' : '🎨'}
        </button>
        {!section.isDefault && (
          <button className="btn-delete-section" onClick={handleDelete} title="Удалить раздел">
            ✕
          </button>
        )}
      </div>

      {section.isDefault && <div className="section-default-badge">Default</div>}

      <div className="section-card-name">{section.name}</div>
      <div className="section-card-count">
        {section.showCount} {section.showCount === 1 ? 'шоу' : 'шоу'}
        {' · '}
        {section.episodeCount} {section.episodeCount === 1 ? 'видео' : 'видео'}
      </div>
    </div>
  )
}
