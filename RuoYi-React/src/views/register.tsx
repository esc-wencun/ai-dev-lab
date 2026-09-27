// 注册页 —— 对位基准 register.vue（验证码 + 规则'0' 密码校验 + 成功跳登录）
// 5.0.0 完善视觉；本版为可工作骨架

import { useCallback, useEffect, useState } from 'react'
import { useNavigate } from 'react-router'
import { Alert, Button, Form, Input } from 'antd'
import { getCodeImg, register } from '@/api/login'
import { pwdValidatorFactory } from '@/utils/passwordRule'

interface RegisterForm {
  username: string
  password: string
  confirmPassword: string
  code?: string
}

export default function Register() {
  const navigate = useNavigate()
  const [form] = Form.useForm<RegisterForm>()
  const [codeUrl, setCodeUrl] = useState('')
  const [uuid, setUuid] = useState('')
  const [captchaEnabled, setCaptchaEnabled] = useState(true)
  const [tip, setTip] = useState('')
  const [loading, setLoading] = useState(false)

  const getCaptcha = useCallback(async () => {
    const res = (await getCodeImg()) as unknown as { captchaEnabled?: boolean; img: string; uuid: string }
    setCaptchaEnabled(res.captchaEnabled !== false)
    setCodeUrl('data:image/gif;base64,' + res.img)
    setUuid(res.uuid)
  }, [])

  useEffect(() => {
    void getCaptcha()
  }, [getCaptcha])

  const onFinish = async (values: RegisterForm) => {
    setLoading(true)
    setTip('')
    try {
      await register({
        username: values.username,
        password: values.password,
        confirmPassword: values.confirmPassword,
        code: captchaEnabled ? values.code : undefined,
        uuid: captchaEnabled ? uuid : undefined,
      })
      navigate('/login', { replace: true })
    } catch {
      setTip('注册失败，请重试')
      if (captchaEnabled) void getCaptcha()
    } finally {
      setLoading(false)
    }
  }

  const pwdValidator = pwdValidatorFactory('register')

  return (
    <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: '100vh', background: '#2d3a4b' }}>
      <div style={{ width: 400, background: '#fff', borderRadius: 8, padding: '32px 40px 24px' }}>
        <h2 style={{ textAlign: 'center', marginTop: 0 }}>注册</h2>
        {tip && <Alert type="error" message={tip} style={{ marginBottom: 16 }} />}
        <Form form={form} onFinish={onFinish} size="large">
          <Form.Item name="username" rules={[{ required: true, min: 2, max: 20, message: '账号长度必须介于 2 和 20 之间' }]}>
            <Input placeholder="账号" />
          </Form.Item>
          <Form.Item name="password" rules={[{ validator: pwdValidator }]}>
            <Input.Password placeholder="密码" />
          </Form.Item>
          <Form.Item
            name="confirmPassword"
            dependencies={['password']}
            rules={[
              { required: true, message: '请再次输入您的密码' },
              ({ getFieldValue }) => ({
                validator(_, value) {
                  if (!value || getFieldValue('password') === value) return Promise.resolve()
                  return Promise.reject(new Error('两次输入的密码不一致'))
                },
              }),
            ]}
          >
            <Input.Password placeholder="确认密码" />
          </Form.Item>
          {captchaEnabled && (
            <Form.Item name="code" rules={[{ required: true, message: '请输入验证码' }]}>
              <div style={{ display: 'flex', gap: 8 }}>
                <Input placeholder="验证码" />
                <img src={codeUrl} onClick={() => void getCaptcha()} style={{ height: 40, cursor: 'pointer' }} alt="验证码" />
              </div>
            </Form.Item>
          )}
          <Button type="primary" htmlType="submit" block loading={loading}>
            注 册
          </Button>
          <Button type="link" block onClick={() => navigate('/login')}>
            使用已有账户登录
          </Button>
        </Form>
      </div>
    </div>
  )
}
