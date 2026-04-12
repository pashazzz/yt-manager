import type { Episode } from '../types'

interface Props {
  episodes: Episode[]
  currentId: string
  onSelect: (episode: Episode) => void
  onAddVideo?: () => void
}

function fmtDuration(sec: number): string {
  if (!sec) return ''
  const h = Math.floor(sec / 3600)
  const m = Math.floor((sec % 3600) / 60)
  const s = Math.floor(sec % 60)
  if (h > 0) return `${h}:${String(m).padStart(2, '0')}:${String(s).padStart(2, '0')}`
  return `${m}:${String(s).padStart(2, '0')}`
}

function thumbUrl(videoId: string) {
  return `https://img.youtube.com/vi/${videoId}/mqdefault.jpg`
}

export default function EpisodeList({ episodes, currentId, onSelect, onAddVideo }: Props) {
  const watchedCount = episodes.filter(e => e.isWatched).length

  return (
    <aside className="episode-list-sidebar">
      <div className="episode-list-header">
        <div className="episode-list-title">Эпизоды</div>
        <div className="episode-list-count">
          {watchedCount} / {episodes.length} просмотрено
        </div>
      </div>

      <div className="episode-list-scroll">
        {episodes.map(ep => {
          const progress =
            ep.duration > 0 ? Math.min(100, (ep.currentTime / ep.duration) * 100) : 0
          const isActive = ep.id === currentId

          return (
            <div
              key={ep.id}
              className={`episode-item ${isActive ? 'active' : ''}`}
              onClick={() => onSelect(ep)}
              title={ep.title}
            >
              <div className="episode-thumb">
                <img src={thumbUrl(ep.videoId)} alt={ep.title} loading="lazy" />
                {ep.isWatched ? (
                  <div className="episode-thumb-watched">✓</div>
                ) : progress > 0 ? (
                  <div className="episode-thumb-progress">
                    <div
                      className="episode-thumb-progress-bar"
                      style={{ width: `${progress}%` }}
                    />
                  </div>
                ) : null}
              </div>

              <div className="episode-item-info">
                <div className="episode-item-index">Эп. {ep.orderIndex + 1}</div>
                <div className="episode-item-title">{ep.title}</div>
                {ep.duration > 0 && (
                  <div className="episode-item-duration">{fmtDuration(ep.duration)}</div>
                )}
              </div>
            </div>
          )
        })}
        {onAddVideo && (
            <div className="episode-list-add" style={{ padding: '10px' }}>
              <button className="btn-ghost" style={{ width: '100%', fontSize: '0.85rem' }} onClick={onAddVideo}>
                + Добавить видео
              </button>
            </div>
        )}
      </div>
    </aside>
  )
}
