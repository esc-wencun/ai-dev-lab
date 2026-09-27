// 字典管理 —— 对位基准 views/system/dict（类型列表 + 行内数据抽屉）
// 完整 CRUD + DictTag(sys_normal_disable) + 刷新缓存；数据列表嵌在抽屉内

import { useEffect, useState } from 'react'
import { Button, Card, Form, Input, Modal, Select, Space, Table } from 'antd'
import { PlusOutlined, DeleteOutlined, SearchOutlined, ReloadOutlined, RedoOutlined } from '@ant-design/icons'
import { listType, getType, addType, updateType, delType, refreshCache } from '@/api/system/dict/type'
import { useCrud } from '@/hooks/useCrud'
import { useDict } from '@/utils/dict'
import DictTag from '@/components/DictTag'
import Auth from '@/components/Auth'
import DictDataDrawer from './DictDataDrawer'

interface DictTypeRow {
  [k: string]: unknown
  dictId: number
  dictName: string
  dictType: string
  status: string
  remark?: string
}

export default function Dict() {
  const crud = useCrud<DictTypeRow, { pageNum: number; pageSize: number; dictName?: string; dictType?: string; status?: string }>({
    listApi: listType,
    delApi: delType,
    defaultQuery: { pageNum: 1, pageSize: 10 },
  })
  const dicts = useDict('sys_normal_disable')
  const [form] = Form.useForm()
  const [open, setOpen] = useState(false)
  const [editId, setEditId] = useState<number | null>(null)
  // 抽屉状态
  const [drawerOpen, setDrawerOpen] = useState(false)
  const [currentType, setCurrentType] = useState<DictTypeRow | null>(null)

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
    const res = (await getType(id)) as unknown as { data: Partial<DictTypeRow> }
    setEditId(id)
    form.setFieldsValue(res.data)
    setOpen(true)
  }

  const submit = async () => {
    const values = await form.validateFields()
    if (editId) await updateType({ ...values, dictId: editId })
    else await addType(values)
    setOpen(false)
    void crud.getList()
  }

  const handleRefreshCache = () => {
    Modal.confirm({
      title: '系统提示', content: '是否确认刷新字典缓存?', okText: '确定', cancelText: '取消',
      onOk: async () => { await refreshCache() },
    })
  }

  const typeColumns = [
    { title: '字典编号', dataIndex: 'dictId', width: 90 },
    { title: '字典名称', dataIndex: 'dictName' },
    {
      title: '字典类型', dataIndex: 'dictType',
      render: (v: string, row: DictTypeRow) => (
        <Button type="link" size="small" onClick={() => { setCurrentType(row); setDrawerOpen(true) }}>{v}</Button>
      ),
    },
    {
      title: '状态', dataIndex: 'status', width: 80,
      render: (v: string) => <DictTag options={dicts.sys_normal_disable || []} value={v} />,
    },
    { title: '备注', dataIndex: 'remark' },
    {
      title: '操作', width: 150,
      render: (_: unknown, row: DictTypeRow) => (
        <Space>
          <Auth permissions={['system:dict:edit']}>
            <Button type="link" size="small" onClick={() => void openEdit(row.dictId)}>修改</Button>
          </Auth>
          <Auth permissions={['system:dict:remove']}>
            <Button type="link" size="small" danger onClick={() => crud.handleDelete(String(row.dictId), [row.dictId])}>删除</Button>
          </Auth>
        </Space>
      ),
    },
  ]

  return (
    <div style={{ padding: 16 }}>
      <Card>
        <Form layout="inline" onFinish={() => crud.handleQuery()} style={{ marginBottom: 16 }}>
          <Form.Item label="字典名称">
            <Input allowClear onChange={(e) => crud.setQuery((q) => ({ ...q, dictName: e.target.value }))} />
          </Form.Item>
          <Form.Item label="字典类型">
            <Input allowClear onChange={(e) => crud.setQuery((q) => ({ ...q, dictType: e.target.value }))} />
          </Form.Item>
          <Form.Item>
            <Space>
              <Button type="primary" htmlType="submit" icon={<SearchOutlined />}>搜索</Button>
              <Button icon={<ReloadOutlined />} onClick={crud.resetQuery}>重置</Button>
            </Space>
          </Form.Item>
        </Form>

        <Space style={{ marginBottom: 16 }}>
          <Auth permissions={['system:dict:add']}>
            <Button type="primary" icon={<PlusOutlined />} onClick={openAdd}>新增</Button>
          </Auth>
          <Auth permissions={['system:dict:remove']}>
            <Button type="primary" danger icon={<DeleteOutlined />} disabled={!crud.multiple}
              onClick={() => crud.handleDelete(crud.ids.join(','))}>删除</Button>
          </Auth>
          <Auth permissions={['system:dict:remove']}>
            <Button icon={<RedoOutlined />} onClick={handleRefreshCache}>刷新缓存</Button>
          </Auth>
        </Space>

        <Table
          rowKey="dictId" columns={typeColumns} dataSource={crud.rows} loading={crud.loading}
          pagination={{ total: crud.total, pageSize: crud.query.pageSize, current: crud.query.pageNum,
            showSizeChanger: true, showTotal: (t) => `共 ${t} 条`,
            onChange: (p, size) => void crud.getList({ pageNum: p, pageSize: size }) }}
          scroll={{ x: 'max-content' }}
        />
      </Card>

      {/* 字典类型 新增/修改弹窗 */}
      <Modal title={editId ? '修改字典类型' : '添加字典类型'} open={open} onOk={() => void submit()} onCancel={() => setOpen(false)} destroyOnHidden>
        <Form form={form} layout="vertical">
          <Form.Item name="dictName" label="字典名称" rules={[{ required: true, message: '字典名称不能为空' }]}>
            <Input />
          </Form.Item>
          <Form.Item name="dictType" label="字典类型" rules={[{ required: true, message: '字典类型不能为空' }]}>
            <Input />
          </Form.Item>
          <Form.Item name="status" label="状态" initialValue="0">
            <Select options={(dicts.sys_normal_disable || []).map((d) => ({ label: d.label, value: d.value }))} />
          </Form.Item>
          <Form.Item name="remark" label="备注"><Input.TextArea /></Form.Item>
        </Form>
      </Modal>

      {/* 字典数据抽屉 */}
      <DictDataDrawer
        open={drawerOpen}
        dictType={currentType?.dictType || ''}
        title={`字典数据 — ${currentType?.dictName || ''}`}
        onClose={() => setDrawerOpen(false)}
      />
    </div>
  )
}
