import { useState, useRef } from 'react'
import type { Tag } from '../types'

interface Props {
  tags?: Tag[]
  defaultTagId?: string
  onCreated: (url: string, tagIds: string[]) => void
  onClose: () => void
  loading?: boolean
}

export default function AddVideoModal({ tags = [], defaultTagId = '', onCreated, onClose, loading }: Props) {
  const [url, setUrl] = useState('')
  const [tagIds, setTagIds] = useState<string[]>(defaultTagId ? [defaultTagId] : [])
  const inputRef = useRef<HTMLInputElement>(null)

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    if (!url.trim()) return
    onCreated(url.trim(), tagIds)
  }

  const handleBackdropClick = (e: React.MouseEvent) => {
    if (e.target === e.currentTarget && !loading) onClose()
  }

  return (
    <div className="modal-backdrop" onClick={handleBackdropClick}>
      <div className="modal">
        <div className="modal-title">Добавить видео</div>
        <div className="modal-subtitle">
          Вставь ссылку на YouTube видео для добавления.
        </div>

        <form onSubmit={handleSubmit}>
          <input
            ref={inputRef}
            className="modal-input"
            type="url"
            placeholder="https://youtube.com/watch?v=…"
            value={url}
            onChange={e => setUrl(e.target.value)}
            disabled={loading}
            autoFocus
          />

          {tags.length > 0 && (
            <div className="modal-field" style={{ marginTop: 16 }}>
              <label className="modal-label">Теги</label>
              <div className="tag-checkbox-grid">
                {tags.map(t => (
                  <label key={t.id} className="tag-checkbox-item">
                    <input
                      type="checkbox"
                      checked={(tagIds || []).includes(t.id)}
                      onChange={e => {
                        const checked = e.target.checked
                        if (checked) {
                          setTagIds(prev => [...prev, t.id])
                        } else {
                          setTagIds(prev => prev.filter(id => id !== t.id))
                        }
                      }}
                    />
                    <span>{t.name}</span>
                  </label>
                ))}
              </div>
            </div>
          )}

          <div className="modal-actions" style={{ marginTop: 24 }}>
            <button type="button" className="btn-ghost" onClick={onClose} disabled={loading}>
              Отмена
            </button>
            <button type="submit" className="btn-primary" disabled={loading || !url.trim()}>
              {loading ? 'Загружаем…' : 'Добавить'}
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}
