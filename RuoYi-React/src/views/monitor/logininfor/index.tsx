// 登录日志 —— 对位基准 monitor/logininfor（查询 + 解锁账户 + 删除/清空/导出）

import { useEffect } from 'react'
import { Button, Card, Form, Input, Modal, Select, Space, Table } from 'antd'
import { DeleteOutlined, SearchOutlined, ReloadOutlined, ExportOutlined, UnlockOutlined } from '@ant-design/icons'
import { list, delLogininfor, unlockLogininfor, cleanLogininfor } from '@/api/monitor/logininfor'
import { useCrud } from '@/hooks/useCrud'
import { useDict } from '@/utils/dict'
import DictTag from '@/components/DictTag'
import { parseTime } from '@/utils/ruoyi'
import { message } from 'antd'

interface LoginRow {
  [k: string]: unknown
  infoId: number
  userName?: string
  ipaddr?: string
  loginLocation?: string
  browser?: string
  os?: string
  status: string
  msg?: string
  loginTime?: number
}

export default function Logininfor() {
  const dicts = useDict('sys_common_status')
  const crud = useCrud<LoginRow, { pageNum: number; pageSize: number; userName?: string; ipaddr?: string; status?: string }>({
    listApi: list,
    delApi: delLogininfor,
    exportUrl: '/monitor/logininfor/export',
    defaultQuery: { pageNum: 1, pageSize: 10 },
  })

  useEffect(() => { void crud.getList() /* eslint-disable-line react-hooks/exhaustive-deps */ }, [])

  const handleUnlock = (row: LoginRow) => {
    void unlockLogininfor(String(row.userName)).then(() => message.success('解锁成功'))
  }

  const handleClean = () => {
    Modal.confirm({
      title: '系统提示', content: '是否确认清空所有登录日志数据项?', okText: '确定', cancelText: '取消',
      onOk: async () => { await cleanLogininfor(); void crud.getList() },
    })
  }

  const columns = [
    { title: '日志编号', dataIndex: 'infoId', width: 90 },
    { title: '登录账号', dataIndex: 'userName' },
    { title: '登录地址', dataIndex: 'ipaddr' },
    { title: '登录地点', dataIndex: 'loginLocation' },
    { title: '浏览器', dataIndex: 'browser' },
    { title: '操作系统', dataIndex: 'os' },
    { title: '状态', dataIndex: 'status', width: 80,
      render: (v: string) => <DictTag options={dicts.sys_common_status || []} value={v} /> },
    { title: '登录时间', dataIndex: 'loginTime', width: 160, render: (v: number) => parseTime(v) },
    {
      title: '操作', width: 90,
      render: (_: unknown, row: LoginRow) => (
        <Button type="link" size="small" icon={<UnlockOutlined />} onClick={() => handleUnlock(row)}>解锁</Button>
      ),
    },
  ]

  return (
    <div style={{ padding: 16 }}>
      <Card>
        <Form layout="inline" onFinish={() => crud.handleQuery()} style={{ marginBottom: 16 }}>
          <Form.Item label="登录地址">
            <Input allowClear onChange={(e) => crud.setQuery((q) => ({ ...q, ipaddr: e.target.value }))} />
          </Form.Item>
          <Form.Item label="登录账号">
            <Input allowClear onChange={(e) => crud.setQuery((q) => ({ ...q, userName: e.target.value }))} />
          </Form.Item>
          <Form.Item label="状态">
            <Select allowClear style={{ width: 120 }}
              options={(dicts.sys_common_status || []).map((d) => ({ label: d.label, value: d.value }))}
              onChange={(v) => crud.setQuery((q) => ({ ...q, status: v }))} />
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
          rowKey="infoId" columns={columns} dataSource={crud.rows} loading={crud.loading}
          rowSelection={{ onChange: (keys) => crud.handleSelectionChange([...keys] as (string | number)[], []) }}
          pagination={{ total: crud.total, pageSize: crud.query.pageSize, current: crud.query.pageNum,
            showTotal: (t) => `共 ${t} 条`, onChange: (p, size) => void crud.getList({ pageNum: p, pageSize: size } as never) }}
          scroll={{ x: 'max-content' }}
        />
      </Card>
    </div>
  )
}
