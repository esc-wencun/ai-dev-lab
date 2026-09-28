// 菜单管理 —— 对位基准 views/system/menu/index.vue
// 树表 + IconSelect 图标选择 + 三种菜单类型 + 批量排序(updateMenuSort)
// 3.0.0 批次 C：工具行接入 RightToolbar（搜索折叠 + 刷新；树表无分页不接 Pagination）

import { useCallback, useEffect, useRef, useState } from 'react'
import { Button, Card, Form, Input, InputNumber, Modal, Radio, Space, Table, Popover, TreeSelect, message } from 'antd'
import { PlusOutlined, SearchOutlined, ReloadOutlined, OrderedListOutlined } from '@ant-design/icons'
import { listMenu, getMenu, addMenu, updateMenu, delMenu, updateMenuSort } from '@/api/system/menu'
import Auth from '@/components/Auth'
import IconSelect from '@/components/IconSelect'
import RightToolbar from '@/components/RightToolbar'
import { handleTree } from '@/utils/ruoyi'

interface MenuRow {
  [k: string]: unknown
  menuId: number
  parentId: number
  menuName: string
  orderNum: number
  path: string
  component?: string
  menuType: string
  visible?: string
  status?: string
  perms?: string
  icon?: string
  children?: MenuRow[]
}

export default function MenuPage() {
  const [rows, setRows] = useState<MenuRow[]>([])
  const [loading, setLoading] = useState(false)
  const [keyword, setKeyword] = useState('')
  const [form] = Form.useForm()
  const [open, setOpen] = useState(false)
  const [editId, setEditId] = useState<number | null>(null)
  const [menuType, setMenuType] = useState('M')
  const [treeData, setTreeData] = useState<{ title: string; value: number; children?: unknown[] }[]>([])
  // 搜索表单显隐（对位基准 showSearch ref，配合 RightToolbar 折叠）
  const [showSearch, setShowSearch] = useState(true)

  const load = useCallback(async () => {
    setLoading(true)
    try {
      const res = (await listMenu()) as unknown as { data: MenuRow[] }
      const flat = res.data || []
      const filtered = keyword
        ? flat.filter((m) => m.menuName.includes(keyword))
        : flat
      setRows(handleTree(filtered, 'menuId', 'parentId', 'children'))
      // 记录原始排序（保存时对比找改动行）
      const orders: Record<number, number> = {}
      ;(function record(list: MenuRow[]) {
        list.forEach((i) => { orders[i.menuId] = i.orderNum; if (i.children) record(i.children) })
      })(handleTree(filtered, 'menuId', 'parentId', 'children'))
      originalOrdersRef.current = orders
    } finally {
      setLoading(false)
    }
  }, [keyword])

  useEffect(() => {
    void load()
  }, [load])

  const toTreeData = (items: MenuRow[]): { title: string; value: number; children?: unknown[] }[] =>
    items.map((m) => ({ title: m.menuName, value: m.menuId, children: m.children ? toTreeData(m.children) : undefined }))

  const openAdd = async (parentId?: number) => {
    setEditId(null)
    form.resetFields()
    form.setFieldsValue({ parentId, orderNum: 0, menuType: 'M', visible: '0', status: '0' })
    const res = (await listMenu()) as unknown as { data: MenuRow[] }
    setTreeData(toTreeData(handleTree(res.data || [], 'menuId', 'parentId', 'children')))
    setMenuType('M')
    setOpen(true)
  }

  const openEdit = async (id: number) => {
    const res = (await getMenu(id)) as unknown as { data: Partial<MenuRow> }
    const all = (await listMenu()) as unknown as { data: MenuRow[] }
    setTreeData(toTreeData(handleTree(all.data || [], 'menuId', 'parentId', 'children')))
    setEditId(id)
    form.setFieldsValue(res.data)
    setMenuType(String(res.data.menuType || 'M'))
    setOpen(true)
  }

  const submit = async () => {
    const values = await form.validateFields()
    if (editId) await updateMenu({ ...values, menuId: editId })
    else await addMenu(values)
    setOpen(false)
    void load()
  }

  const handleDeleteOne = (row: MenuRow) => {
    Modal.confirm({
      title: '系统提示', content: `是否确认删除名称为"${row.menuName}"的数据项？`, okText: '确定', cancelText: '取消',
      onOk: async () => { await delMenu(row.menuId); void load() },
    })
  }

  // 行内排序编辑（对位基准：排序列 InputNumber 可编辑 + 原始值记录）
  const [editOrders, setEditOrders] = useState<Record<number, number>>({})
  const originalOrdersRef = useRef<Record<number, number>>({})

  // 保存排序（对位基准 handleSaveSort：只提交有改动的行；Java 端 Map<String,String> 收逗号拼接串）
  const handleSaveSort = () => {
    const menuIds: number[] = []
    const orderNums: number[] = []
    const collectChanged = (list: MenuRow[]) => {
      list.forEach((item) => {
        const edited = editOrders[item.menuId]
        if (edited !== undefined && String(originalOrdersRef.current[item.menuId]) !== String(edited)) {
          menuIds.push(item.menuId)
          orderNums.push(edited)
        }
        if (item.children) collectChanged(item.children)
      })
    }
    collectChanged(rows)
    if (menuIds.length === 0) {
      message.warning('未检测到排序修改')
      return
    }
    Modal.confirm({
      title: '系统提示', okText: '确定', cancelText: '取消',
      onOk: async () => {
        await updateMenuSort({ menuIds: menuIds.join(','), orderNums: orderNums.join(',') })
        message.success('排序保存成功')
        setEditOrders({})
        void load()
      },
    })
  }

  const columns = [
    { title: '菜单名称', dataIndex: 'menuName' },
    { title: '图标', dataIndex: 'icon', width: 70,
      render: (v: string) => (v ? <span className={'icon-' + v} /> : null) },
    { title: '排序', dataIndex: 'orderNum', width: 120,
      render: (_: unknown, row: MenuRow) => (
        <InputNumber
          size="small" min={0} style={{ width: 88 }}
          value={editOrders[row.menuId] ?? row.orderNum}
          onChange={(v) => setEditOrders((prev) => ({ ...prev, [row.menuId]: Number(v) }))}
        />
      ) },
    { title: '权限标识', dataIndex: 'perms' },
    { title: '组件路径', dataIndex: 'component' },
    {
      title: '操作', width: 210,
      render: (_: unknown, row: MenuRow) => (
        <Space>
          <Auth permissions={['system:menu:edit']}>
            <Button type="link" size="small" onClick={() => void openEdit(row.menuId)}>修改</Button>
          </Auth>
          <Auth permissions={['system:menu:add']}>
            <Button type="link" size="small" onClick={() => void openAdd(row.menuId)}>新增下级</Button>
          </Auth>
          <Auth permissions={['system:menu:remove']}>
            <Button type="link" size="small" danger onClick={() => handleDeleteOne(row)}>删除</Button>
          </Auth>
        </Space>
      ),
    },
  ]

  return (
    <div style={{ padding: 16 }}>
      <Card>
        <Form layout="inline" style={{ marginBottom: 16, display: showSearch ? undefined : 'none' }}>
          <Form.Item label="菜单名称">
            <Input allowClear onChange={(e) => setKeyword(e.target.value)} />
          </Form.Item>
          <Form.Item>
            <Space>
              <Button type="primary" icon={<SearchOutlined />} onClick={() => void load()}>搜索</Button>
              <Button icon={<ReloadOutlined />} onClick={() => { setKeyword(''); void load() }}>重置</Button>
            </Space>
          </Form.Item>
        </Form>

        {/* 工具行：操作按钮 + RightToolbar（搜索折叠 + 刷新；列显隐未用不传 columns） */}
        <Space style={{ marginBottom: 16, display: 'flex' }}>
          <Auth permissions={['system:menu:add']}>
            <Button type="primary" icon={<PlusOutlined />} onClick={() => void openAdd()}>新增</Button>
          </Auth>
          <Auth permissions={['system:menu:edit']}>
            <Button type="primary" ghost icon={<OrderedListOutlined />} onClick={handleSaveSort}>保存排序</Button>
          </Auth>
          <RightToolbar showSearch={showSearch} onShowSearchChange={setShowSearch} onRefresh={() => void load()} />
        </Space>

        <Table
          rowKey="menuId"
          columns={columns}
          dataSource={rows}
          loading={loading}
          expandable={{ defaultExpandAllRows: true }}
          pagination={false}
          scroll={{ x: 'max-content' }}
        />
      </Card>

      <Modal title={editId ? '修改菜单' : '添加菜单'} open={open} width={640} onOk={() => void submit()} onCancel={() => setOpen(false)} destroyOnHidden>
        <Form form={form} layout="vertical">
          <Form.Item name="parentId" label="上级菜单">
            <TreeSelect treeData={treeData as never} treeDefaultExpandAll placeholder="请选择上级菜单" allowClear />
          </Form.Item>
          <Form.Item name="menuType" label="菜单类型" initialValue="M">
            <Radio.Group
              onChange={(e) => setMenuType(e.target.value)}
              options={[
                { label: '目录', value: 'M' },
                { label: '菜单', value: 'C' },
                { label: '按钮', value: 'F' },
              ]}
            />
          </Form.Item>
          <Form.Item name="icon" label="菜单图标">
            <Popover trigger="click" content={<IconSelect activeIcon={form.getFieldValue('icon')} onSelect={(n) => form.setFieldValue('icon', n)} />}>
              <Input readOnly placeholder="点击选择图标" />
            </Popover>
          </Form.Item>
          <Form.Item name="menuName" label="菜单名称" rules={[{ required: true, message: '菜单名称不能为空' }]}>
            <Input />
          </Form.Item>
          {menuType !== 'F' && (
            <Form.Item name="orderNum" label="显示排序" initialValue={0}>
              <InputNumber min={0} style={{ width: '100%' }} />
            </Form.Item>
          )}
          {menuType !== 'F' && (
            <Form.Item name="path" label="路由地址" rules={[{ required: true, message: '路由地址不能为空' }]}>
              <Input />
            </Form.Item>
          )}
          {menuType === 'C' && (
            <Form.Item name="component" label="组件路径">
              <Input placeholder="如 system/user/index" />
            </Form.Item>
          )}
          {menuType !== 'F' && (
            <Form.Item name="visible" label="显示状态" initialValue="0">
              <Radio.Group options={[{ label: '显示', value: '0' }, { label: '隐藏', value: '1' }]} />
            </Form.Item>
          )}
          <Form.Item name="perms" label="权限标识">
            <Input placeholder="如 system:user:list" />
          </Form.Item>
          <Form.Item name="status" label="菜单状态" initialValue="0">
            <Radio.Group options={[{ label: '正常', value: '0' }, { label: '停用', value: '1' }]} />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}
