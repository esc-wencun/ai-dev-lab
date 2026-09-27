import { UnauthorizedPage } from './pages'

// 薄壳：glob 约定 'error/401' → 本文件
export default function Error401() {
  return <UnauthorizedPage />
}
