// 用户管理 —— 对位基准 views/system/user/index.vue
// 左部门树（TreePanel 简版）+ 用户 CRUD + 重置密码 + 导入导出 + 详情 + 分配角色子页

import { useCallback, useEffect, useState } from 'react'
import { Button, Card, Col, Form, Input, Modal, Row, Select, Space, Table, Tree } from 'antd'
import { PlusOutlined, SearchOutlined, ReloadOutlined, ExportOutlined, UploadOutlined } from '@ant-design/icons'
import {
  listUser, getUser, addUser, updateUser, delUser, resetUserPwd,
  changeUserStatus, deptTreeSelect,
} from '@/api/system/user'
import { useCrud } from '@/hooks/useCrud'
import { useDict } from '@/utils/dict'
import DictTag from '@/components/DictTag'
import Auth from '@/components/Auth'
import { parseTime } from '@/utils/ruoyi'
import { pwdValidatorFactory } from '@/utils/passwordRule'

interface UserRow {
  [k: string]: unknown
  userId: number
  userName: string
  nickName: string
  dept?: { deptName: string }
  phonenumber?: string
  status: string
  createTime?: string
}

export default function User() {
  const dicts = useDict('sys_normal_disable', 'sys_user_sex')
  const [deptId, setDeptId] = useState<number | undefined>()
  const crud = useCrud<UserRow, { pageNum: number; pageSize: number; userName?: string; phonenumber?: string; status?: string; deptId?: number }>({
    listApi: listUser,
    delApi: delUser,
    exportUrl: '/system/user/export',
    defaultQuery: { pageNum: 1, pageSize: 10 },
  })
  const [deptTree, setDeptTree] = useState<{ title: string; key: number; children?: unknown[] }[]>([])
  const [form] = Form.useForm()
  const [open, setOpen] = useState(false)
  const [editId, setEditId] = useState<number | null>(null)
  const [postOptions, setPostOptions] = useState<{ label: string; value: number }[]>([])
  const [roleOptions, setRoleOptions] = useState<{ label: string; value: number; disabled?: boolean }[]>([])

  const load = useCallback(() => { void crud.getList({ deptId } as never) }, [crud.getList, deptId])

  useEffect(() => {
    void deptTreeSelect().then((res) => {
      const r = res as unknown as { data: { id: number; label: string; children?: unknown[] }[] }
      setDeptTree(toAntTree(r.data || []) as never)
    })
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  useEffect(() => { load() }, [load])

  const toAntTree = (items: { id: number; label: string; children?: unknown[] }[]): unknown[] =>
    items.map((i) => ({ title: i.label, key: i.id, children: i.children ? toAntTree(i.children as never) : undefined })) as never

  const openAdd = async () => {
    setEditId(null)
    form.resetFields()
    form.setFieldsValue({ status: '0', sex: '0', postIds: [], roleIds: [] })
    // 空参 getUser 返回角色/岗位选项（尾斜杠契约）
    const res = (await getUser()) as unknown as { posts: { postId: number; postName: string }[]; roles: { roleId: number; roleName: string; disabled?: boolean }[] }
    setPostOptions((res.posts || []).map((p) => ({ label: p.postName, value: p.postId })))
    setRoleOptions((res.roles || []).map((r) => ({ label: r.roleName, value: r.roleId, disabled: r.disabled === true })))
    setOpen(true)
  }

  const openEdit = async (id: number) => {
    const res = (await getUser(id)) as unknown as {
      data: Partial<UserRow> & { postIds?: number[]; roleIds?: number[]; roles?: { roleId: number; roleName: string; disabled?: boolean }[] }
      posts: { postId: number; postName: string }[]
    }
    setEditId(id)
    form.setFieldsValue({ ...res.data, postIds: res.data.postIds, roleIds: res.data.roleIds })
    setPostOptions((res.posts || []).map((p) => ({ label: p.postName, value: p.postId })))
    setRoleOptions((res.data.roles || []).map((r) => ({ label: r.roleName, value: r.roleId, disabled: r.disabled === true })))
    setOpen(true)
  }

  const submit = async () => {
    const values = await form.validateFields()
    if (editId) await updateUser({ ...values, userId: editId })
    else await addUser(values)
    setOpen(false)
    void load()
  }

  const toggleStatus = (row: UserRow) => {
    const next = row.status === '0' ? '1' : '0'
    Modal.confirm({
      title: '系统提示', content: `是否确认改变"${row.userName}"的状态?`, okText: '确定', cancelText: '取消',
      onOk: async () => { await changeUserStatus(row.userId, next); void load() },
    })
  }

  // 重置密码（prompt 等价：Modal + Input + 密码校验）
  const handleResetPwd = (row: UserRow) => {
    let newPwd = ''
    Modal.confirm({
      title: '系统提示',
      content: (
        <div>
          <p>请输入"{row.userName}"的新密码（6-20位）：</p>
          <Input.Password onChange={(e) => { newPwd = e.target.value }} />
        </div>
      ),
      okText: '确定', cancelText: '取消',
      onOk: async () => {
        if (!newPwd || newPwd.length < 6 || newPwd.length > 20) {
          throw new Error('密码长度必须介于 6 和 20 之间')
        }
        await resetUserPwd(row.userId, newPwd)
      },
    })
  }

  const handleImport = () => {
    // ExcelImportDialog 简化内联：选择文件 + updateSupport + 提交
    const input = document.createElement('input')
    input.type = 'file'
    input.accept = '.xlsx,.xls'
    input.onchange = () => {
      const file = input.files?.[0]
      if (!file) return
      Modal.confirm({
        title: '系统提示',
        content: `是否确认导入文件 ${file.name}？已存在用户不更新。`,
        okText: '确定', cancelText: '取消',
        onOk: async () => {
          const form_ = new FormData()
          form_.append('file', file)
          form_.append('updateSupport', '0')
          const { default: request } = await import('@/utils/request')
          await request.post('/system/user/importData?updateSupport=0', form_, {
            headers: { 'Content-Type': 'multipart/form-data' },
          })
          void load()
        },
      })
    }
    input.click()
  }

  const columns = [
    { title: '用户编号', dataIndex: 'userId', width: 90 },
    { title: '用户名称', dataIndex: 'nickName' },
    { title: '用户账号', dataIndex: 'userName' },
    { title: '部门', dataIndex: ['dept', 'deptName'] },
    { title: '手机号码', dataIndex: 'phonenumber' },
    {
      title: '状态', dataIndex: 'status', width: 80,
      render: (v: string, row: UserRow) => (
        <Button type="link" size="small" onClick={() => toggleStatus(row)}>
          <DictTag options={dicts.sys_normal_disable || []} value={v} />
        </Button>
      ),
    },
    { title: '创建时间', dataIndex: 'createTime', width: 160, render: (v: string) => parseTime(v) },
    {
      title: '操作', width: 260,
      render: (_: unknown, row: UserRow) => (
        <Space>
          <Auth permissions={['system:user:edit']}>
            <Button type="link" size="small" onClick={() => void openEdit(row.userId)}>修改</Button>
          </Auth>
          <Auth permissions={['system:user:edit']}>
            <Button type="link" size="small" onClick={() => handleResetPwd(row)}>重置密码</Button>
          </Auth>
          <Auth permissions={['system:user:edit']}>
            <Button type="link" size="small" onClick={() => window.open(`/system/user-auth/role/${row.userId}`, '_self')}>分配角色</Button>
          </Auth>
          <Auth permissions={['system:user:remove']}>
            <Button type="link" size="small" danger onClick={() => crud.handleDelete(String(row.userId), [row.userId])}>删除</Button>
          </Auth>
        </Space>
      ),
    },
  ]

  return (
    <div style={{ padding: 16 }}>
      <Row gutter={16}>
        {/* 部门树 */}
        <Col xs={24} md={5}>
          <Card size="small" title="部门列表">
            <Tree
              treeData={deptTree as never}
              defaultExpandAll
              onSelect={(keys) => {
                setDeptId(keys[0] as number | undefined)
                void crud.getList({ deptId: keys[0] as number, pageNum: 1 } as never)
              }}
            />
          </Card>
        </Col>
        <Col xs={24} md={19}>
          <Card>
            <Form layout="inline" onFinish={() => crud.handleQuery()} style={{ marginBottom: 16 }}>
              <Form.Item label="用户名称">
                <Input allowClear onChange={(e) => crud.setQuery((q) => ({ ...q, userName: e.target.value }))} />
              </Form.Item>
              <Form.Item label="手机号码">
                <Input allowClear onChange={(e) => crud.setQuery((q) => ({ ...q, phonenumber: e.target.value }))} />
              </Form.Item>
              <Form.Item>
                <Space>
                  <Button type="primary" htmlType="submit" icon={<SearchOutlined />}>搜索</Button>
                  <Button icon={<ReloadOutlined />} onClick={crud.resetQuery}>重置</Button>
                </Space>
              </Form.Item>
            </Form>

            <Space style={{ marginBottom: 16 }}>
              <Auth permissions={['system:user:add']}>
                <Button type="primary" icon={<PlusOutlined />} onClick={() => void openAdd()}>新增</Button>
              </Auth>
              <Auth permissions={['system:user:edit']}>
                <Button icon={<UploadOutlined />} onClick={handleImport}>导入</Button>
              </Auth>
              <Auth permissions={['system:user:export']}>
                <Button icon={<ExportOutlined />} onClick={() => crud.handleExport()}>导出</Button>
              </Auth>
            </Space>

            <Table
              rowKey="userId" columns={columns} dataSource={crud.rows} loading={crud.loading}
              rowSelection={{ onChange: (keys) => crud.handleSelectionChange([...keys] as (string | number)[], []) }}
              pagination={{ total: crud.total, pageSize: crud.query.pageSize, current: crud.query.pageNum,
                showSizeChanger: true, showTotal: (t) => `共 ${t} 条`,
                onChange: (p, size) => void crud.getList({ pageNum: p, pageSize: size } as never) }}
              scroll={{ x: 'max-content' }}
            />
          </Card>
        </Col>
      </Row>

      {/* 用户新增/修改弹窗 */}
      <Modal title={editId ? '修改用户' : '添加用户'} open={open} width={640} onOk={() => void submit()} onCancel={() => setOpen(false)} destroyOnHidden>
        <Form form={form} layout="vertical">
          <Form.Item name="nickName" label="用户昵称" rules={[{ required: true, message: '用户昵称不能为空' }]}>
            <Input />
          </Form.Item>
          <Form.Item name="deptId" label="归属部门">
            <Tree treeData={deptTree as never} defaultExpandAll
              onSelect={(keys) => form.setFieldValue('deptId', keys[0])} />
          </Form.Item>
          <Form.Item name="phonenumber" label="手机号码">
            <Input />
          </Form.Item>
          <Form.Item name="email" label="邮箱">
            <Input />
          </Form.Item>
          <Form.Item name="userName" label="用户账号" rules={[{ required: true, message: '用户账号不能为空' }]}>
            <Input disabled={!!editId} />
          </Form.Item>
          <Form.Item name="password" label="登录密码" rules={editId ? [] : [{ validator: pwdValidatorFactory('login') }]}>
            <Input.Password disabled={!!editId} />
          </Form.Item>
          <Form.Item name="sex" label="性别" initialValue="0">
            <Select options={(dicts.sys_user_sex || []).map((d) => ({ label: d.label, value: d.value }))} />
          </Form.Item>
          <Form.Item name="status" label="状态" initialValue="0">
            <RadioGroup />
          </Form.Item>
          <Form.Item name="postIds" label="岗位">
            <Select mode="multiple" options={postOptions} />
          </Form.Item>
          <Form.Item name="roleIds" label="角色">
            <Select mode="multiple" options={roleOptions} />
          </Form.Item>
          <Form.Item name="remark" label="备注"><Input.TextArea /></Form.Item>
        </Form>
      </Modal>

    </div>
  )
}

function RadioGroup() {
  return (
    <Select
      options={[{ label: '正常', value: '0' }, { label: '停用', value: '1' }]}
    />
  )
}
