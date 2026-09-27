// 操作日志 —— 对位基准 monitor/operlog（只读 + 详情分区弹窗 + 删除/清空/导出）

import { useEffect, useState } from 'react'
import { Button, Card, Form, Input, Modal, Select, Space, Table } from 'antd'
import { DeleteOutlined, SearchOutlined, ReloadOutlined, ExportOutlined } from '@ant-design/icons'
import { list, delOperlog, cleanOperlog } from '@/api/monitor/operlog'
import { useCrud } from '@/hooks/useCrud'
import { useDict } from '@/utils/dict'
import DictTag from '@/components/DictTag'
import { parseTime } from '@/utils/ruoyi'

interface OperRow {
  [k: string]: unknown
  operId: number
  title?: string
  businessType?: number
  method?: string
  requestMethod?: string
  operName?: string
  deptName?: string
  operUrl?: string
  operIp?: string
  operLocation?: string
  operParam?: string
  jsonResult?: string
  status: string
  errorMsg?: string
  operTime?: number
}

export default function Operlog() {
  const dicts = useDict('sys_oper_type', 'sys_common_status')
  const crud = useCrud<OperRow, { pageNum: number; pageSize: number; title?: string; operName?: string; businessType?: string; status?: string }>({
    listApi: list,
    delApi: delOperlog,
    exportUrl: '/monitor/operlog/export',
    defaultQuery: { pageNum: 1, pageSize: 10 },
  })
  const [detail, setDetail] = useState<OperRow | null>(null)

  useEffect(() => { void crud.getList() /* eslint-disable-line react-hooks/exhaustive-deps */ }, [])

  const handleClean = () => {
    Modal.confirm({
      title: '系统提示', content: '是否确认清空所有操作日志数据项?', okText: '确定', cancelText: '取消',
      onOk: async () => { await cleanOperlog(); void crud.getList() },
    })
  }

  const columns = [
    { title: '日志编号', dataIndex: 'operId', width: 90 },
    { title: '系统模块', dataIndex: 'title' },
    { title: '操作类型', dataIndex: 'businessType', width: 100,
      render: (v: number) => <DictTag options={dicts.sys_oper_type || []} value={v} /> },
    { title: '请求方式', dataIndex: 'requestMethod', width: 90 },
    { title: '操作人员', dataIndex: 'operName' },
    { title: '操作地点', dataIndex: 'operLocation' },
    { title: '状态', dataIndex: 'status', width: 80,
      render: (v: string) => <DictTag options={dicts.sys_common_status || []} value={v} /> },
    { title: '操作时间', dataIndex: 'operTime', width: 160, render: (v: number) => parseTime(v) },
    {
      title: '操作', width: 90,
      render: (_: unknown, row: OperRow) => <Button type="link" size="small" onClick={() => setDetail(row)}>详情</Button>,
    },
  ]

  return (
    <div style={{ padding: 16 }}>
      <Card>
        <Form layout="inline" onFinish={() => crud.handleQuery()} style={{ marginBottom: 16 }}>
          <Form.Item label="系统模块">
            <Input allowClear onChange={(e) => crud.setQuery((q) => ({ ...q, title: e.target.value }))} />
          </Form.Item>
          <Form.Item label="操作人员">
            <Input allowClear onChange={(e) => crud.setQuery((q) => ({ ...q, operName: e.target.value }))} />
          </Form.Item>
          <Form.Item label="类型">
            <Select allowClear style={{ width: 140 }}
              options={(dicts.sys_oper_type || []).map((d) => ({ label: d.label, value: d.value }))}
              onChange={(v) => crud.setQuery((q) => ({ ...q, businessType: v }))} />
          </Form.Item>
          <Form.Item>
            <Space>
              <Button type="primary" htmlType="submit" icon={<SearchOutlined />}>搜索</Button>
              <Button icon={<ReloadOutlined />} onClick={crud.resetQuery}>重置</Button>
            </Space>
          </Form.Item>
        </Form>

        <Space style={{ marginBottom: 16 }}>
          <Button type="primary" danger icon={<DeleteOutlined />} disabled={!crud.single}
            onClick={() => crud.handleDelete(crud.ids.join(','))}>删除</Button>
          <Button type="primary" danger onClick={handleClean}>清空</Button>
          <Button style={{ color: '#e6a23c', borderColor: '#e6a23c' }} icon={<ExportOutlined />}
            onClick={() => crud.handleExport()}>导出</Button>
        </Space>

        <Table
          rowKey="operId" columns={columns} dataSource={crud.rows} loading={crud.loading}
          rowSelection={{ onChange: (keys) => crud.handleSelectionChange([...keys] as (string | number)[], []) }}
          pagination={{ total: crud.total, pageSize: crud.query.pageSize, current: crud.query.pageNum,
            showTotal: (t) => `共 ${t} 条`, onChange: (p, size) => void crud.getList({ pageNum: p, pageSize: size } as never) }}
          scroll={{ x: 'max-content' }}
        />

        <Modal title="操作日志详情" open={!!detail} footer={null} onCancel={() => setDetail(null)} width={720}>
          {detail && (
            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 8, fontSize: 13 }}>
              <div>操作模块：{detail.title}</div>
              <div>操作类型：{detail.businessType}</div>
              <div>操作人员：{detail.operName}</div>
              <div>操作地址：{detail.operIp}</div>
              <div>请求地址：{detail.operUrl}</div>
              <div>请求方式：{detail.requestMethod}</div>
              <div style={{ gridColumn: '1 / -1' }}>操作方法：{detail.method}</div>
              <div style={{ gridColumn: '1 / -1' }}>
                请求参数：<pre style={{ whiteSpace: 'pre-wrap', background: '#f5f5f5', padding: 8 }}>{detail.operParam}</pre>
              </div>
              {detail.errorMsg && (
                <div style={{ gridColumn: '1 / -1' }}>
                  异常信息：<pre style={{ whiteSpace: 'pre-wrap', background: '#fff2f0', padding: 8 }}>{detail.errorMsg}</pre>
                </div>
              )}
            </div>
          )}
        </Modal>
      </Card>
    </div>
  )
}
