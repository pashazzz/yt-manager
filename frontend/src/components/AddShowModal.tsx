import { useState, useRef } from 'react'
import { api } from '../api/client'
import type { Show, Section } from '../types'

interface Props {
  sections: Section[]
  defaultSectionId: string
  onCreated: (show: Show) => void
  onClose: () => void
}

export default function AddShowModal({ sections, defaultSectionId, onCreated, onClose }: Props) {
  const [url, setUrl] = useState('')
  const [sectionId, setSectionId] = useState(defaultSectionId)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const inputRef = useRef<HTMLInputElement>(null)

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    const trimmed = url.trim()
    if (!trimmed) return

    setError('')
    setLoading(true)
    try {
      const res = await api.createShow(trimmed, sectionId)
      onCreated(res.show)
      onClose()
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : 'Что-то пошло не так')
    } finally {
      setLoading(false)
    }
  }

  const handleBackdropClick = (e: React.MouseEvent) => {
    if (e.target === e.currentTarget && !loading) onClose()
  }

  return (
    <div className="modal-backdrop" onClick={handleBackdropClick}>
      <div className="modal">
        <div className="modal-title">Добавить плейлист</div>
        <div className="modal-subtitle">
          Вставь ссылку на YouTube-плейлист — эпизоды загрузятся автоматически.
          На большие плейлисты может уйти до минуты.
        </div>

        <form onSubmit={handleSubmit}>
          <input
            ref={inputRef}
            className="modal-input"
            type="url"
            placeholder="https://youtube.com/playlist?list=…"
            value={url}
            onChange={e => setUrl(e.target.value)}
            disabled={loading}
            autoFocus
          />

          {sections.length > 1 && (
            <div className="modal-field">
              <label className="modal-label">Раздел</label>
              <select
                className="modal-select"
                value={sectionId}
                onChange={e => setSectionId(e.target.value)}
                disabled={loading}
              >
                {sections.map(s => (
                  <option key={s.id} value={s.id}>
                    {s.name}{s.isDefault ? ' (по умолчанию)' : ''}
                  </option>
                ))}
              </select>
            </div>
          )}

          {error && <div className="modal-error">{error}</div>}
          {loading && (
            <div className="modal-loading">
              <div className="spinner" style={{ width: 20, height: 20, borderWidth: 2 }} />
              Загружаем список эпизодов…
            </div>
          )}

          <div className="modal-actions">
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
