import { NotFoundPage, UnauthorizedPage } from './pages'

// 薄壳：glob 约定 'error/404' → 本文件
export default function Error404() {
  return <NotFoundPage />
}

export { UnauthorizedPage as Error401Impl }
