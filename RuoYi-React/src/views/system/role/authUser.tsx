// 分配用户（角色维度）—— 对位基准 views/system/role/authUser.vue（路由 /system/role-auth/user/:roleId）
// 已授权列表 + 取消授权 + 新增授权（未授权列表多选 selectAll）

import { useEffect, useState } from 'react'
import { useNavigate, useParams, useSearchParams } from 'react-router'
import { Button, Card, Input, Modal, Space, Table, message } from 'antd'
import {
  allocatedUserList, unallocatedUserList, authUserCancel, authUserCancelAll, authUserSelectAll,
} from '@/api/system/role'
import Auth from '@/components/Auth'

interface UserItem {
  [k: string]: unknown
  userId: number
  userName: string
  nickName: string
  phonenumber?: string
}

export default function AuthUser() {
  const { roleId } = useParams()
  const navigate = useNavigate()
  const [search] = useSearchParams()
  const [rows, setRows] = useState<UserItem[]>([])
  const [total, setTotal] = useState(0)
  const [checked, setChecked] = useState<number[]>([])
  const [loading, setLoading] = useState(false)
  const [query, setQuery] = useState({ pageNum: 1, pageSize: 10, userName: search.get('userName') || undefined })
  const [addOpen, setAddOpen] = useState(false)
  const [addRows, setAddRows] = useState<UserItem[]>([])
  const [addTotal, setAddTotal] = useState(0)
  const [addChecked, setAddChecked] = useState<number[]>([])
  const [addQuery, setAddQuery] = useState({ pageNum: 1, pageSize: 10, userName: '' })

  const load = async (q = query) => {
    setLoading(true)
    try {
      const res = (await allocatedUserList({ ...q, roleId })) as unknown as { rows: UserItem[]; total: number }
      setRows(res.rows || [])
      setTotal(res.total || 0)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    void load()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [query])

  const openAdd = async () => {
    const res = (await unallocatedUserList({ ...addQuery, roleId })) as unknown as { rows: UserItem[]; total: number }
    setAddRows(res.rows || [])
    setAddTotal(res.total || 0)
    setAddChecked([])
    setAddOpen(true)
  }

  const submitAdd = async () => {
    await authUserSelectAll({ roleId, userIds: addChecked.join(',') })
    message.success('授权成功')
    setAddOpen(false)
    void load()
  }

  const cancelOne = async (row: UserItem) => {
    await authUserCancel({ userId: row.userId, roleId })
    message.success('取消授权成功')
    void load()
  }

  const cancelAll = () => {
    Modal.confirm({
      title: '系统提示', content: `是否取消选中用户授权数据项？`, okText: '确定', cancelText: '取消',
      onOk: async () => {
        await authUserCancelAll({ roleId, userIds: checked.join(',') })
        message.success('取消授权成功')
        setChecked([])
        void load()
      },
    })
  }

  const columns = [
    { title: '用户名称', dataIndex: 'nickName' },
    { title: '用户账号', dataIndex: 'userName' },
    { title: '手机号码', dataIndex: 'phonenumber' },
    {
      title: '操作', width: 110,
      render: (_: unknown, row: UserItem) => (
        <Button type="link" size="small" danger onClick={() => void cancelOne(row)}>取消授权</Button>
      ),
    },
  ]

  return (
    <div style={{ padding: 16 }}>
      <Card title={`分配用户（角色：#${roleId}）`}>
        <Space style={{ marginBottom: 16 }}>
          <Auth permissions={['system:role:edit']}>
            <Button type="primary" onClick={() => void openAdd()}>新增用户</Button>
          </Auth>
          <Auth permissions={['system:role:edit']}>
            <Button danger disabled={checked.length === 0} onClick={cancelAll}>批量取消授权</Button>
          </Auth>
          <Button onClick={() => void navigate('/system/role')}>返回</Button>
        </Space>
        <Table
          rowKey="userId" columns={columns} dataSource={rows} loading={loading}
          rowSelection={{ selectedRowKeys: checked, onChange: (k) => setChecked([...k] as number[]) }}
          pagination={{ total, pageSize: query.pageSize, current: query.pageNum,
            onChange: (p, size) => setQuery({ ...query, pageNum: p, pageSize: size }) }}
        />
      </Card>

      <Modal title="新增授权用户" open={addOpen} width={640} onOk={() => void submitAdd()} onCancel={() => setAddOpen(false)}>
        <Input.Search placeholder="请输入用户账号" style={{ marginBottom: 12 }}
          onSearch={(v) => { setAddQuery({ ...addQuery, userName: v, pageNum: 1 }); void loadAdd(v) }} />
        <Table
          rowKey="userId"
          dataSource={addRows}
          pagination={{ total: addTotal, pageSize: 10, current: addQuery.pageNum,
            onChange: (p) => setAddQuery({ ...addQuery, pageNum: p }) }}
          rowSelection={{ selectedRowKeys: addChecked, onChange: (k) => setAddChecked([...k] as number[]) }}
        >
        </Table>
      </Modal>
    </div>
  )

  async function loadAdd(kw = addQuery.userName) {
    const res = (await unallocatedUserList({ ...addQuery, userName: kw || undefined, roleId })) as unknown as { rows: UserItem[]; total: number }
    setAddRows(res.rows || [])
    setAddTotal(res.total || 0)
  }
}
