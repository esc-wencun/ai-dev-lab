// ReadUsersDialog —— 公告已读用户弹窗（对位基准 notice/ReadUsers.vue）

import { useEffect, useState } from 'react'
import { Input, Modal, Table } from 'antd'
import { listNoticeReadUsers } from '@/api/system/notice'

interface ReadUserRow {
  [k: string]: unknown
  userName?: string
  nickName?: string
  readTime?: string
}

interface Props {
  open: boolean
  noticeId?: number
  onClose: () => void
}

export default function ReadUsersDialog({ open, noticeId, onClose }: Props) {
  const [rows, setRows] = useState<ReadUserRow[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(false)
  const [keyword, setKeyword] = useState('')
  const [page, setPage] = useState({ pageNum: 1, pageSize: 10 })

  const load = async (p = page, kw = keyword) => {
    if (!noticeId) return
    setLoading(true)
    try {
      const res = (await listNoticeReadUsers({ ...p, noticeId, nickName: kw || undefined })) as unknown as {
        rows: ReadUserRow[]
        total: number
      }
      setRows(res.rows || [])
      setTotal(res.total || 0)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    if (open) void load()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [open, page])

  const columns = [
    { title: '用户账号', dataIndex: 'userName' },
    { title: '用户名称', dataIndex: 'nickName' },
    { title: '阅读时间', dataIndex: 'readTime' },
  ]

  return (
    <Modal title="阅读用户" open={open} footer={null} onCancel={onClose} width={640}>
      <Input.Search
        placeholder="请输入用户名称"
        allowClear
        style={{ marginBottom: 12 }}
        onSearch={(v) => {
          setKeyword(v)
          setPage({ ...page, pageNum: 1 })
          void load({ ...page, pageNum: 1 }, v)
        }}
      />
      <Table
        rowKey={(r) => String(r.userName)}
        columns={columns}
        dataSource={rows}
        loading={loading}
        pagination={{ total, pageSize: 10, current: page.pageNum, showTotal: (t) => `共 ${t} 条`,
          onChange: (p) => setPage({ ...page, pageNum: p }) }}
      />
    </Modal>
  )
}
