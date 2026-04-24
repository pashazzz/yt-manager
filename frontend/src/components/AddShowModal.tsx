import { useState, useRef } from 'react'
import { api } from '../api/client'
import type { Show, Tag } from '../types'
import { detectProvider, providerLabel } from '../utils/providers'

interface Props {
  tags: Tag[]
  defaultTagId: string
  onCreated: (show: Show) => void
  onClose: () => void
}

export default function AddShowModal({ tags, defaultTagId, onCreated, onClose }: Props) {
  const [isCustom, setIsCustom] = useState(false)
  const [url, setUrl] = useState('')
  const [title, setTitle] = useState('')
  const [tagIds, setTagIds] = useState<string[]>([defaultTagId])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const inputRef = useRef<HTMLInputElement>(null)

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    const reqUrl = isCustom ? '' : url.trim()
    const reqTitle = isCustom ? title.trim() : ''

    if (!isCustom && !reqUrl) return
    if (isCustom && !reqTitle) return

    setError('')
    setLoading(true)
    try {
      const res = await api.createShow(reqUrl, tagIds, reqTitle)
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
          Вставь ссылку на плейлист — эпизоды загрузятся автоматически.
          На большие плейлисты может уйти до минуты.
        </div>

        <div className="modal-tabs" style={{ display: 'flex', gap: 10, marginBottom: 16 }}>
          <button type="button" className={`btn-ghost ${!isCustom ? 'active' : ''}`} onClick={() => setIsCustom(false)}>По ссылке</button>
          <button type="button" className={`btn-ghost ${isCustom ? 'active' : ''}`} onClick={() => setIsCustom(true)}>Пустое шоу</button>
        </div>

        <form onSubmit={handleSubmit}>
          {!isCustom ? (
            <>
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
              {url.trim() && (
                <div className="modal-hint" style={{ fontSize: '0.8rem', marginTop: 4, opacity: 0.75 }}>
                  {detectProvider(url)
                    ? `Источник: ${providerLabel(detectProvider(url))}`
                    : 'Неизвестный источник'}
                </div>
              )}
            </>
          ) : (
            <input
              className="modal-input"
              type="text"
              placeholder="Название шоу"
              value={title}
              onChange={e => setTitle(e.target.value)}
              disabled={loading}
              autoFocus
            />
          )}

          {tags.length > 0 && (
            <div className="modal-field">
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
                          // Предотвращаем пустой список, если это логически важно, 
                          // или просто позволяем удалять
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
            <button type="submit" className="btn-primary" disabled={loading || (isCustom ? !title.trim() : !url.trim())}>
              {loading ? 'Загружаем…' : 'Добавить'}
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}
