import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import TagsPage from './pages/TagsPage'
import ShowsPage from './pages/ShowsPage'
import ShowPage from './pages/ShowPage'

export default function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<TagsPage />} />
        <Route path="/tags/:tagId" element={<ShowsPage />} />
        <Route path="/shows/:id" element={<ShowPage />} />
        {/* legacy redirect */}
        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
    </BrowserRouter>
  )
}
