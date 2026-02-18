import { BrowserRouter, Routes, Route } from 'react-router-dom'
import Layout from './components/Layout'
import Dashboard from './pages/Dashboard'
import Transactions from './pages/Transactions'
import Pending from './pages/Pending'
import Propose from './pages/Propose'
import Nodes from './pages/Nodes'

export default function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route element={<Layout />}>
          <Route path="/" element={<Dashboard />} />
          <Route path="/transactions" element={<Transactions />} />
          <Route path="/pending" element={<Pending />} />
          <Route path="/propose" element={<Propose />} />
          <Route path="/nodes" element={<Nodes />} />
        </Route>
      </Routes>
    </BrowserRouter>
  )
}
