import { Navigate, Route, Routes } from 'react-router-dom'
import { Layout } from './components/Layout'
import { Overview } from './pages/Overview'
import { Nodes } from './pages/Nodes'
import { Pods } from './pages/Pods'
import { Deployments } from './pages/Deployments'
import { Services } from './pages/Services'
import { Events } from './pages/Events'
import { Pasture } from './pages/Pasture'

export function App() {
  return (
    <Routes>
      <Route element={<Layout />}>
        <Route index element={<Overview />} />
        <Route path="nodes" element={<Nodes />} />
        <Route path="pods" element={<Pods />} />
        <Route path="deployments" element={<Deployments />} />
        <Route path="services" element={<Services />} />
        <Route path="events" element={<Events />} />
        <Route path="pasture" element={<Pasture />} />
        <Route path="*" element={<Navigate to="/" replace />} />
      </Route>
    </Routes>
  )
}
