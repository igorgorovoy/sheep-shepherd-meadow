import { useNavigate } from 'react-router-dom'

export function useRowNavigate() {
  const navigate = useNavigate()
  return (path: string) => (e: React.MouseEvent) => {
    const target = e.target as HTMLElement
    if (target.closest('a, button, input, select, textarea')) return
    navigate(path)
  }
}
