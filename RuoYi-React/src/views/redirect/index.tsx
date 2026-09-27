// redirect 中转页 —— 页签刷新机制依赖（$tab.refreshPage → /redirect/path → 立即 replace 回原路径）

import { useEffect } from 'react'
import { useNavigate, useLocation } from 'react-router'

export default function Redirect() {
  const navigate = useNavigate()
  const location = useLocation()
  const params = new URLSearchParams(location.search)

  useEffect(() => {
    const path = location.pathname.replace(/^\/redirect/, '')
    const query = params.toString()
    navigate((path || '/index') + (query ? '?' + query : ''), { replace: true })
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  return null
}
