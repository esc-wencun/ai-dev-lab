// 通知公告 —— 对位基准 views/system/notice/index.vue
// CRUD + RichEditor 富文本（HTML 存储）+ 详情弹窗 + ReadUsers 已读用户

import { useEffect, useState } from 'react'
import { Button, Card, Form, Input, Modal, Radio, Select, Space, Table, Typography } from 'antd'
import { PlusOutlined, DeleteOutlined, SearchOutlined, ReloadOutlined } from '@ant-design/icons'
import { listNotice, getNotice, addNotice, updateNotice, delNotice } from '@/api/system/notice'
import { useCrud } from '@/hooks/useCrud'
import { useDict } from '@/utils/dict'
import DictTag from '@/components/DictTag'
import Auth from '@/components/Auth'
import RichEditor from '@/components/RichEditor'
import ReadUsersDialog from './ReadUsers'
import { parseTime } from '@/utils/ruoyi'

interface NoticeRow {
  [k: string]: unknown
  noticeId: number
  noticeTitle: string
  noticeType: string
  status: string
  noticeContent?: string
  createTime?: string
  createBy?: string
}

export default function Notice() {
  const dicts = useDict('sys_notice_status', 'sys_notice_type')
  const crud = useCrud<NoticeRow, { pageNum: number; pageSize: number; noticeTitle?: string; createBy?: string; noticeType?: string }>({
    listApi: listNotice,
    delApi: delNotice,
    defaultQuery: { pageNum: 1, pageSize: 10 },
  })
  const [form] = Form.useForm()
  const [open, setOpen] = useState(false)
  const [editId, setEditId] = useState<number | null>(null)
  // 详情弹窗
  const [detail, setDetail] = useState<NoticeRow | null>(null)
  // 已读用户弹窗
  const [readUsers, setReadUsers] = useState<{ open: boolean; noticeId?: number }>({ open: false })

  useEffect(() => {
    void crud.getList()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  const openAdd = () => {
    setEditId(null)
    form.resetFields()
    form.setFieldValue('noticeType', '1')
    setOpen(true)
  }

  const openEdit = async (id: number) => {
    const res = (await getNotice(id)) as unknown as { data: Partial<NoticeRow> }
    setEditId(id)
    form.setFieldsValue(res.data)
    setOpen(true)
  }

  const submit = async () => {
    const values = await form.validateFields()
    if (editId) await updateNotice({ ...values, noticeId: editId })
    else await addNotice(values)
    setOpen(false)
    void crud.getList()
  }

  const openDetail = async (id: number) => {
    const res = (await getNotice(id)) as unknown as { data: NoticeRow }
    setDetail(res.data)
  }

  const columns = [
    { title: '序号', dataIndex: 'noticeId', width: 70 },
    { title: '公告标题', dataIndex: 'noticeTitle',
      render: (v: string, row: NoticeRow) => <Button type="link" size="small" onClick={() => void openDetail(row.noticeId)}>{v}</Button> },
    { title: '公告类型', dataIndex: 'noticeType', width: 90,
      render: (v: string) => <DictTag options={dicts.sys_notice_type || []} value={v} /> },
    { title: '状态', dataIndex: 'status', width: 80,
      render: (v: string) => <DictTag options={dicts.sys_notice_status || []} value={v} /> },
    { title: '创建者', dataIndex: 'createBy', width: 90 },
    { title: '创建时间', dataIndex: 'createTime', width: 160, render: (v: string) => parseTime(v) },
    {
      title: '操作', width: 230,
      render: (_: unknown, row: NoticeRow) => (
        <Space>
          <Auth permissions={['system:notice:edit']}>
            <Button type="link" size="small" onClick={() => void openEdit(row.noticeId)}>修改</Button>
          </Auth>
          <Auth permissions={['system:notice:list']}>
            <Button type="link" size="small" onClick={() => setReadUsers({ open: true, noticeId: row.noticeId })}>阅读用户</Button>
          </Auth>
          <Auth permissions={['system:notice:remove']}>
            <Button type="link" size="small" danger onClick={() => crud.handleDelete(String(row.noticeId), [row.noticeId])}>删除</Button>
          </Auth>
        </Space>
      ),
    },
  ]

  return (
    <div style={{ padding: 16 }}>
      <Card>
        <Form layout="inline" onFinish={() => crud.handleQuery()} style={{ marginBottom: 16 }}>
          <Form.Item label="公告标题">
            <Input allowClear onChange={(e) => crud.setQuery((q) => ({ ...q, noticeTitle: e.target.value }))} />
          </Form.Item>
          <Form.Item label="操作人员">
            <Input allowClear onChange={(e) => crud.setQuery((q) => ({ ...q, createBy: e.target.value }))} />
          </Form.Item>
          <Form.Item label="类型">
            <Select allowClear style={{ width: 120 }}
              options={(dicts.sys_notice_type || []).map((d) => ({ label: d.label, value: d.value }))}
              onChange={(v) => crud.setQuery((q) => ({ ...q, noticeType: v }))} />
          </Form.Item>
          <Form.Item>
            <Space>
              <Button type="primary" htmlType="submit" icon={<SearchOutlined />}>搜索</Button>
              <Button icon={<ReloadOutlined />} onClick={crud.resetQuery}>重置</Button>
            </Space>
          </Form.Item>
        </Form>

        <Space style={{ marginBottom: 16 }}>
          <Auth permissions={['system:notice:add']}>
            <Button type="primary" icon={<PlusOutlined />} onClick={openAdd}>新增</Button>
          </Auth>
          <Auth permissions={['system:notice:remove']}>
            <Button type="primary" danger icon={<DeleteOutlined />} disabled={!crud.multiple}
              onClick={() => crud.handleDelete(crud.ids.join(','))}>删除</Button>
          </Auth>
        </Space>

        <Table
          rowKey="noticeId" columns={columns} dataSource={crud.rows} loading={crud.loading}
          pagination={{ total: crud.total, pageSize: crud.query.pageSize, current: crud.query.pageNum,
            showSizeChanger: true, showTotal: (t) => `共 ${t} 条`,
            onChange: (p, size) => void crud.getList({ pageNum: p, pageSize: size }) }}
          scroll={{ x: 'max-content' }}
        />
      </Card>

      {/* 新增/修改弹窗（含富文本） */}
      <Modal title={editId ? '修改公告' : '添加公告'} open={open} width={760}
        onOk={() => void submit()} onCancel={() => setOpen(false)} destroyOnHidden>
        <Form form={form} layout="vertical">
          <Form.Item name="noticeTitle" label="公告标题" rules={[{ required: true, message: '公告标题不能为空' }]}>
            <Input />
          </Form.Item>
          <Form.Item name="noticeType" label="公告类型" rules={[{ required: true, message: '公告类型不能为空' }]}>
            <Radio.Group options={(dicts.sys_notice_type || []).map((d) => ({ label: d.label, value: d.value }))} />
          </Form.Item>
          <Form.Item name="status" label="状态" initialValue="0">
            <Radio.Group options={(dicts.sys_notice_status || []).map((d) => ({ label: d.label, value: d.value }))} />
          </Form.Item>
          <Form.Item name="noticeContent" label="内容">
            <RichEditor height={220} />
          </Form.Item>
        </Form>
      </Modal>

      {/* 详情弹窗（HTML 渲染对位 v-html / dangerouslySetInnerHTML，deviations #14） */}
      <Modal title="公告详情" open={!!detail} footer={null} onCancel={() => setDetail(null)} width={720}>
        {detail && (
          <div style={{ padding: 8 }}>
            <Typography.Title level={4} style={{ textAlign: 'center' }}>{detail.noticeTitle}</Typography.Title>
            <div style={{ textAlign: 'center', marginBottom: 16, color: '#999' }}>
              {detail.createBy && <span>创建人：{detail.createBy}　</span>}
              {detail.createTime && <span>{parseTime(detail.createTime)}</span>}
            </div>
            <div dangerouslySetInnerHTML={{ __html: detail.noticeContent || '' }} />
          </div>
        )}
      </Modal>

      <ReadUsersDialog open={readUsers.open} noticeId={readUsers.noticeId} onClose={() => setReadUsers({ open: false })} />
    </div>
  )
}
