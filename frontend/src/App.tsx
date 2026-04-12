import { BrowserRouter, Routes, Route } from 'react-router-dom'
import ShowsPage from './pages/ShowsPage'
import ShowPage from './pages/ShowPage'

export default function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<ShowsPage />} />
        <Route path="/shows/:id" element={<ShowPage />} />
      </Routes>
    </BrowserRouter>
  )
}
