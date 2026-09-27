// 岗位管理 —— 对位基准 views/system/post/index.vue
// 查询（编码/名称/状态）/ 新增 / 修改 / 删除 / 批量删除 + DictTag(sys_normal_disable) + Auth

import { useEffect, useState } from 'react'
import { Button, Card, Form, Input, InputNumber, Modal, Select, Space, Table } from 'antd'
import { PlusOutlined, DeleteOutlined, SearchOutlined, ReloadOutlined } from '@ant-design/icons'
import { listPost, getPost, addPost, updatePost, delPost } from '@/api/system/post'
import { useCrud } from '@/hooks/useCrud'
import { useDict } from '@/utils/dict'
import DictTag from '@/components/DictTag'
import Auth from '@/components/Auth'

interface PostRow {
  [k: string]: unknown
  postId: number
  postCode: string
  postName: string
  postSort: number
  status: string
  remark?: string
}

export default function Post() {
  const dicts = useDict('sys_normal_disable')
  const crud = useCrud<PostRow, { pageNum: number; pageSize: number; postCode?: string; postName?: string; status?: string }>({
    listApi: listPost,
    delApi: delPost,
    exportUrl: '/system/post/export',
    defaultQuery: { pageNum: 1, pageSize: 10 },
  })
  const [form] = Form.useForm()
  const [open, setOpen] = useState(false)
  const [editId, setEditId] = useState<number | null>(null)

  useEffect(() => {
    void crud.getList()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  const openAdd = () => {
    setEditId(null)
    form.resetFields()
    setOpen(true)
  }

  const openEdit = async (id: number) => {
    const res = (await getPost(id)) as unknown as { data: Partial<PostRow> }
    setEditId(id)
    form.setFieldsValue(res.data)
    setOpen(true)
  }

  const submit = async () => {
    const values = await form.validateFields()
    if (editId) {
      await updatePost({ ...values, postId: editId })
    } else {
      await addPost(values)
    }
    setOpen(false)
    void crud.getList()
  }

  const columns = [
    { title: '岗位编号', dataIndex: 'postId', width: 90 },
    { title: '岗位编码', dataIndex: 'postCode' },
    { title: '岗位名称', dataIndex: 'postName' },
    { title: '岗位排序', dataIndex: 'postSort', width: 90 },
    {
      title: '状态', dataIndex: 'status', width: 90,
      render: (v: string) => <DictTag options={dicts.sys_normal_disable || []} value={v} />,
    },
    {
      title: '操作', width: 150,
      render: (_: unknown, row: PostRow) => (
        <Space>
          <Auth permissions={['system:post:edit']}>
            <Button type="link" size="small" onClick={() => void openEdit(row.postId)}>修改</Button>
          </Auth>
          <Auth permissions={['system:post:remove']}>
            <Button type="link" size="small" danger onClick={() => crud.handleDelete(String(row.postId), [row.postId])}>删除</Button>
          </Auth>
        </Space>
      ),
    },
  ]

  return (
    <div style={{ padding: 16 }}>
      <Card>
        <Form layout="inline" onFinish={() => crud.handleQuery()} style={{ marginBottom: 16 }}>
          <Form.Item label="岗位编码">
            <Input allowClear onChange={(e) => crud.setQuery((q) => ({ ...q, postCode: e.target.value }))} />
          </Form.Item>
          <Form.Item label="岗位名称">
            <Input allowClear onChange={(e) => crud.setQuery((q) => ({ ...q, postName: e.target.value }))} />
          </Form.Item>
          <Form.Item label="状态">
            <Select allowClear style={{ width: 120 }} placeholder="请选择"
              options={(dicts.sys_normal_disable || []).map((d) => ({ label: d.label, value: d.value }))}
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
          <Auth permissions={['system:post:add']}>
            <Button type="primary" icon={<PlusOutlined />} onClick={openAdd}>新增</Button>
          </Auth>
          <Auth permissions={['system:post:remove']}>
            <Button type="primary" danger icon={<DeleteOutlined />} disabled={!crud.multiple}
              onClick={() => crud.handleDelete(crud.ids.join(','))}>删除</Button>
          </Auth>
          <Auth permissions={['system:post:export']}>
            <Button style={{ color: '#e6a23c', borderColor: '#e6a23c' }} onClick={() => crud.handleExport()}>导出</Button>
          </Auth>
        </Space>

        <Table
          rowKey="postId"
          columns={columns}
          dataSource={crud.rows}
          loading={crud.loading}
          pagination={{
            total: crud.total, pageSize: crud.query.pageSize, current: crud.query.pageNum,
            showSizeChanger: true, showTotal: (t) => `共 ${t} 条`,
            onChange: (p, size) => void crud.getList({ pageNum: p, pageSize: size }),
          }}
          scroll={{ x: 'max-content' }}
        />
      </Card>

      <Modal title={editId ? '修改岗位' : '添加岗位'} open={open} onOk={() => void submit()} onCancel={() => setOpen(false)} destroyOnHidden>
        <Form form={form} layout="vertical">
          <Form.Item name="postName" label="岗位名称" rules={[{ required: true, message: '岗位名称不能为空' }]}>
            <Input />
          </Form.Item>
          <Form.Item name="postCode" label="岗位编码" rules={[{ required: true, message: '岗位编码不能为空' }]}>
            <Input />
          </Form.Item>
          <Form.Item name="postSort" label="岗位顺序" initialValue={0} rules={[{ required: true, message: '岗位顺序不能为空' }]}>
            <InputNumber min={0} style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item name="status" label="状态" initialValue="0">
            <Select options={(dicts.sys_normal_disable || []).map((d) => ({ label: d.label, value: d.value }))} />
          </Form.Item>
          <Form.Item name="remark" label="备注"><Input.TextArea /></Form.Item>
        </Form>
      </Modal>
    </div>
  )
}
