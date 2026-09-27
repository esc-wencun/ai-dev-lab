// 参数设置 —— 完整 CRUD（对位基准 views/system/config/index.vue）
// 查询/新增/修改/删除/批量删除/刷新缓存 + Auth 按钮权限 + useCrud 范式

import { useEffect, useState } from 'react'
import { Button, Card, Form, Input, Modal, Space, Table } from 'antd'
import { PlusOutlined, DeleteOutlined, ReloadOutlined, SearchOutlined, RedoOutlined } from '@ant-design/icons'
import { listConfig, getConfig, addConfig, updateConfig, delConfig, refreshCache } from '@/api/system/config'
import { useCrud } from '@/hooks/useCrud'
import Auth from '@/components/Auth'

interface ConfigRow {
  [k: string]: unknown
  configId: number
  configName: string
  configKey: string
  configValue: string
  configType: string
  remark?: string
}

interface ConfigQuery {
  pageNum: number
  pageSize: number
  configName?: string
  configKey?: string
  configType?: string
}

export default function Config() {
  const crud = useCrud<ConfigRow, ConfigQuery>({
    listApi: listConfig,
    delApi: delConfig,
    exportUrl: '/system/config/export',
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
    const res = (await getConfig(id)) as unknown as { data: Partial<ConfigRow> }
    setEditId(id)
    form.setFieldsValue(res.data)
    setOpen(true)
  }

  const submit = async () => {
    const values = await form.validateFields()
    if (editId) {
      await updateConfig({ ...values, configId: editId })
    } else {
      await addConfig(values)
    }
    setOpen(false)
    void crud.getList()
  }

  const handleRefreshCache = () => {
    Modal.confirm({
      title: '系统提示',
      content: '是否确认刷新参数缓存?',
      okText: '确定',
      cancelText: '取消',
      onOk: async () => {
        await refreshCache()
      },
    })
  }

  const columns = [
    { title: '参数主键', dataIndex: 'configId', width: 90 },
    { title: '参数名称', dataIndex: 'configName' },
    { title: '参数键名', dataIndex: 'configKey' },
    { title: '参数键值', dataIndex: 'configValue' },
    {
      title: '操作', width: 160,
      render: (_: unknown, row: ConfigRow) => (
        <Space>
          <Auth permissions={['system:config:edit']}>
            <Button type="link" size="small" onClick={() => void openEdit(row.configId)}>修改</Button>
          </Auth>
          <Auth permissions={['system:config:remove']}>
            <Button type="link" size="small" danger
              onClick={() => crud.handleDelete(String(row.configId), [row.configId])}>删除</Button>
          </Auth>
        </Space>
      ),
    },
  ]

  return (
    <div style={{ padding: 16 }}>
      <Card>
        <Form layout="inline" onFinish={() => crud.handleQuery()} style={{ marginBottom: 16 }}>
          <Form.Item name="configName" label="参数名称">
            <Input placeholder="请输入参数名称" allowClear onChange={(e) => crud.setQuery((q) => ({ ...q, configName: e.target.value }))} />
          </Form.Item>
          <Form.Item name="configKey" label="参数键名">
            <Input placeholder="请输入参数键名" allowClear onChange={(e) => crud.setQuery((q) => ({ ...q, configKey: e.target.value }))} />
          </Form.Item>
          <Form.Item>
            <Space>
              <Button type="primary" htmlType="submit" icon={<SearchOutlined />}>搜索</Button>
              <Button icon={<ReloadOutlined />} onClick={crud.resetQuery}>重置</Button>
            </Space>
          </Form.Item>
        </Form>

        <Space style={{ marginBottom: 16 }}>
          <Auth permissions={['system:config:add']}>
            <Button type="primary" icon={<PlusOutlined />} onClick={openAdd}>新增</Button>
          </Auth>
          <Auth permissions={['system:config:remove']}>
            <Button type="primary" danger icon={<DeleteOutlined />} disabled={!crud.multiple}
              onClick={() => crud.handleDelete(crud.ids.join(','))}>删除</Button>
          </Auth>
          <Auth permissions={['system:config:remove']}>
            <Button icon={<RedoOutlined />} onClick={handleRefreshCache}>刷新缓存</Button>
          </Auth>
        </Space>

        <Table
          rowKey="configId"
          columns={columns}
          dataSource={crud.rows}
          loading={crud.loading}
          pagination={{
            total: crud.total,
            pageSize: crud.query.pageSize,
            current: crud.query.pageNum,
            showSizeChanger: true,
            showTotal: (t) => `共 ${t} 条`,
            onChange: (p, size) => void crud.getList({ pageNum: p, pageSize: size } as ConfigQuery),
          }}
          scroll={{ x: 'max-content' }}
        />
      </Card>

      <Modal
        title={editId ? '修改参数' : '添加参数'}
        open={open}
        onOk={() => void submit()}
        onCancel={() => setOpen(false)}
        destroyOnHidden
      >
        <Form form={form} layout="vertical">
          <Form.Item name="configName" label="参数名称" rules={[{ required: true, message: '参数名称不能为空' }]}>
            <Input />
          </Form.Item>
          <Form.Item name="configKey" label="参数键名" rules={[{ required: true, message: '参数键名不能为空' }]}>
            <Input />
          </Form.Item>
          <Form.Item name="configValue" label="参数键值" rules={[{ required: true, message: '参数键值不能为空' }]}>
            <Input />
          </Form.Item>
          <Form.Item name="configType" label="系统内置" initialValue="N">
            <Input />
          </Form.Item>
          <Form.Item name="remark" label="备注">
            <Input.TextArea />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}
