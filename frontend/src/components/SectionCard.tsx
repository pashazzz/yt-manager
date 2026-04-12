import { useNavigate } from 'react-router-dom'
import type { Section } from '../types'

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
  section: Section
  showCount: number
  index: number
  onDelete: (id: string) => void
}

export default function SectionCard({ section, showCount, index, onDelete }: Props) {
  const navigate = useNavigate()
  const gradient = GRADIENTS[index % GRADIENTS.length]

  const handleDelete = (e: React.MouseEvent) => {
    e.stopPropagation()
    if (confirm(`Удалить раздел «${section.name}»?\nШоу переедут в раздел Default.`)) {
      onDelete(section.id)
    }
  }

  return (
    <div
      className="section-card"
      style={{ background: gradient }}
      onClick={() => navigate(`/sections/${section.id}`)}
    >
      {!section.isDefault && (
        <button className="btn-delete-section" onClick={handleDelete} title="Удалить раздел">
          ✕
        </button>
      )}
      {section.isDefault && <div className="section-default-badge">Default</div>}

      <div className="section-card-name">{section.name}</div>
      <div className="section-card-count">
        {showCount} {showCount === 1 ? 'шоу' : 'шоу'}
      </div>
    </div>
  )
}
