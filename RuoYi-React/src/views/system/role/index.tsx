// 角色管理 —— 对位基准 views/system/role/index.vue
// 树表 CRUD + 菜单权限树（半选提交）+ 数据权限 + 状态 + 导出
// 3.0.0 批次 C：内联分页/工具行替换为 Pagination / RightToolbar 组件消费（行为等价）

import { useCallback, useEffect, useState } from 'react'
import { Button, Card, Form, Input, InputNumber, Modal, Radio, Select, Space, Table, Tree } from 'antd'
import { PlusOutlined, SearchOutlined, ReloadOutlined } from '@ant-design/icons'
import { listRole, getRole, addRole, updateRole, delRole, changeRoleStatus, deptTreeSelect, dataScope } from '@/api/system/role'
import { roleMenuTreeselect } from '@/api/system/menu'
import { useCrud } from '@/hooks/useCrud'
import { useDict } from '@/utils/dict'
import DictTag from '@/components/DictTag'
import Auth from '@/components/Auth'
import Pagination from '@/components/Pagination'
import RightToolbar from '@/components/RightToolbar'

interface RoleRow {
  [k: string]: unknown
  roleId: number
  roleName: string
  roleKey: string
  roleSort: number
  status: string
  remark?: string
  menuCheckStrictly?: boolean
  deptCheckStrictly?: boolean
}

export default function Role() {
  const dicts = useDict('sys_normal_disable')
  const crud = useCrud<RoleRow, { pageNum: number; pageSize: number; roleName?: string; roleKey?: string; status?: string }>({
    listApi: listRole,
    delApi: delRole,
    exportUrl: '/system/role/export',
    defaultQuery: { pageNum: 1, pageSize: 10 },
  })
  const [form] = Form.useForm()
  const [open, setOpen] = useState(false)
  const [editId, setEditId] = useState<number | null>(null)
  // 菜单权限
  const [menuOpen, setMenuOpen] = useState(false)
  const [menuTree, setMenuTree] = useState<{ title: string; key: number; children?: unknown[] }[]>([])
  const [checkedKeys, setCheckedKeys] = useState<number[]>([])
  const [halfChecked, setHalfChecked] = useState<number[]>([])
  const [currentRole, setCurrentRole] = useState<RoleRow | null>(null)
  const [, setMenuCheckStrictly] = useState(true)
  // 数据权限
  const [scopeOpen, setScopeOpen] = useState(false)
  const [scopeTree, setScopeTree] = useState<{ title: string; key: number; children?: unknown[] }[]>([])
  const [scopeChecked, setScopeChecked] = useState<number[]>([])
  const [scopeHalf, setScopeHalf] = useState<number[]>([])
  const [scopeForm] = Form.useForm()
  // 搜索表单显隐（对位基准 showSearch ref，配合 RightToolbar 折叠）
  const [showSearch, setShowSearch] = useState(true)

  const load = useCallback(() => { void crud.getList() }, [crud.getList]) // eslint-disable-line

  useEffect(() => { load() }, [load])

  const toAntTree = (items: { id: number; label: string; children?: unknown[] }[]): { title: string; key: number; children?: unknown[] }[] =>
    items.map((i) => ({ title: i.label, key: i.id, children: i.children ? toAntTree(i.children as never) : undefined })) as never

  // 菜单权限弹窗（回显 checked = menuIds - halfChecked；父子不关联时全量回显）
  const openMenuDialog = async (row: RoleRow) => {
    setCurrentRole(row)
    setMenuCheckStrictly(row.menuCheckStrictly !== false)
    const res = (await roleMenuTreeselect(row.roleId)) as unknown as {
      menus: { id: number; label: string; children?: unknown[] }[]
      checkedKeys: number[]
    }
    setMenuTree(toAntTree(res.menus || []) as never)
    // 后端 checkedKeys 含半选父节点；menuCheckStrictly（父子联动）时父节点会被 antd 自动推导
    setCheckedKeys(res.checkedKeys || [])
    setMenuOpen(true)
  }

  // 提交：checked + halfChecked 合并（对位基准 getCheckedKeys+getHalfCheckedKeys）
  const submitMenu = async () => {
    if (!currentRole) return
    await updateRole({
      roleId: currentRole.roleId,
      roleName: currentRole.roleName,
      roleKey: currentRole.roleKey,
      roleSort: currentRole.roleSort,
      status: currentRole.status,
      // 字段名对位 Java SysRole.menuIds（基准 v-model form.menuIds）；半选一并提交（getMenuAllCheckedKeys 等价）
      menuIds: [...checkedKeys, ...halfChecked].filter((v, i, a) => a.indexOf(v) === i),
    })
    setMenuOpen(false)
    void crud.getList()
  }

  // 数据权限弹窗
  const openScopeDialog = async (row: RoleRow) => {
    setCurrentRole(row)
    const res = (await deptTreeSelect(row.roleId)) as unknown as {
      depts: { id: number; label: string; children?: unknown[] }[]
      checkedKeys: number[]
    }
    setScopeTree(toAntTree(res.depts || []) as never)
    setScopeChecked(res.checkedKeys || [])
    scopeForm.setFieldsValue({ dataScope: String(row.dataScope ?? '1') })
    setScopeOpen(true)
  }

  const submitScope = async () => {
    if (!currentRole) return
    const values = await scopeForm.validateFields()
    await dataScope({
      roleId: currentRole.roleId,
      dataScope: values.dataScope,
      // 字段名对位 Java SysRole.deptIds（基准 form.deptIds）；自定义权限时提交全选+半选
      deptIds: values.dataScope === '2' ? [...scopeChecked, ...scopeHalf].filter((v, i, a) => a.indexOf(v) === i) : [],
    })
    setScopeOpen(false)
  }

  const openAdd = () => {
    setEditId(null)
    form.resetFields()
    form.setFieldsValue({ roleSort: 0, status: '0' })
    setMenuOpen(true)
    setOpen(true)
    void roleMenuTreeselect(0).then((res) => {
      const r = res as unknown as { menus: { id: number; label: string; children?: unknown[] }[] }
      setMenuTree(toAntTree(r.menus || []) as never)
      setCheckedKeys([])
    })
  }

  const openEdit = async (id: number) => {
    const res = (await getRole(id)) as unknown as { data: Partial<RoleRow> }
    setEditId(id)
    form.setFieldsValue(res.data)
    setOpen(true)
  }

  const submit = async () => {
    const values = await form.validateFields()
    if (editId) await updateRole({ ...values, roleId: editId })
    else await addRole(values)
    setOpen(false)
    void crud.getList()
  }

  const toggleStatus = (row: RoleRow) => {
    const next = row.status === '0' ? '1' : '0'
    Modal.confirm({
      title: '系统提示', content: `是否确认改变"${row.roleName}"的状态?`, okText: '确定', cancelText: '取消',
      onOk: async () => {
        await changeRoleStatus(row.roleId, next)
        void crud.getList()
      },
    })
  }

  const columns = [
    { title: '角色编号', dataIndex: 'roleId', width: 90 },
    { title: '角色名称', dataIndex: 'roleName' },
    { title: '权限字符', dataIndex: 'roleKey' },
    { title: '显示顺序', dataIndex: 'roleSort', width: 90 },
    { title: '状态', dataIndex: 'status', width: 80,
      render: (v: string, row: RoleRow) => (
        <Button type="link" size="small" onClick={() => toggleStatus(row)}>
          <DictTag options={dicts.sys_normal_disable || []} value={v} />
        </Button>
      ) },
    {
      title: '操作', width: 300,
      render: (_: unknown, row: RoleRow) => (
        <Space>
          <Auth permissions={['system:role:edit']}>
            <Button type="link" size="small" onClick={() => void openEdit(row.roleId)}>修改</Button>
          </Auth>
          <Auth permissions={['system:role:edit']}>
            <Button type="link" size="small" onClick={() => void openMenuDialog(row)}>菜单权限</Button>
          </Auth>
          <Auth permissions={['system:role:edit']}>
            <Button type="link" size="small" onClick={() => void openScopeDialog(row)}>数据权限</Button>
          </Auth>
          <Auth permissions={['system:role:edit']}>
            <Button type="link" size="small" onClick={() => window.open(`/system/role-auth/user/${row.roleId}`, '_self')}>分配用户</Button>
          </Auth>
          <Auth permissions={['system:role:remove']}>
            <Button type="link" size="small" danger onClick={() => crud.handleDelete(String(row.roleId), [row.roleId])}>删除</Button>
          </Auth>
        </Space>
      ),
    },
  ]

  return (
    <div style={{ padding: 16 }}>
      <Card>
        <Form layout="inline" onFinish={() => crud.handleQuery()} style={{ marginBottom: 16, display: showSearch ? undefined : 'none' }}>
          <Form.Item label="角色名称">
            <Input allowClear onChange={(e) => crud.setQuery((q) => ({ ...q, roleName: e.target.value }))} />
          </Form.Item>
          <Form.Item label="权限字符">
            <Input allowClear onChange={(e) => crud.setQuery((q) => ({ ...q, roleKey: e.target.value }))} />
          </Form.Item>
          <Form.Item label="状态">
            <Select allowClear style={{ width: 120 }}
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

        {/* 工具行：操作按钮 + RightToolbar（搜索折叠 + 刷新；列显隐未用不传 columns） */}
        <Space style={{ marginBottom: 16, display: 'flex' }}>
          <Auth permissions={['system:role:add']}>
            <Button type="primary" icon={<PlusOutlined />} onClick={openAdd}>新增</Button>
          </Auth>
          <Auth permissions={['system:role:export']}>
            <Button style={{ color: '#e6a23c', borderColor: '#e6a23c' }} onClick={() => crud.handleExport()}>导出</Button>
          </Auth>
          <RightToolbar showSearch={showSearch} onShowSearchChange={setShowSearch} onRefresh={() => void crud.getList()} />
        </Space>

        <Table
          rowKey="roleId" columns={columns} dataSource={crud.rows} loading={crud.loading}
          rowSelection={{ onChange: (keys) => crud.handleSelectionChange([...keys] as (string | number)[], []) }}
          pagination={false}
          scroll={{ x: 'max-content' }}
        />
        {crud.total > 0 && (
          <Pagination
            total={crud.total}
            page={crud.query.pageNum}
            pageSize={crud.query.pageSize}
            onChange={(p, size) => void crud.getList({ pageNum: p, pageSize: size })}
          />
        )}
      </Card>

      {/* 新增/修改角色 */}
      <Modal title={editId ? '修改角色' : '添加角色'} open={open} width={640} onOk={() => void submit()} onCancel={() => setOpen(false)} destroyOnHidden>
        <Form form={form} layout="vertical">
          <Form.Item name="roleName" label="角色名称" rules={[{ required: true, message: '角色名称不能为空' }]}>
            <Input />
          </Form.Item>
          <Form.Item name="roleKey" label="权限字符" rules={[{ required: true, message: '权限字符不能为空' }]}>
            <Input />
          </Form.Item>
          <Form.Item name="roleSort" label="显示顺序" initialValue={0}>
            <InputNumber min={0} style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item name="status" label="状态" initialValue="0">
            <Radio.Group options={[{ label: '正常', value: '0' }, { label: '停用', value: '1' }]} />
          </Form.Item>
          <Form.Item name="remark" label="备注"><Input.TextArea /></Form.Item>
        </Form>
      </Modal>

      {/* 菜单权限弹窗（Tree checkable + 半选收集） */}
      <Modal title={`菜单权限 — ${currentRole?.roleName || ''}`} open={menuOpen} width={520}
        onOk={() => void submitMenu()} onCancel={() => setMenuOpen(false)} destroyOnHidden>
        <Tree
          checkable
          defaultExpandAll
          treeData={menuTree as never}
          checkedKeys={checkedKeys as never}
          onCheck={(checked, info) => {
            setCheckedKeys([...(checked as number[])] )
            setHalfChecked([...(info.halfCheckedKeys as number[])] )
          }}
        />
      </Modal>

      {/* 数据权限弹窗 */}
      <Modal title={`数据权限 — ${currentRole?.roleName || ''}`} open={scopeOpen} width={520}
        onOk={() => void submitScope()} onCancel={() => setScopeOpen(false)} destroyOnHidden>
        <Form form={scopeForm} layout="vertical">
          <Form.Item name="dataScope" label="权限范围">
            <Select options={[
              { label: '全部数据权限', value: '1' },
              { label: '自定数据权限', value: '2' },
              { label: '本部门数据权限', value: '3' },
              { label: '本部门及以下数据权限', value: '4' },
              { label: '仅本人数据权限', value: '5' },
            ]} />
          </Form.Item>
          {scopeForm.getFieldValue('dataScope') === '2' && (
            <Form.Item label="数据权限">
              <Tree checkable defaultExpandAll treeData={scopeTree as never}
                checkedKeys={scopeChecked as never}
                onCheck={(checked, info) => {
                  setScopeChecked([...(checked as number[])])
                  setScopeHalf([...(info.halfCheckedKeys as number[])])
                }} />
            </Form.Item>
          )}
        </Form>
      </Modal>
    </div>
  )
}
