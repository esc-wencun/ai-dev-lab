// 在线用户 —— 对位基准 monitor/online（只读列表 + 强退）

import { useCallback, useEffect, useState } from 'react'
import { Button, Card, Form, Input, Space, Table } from 'antd'
import { SearchOutlined, ReloadOutlined } from '@ant-design/icons'
import { list, forceLogout } from '@/api/monitor/online'
import { Modal } from 'antd'
import { parseTime } from '@/utils/ruoyi'
import Auth from '@/components/Auth'

interface OnlineRow {
  [k: string]: unknown
  tokenId: string
  userName: string
  deptName?: string
  ipaddr?: string
  loginLocation?: string
  browser?: string
  os?: string
  loginTime?: number
}

export default function Online() {
  const [rows, setRows] = useState<OnlineRow[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(false)
  const [query, setQuery] = useState<{ pageNum: number; pageSize: number; ipaddr?: string; userName?: string }>({
    pageNum: 1, pageSize: 10,
  })

  const load = useCallback(async (q = query) => {
    setLoading(true)
    try {
      const res = (await list(q)) as unknown as { rows: OnlineRow[]; total: number }
      setRows(res.rows || [])
      setTotal(res.total || 0)
    } finally {
      setLoading(false)
    }
  }, [query])

  useEffect(() => { void load() }, [load])

  const handleForce = (row: OnlineRow) => {
    Modal.confirm({
      title: '系统提示', content: `是否确认强退"${row.userName}"这位用户?`, okText: '确定', cancelText: '取消',
      onOk: async () => { await forceLogout(row.tokenId); void load() },
    })
  }

  const columns = [
    { title: '会话编号', dataIndex: 'tokenId', render: (v: string) => v.slice(0, 12) + '...' },
    { title: '登录名称', dataIndex: 'userName' },
    { title: '部门名称', dataIndex: 'deptName' },
    { title: '主机', dataIndex: 'ipaddr' },
    { title: '登录地点', dataIndex: 'loginLocation' },
    { title: '浏览器', dataIndex: 'browser' },
    { title: '登录时间', dataIndex: 'loginTime', width: 160, render: (v: number) => parseTime(v) },
    {
      title: '操作', width: 90,
      render: (_: unknown, row: OnlineRow) => (
        <Auth permissions={['monitor:online:forceLogout']}>
          <Button type="link" size="small" danger onClick={() => handleForce(row)}>强退</Button>
        </Auth>
      ),
    },
  ]

  return (
    <div style={{ padding: 16 }}>
      <Card>
        <Form layout="inline" style={{ marginBottom: 16 }}>
          <Form.Item label="登录地址">
            <Input allowClear onChange={(e) => setQuery({ ...query, ipaddr: e.target.value })} />
          </Form.Item>
          <Form.Item label="登录名称">
            <Input allowClear onChange={(e) => setQuery({ ...query, userName: e.target.value })} />
          </Form.Item>
          <Form.Item>
            <Space>
              <Button type="primary" icon={<SearchOutlined />} onClick={() => void load({ ...query, pageNum: 1 })}>搜索</Button>
              <Button icon={<ReloadOutlined />} onClick={() => { setQuery({ pageNum: 1, pageSize: 10 }); void load() }}>重置</Button>
            </Space>
          </Form.Item>
        </Form>
        <Table
          rowKey="tokenId" columns={columns} dataSource={rows} loading={loading}
          pagination={{ total, pageSize: query.pageSize, current: query.pageNum,
            showTotal: (t) => `共 ${t} 条`, onChange: (p, size) => { setQuery({ ...query, pageNum: p, pageSize: size }); void load({ ...query, pageNum: p, pageSize: size }) } }}
          scroll={{ x: 'max-content' }}
        />
      </Card>
    </div>
  )
}
