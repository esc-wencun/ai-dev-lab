// 缓存列表 —— 对位基准 monitor/cache/list（三栏联动：缓存名 → 键名 → 内容 + 三级清理）

import { useCallback, useEffect, useState } from 'react'
import { Button, Card, Col, Modal, Row, Table, Typography } from 'antd'
import { DeleteOutlined } from '@ant-design/icons'
import { listCacheName, listCacheKey, getCacheValue, clearCacheName, clearCacheKey, clearCacheAll } from '@/api/monitor/cache'

interface CacheNameRow { cacheName: string; remark: string }
interface CacheKeyRow { cacheKey: string; cacheValue: string; remark?: string }

export default function CacheList() {
  const [names, setNames] = useState<CacheNameRow[]>([])
  const [keys, setKeys] = useState<CacheKeyRow[]>([])
  const [totalKeys, setTotalKeys] = useState(0)
  const [nowCacheName, setNowCacheName] = useState('')
  const [nowKey, setNowKey] = useState('')
  const [keyValue, setKeyValue] = useState('')

  const loadNames = useCallback(() => {
    void listCacheName().then((res) => {
      const rows = (res as unknown as { data: CacheNameRow[] }).data || []
      setNames(rows)
    })
  }, [])

  useEffect(() => { loadNames() }, [loadNames])

  const handleKeyClick = (cacheName: string, key: string) => {
    setNowKey(key)
    void getCacheValue(cacheName, key).then((res) => {
      setKeyValue(String((res as unknown as { data: string }).data ?? ''))
    })
  }

  const confirm = (content: string, run: () => Promise<unknown>) => {
    Modal.confirm({ title: '系统提示', content, okText: '确定', cancelText: '取消', onOk: run })
  }

  return (
    <div style={{ padding: 16 }}>
      <Row gutter={[16, 16]}>
        <Col xs={24} md={7}>
          <Card size="small" title="缓存列表" extra={
            <Button type="link" danger icon={<DeleteOutlined />} onClick={() =>
              confirm('是否确认清除全部缓存?', async () => { await clearCacheAll(); loadNames() })}>清理全部</Button>
          }>
            <Table
              rowKey="cacheName"
              dataSource={names}
              pagination={false}
              onRow={(record) => ({
                onClick: () => {
                  setNowCacheName(record.cacheName)
                  void listCacheKey(record.cacheName).then((res) => {
                    const rows = (res as unknown as { data: CacheKeyRow[] }).data || []
                    setKeys(rows)
                    setTotalKeys(rows.length)
                  })
                },
              })}
              columns={[{ title: '缓存名称', dataIndex: 'cacheName' }, { title: '备注', dataIndex: 'remark' }]}
            />
          </Card>
        </Col>

        <Col xs={24} md={7}>
          <Card size="small" title={`键名列表${nowCacheName ? ` — ${nowCacheName}（${totalKeys}）` : ''}`} extra={
            nowCacheName ? (
              <Button type="link" danger icon={<DeleteOutlined />} onClick={() =>
                confirm(`是否确认清除${nowCacheName}缓存?`, async () => {
                  await clearCacheName(nowCacheName); setKeys([]); setTotalKeys(0); loadNames()
                })}>清理</Button>
            ) : null
          }>
            <Table
              rowKey="cacheKey"
              dataSource={keys}
              pagination={false}
              onRow={(record) => ({ onClick: () => handleKeyClick(nowCacheName, record.cacheKey) })}
              columns={[{ title: '键名', dataIndex: 'cacheKey', ellipsis: true }]}
            />
          </Card>
        </Col>
        <Col xs={24} md={10}>
          <Card size="small" title={`缓存内容${nowKey ? ` — ${nowKey}` : ''}`} extra={
            nowKey ? (
              <Button type="link" danger icon={<DeleteOutlined />} onClick={() =>
                confirm(`是否确认清除${nowKey}缓存?`, async () => {
                  await clearCacheKey(nowKey); setKeyValue(''); setNowKey('')
                  void listCacheKey(nowCacheName).then((res) => {
                    const rows = (res as unknown as { data: CacheKeyRow[] }).data || []
                    setKeys(rows); setTotalKeys(rows.length)
                  })
                })}>清理</Button>
            ) : null
          }>
            <Typography.Paragraph copyable style={{ maxHeight: 480, overflowY: 'auto' }}>
              <pre style={{ whiteSpace: 'pre-wrap', margin: 0 }}>{keyValue || '点击左侧键名查看内容'}</pre>
            </Typography.Paragraph>
          </Card>
        </Col>
      </Row>
    </div>
  )
}
