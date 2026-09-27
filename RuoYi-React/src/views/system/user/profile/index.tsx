// 个人中心 —— 完整版（对位基准 views/system/user/profile/index.vue）
// 左信息卡（头像裁剪入口）+ 右 Tabs（基本资料 / 修改密码）；:activeTab? 直达

import { useEffect, useState } from 'react'
import { useNavigate, useParams } from 'react-router'
import {
  Avatar, Button, Card, Col, Form, Input, message, Row, Select, Tabs, Typography,
} from 'antd'
import { UserOutlined } from '@ant-design/icons'
import { getUserProfile, updateUserProfile, updateUserPwd } from '@/api/system/user'
import { useAppDispatch, useAppSelector } from '@/store/hooks'
import { getInfo } from '@/store/modules/user'
import { infoPwdValidator } from './passwordRules'
import AvatarCropper from './AvatarCropper'

interface ProfileResp {
  data?: { userName?: string; nickName?: string; phonenumber?: string; email?: string; sex?: string; createTime?: string }
  roleGroup?: string
  postGroup?: string
}

export default function Profile() {
  const dispatch = useAppDispatch()
  const navigate = useNavigate()
  const params = useParams()
  const storeUser = useAppSelector((s) => s.user)
  const [profile, setProfile] = useState<ProfileResp>({})
  const [userForm] = Form.useForm()
  const [pwdForm] = Form.useForm()
  const [cropperOpen, setCropperOpen] = useState(false)
  const [avatarUrl, setAvatarUrl] = useState(storeUser.avatar)
  const [activeTab, setActiveTab] = useState(params.activeTab === 'resetPwd' ? 'resetPwd' : 'userinfo')

  useEffect(() => {
    void getUserProfile().then((res) => {
      const p = res as unknown as ProfileResp
      setProfile(p)
      userForm.setFieldsValue({
        nickName: p.data?.nickName,
        phonenumber: p.data?.phonenumber,
        email: p.data?.email,
        sex: p.data?.sex || '0',
      })
    })
  }, [userForm])

  const saveUser = async () => {
    const values = await userForm.validateFields()
    await updateUserProfile(values)
    message.success('修改成功')
    void dispatch(getInfo())
  }

  const savePwd = async () => {
    const values = await pwdForm.validateFields()
    await updateUserPwd(values.oldPassword, values.newPassword)
    message.success('密码修改成功，请重新登录')
    pwdForm.resetFields()
    navigate('/login', { replace: true })
  }


  const u = profile.data || {}
  return (
    <div style={{ padding: 16 }}>
      <Row gutter={16}>
        <Col xs={24} md={6}>
          <Card size="small">
            <div style={{ textAlign: 'center', marginBottom: 12 }}>
              <Avatar size={96} src={avatarUrl || undefined} icon={!avatarUrl && <UserOutlined />} />
              <div>
                <Button type="link" size="small" onClick={() => setCropperOpen(true)}>修改头像</Button>
              </div>
            </div>
            <Typography.Text strong style={{ display: 'block', marginBottom: 8 }}>
              {u.nickName || storeUser.nickName}
            </Typography.Text>
            <Typography.Text type="secondary" style={{ display: 'block' }}>手机号码：{u.phonenumber || '-'}</Typography.Text>
            <Typography.Text type="secondary" style={{ display: 'block' }}>用户邮箱：{u.email || '-'}</Typography.Text>
            <Typography.Text type="secondary" style={{ display: 'block' }}>所属角色：{profile.roleGroup || '-'}</Typography.Text>
            <Typography.Text type="secondary" style={{ display: 'block' }}>所属岗位：{profile.postGroup || '-'}</Typography.Text>
            <Typography.Text type="secondary" style={{ display: 'block' }}>创建日期：{u.createTime || '-'}</Typography.Text>
          </Card>
        </Col>
        <Col xs={24} md={18}>
          <Card size="small">
            <Tabs
              activeKey={activeTab}
              onChange={(k) => {
                setActiveTab(k)
                window.history.replaceState(null, '', k === 'resetPwd' ? '/user/profile/resetPwd' : '/user/profile')
              }}
              items={[
                {
                  key: 'userinfo',
                  label: '基本资料',
                  children: (
                    <Form form={userForm} layout="vertical" style={{ maxWidth: 420 }}>
                      <Form.Item name="nickName" label="用户昵称" rules={[{ required: true, message: '用户昵称不能为空' }]}>
                        <Input maxLength={30} />
                      </Form.Item>
                      <Form.Item name="phonenumber" label="手机号码" rules={[
                        { required: true, message: '手机号码不能为空' },
                        { pattern: /^1[3|4|5|6|7|8|9][0-9]\d{8}$/, message: '请输入正确的手机号码' },
                      ]}>
                        <Input maxLength={11} />
                      </Form.Item>
                      <Form.Item name="email" label="邮箱" rules={[
                        { required: true, message: '邮箱不能为空' },
                        { type: 'email', message: '请输入正确的邮箱地址' },
                      ]}>
                        <Input maxLength={50} />
                      </Form.Item>
                      <Form.Item name="sex" label="性别">
                        <Select options={[{ label: '男', value: '0' }, { label: '女', value: '1' }]} />
                      </Form.Item>
                      <Button type="primary" onClick={() => void saveUser()}>保存</Button>
                    </Form>
                  ),
                },
                {
                  key: 'resetPwd',
                  label: '修改密码',
                  children: (
                    <Form form={pwdForm} layout="vertical" style={{ maxWidth: 420 }}>
                      <Form.Item name="oldPassword" label="旧密码" rules={[{ required: true, message: '旧密码不能为空' }]}>
                        <Input.Password />
                      </Form.Item>
                      <Form.Item name="newPassword" label="新密码" rules={[{ required: true, validator: infoPwdValidator }]}>
                        <Input.Password />
                      </Form.Item>
                      <Form.Item name="confirm" label="确认密码" dependencies={['newPassword']} rules={[
                        { required: true, message: '确认密码不能为空' },
                        ({ getFieldValue }) => ({
                          validator(_, v) {
                            if (!v || getFieldValue('newPassword') === v) return Promise.resolve()
                            return Promise.reject(new Error('两次输入的密码不一致'))
                          },
                        }),
                      ]}>
                        <Input.Password />
                      </Form.Item>
                      <Button type="primary" onClick={() => void savePwd()}>保存</Button>
                    </Form>
                  ),
                },
              ]}
            />
          </Card>
        </Col>
      </Row>

      <AvatarCropper
        open={cropperOpen}
        onClose={() => setCropperOpen(false)}
        onUploaded={(url) => setAvatarUrl(url)}
      />
    </div>
  )
}
