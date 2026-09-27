// DictDataDrawer —— 字典数据抽屉（对位基准 dict/detail.vue，700px Drawer 内嵌数据 CRUD）
// 独立路由页 /system/dict-data/index/:dictId 也消费同一组件（两用，对齐基准）

import { useCallback, useEffect, useState } from 'react'
import { Button, Drawer, Form, Input, InputNumber, Modal, Select, Space, Table } from 'antd'
import { PlusOutlined, SearchOutlined, ReloadOutlined } from '@ant-design/icons'
import { listData, getData, addData, updateData, delData } from '@/api/system/dict/data'
import { useDict } from '@/utils/dict'
import DictTag from '@/components/DictTag'

interface DictDataRow {
  [k: string]: unknown
  dictCode: number
  dictLabel: string
  dictValue: string
  dictSort: number
  listClass?: string
  status: string
  remark?: string
}

interface Props {
  open: boolean
  dictType: string
  title?: string
  onClose: () => void
}

const TAG_STYLE_OPTIONS = [
  { label: '默认', value: 'default' },
  { label: '主要', value: 'primary' },
  { label: '成功', value: 'success' },
  { label: '信息', value: 'info' },
  { label: '警告', value: 'warning' },
  { label: '危险', value: 'danger' },
]

export default function DictDataDrawer({ open, dictType, title, onClose }: Props) {
  const dicts = useDict('sys_normal_disable')
  const [rows, setRows] = useState<DictDataRow[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(false)
  const [keyword, setKeyword] = useState('')
  const [form] = Form.useForm()
  const [modalOpen, setModalOpen] = useState(false)
  const [editId, setEditId] = useState<number | null>(null)

  const load = useCallback(async () => {
    if (!dictType) return
    setLoading(true)
    try {
      const res = (await listData({ pageNum: 1, pageSize: 50, dictType, dictLabel: keyword || undefined })) as unknown as {
        rows: DictDataRow[]; total: number
      }
      setRows(res.rows || [])
      setTotal(res.total || 0)
    } finally {
      setLoading(false)
    }
  }, [dictType, keyword])

  useEffect(() => {
    if (open) void load()
  }, [open, load])

  const openAdd = () => {
    setEditId(null); form.resetFields(); setModalOpen(true)
  }
  const openEdit = async (id: number) => {
    const res = (await getData(id)) as unknown as { data: Partial<DictDataRow> }
    setEditId(id); form.setFieldsValue(res.data); setModalOpen(true)
  }
  const submit = async () => {
    const values = await form.validateFields()
    const payload = { ...values, dictType }
    if (editId) await updateData({ ...payload, dictCode: editId })
    else await addData(payload)
    setModalOpen(false)
    void load()
  }
  const handleDelete = (code: number) => {
    Modal.confirm({
      title: '系统提示', content: `是否确认删除编号为"${code}"的数据项？`, okText: '确定', cancelText: '取消',
      onOk: async () => { await delData(code); void load() },
    })
  }

  const columns = [
    { title: '字典编码', dataIndex: 'dictCode', width: 90 },
    { title: '字典标签', dataIndex: 'dictLabel' },
    { title: '字典键值', dataIndex: 'dictValue' },
    { title: '排序', dataIndex: 'dictSort', width: 70 },
    {
      title: '状态', dataIndex: 'status', width: 80,
      render: (v: string) => <DictTag options={dicts.sys_normal_disable || []} value={v} />,
    },
    {
      title: '操作', width: 140,
      render: (_: unknown, row: DictDataRow) => (
        <Space>
          <Button type="link" size="small" onClick={() => void openEdit(row.dictCode)}>修改</Button>
          <Button type="link" size="small" danger onClick={() => handleDelete(row.dictCode)}>删除</Button>
        </Space>
      ),
    },
  ]

  return (
    <Drawer title={title || '字典数据'} open={open} onClose={onClose} width={720} destroyOnHidden>
      <Form layout="inline" style={{ marginBottom: 12 }}>
        <Form.Item label="数据标签">
          <Input allowClear value={keyword} onChange={(e) => setKeyword(e.target.value)} />
        </Form.Item>
        <Form.Item>
          <Space>
            <Button type="primary" icon={<SearchOutlined />} onClick={() => void load()}>搜索</Button>
            <Button icon={<ReloadOutlined />} onClick={() => { setKeyword(''); void load() }}>重置</Button>
          </Space>
        </Form.Item>
      </Form>
      <Space style={{ marginBottom: 12 }}>
        <Button type="primary" icon={<PlusOutlined />} onClick={openAdd}>新增</Button>
      </Space>
      <Table
        rowKey="dictCode" columns={columns} dataSource={rows} loading={loading}
        pagination={{ total, pageSize: 50, showTotal: (t) => `共 ${t} 条` }}
        scroll={{ x: 'max-content' }}
      />
      <Modal title={editId ? '修改字典数据' : '添加字典数据'} open={modalOpen}
        onOk={() => void submit()} onCancel={() => setModalOpen(false)} destroyOnHidden>
        <Form form={form} layout="vertical">
          <Form.Item name="dictLabel" label="数据标签" rules={[{ required: true, message: '数据标签不能为空' }]}>
            <Input />
          </Form.Item>
          <Form.Item name="dictValue" label="数据键值" rules={[{ required: true, message: '数据键值不能为空' }]}>
            <Input />
          </Form.Item>
          <Form.Item name="dictSort" label="显示排序" initialValue={0}>
            <InputNumber min={0} style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item name="listClass" label="样式属性">
            <Select allowClear options={TAG_STYLE_OPTIONS} />
          </Form.Item>
          <Form.Item name="status" label="状态" initialValue="0">
            <Select options={(dicts.sys_normal_disable || []).map((d) => ({ label: d.label, value: d.value }))} />
          </Form.Item>
          <Form.Item name="remark" label="备注"><Input.TextArea /></Form.Item>
        </Form>
      </Modal>
    </Drawer>
  )
}
