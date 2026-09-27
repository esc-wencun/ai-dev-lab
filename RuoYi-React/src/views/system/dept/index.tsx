// 部门管理 —— 对位基准 views/system/dept/index.vue
// 树表 + 展开/折叠切换 + 批量排序(updateDeptSort) + excludeChild + DictTag

import { useCallback, useEffect, useState } from 'react'
import { Button, Card, Form, Input, Modal, InputNumber, Radio, Space, Table } from 'antd'
import { PlusOutlined, SearchOutlined, ReloadOutlined, OrderedListOutlined, ColumnWidthOutlined } from '@ant-design/icons'
import { listDept, listDeptExcludeChild, getDept, addDept, updateDept, delDept, updateDeptSort } from '@/api/system/dept'
import { useDict } from '@/utils/dict'
import DictTag from '@/components/DictTag'
import Auth from '@/components/Auth'
import { handleTree } from '@/utils/ruoyi'

interface DeptRow {
  [k: string]: unknown
  deptId: number
  parentId: number
  deptName: string
  orderNum: number
  leader?: string
  status: string
  children?: DeptRow[]
}

export default function Dept() {
  const dicts = useDict('sys_normal_disable')
  const [rows, setRows] = useState<DeptRow[]>([])
  const [loading, setLoading] = useState(false)
  const [expanded, setExpanded] = useState<React.Key[]>([])
  const [expandAll, setExpandAll] = useState(true)
  const [keyword, setKeyword] = useState('')
  const [checkedKeys, setCheckedKeys] = useState<React.Key[]>([])
  const [form] = Form.useForm()
  const [open, setOpen] = useState(false)
  const [editId, setEditId] = useState<number | null>(null)

  const load = useCallback(async () => {
    setLoading(true)
    try {
      const res = (await listDept()) as unknown as { data: DeptRow[] }
      const flat = res.data || []
      // 查询关键词时基准行为是前端过滤
      const filtered = keyword
        ? flat.filter((d) => d.deptName.includes(keyword) || String(d.status).includes(keyword))
        : flat
      setRows(handleTree(filtered, 'deptId', 'parentId', 'children'))
    } finally {
      setLoading(false)
    }
  }, [keyword])

  useEffect(() => {
    void load()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  // 展开/折叠全部
  const toggleExpand = () => {
    const next = !expandAll
    setExpandAll(next)
    if (next) {
      const keys: React.Key[] = []
      const walk = (items: DeptRow[]) => items.forEach((i) => { keys.push(i.deptId); if (i.children) walk(i.children) })
      walk(rows)
      setExpanded(keys)
    } else {
      setExpanded([])
    }
  }

  const openAdd = (parentId?: number) => {
    setEditId(null)
    form.resetFields()
    form.setFieldsValue({ parentId, orderNum: 0, status: '0' })
    setOpen(true)
  }

  const openEdit = async (id: number) => {
    // 排除自身及子级（excludeChild 契约）
    await listDeptExcludeChild(id)
    const res = (await getDept(id)) as unknown as { data: Partial<DeptRow> }
    setEditId(id)
    form.setFieldsValue(res.data)
    setOpen(true)
  }

  const submit = async () => {
    const values = await form.validateFields()
    if (editId) await updateDept({ ...values, deptId: editId })
    else await addDept(values)
    setOpen(false)
    void load()
  }

  const handleDeleteOne = (row: DeptRow) => {
    Modal.confirm({
      title: '系统提示', content: `是否确认删除名称为"${row.deptName}"的数据项？`, okText: '确定', cancelText: '取消',
      onOk: async () => { await delDept(row.deptId); void load() },
    })
  }

  // 批量排序（勾选行 + Sort → updateDeptSort：deptIds + orderNums）
  const handleSort = () => {
    if (checkedKeys.length === 0) return
    const flat = collectFlat(rows)
    const chosen = checkedKeys.map((k) => flat.find((d) => d.deptId === k)).filter(Boolean) as DeptRow[]
    Modal.confirm({
      title: '系统提示',
      content: `已选择 ${chosen.length} 行，将按当前显示顺序保存排序，是否继续？`,
      okText: '确定', cancelText: '取消',
      onOk: async () => {
        await updateDeptSort({ deptIds: chosen.map((d) => d.deptId), orderNums: chosen.map((d) => d.orderNum) })
      },
    })
  }

  function collectFlat(items: DeptRow[]): DeptRow[] {
    const out: DeptRow[] = []
    const walk = (list: DeptRow[]) => list.forEach((i) => { out.push(i); if (i.children) walk(i.children) })
    walk(items)
    return out
  }

  const columns = [
    { title: '部门名称', dataIndex: 'deptName' },
    { title: '排序', dataIndex: 'orderNum', width: 70 },
    { title: '负责人', dataIndex: 'leader', width: 90 },
    { title: '状态', dataIndex: 'status', width: 80,
      render: (v: string) => <DictTag options={dicts.sys_normal_disable || []} value={v} /> },
    {
      title: '操作', width: 200,
      render: (_: unknown, row: DeptRow) => (
        <Space>
          <Auth permissions={['system:dept:edit']}>
            <Button type="link" size="small" onClick={() => void openEdit(row.deptId)}>修改</Button>
          </Auth>
          <Auth permissions={['system:dept:add']}>
            <Button type="link" size="small" onClick={() => openAdd(row.deptId)}>新增下级</Button>
          </Auth>
          <Auth permissions={['system:dept:remove']}>
            <Button type="link" size="small" danger onClick={() => handleDeleteOne(row)}>删除</Button>
          </Auth>
        </Space>
      ),
    },
  ]

  return (
    <div style={{ padding: 16 }}>
      <Card>
        <Form layout="inline" style={{ marginBottom: 16 }}>
          <Form.Item label="部门名称">
            <Input allowClear onChange={(e) => setKeyword(e.target.value)} />
          </Form.Item>
          <Form.Item>
            <Space>
              <Button type="primary" icon={<SearchOutlined />} onClick={() => void load()}>搜索</Button>
              <Button icon={<ReloadOutlined />} onClick={() => { setKeyword(''); void load() }}>重置</Button>
            </Space>
          </Form.Item>
        </Form>

        <Space style={{ marginBottom: 16 }}>
          <Auth permissions={['system:dept:add']}>
            <Button type="primary" icon={<PlusOutlined />} onClick={() => openAdd()}>新增</Button>
          </Auth>
          <Auth permissions={['system:dept:edit']}>
            <Button icon={<OrderedListOutlined />} disabled={checkedKeys.length === 0} onClick={handleSort}>批量排序</Button>
          </Auth>
          <Button icon={<ColumnWidthOutlined />} onClick={toggleExpand}>
            {expandAll ? '折叠' : '展开'}
          </Button>
        </Space>

        <Table
          rowKey="deptId"
          columns={columns}
          dataSource={rows}
          loading={loading}
          rowSelection={{ checkStrictly: false, onChange: (keys) => setCheckedKeys(keys as React.Key[]) }}
          expandable={{
            expandedRowKeys: expanded,
            onExpandedRowsChange: (keys) => setExpanded([...keys]),
            defaultExpandAllRows: true,
          }}
          pagination={false}
          scroll={{ x: 'max-content' }}
        />
      </Card>

      {/* 部门弹窗 */}
      <Modal title={editId ? '修改部门' : '添加部门'} open={open} onOk={() => void submit()} onCancel={() => setOpen(false)} destroyOnHidden>
        <Form form={form} layout="vertical">
          <Form.Item name="parentId" label="上级部门">
            <InputNumber style={{ width: '100%' }} min={0} />
          </Form.Item>
          <Form.Item name="deptName" label="部门名称" rules={[{ required: true, message: '部门名称不能为空' }]}>
            <Input />
          </Form.Item>
          <Form.Item name="orderNum" label="显示排序" initialValue={0}>
            <InputNumber min={0} style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item name="leader" label="负责人"><Input /></Form.Item>
          <Form.Item name="status" label="状态" initialValue="0">
            <Radio.Group options={[{ label: '正常', value: '0' }, { label: '停用', value: '1' }]} />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}
