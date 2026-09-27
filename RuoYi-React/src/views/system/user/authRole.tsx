// 分配角色 —— 对位基准 views/system/user/authRole.vue（路由 /system/user-auth/role/:userId）

import { useEffect, useState } from 'react'
import { useNavigate, useParams } from 'react-router'
import { Button, Card, Space, Table, message } from 'antd'
import { getAuthRole, updateAuthRole } from '@/api/system/user'
import Auth from '@/components/Auth'

interface RoleItem {
  [k: string]: unknown
  roleId: number
  roleName: string
  roleKey: string
  status: string
  flag?: boolean
}

export default function AuthRole() {
  const { userId } = useParams()
  const navigate = useNavigate()
  const [rows, setRows] = useState<RoleItem[]>([])
  const [checked, setChecked] = useState<number[]>([])
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    if (!userId) return
    setLoading(true)
    void getAuthRole(userId)
      .then((res) => {
        const r = res as unknown as { roles: RoleItem[] }
        setRows(r.roles || [])
        setChecked((r.roles || []).filter((x) => x.flag).map((x) => x.roleId))
      })
      .finally(() => setLoading(false))
  }, [userId])

  const submit = async () => {
    await updateAuthRole({ userId, roleIds: checked.join(',') })
    message.success('授权成功')
    void navigate('/system/user')
  }

  const columns = [
    { title: '角色编号', dataIndex: 'roleId', width: 90 },
    { title: '角色名称', dataIndex: 'roleName' },
    { title: '权限字符', dataIndex: 'roleKey' },
    { title: '状态', dataIndex: 'status', width: 80, render: (v: string) => (v === '0' ? '正常' : '停用') },
  ]

  return (
    <div style={{ padding: 16 }}>
      <Card title={`分配角色（用户：#${userId}）`}>
        <Space style={{ marginBottom: 16 }}>
          <Auth permissions={['system:user:edit']}>
            <Button type="primary" onClick={() => void submit()}>提交</Button>
          </Auth>
          <Button onClick={() => void navigate('/system/user')}>返回</Button>
        </Space>
        <Table
          rowKey="roleId"
          columns={columns}
          dataSource={rows}
          loading={loading}
          pagination={false}
          rowSelection={{
            selectedRowKeys: checked,
            onChange: (keys) => setChecked([...keys] as number[]),
          }}
        />
      </Card>
    </div>
  )
}
