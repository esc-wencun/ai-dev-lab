// HeaderSearch —— 对位基准 HeaderSearch（菜单模糊搜索：fuse.js + 键盘导航）

import { useMemo, useState } from 'react'
import { useNavigate } from 'react-router'
import { Input, Popover } from 'antd'
import { SearchOutlined } from '@ant-design/icons'
import Fuse from 'fuse.js'
import { useAppSelector } from '@/store/hooks'

interface SearchItem {
  path: string
  title: string
  isHttp?: boolean
}

export default function HeaderSearch() {
  const navigate = useNavigate()
  const sidebarRoutes = useAppSelector((s) => s.permission.sidebarRoutes)
  const [open, setOpen] = useState(false)
  const [keyword, setKeyword] = useState('')

  // 菜单树扁平化（对位基准 defaultRoutes 递归扁平化）
  const flatMenus = useMemo<SearchItem[]>(() => {
    const out: SearchItem[] = []
    const walk = (items: typeof sidebarRoutes, parentTitle?: string) => {
      for (const r of items) {
        if (r.hidden) continue
        const title = parentTitle ? `${parentTitle} / ${r.meta?.title || r.path}` : r.meta?.title || r.path
        const isHttp = !!r.meta?.link && /^https?:\/\//.test(r.meta.link)
        if (!r.children || r.children.filter((c) => !c.hidden).length === 0) {
          out.push({ path: r.path, title, isHttp })
        } else {
          walk(r.children, title)
        }
      }
    }
    walk(sidebarRoutes)
    return out
  }, [sidebarRoutes])

  const results = useMemo(() => {
    if (!keyword) return flatMenus.slice(0, 8)
    const fuse = new Fuse(flatMenus, { threshold: 0.2, keys: [{ name: 'title', weight: 0.7 }, { name: 'path', weight: 0.3 }] })
    return fuse.search(keyword).map((r) => r.item).slice(0, 8)
  }, [keyword, flatMenus])

  const go = (item: SearchItem) => {
    setOpen(false)
    setKeyword('')
    if (item.isHttp) {
      window.open(item.path, '_blank', 'noopener')
    } else {
      void navigate(item.path)
    }
  }

  return (
    <Popover
      open={open}
      onOpenChange={setOpen}
      trigger="click"
      placement="bottomRight"
      content={
        <div style={{ width: 320 }}>
          <Input
            autoFocus
            placeholder="搜索菜单"
            value={keyword}
            onChange={(e) => setKeyword(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === 'Enter' && results[0]) go(results[0])
              if (e.key === 'Escape') setOpen(false)
            }}
            allowClear
          />
          <div style={{ marginTop: 8 }}>
            {results.map((r) => (
              <div key={r.path} onClick={() => go(r)} style={{ padding: '6px 8px', cursor: 'pointer', borderRadius: 4 }}
                onMouseEnter={(e) => (e.currentTarget.style.background = '#f5f5f5')}
                onMouseLeave={(e) => (e.currentTarget.style.background = 'transparent')}
              >
                {r.title}
              </div>
            ))}
          </div>
        </div>
      }
    >
      <span style={{ cursor: 'pointer', padding: '0 12px', display: 'inline-flex', alignItems: 'center' }}>
        <SearchOutlined />
      </span>
    </Popover>
  )
}
