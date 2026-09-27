// 调度日志 —— 对位基准 monitor/job/log.vue（子路由 /monitor/job-log/index/:jobId）

import { useEffect, useState } from 'react'
import { useParams } from 'react-router'
import { Button, Card, Form, Input, Modal, Select, Space, Table } from 'antd'
import { DeleteOutlined, SearchOutlined, ReloadOutlined, ExportOutlined } from '@ant-design/icons'
import { listJobLog, delJobLog, cleanJobLog } from '@/api/monitor/jobLog'
import { useCrud } from '@/hooks/useCrud'
import { useDict } from '@/utils/dict'
import DictTag from '@/components/DictTag'
import Auth from '@/components/Auth'
import { parseTime } from '@/utils/ruoyi'

interface JobLogRow {
  [k: string]: unknown
  jobLogId: number
  jobName?: string
  jobGroup?: string
  invokeTarget?: string
  jobMessage?: string
  status: string
  exceptionInfo?: string
  createTime?: number
}

export default function JobLog() {
  const dicts = useDict('sys_common_status', 'sys_job_group')
  const { jobId } = useParams()
  const crud = useCrud<JobLogRow, { pageNum: number; pageSize: number; jobName?: string; jobGroup?: string; status?: string; jobId?: number }>({
    listApi: listJobLog,
    delApi: delJobLog,
    exportUrl: '/monitor/jobLog/export',
    defaultQuery: { pageNum: 1, pageSize: 10, jobId: jobId ? Number(jobId) : undefined },
  })
  const [detail, setDetail] = useState<JobLogRow | null>(null)

  useEffect(() => { void crud.getList() /* eslint-disable-line react-hooks/exhaustive-deps */ }, [])

  const handleClean = () => {
    Modal.confirm({
      title: '系统提示', content: '是否确认清空所有调度日志数据项?', okText: '确定', cancelText: '取消',
      onOk: async () => { await cleanJobLog(); void crud.getList() },
    })
  }

  const columns = [
    { title: '日志编号', dataIndex: 'jobLogId', width: 90 },
    { title: '任务名称', dataIndex: 'jobName' },
    { title: '任务组名', dataIndex: 'jobGroup', width: 100,
      render: (v: string) => <DictTag options={dicts.sys_job_group || []} value={v} /> },
    { title: '调用目标', dataIndex: 'invokeTarget' },
    { title: '日志信息', dataIndex: 'jobMessage' },
    { title: '状态', dataIndex: 'status', width: 80,
      render: (v: string) => <DictTag options={dicts.sys_common_status || []} value={v} /> },
    { title: '执行时间', dataIndex: 'createTime', width: 160, render: (v: number) => parseTime(v) },
    {
      title: '操作', width: 150,
      render: (_: unknown, row: JobLogRow) => (
        <Space>
          <Button type="link" size="small" onClick={() => setDetail(row)}>详情</Button>
          <Auth permissions={['monitor:job:remove']}>
            <Button type="link" size="small" danger onClick={() => crud.handleDelete(String(row.jobLogId), [row.jobLogId])}>删除</Button>
          </Auth>
        </Space>
      ),
    },
  ]

  return (
    <div style={{ padding: 16 }}>
      <Card>
        <Form layout="inline" onFinish={() => crud.handleQuery()} style={{ marginBottom: 16 }}>
          <Form.Item label="任务名称">
            <Input allowClear onChange={(e) => crud.setQuery((q) => ({ ...q, jobName: e.target.value }))} />
          </Form.Item>
          <Form.Item label="任务组名">
            <Select allowClear style={{ width: 140 }}
              options={(dicts.sys_job_group || []).map((d) => ({ label: d.label, value: d.value }))}
              onChange={(v) => crud.setQuery((q) => ({ ...q, jobGroup: v }))} />
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
          <Auth permissions={['monitor:job:remove']}>
            <Button type="primary" danger icon={<DeleteOutlined />} disabled={!crud.multiple}
              onClick={() => crud.handleDelete(crud.ids.join(','))}>删除</Button>
          </Auth>
          <Auth permissions={['monitor:job:remove']}>
            <Button type="primary" danger onClick={handleClean}>清空</Button>
          </Auth>
          <Auth permissions={['monitor:job:export']}>
            <Button style={{ color: '#e6a23c', borderColor: '#e6a23c' }} icon={<ExportOutlined />}
              onClick={() => crud.handleExport()}>导出</Button>
          </Auth>
        </Space>

        <Table
          rowKey="jobLogId" columns={columns} dataSource={crud.rows} loading={crud.loading}
          rowSelection={{ onChange: (keys) => crud.handleSelectionChange([...keys] as (string | number)[], []) }}
          pagination={{ total: crud.total, pageSize: crud.query.pageSize, current: crud.query.pageNum,
            showTotal: (t) => `共 ${t} 条`, onChange: (p, size) => void crud.getList({ pageNum: p, pageSize: size } as never) }}
          scroll={{ x: 'max-content' }}
        />

        <Modal title="调度日志详情" open={!!detail} footer={null} onCancel={() => setDetail(null)} width={680}>
          {detail && (
            <div style={{ fontSize: 13, lineHeight: 2 }}>
              <div>任务编号：{detail.jobLogId}</div>
              <div>任务名称：{detail.jobName}</div>
              <div>任务组名：{detail.jobGroup}</div>
              <div>调用目标：{detail.invokeTarget}</div>
              <div>日志信息：{detail.jobMessage}</div>
              <div>执行状态：{detail.status === '0' ? '成功' : '失败'}</div>
              {detail.exceptionInfo && (
                <div>异常信息：<pre style={{ whiteSpace: 'pre-wrap', background: '#fff2f0', padding: 8 }}>{detail.exceptionInfo}</pre></div>
              )}
            </div>
          )}
        </Modal>
      </Card>
    </div>
  )
}
