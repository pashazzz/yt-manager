import type { Provider } from '../types'

// Грубый автодетект по URL — должен быть синхронизирован с internal/providers/*.go.
export function detectProvider(url: string): Provider | null {
  const u = url.toLowerCase().trim()
  if (!u) return null
  if (u.includes('youtube.com/') || u.includes('youtu.be/') || u.includes('youtube-nocookie.com/')) {
    return 'youtube'
  }
  if (u.includes('rutube.ru/') || u.includes('rutube.com/')) {
    return 'rutube'
  }
  return null
}

export function providerLabel(p: Provider | null): string {
  switch (p) {
    case 'youtube': return 'YouTube'
    case 'rutube': return 'Rutube'
    default: return ''
  }
}
