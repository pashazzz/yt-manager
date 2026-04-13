import { DndContext, closestCenter, DragEndEvent, PointerSensor, useSensor, useSensors } from '@dnd-kit/core'
import { SortableContext, arrayMove, rectSortingStrategy, useSortable } from '@dnd-kit/sortable'
import { CSS } from '@dnd-kit/utilities'

import { useEffect, useState } from 'react'
import { api } from '../api/client'
import type { SectionInfo } from '../types'
import SectionCard from '../components/SectionCard'

function SortableSection({ section, index, onDelete, onToggleThumb }: any) {
  const { attributes, listeners, setNodeRef, transform, transition, isDragging } = useSortable({ id: section.id })
  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
    opacity: isDragging ? 0.8 : 1,
    zIndex: isDragging ? 10 : 1,
    cursor: isDragging ? 'grabbing' : 'grab'
  }

  return (
    <div ref={setNodeRef} style={style} {...attributes} {...listeners}>
      <SectionCard
        section={section}
        index={index}
        onDelete={onDelete}
        onToggleThumb={onToggleThumb}
      />
    </div>
  )
}


export default function SectionsPage() {
  const [sections, setSections] = useState<SectionInfo[]>([])
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
      const list = await api.getSections()
      setSections(list)
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
      const s = await api.createSection(name)
      // Временно добавляем как SectionInfo для UI
      setSections(prev => [...prev, { ...s, showCount: 0, episodeCount: 0, firstVideoId: '' }])
      setNewName('')
      setShowForm(false)
    } catch (err: unknown) {
      alert(err instanceof Error ? err.message : 'Ошибка создания раздела')
    } finally {
      setCreating(false)
    }
  }

  async function handleDelete(id: string) {
    try {
      await api.deleteSection(id)
      setSections(prev => prev.filter(s => s.id !== id))
      // Перезагружаем счётчики (шоу могли переехать в Default)
      load()
    } catch (err: unknown) {
      alert(err instanceof Error ? err.message : 'Ошибка удаления')
    }
  }

  async function handleToggleThumb(id: string, useThumb: boolean) {
    try {
      await api.updateSectionSettings(id, useThumb)
      setSections(prev => prev.map(s => s.id === id ? { ...s, useThumb } : s))
    } catch (err: unknown) {
      alert(err instanceof Error ? err.message : 'Ошибка обновления')
    }
  }

  function handleDragEnd(event: DragEndEvent) {
    const { active, over } = event
    if (over && active.id !== over.id) {
      setSections(items => {
        const oldIndex = items.findIndex(i => i.id === active.id)
        const newIndex = items.findIndex(i => i.id === over.id)
        const newItems = arrayMove(items, oldIndex, newIndex)

        api.reorderSections(newItems.map(i => i.id)).catch(console.error)
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
          Новый раздел
        </button>
      </header>

      <main className="shows-content">
        {showForm && (
          <form className="new-section-form" onSubmit={handleCreate}>
            <input
              className="modal-input"
              style={{ marginBottom: 0 }}
              placeholder="Название раздела"
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

        {sections.length === 0 ? (
          <div className="empty-state">
            <div className="empty-state-icon">📂</div>
            <h3>Нет разделов</h3>
            <p>Разделы создаются автоматически при первом добавлении шоу</p>
          </div>
        ) : (
          <>
            <h2 className="shows-section-title">Разделы</h2>
            <DndContext sensors={sensors} collisionDetection={closestCenter} onDragEnd={handleDragEnd}>
              <SortableContext items={sections} strategy={rectSortingStrategy}>
                <div className="sections-grid">
                  {sections.map((s, i) => (
                    <SortableSection
                      key={s.id}
                      section={s}
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
