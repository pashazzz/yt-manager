import { DndContext, closestCenter, DragEndEvent, PointerSensor, useSensor, useSensors } from '@dnd-kit/core'
import { SortableContext, arrayMove, rectSortingStrategy, useSortable } from '@dnd-kit/sortable'
import { CSS } from '@dnd-kit/utilities'

import { useEffect, useState } from 'react'
import { api } from '../api/client'
import type { TagInfo } from '../types'
import TagCard from '../components/TagCard'

function SortableTag({ tag, index, onDelete, onToggleThumb }: any) {
  const { attributes, listeners, setNodeRef, transform, transition, isDragging } = useSortable({ id: tag.id })
  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
    opacity: isDragging ? 0.8 : 1,
    zIndex: isDragging ? 10 : 1,
    cursor: isDragging ? 'grabbing' : 'grab'
  }

  return (
    <div ref={setNodeRef} style={style} {...attributes} {...listeners}>
      <TagCard
        tag={tag}
        index={index}
        onDelete={onDelete}
        onToggleThumb={onToggleThumb}
      />
    </div>
  )
}


export default function TagsPage() {
  const [tags, setTags] = useState<TagInfo[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [newName, setNewName] = useState('')
  const [creating, setCreating] = useState(false)
  const [showForm, setShowForm] = useState(false)

  const sensors = useSensors(
    useSensor(PointerSensor, {
      activationConstraint: {
        distance: 5,
      },
    })
  )

  useEffect(() => { load() }, [])

  async function load() {
    try {
      setLoading(true)
      const list = await api.getTags()
      setTags(list)
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : 'Ошибка загрузки')
    } finally {
      setLoading(false)
    }
  }

  async function handleCreate(e: React.FormEvent) {
    e.preventDefault()
    const name = newName.trim()
    if (!name) return
    setCreating(true)
    try {
      const s = await api.createTag(name)
      // Временно добавляем как TagInfo для UI
      setTags(prev => [...prev, { ...s, showCount: 0, episodeCount: 0, firstVideoId: '' }])
      setNewName('')
      setShowForm(false)
    } catch (err: unknown) {
      alert(err instanceof Error ? err.message : 'Ошибка создания тега')
    } finally {
      setCreating(false)
    }
  }

  async function handleDelete(id: string) {
    try {
      await api.deleteTag(id)
      setTags(prev => prev.filter(s => s.id !== id))
      // Перезагружаем счётчики (элементы могли изменить принадлежность)
      load()
    } catch (err: unknown) {
      alert(err instanceof Error ? err.message : 'Ошибка удаления')
    }
  }

  async function handleToggleThumb(id: string, useThumb: boolean) {
    try {
      await api.updateTagSettings(id, useThumb)
      setTags(prev => prev.map(s => s.id === id ? { ...s, useThumb } : s))
    } catch (err: unknown) {
      alert(err instanceof Error ? err.message : 'Ошибка обновления')
    }
  }

  function handleDragEnd(event: DragEndEvent) {
    const { active, over } = event
    if (over && active.id !== over.id) {
      setTags(items => {
        const oldIndex = items.findIndex(i => i.id === active.id)
        const newIndex = items.findIndex(i => i.id === over.id)
        const newItems = arrayMove(items, oldIndex, newIndex)

        api.reorderTags(newItems.map(i => i.id)).catch(console.error)
        return newItems
      })
    }
  }

  if (loading) return <div className="page-loader"><div className="spinner" /></div>
  if (error) return (
    <div className="page-error">
      <span>⚠ {error}</span>
      <button className="btn-primary" onClick={load}>Повторить</button>
    </div>
  )

  return (
    <div className="shows-page">
      <header className="app-header">
        <span className="app-logo">▶ YT Manager</span>
        <button className="btn-add" onClick={() => setShowForm(v => !v)}>
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5">
            <path d="M12 5v14M5 12h14" strokeLinecap="round" />
          </svg>
          Новый тег
        </button>
      </header>

      <main className="shows-content">
        {showForm && (
          <form className="new-tag-form" onSubmit={handleCreate}>
            <input
              className="modal-input"
              style={{ marginBottom: 0 }}
              placeholder="Название тега"
              value={newName}
              onChange={e => setNewName(e.target.value)}
              autoFocus
              disabled={creating}
            />
            <button type="submit" className="btn-primary" disabled={creating || !newName.trim()}>
              {creating ? '…' : 'Создать'}
            </button>
            <button type="button" className="btn-ghost" onClick={() => setShowForm(false)}>
              Отмена
            </button>
          </form>
        )}

        {tags.length === 0 ? (
          <div className="empty-state">
            <div className="empty-state-icon">🏷</div>
            <h3>Нет тегов</h3>
            <p>Теги помогают группировать ваши шоу и видео</p>
          </div>
        ) : (
          <>
            <h2 className="shows-tag-title">Теги</h2>
            <DndContext sensors={sensors} collisionDetection={closestCenter} onDragEnd={handleDragEnd}>
              <SortableContext items={tags} strategy={rectSortingStrategy}>
                <div className="tags-grid">
                  {tags.map((s, i) => (
                    <SortableTag
                      key={s.id}
                      tag={s}
                      index={i}
                      onDelete={handleDelete}
                      onToggleThumb={handleToggleThumb}
                    />
                  ))}
                </div>
              </SortableContext>
            </DndContext>
          </>
        )}
      </main>
    </div>
  )
}
