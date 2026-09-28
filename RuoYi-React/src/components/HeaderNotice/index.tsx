// HeaderNotice 通知公告铃铛 —— 对位基准 layout/components/HeaderNotice/index.vue + DetailView.vue
// 铃铛 + 未读数 Badge；hover 弹出列表（类型 tag + 标题 + 时间）；
// 点击条目：未读 → markNoticeRead 本地标记 + 打开详情 Drawer（getNotice 取 HTML 内容 dangerouslySetInnerHTML）
// 全部已读 → markNoticeReadAll(所有 id 逗号拼接)

import { useEffect, useState } from 'react'
import { Badge, Drawer, Popover, Spin, Tag } from 'antd'
import { BellOutlined, LoadingOutlined, FileTextOutlined } from '@ant-design/icons'
import { listNoticeTop, markNoticeRead, markNoticeReadAll, getNotice } from '@/api/system/notice'

interface NoticeItem {
  noticeId: number
  noticeTitle: string
  noticeType: string
  isRead?: boolean
  createTime?: string
}

interface NoticeDetail {
  noticeId: number
  noticeTitle: string
  noticeType: string
  status?: string | number
  noticeContent?: string
  createBy?: string
  createTime?: string
}

export default function HeaderNotice() {
  const [noticeList, setNoticeList] = useState<NoticeItem[]>([])
  const [unreadCount, setUnreadCount] = useState(0)
  const [loading, setLoading] = useState(false)
  const [open, setOpen] = useState(false)

  // 详情 Drawer
  const [detail, setDetail] = useState<NoticeDetail | null>(null)
  const [detailOpen, setDetailOpen] = useState(false)
  const [detailLoading, setDetailLoading] = useState(false)

  // 加载顶部公告列表（对位 loadNoticeTop：unreadCount 优先取后端字段，否则本地计数）
  const loadNoticeTop = () => {
    setLoading(true)
    listNoticeTop()
      .then((res) => {
        const body = res as unknown as { data?: NoticeItem[]; unreadCount?: number }
        const data = body.data || []
        setNoticeList(data)
        setUnreadCount(body.unreadCount !== undefined ? body.unreadCount : data.filter((n) => !n.isRead).length)
      })
      .finally(() => setLoading(false))
  }

  useEffect(() => {
    loadNoticeTop()
  }, [])

  // 预览公告详情（对位 previewNotice：未读则标记 + 打开详情）
  const previewNotice = (item: NoticeItem) => {
    if (!item.isRead) {
      markNoticeRead(item.noticeId).catch(() => undefined)
      setNoticeList((list) => list.map((n) => (n.noticeId === item.noticeId ? { ...n, isRead: true } : n)))
      setUnreadCount((c) => Math.max(0, c - 1))
    }
    setDetailOpen(true)
    setDetailLoading(true)
    setDetail(null)
    getNotice(item.noticeId)
      .then((res) => setDetail((res as unknown as { data: NoticeDetail }).data))
      .catch(() => setDetail(null))
      .finally(() => setDetailLoading(false))
  }

  // 全部已读（对位 markAllRead：所有 id 逗号拼接）
  const markAllRead = () => {
    const ids = noticeList.map((n) => n.noticeId).join(',')
    if (!ids) return
    markNoticeReadAll(ids).catch(() => undefined)
    setNoticeList((list) => list.map((n) => ({ ...n, isRead: true })))
    setUnreadCount(0)
  }

  const content = (
    // 宽度定死 320(与基准 :width="320" 一致);不动 margin——antd Popover 用 styles.body
    // 去 padding 即可,负 margin 会把头部拉出弹层边界(样式错乱根因)
    <div className="notice-popover-panel">
      {/* 头部：通知公告 / 全部已读 */}
      <div className="notice-header">
        <span className="notice-title">通知公告</span>
        <span className="notice-mark-all" onClick={markAllRead}>全部已读</span>
      </div>
      {loading ? (
        <div className="notice-empty"><LoadingOutlined /> 加载中...</div>
      ) : noticeList.length === 0 ? (
        <div className="notice-empty"><FileTextOutlined style={{ fontSize: 24, display: 'block', marginBottom: 6 }} />暂无公告</div>
      ) : (
        noticeList.map((item) => (
          <div key={item.noticeId} className={`notice-item${item.isRead ? ' is-read' : ''}`} onClick={() => previewNotice(item)}>
            <Tag color={item.noticeType === '1' ? 'warning' : 'success'} style={{ flexShrink: 0 }}>
              {item.noticeType === '1' ? '通知' : '公告'}
            </Tag>
            <span className="notice-item-title">{item.noticeTitle}</span>
            <span className="notice-item-date">{item.createTime}</span>
          </div>
        ))
      )}
    </div>
  )

  const isNormal = (status?: string | number) => status === '0' || status === 0

  return (
    <div>
      <Popover
        content={content}
        trigger="hover"
        open={open}
        placement="bottomRight"
        styles={{ body: { padding: 0, width: 320 } }}
        onOpenChange={setOpen}
      >
        <span className="notice-trigger">
          <Badge count={unreadCount} size="small">
            <BellOutlined style={{ fontSize: 16 }} />
          </Badge>
        </span>
      </Popover>

      {/* 详情抽屉（对位 DetailView.vue：类型 tag / 标题 / 元信息 / HTML 正文） */}
      <Drawer
        title="公告详情"
        placement="right"
        width="50%"
        open={detailOpen}
        onClose={() => { setDetailOpen(false); setDetail(null) }}
      >
        {detailLoading ? (
          <div style={{ textAlign: 'center', padding: 60 }}><Spin size="large" /></div>
        ) : !detail ? (
          <div style={{ textAlign: 'center', padding: 60, color: '#999' }}>暂无数据</div>
        ) : (
          <div className="notice-page">
            <div className="notice-type-wrap">
              <span className={`notice-type-tag ${detail.noticeType === '2' ? 'type-announce' : 'type-notify'}`}>
                {detail.noticeType === '1' ? '通知' : detail.noticeType === '2' ? '公告' : '消息'}
              </span>
            </div>
            <h1 className="notice-title">{detail.noticeTitle}</h1>
            <div className="notice-meta">
              <span className="meta-item">
                <span className="meta-label">创建者：</span>{detail.createBy || '—'}
              </span>
              <span className="meta-item">
                <span className="meta-label">时间：</span>{detail.createTime || '—'}
              </span>
              <span className="meta-item">
                <span className={`status-dot ${isNormal(detail.status) ? 'status-ok' : 'status-off'}`} />
                {isNormal(detail.status) ? '正常' : '已关闭'}
              </span>
            </div>
            <div className="notice-divider">
              <span className="notice-divider-dot" />
            </div>
            <div className="notice-body">
              {detail.noticeContent ? (
                <div className="notice-content" dangerouslySetInnerHTML={{ __html: detail.noticeContent }} />
              ) : (
                <div style={{ textAlign: 'center', color: '#999', padding: 32 }}>暂无内容</div>
              )}
            </div>
          </div>
        )}
      </Drawer>
    </div>
  )
}
