import { Navigate, Route, Routes } from 'react-router-dom'
import { Layout } from './components/Layout'
import { DeploymentDetail } from './pages/detail/DeploymentDetail'
import { NodeDetail } from './pages/detail/NodeDetail'
import { PodDetail } from './pages/detail/PodDetail'
import { ServiceDetail } from './pages/detail/ServiceDetail'
import { MeadowOverview } from './pages/meadow/MeadowOverview'
import { RepoDetail } from './pages/meadow/RepoDetail'
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
        <Route path="nodes/:name" element={<NodeDetail />} />
        <Route path="pods" element={<Pods />} />
        <Route path="pods/:ns/:name" element={<PodDetail />} />
        <Route path="deployments" element={<Deployments />} />
        <Route path="deployments/:ns/:name" element={<DeploymentDetail />} />
        <Route path="services" element={<Services />} />
        <Route path="services/:ns/:name" element={<ServiceDetail />} />
        <Route path="events" element={<Events />} />
        <Route path="pasture" element={<Pasture />} />
        <Route path="meadow" element={<MeadowOverview />} />
        <Route path="meadow/repos/:name" element={<RepoDetail />} />
        <Route path="*" element={<Navigate to="/" replace />} />
      </Route>
    </Routes>
  )
}
