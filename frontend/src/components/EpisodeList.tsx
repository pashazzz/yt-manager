import type { CSSProperties } from 'react'
import { DndContext, closestCenter, DragEndEvent, PointerSensor, useSensor, useSensors } from '@dnd-kit/core'
import { SortableContext, arrayMove, verticalListSortingStrategy, useSortable } from '@dnd-kit/sortable'
import { CSS } from '@dnd-kit/utilities'

import type { Episode, Tag } from '../types'
import { thumbForEpisode } from '../utils/thumbnails'
import { fmtDuration } from '../utils/format'

interface Props {
  episodes: Episode[]
  currentId: string
  onSelect: (episode: Episode) => void
  onToggleWatched?: (episode: Episode) => void
  onAddVideo?: () => void
  isReorderable?: boolean
  onReorder?: (episodes: Episode[]) => void
  variant?: 'sidebar' | 'inline'
  tags?: Tag[]
  onMove?: (episodeId: string, tagIds: string[]) => void
  onDelete?: (episodeId: string) => void
}

interface EpisodeItemProps {
  ep: Episode
  index: number
  isActive: boolean
  onSelect: (ep: Episode) => void
  onToggleWatched?: (ep: Episode) => void
  isReorderable?: boolean
  listeners?: Record<string, unknown>
  setNodeRef?: (node: HTMLElement | null) => void
  style?: CSSProperties
  tags?: Tag[]
  onMove?: (episodeId: string, tagIds: string[]) => void
  onDelete?: (episodeId: string) => void
}

function EpisodeItem({ ep, index, isActive, onSelect, onToggleWatched, isReorderable, listeners, setNodeRef, style, tags, onMove, onDelete }: EpisodeItemProps) {
  const progress = ep.duration > 0 ? Math.min(100, (ep.currentTime / ep.duration) * 100) : 0
  const thumb = thumbForEpisode(ep)

  return (
    <div
      ref={setNodeRef}
      style={style}
      className={`episode-item ${isActive ? 'active' : ''}`}
      onClick={() => onSelect(ep)}
      title={ep.title}
    >
      {isReorderable && (
        <div className="episode-drag-handle" {...listeners}>
          ☰
        </div>
      )}
      <div className="episode-thumb">
        {thumb && <img src={thumb} alt={ep.title} loading="lazy" />}
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
        <div className="episode-item-index">Эп. {index + 1}</div>
        <div className="episode-item-title">{ep.title}</div>
        {ep.duration > 0 && (
          <div className="episode-item-duration">{fmtDuration(ep.duration)}</div>
        )}
      </div>

      <div className="episode-item-actions">
        {tags && tags.length > 0 && onMove && (
          <div className="tag-picker">
            <button
              className="btn-icon-ghost"
              title="Управление тегами"
            >
              #
            </button>
            <select
              className="native-dropdown-overlay"
              value=""
              onChange={e => {
                e.stopPropagation()
                const targetTagId = e.target.value
                if (!targetTagId) return

                const currentIds = ep.tagIds || []
                const isAssigned = currentIds.includes(targetTagId)
                const newIds = isAssigned
                  ? currentIds.filter((id: string) => id !== targetTagId)
                  : [...currentIds, targetTagId]

                onMove(ep.id, newIds)
                e.target.value = ''
              }}
              onClick={e => e.stopPropagation()}
            >
              <option value="" disabled>Теги...</option>
              {tags.map((t: Tag) => {
                const isAssigned = (ep.tagIds || []).includes(t.id)
                return (
                  <option key={t.id} value={t.id}>
                    {isAssigned ? '✓ ' : ''}{t.name}
                  </option>
                )
              })}
            </select>
          </div>
        )}

        {onToggleWatched && (
          <div className="episode-action-btn" onClick={e => { e.stopPropagation(); onToggleWatched(ep) }}>
            <button
              title={ep.isWatched ? 'Отметить как непросмотренное' : 'Отметить как просмотренное'}
              className="btn-icon-ghost"
            >
              {ep.isWatched ? '↻' : '✓'}
            </button>
          </div>
        )}

        {onDelete && (
          <div className="episode-action-btn" onClick={e => { e.stopPropagation(); onDelete(ep.id) }}>
            <button
              title="Удалить эпизод"
              className="btn-icon-ghost btn-icon-danger"
            >
              ✕
            </button>
          </div>
        )}
      </div>
    </div>
  )
}

interface SortableEpisodeProps {
  ep: Episode
  index: number
  isActive: boolean
  onSelect: (ep: Episode) => void
  onToggleWatched?: (ep: Episode) => void
  tags?: Tag[]
  onMove?: (episodeId: string, tagIds: string[]) => void
  onDelete?: (episodeId: string) => void
}

function SortableEpisode({ ep, index, isActive, onSelect, onToggleWatched, tags, onMove, onDelete }: SortableEpisodeProps) {
  const { attributes, listeners, setNodeRef, transform, transition, isDragging } = useSortable({ id: ep.id })
  const style: CSSProperties = {
    transform: CSS.Transform.toString(transform),
    transition,
    opacity: isDragging ? 0.8 : 1,
    zIndex: isDragging ? 10 : 1,
  }

  return (
    <EpisodeItem
      ep={ep}
      index={index}
      isActive={isActive}
      onSelect={onSelect}
      onToggleWatched={onToggleWatched}
      isReorderable={true}
      listeners={{ ...attributes, ...listeners }}
      setNodeRef={setNodeRef}
      style={style}
      tags={tags}
      onMove={onMove}
      onDelete={onDelete}
    />
  )
}

export default function EpisodeList({ episodes, currentId, onSelect, onToggleWatched, onAddVideo, isReorderable, onReorder, variant = 'sidebar', tags, onMove, onDelete }: Props) {
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
      const updatedItems = newItems.map((item, index) => ({ ...item, orderIndex: index }))
      onReorder(updatedItems)
    }
  }

  const TagComp = variant === 'inline' ? 'div' : 'aside'
  const className = variant === 'inline' ? 'episode-list-inline' : 'episode-list-sidebar'

  return (
    <TagComp className={className}>
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
              {episodes.map((ep, i) => (
                <SortableEpisode
                  key={ep.id}
                  ep={ep}
                  index={i}
                  isActive={ep.id === currentId}
                  onSelect={onSelect}
                  onToggleWatched={onToggleWatched}
                  tags={tags}
                  onMove={onMove}
                  onDelete={onDelete}
                />
              ))}
            </SortableContext>
          </DndContext>
        ) : (
          episodes.map((ep, i) => (
            <EpisodeItem
              key={ep.id}
              ep={ep}
              index={i}
              isActive={ep.id === currentId}
              onSelect={onSelect}
              onToggleWatched={onToggleWatched}
              isReorderable={false}
              tags={tags}
              onMove={onMove}
              onDelete={onDelete}
            />
          ))
        )}

        {onAddVideo && (
          <div className="episode-list-add">
            <button className="btn-ghost btn-add-video" onClick={onAddVideo}>
              + Добавить видео
            </button>
          </div>
        )}
      </div>
    </TagComp>
  )
}
