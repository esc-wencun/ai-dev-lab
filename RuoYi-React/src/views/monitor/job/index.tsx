// 定时任务 —— 对位基准 monitor/job（CRUD + Crontab 生成器 + 立即执行 + 状态切换 + 详情）
// Crontab 七域组件为 3.0.0 批次 D 大件，此处先提供表达式输入 + 简易常用表达式下拉，完整生成器随后补

import { useEffect, useState } from 'react'
import { Button, Card, Form, Input, Modal, Radio, Select, Space, Table } from 'antd'
import { PlusOutlined, SearchOutlined, ReloadOutlined, CaretRightOutlined } from '@ant-design/icons'
import { listJob, getJob, addJob, updateJob, delJob, changeJobStatus, runJob } from '@/api/monitor/job'
import { useCrud } from '@/hooks/useCrud'
import { useDict } from '@/utils/dict'
import DictTag from '@/components/DictTag'
import Auth from '@/components/Auth'
import { parseTime } from '@/utils/ruoyi'

interface JobRow {
  [k: string]: unknown
  jobId: number
  jobName: string
  jobGroup: string
  invokeTarget: string
  cronExpression: string
  misfirePolicy?: string
  concurrent?: string
  status: string
  remark?: string
}

const COMMON_CRONS = [
  { label: '每分钟（0 * * * * ?）', value: '0 * * * * ?' },
  { label: '每小时（0 0 * * * ?）', value: '0 0 * * * ?' },
  { label: '每天 0 点（0 0 0 * * ?）', value: '0 0 0 * * ?' },
  { label: '每天 8 点（0 0 8 * * ?）', value: '0 0 8 * * ?' },
]

export default function Job() {
  const dicts = useDict('sys_job_group', 'sys_job_status')
  const crud = useCrud<JobRow, { pageNum: number; pageSize: number; jobName?: string; jobGroup?: string; status?: string }>({
    listApi: listJob,
    delApi: delJob,
    defaultQuery: { pageNum: 1, pageSize: 10 },
  })
  const [form] = Form.useForm()
  const [open, setOpen] = useState(false)
  const [editId, setEditId] = useState<number | null>(null)
  const [detail, setDetail] = useState<JobRow | null>(null)

  useEffect(() => { void crud.getList() /* eslint-disable-line react-hooks/exhaustive-deps */ }, [])

  const openAdd = () => {
    setEditId(null)
    form.resetFields()
    form.setFieldsValue({ jobGroup: 'DEFAULT', misfirePolicy: '1', concurrent: '1', status: '0' })
    setOpen(true)
  }

  const openEdit = async (id: number) => {
    const res = (await getJob(id)) as unknown as { data: Partial<JobRow> }
    setEditId(id)
    form.setFieldsValue(res.data)
    setOpen(true)
  }

  const submit = async () => {
    const values = await form.validateFields()
    if (editId) await updateJob({ ...values, jobId: editId })
    else await addJob(values)
    setOpen(false)
    void crud.getList()
  }

  const toggleStatus = (row: JobRow) => {
    const next = row.status === '0' ? '1' : '0'
    Modal.confirm({
      title: '系统提示', content: `是否确认改变"${row.jobName}"的状态?`, okText: '确定', cancelText: '取消',
      onOk: async () => { await changeJobStatus(row.jobId, next); void crud.getList() },
    })
  }

  const handleRun = (row: JobRow) => {
    Modal.confirm({
      title: '系统提示', content: `是否确认立即执行一次"${row.jobName}"任务?`, okText: '确定', cancelText: '取消',
      onOk: async () => { await runJob(row.jobId, row.jobGroup) },
    })
  }

  const columns = [
    { title: '任务编号', dataIndex: 'jobId', width: 90 },
    { title: '任务名称', dataIndex: 'jobName' },
    { title: '任务组名', dataIndex: 'jobGroup', width: 100,
      render: (v: string) => <DictTag options={dicts.sys_job_group || []} value={v} /> },
    { title: '调用目标', dataIndex: 'invokeTarget' },
    { title: 'cron执行表达式', dataIndex: 'cronExpression' },
    { title: '状态', dataIndex: 'status', width: 80,
      render: (v: string, row: JobRow) => (
        <Button type="link" size="small" onClick={() => toggleStatus(row)}>
          <DictTag options={dicts.sys_job_status || []} value={v} />
        </Button>
      ) },
    {
      title: '操作', width: 230,
      render: (_: unknown, row: JobRow) => (
        <Space>
          <Auth permissions={['monitor:job:edit']}>
            <Button type="link" size="small" onClick={() => void openEdit(row.jobId)}>修改</Button>
          </Auth>
          <Auth permissions={['monitor:job:changeStatus']}>
            <Button type="link" size="small" icon={<CaretRightOutlined />} onClick={() => handleRun(row)}>执行一次</Button>
          </Auth>
          <Auth permissions={['monitor:job:remove']}>
            <Button type="link" size="small" danger onClick={() => crud.handleDelete(String(row.jobId), [row.jobId])}>删除</Button>
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
              options={(dicts.sys_job_status || []).map((d) => ({ label: d.label, value: d.value }))}
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
          <Auth permissions={['monitor:job:add']}>
            <Button type="primary" icon={<PlusOutlined />} onClick={openAdd}>新增</Button>
          </Auth>
        </Space>

        <Table
          rowKey="jobId" columns={columns} dataSource={crud.rows} loading={crud.loading}
          pagination={{ total: crud.total, pageSize: crud.query.pageSize, current: crud.query.pageNum,
            showTotal: (t) => `共 ${t} 条`, onChange: (p, size) => void crud.getList({ pageNum: p, pageSize: size } as never) }}
          scroll={{ x: 'max-content' }}
        />
      </Card>

      <Modal title={editId ? '修改任务' : '添加任务'} open={open} width={640} onOk={() => void submit()} onCancel={() => setOpen(false)} destroyOnHidden>
        <Form form={form} layout="vertical">
          <Form.Item name="jobName" label="任务名称" rules={[{ required: true, message: '任务名称不能为空' }]}>
            <Input />
          </Form.Item>
          <Form.Item name="jobGroup" label="任务组名" rules={[{ required: true, message: '任务组名不能为空' }]}>
            <Select options={(dicts.sys_job_group || []).map((d) => ({ label: d.label, value: d.value }))} />
          </Form.Item>
          <Form.Item name="invokeTarget" label="调用目标" rules={[{ required: true, message: '调用目标不能为空' }]}>
            <Input placeholder="如 ryTask.ryParams('ry')" />
          </Form.Item>
          <Form.Item name="cronExpression" label="cron执行表达式" rules={[{ required: true, message: 'cron执行表达式不能为空' }]}>
            <Input placeholder="如 0 0 8 * * ?" />
          </Form.Item>
          <Form.Item name="cronPick" label="常用表达式">
            <Select allowClear options={COMMON_CRONS}
              onChange={(v) => { if (v) form.setFieldValue('cronExpression', v) }} />
          </Form.Item>
          <Form.Item name="misfirePolicy" label="执行策略" initialValue="1">
            <Radio.Group options={[
              { label: '立即执行', value: '1' },
              { label: '放弃执行', value: '2' },
              { label: '恢复执行', value: '3' },
            ]} />
          </Form.Item>
          <Form.Item name="concurrent" label="是否并发" initialValue="1">
            <Radio.Group options={[{ label: '允许', value: '1' }, { label: '禁止', value: '0' }]} />
          </Form.Item>
          <Form.Item name="status" label="状态" initialValue="0">
            <Radio.Group options={[{ label: '正常', value: '0' }, { label: '暂停', value: '1' }]} />
          </Form.Item>
          <Form.Item name="remark" label="备注"><Input.TextArea /></Form.Item>
        </Form>
      </Modal>

      <Modal title="任务详情" open={!!detail} footer={null} onCancel={() => setDetail(null)} width={680}>
        {detail && (
          <div style={{ fontSize: 13, lineHeight: 2 }}>
            <div>任务编号：{detail.jobId}</div>
            <div>任务名称：{detail.jobName}</div>
            <div>任务组名：{detail.jobGroup}</div>
            <div>调用目标：{detail.invokeTarget}</div>
            <div>cron表达式：{detail.cronExpression}</div>
            <div>状态：{detail.status === '0' ? '正常' : '暂停'}</div>
            <div>创建时间：{parseTime(detail.createTime)}</div>
          </div>
        )}
      </Modal>
    </div>
  )
}
