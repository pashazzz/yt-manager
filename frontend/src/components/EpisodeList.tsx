import { DndContext, closestCenter, DragEndEvent, PointerSensor, useSensor, useSensors } from '@dnd-kit/core'
import { SortableContext, arrayMove, verticalListSortingStrategy, useSortable } from '@dnd-kit/sortable'
import { CSS } from '@dnd-kit/utilities'

import type { Episode } from '../types'

interface Props {
  episodes: Episode[]
  currentId: string
  onSelect: (episode: Episode) => void
  onToggleWatched?: (episode: Episode) => void
  onAddVideo?: () => void
  isReorderable?: boolean
  onReorder?: (episodes: Episode[]) => void
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

function EpisodeItem({ ep, isActive, onSelect, onToggleWatched, isReorderable, listeners, setNodeRef, style }: any) {
  const progress = ep.duration > 0 ? Math.min(100, (ep.currentTime / ep.duration) * 100) : 0

  return (
    <div
      ref={setNodeRef}
      style={style}
      className={`episode-item ${isActive ? 'active' : ''}`}
      onClick={() => onSelect(ep)}
      title={ep.title}
    >
      {isReorderable && (
        <div className="episode-drag-handle" {...listeners} style={{ padding: '0 8px', cursor: 'grab', opacity: 0.5 }}>
          ☰
        </div>
      )}
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

      {onToggleWatched && (
        <div className="episode-action-btn" onClick={e => { e.stopPropagation(); onToggleWatched(ep) }}>
          <button 
            title={ep.isWatched ? 'Отметить как непросмотренное' : 'Отметить как просмотренное'}
            style={{ background: 'none', border: 'none', cursor: 'pointer', fontSize: '1rem', opacity: 0.6 }}
          >
            {ep.isWatched ? '✕' : '✓'}
          </button>
        </div>
      )}
    </div>
  )
}

function SortableEpisode({ ep, isActive, onSelect, onToggleWatched }: any) {
  const { attributes, listeners, setNodeRef, transform, transition, isDragging } = useSortable({ id: ep.id })
  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
    opacity: isDragging ? 0.8 : 1,
    zIndex: isDragging ? 10 : 1,
  }

  return (
    <EpisodeItem
      ep={ep}
      isActive={isActive}
      onSelect={onSelect}
      onToggleWatched={onToggleWatched}
      isReorderable={true}
      listeners={{ ...attributes, ...listeners }}
      setNodeRef={setNodeRef}
      style={style}
    />
  )
}

export default function EpisodeList({ episodes, currentId, onSelect, onToggleWatched, onAddVideo, isReorderable, onReorder }: Props) {
  const watchedCount = episodes.filter(e => e.isWatched).length

  const sensors = useSensors(
    useSensor(PointerSensor, {
      activationConstraint: {
        distance: 5,
      },
    })
  )

  const handleDragEnd = (event: DragEndEvent) => {
    if (!onReorder) return
    const { active, over } = event
    if (over && active.id !== over.id) {
      const oldIndex = episodes.findIndex(i => i.id === active.id)
      const newIndex = episodes.findIndex(i => i.id === over.id)
      const newItems = arrayMove(episodes, oldIndex, newIndex)
      // Update orderIndex visually to avoid flashing before server sync
      const updatedItems = newItems.map((item, index) => ({ ...item, orderIndex: index }))
      onReorder(updatedItems)
    }
  }

  return (
    <aside className="episode-list-sidebar">
      <div className="episode-list-header">
        <div className="episode-list-title">Эпизоды</div>
        <div className="episode-list-count">
          {watchedCount} / {episodes.length} просмотрено
        </div>
      </div>

      <div className="episode-list-scroll">
        {isReorderable ? (
          <DndContext sensors={sensors} collisionDetection={closestCenter} onDragEnd={handleDragEnd}>
            <SortableContext items={episodes} strategy={verticalListSortingStrategy}>
              {episodes.map(ep => (
                <SortableEpisode
                  key={ep.id}
                  ep={ep}
                  isActive={ep.id === currentId}
                  onSelect={onSelect}
                  onToggleWatched={onToggleWatched}
                />
              ))}
            </SortableContext>
          </DndContext>
        ) : (
          episodes.map(ep => (
            <EpisodeItem
              key={ep.id}
              ep={ep}
              isActive={ep.id === currentId}
              onSelect={onSelect}
              onToggleWatched={onToggleWatched}
              isReorderable={false}
            />
          ))
        )}
        
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
