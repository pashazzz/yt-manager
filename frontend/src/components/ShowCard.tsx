import { useNavigate } from 'react-router-dom'
import type { Show, Episode } from '../types'

interface Props {
  show: Show
  episodes: Episode[]
  onDelete: (id: string) => void
}

function thumbUrl(videoId: string | undefined) {
  return videoId ? `https://img.youtube.com/vi/${videoId}/mqdefault.jpg` : ''
}

export default function ShowCard({ show, episodes, onDelete }: Props) {
  const navigate = useNavigate()
  const firstEp = episodes[0]
  const watched = episodes.filter(e => e.isWatched).length
  const progress = episodes.length > 0 ? (watched / episodes.length) * 100 : 0

  const handleDelete = (e: React.MouseEvent) => {
    e.stopPropagation()
    if (confirm(`Удалить «${show.title}»?`)) onDelete(show.id)
  }

  return (
    <div className="show-card" onClick={() => navigate(`/shows/${show.id}`)}>
      <div className="show-card-thumb">
        {firstEp && (
          <img src={thumbUrl(firstEp.videoId)} alt={show.title} loading="lazy" />
        )}
        <div className="show-card-overlay">
          <span className="btn-play-overlay">▶ Смотреть</span>
        </div>
      </div>

      <button className="btn-delete-card" onClick={handleDelete} title="Удалить шоу">
        ✕
      </button>

      <div className="show-card-info">
        <div className="show-card-title">{show.title}</div>
        <div className="show-card-meta">
          <span>{episodes.length} эп.</span>
          {watched > 0 && (
            <span className="badge-watched">{watched} просмотрено</span>
          )}
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
