// 登录页 —— 对位 RuoYi-Vue3 src/views/login.vue（5.0.0 的核心页，2.0.0 链路入口先行落地）
// 行为：验证码（captchaEnabled===undefined 视为开启）、明文登录、记住我 RSA、redirect 跳转

import { useEffect, useState } from 'react'
import { useNavigate, useSearchParams } from 'react-router'
import { Button, Checkbox, Form, Input, Alert } from 'antd'
import { LockOutlined, UserOutlined, SafetyCertificateOutlined } from '@ant-design/icons'
import { getCodeImg } from '@/api/login'
import { login as loginThunk } from '@/store/modules/user'
import { unlockScreen } from '@/store/modules/lock'
import { useAppDispatch } from '@/store/hooks'
import { encrypt } from '@/utils/jsencrypt'
import Cookies from 'js-cookie'

interface LoginForm {
  username: string
  password: string
  code?: string
  rememberMe?: boolean
}

export default function Login() {
  const dispatch = useAppDispatch()
  const navigate = useNavigate()
  const [search] = useSearchParams()
  const [form] = Form.useForm<LoginForm>()
  const [captchaEnabled, setCaptchaEnabled] = useState(true)
  const [codeUrl, setCodeUrl] = useState('')
  const [uuid, setUuid] = useState('')
  const [loading, setLoading] = useState(false)
  const [tip, setTip] = useState('')

  const getCaptcha = async () => {
    const res = (await getCodeImg()) as unknown as {
      captchaEnabled?: boolean
      img: string
      uuid: string
    }
    setCaptchaEnabled(res.captchaEnabled !== false)
    setCodeUrl('data:image/gif;base64,' + res.img)
    setUuid(res.uuid)
  }

  // 记住我回显（cookie: username / password(RSA密文) / rememberMe，30 天）
  useEffect(() => {
    const username = Cookies.get('username') || 'admin'
    const password = Cookies.get('password')
    const rememberMe = Cookies.get('rememberMe') === 'true'
    form.setFieldsValue({
      username,
      password: password ? '••••••••' : 'admin123',
    })
    form.setFieldValue('rememberMe', rememberMe)
    void getCaptcha()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  const onFinish = async (values: LoginForm & { rememberMe?: boolean }) => {
    setLoading(true)
    setTip('')
    try {
      const rememberMe = !!values.rememberMe
      // 记住我：cookie 写 RSA 密文（基准 login.vue L113-120 行为）
      if (rememberMe) {
        Cookies.set('username', values.username, { expires: 30 })
        Cookies.set('password', encrypt(values.password) || '', { expires: 30 })
        Cookies.set('rememberMe', 'true', { expires: 30 })
      } else {
        Cookies.remove('username')
        Cookies.remove('password')
        Cookies.remove('rememberMe')
      }
      await dispatch(
        loginThunk({
          username: values.username,
          password: values.password,
          code: captchaEnabled ? values.code : undefined,
          uuid: captchaEnabled ? uuid : undefined,
        }),
      ).unwrap()
      // 登录即解锁（对位基准 user store login 末尾 unlockScreen）
      dispatch(unlockScreen())
      // redirect 参数跳转（保留其他 query——简化为只跳 path，与基准一致场景）
      const redirect = search.get('redirect')
      navigate(redirect ? decodeURIComponent(redirect) : '/index', { replace: true })
    } catch {
      // 后端 msg 已由 request.ts message.error 展示；刷新验证码
      if (captchaEnabled) void getCaptcha()
    } finally {
      setLoading(false)
    }
  }

  return (
    <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: '100vh', background: '#2d3a4b' }}>
      <div style={{ width: 400, background: '#fff', borderRadius: 8, padding: '32px 40px 24px' }}>
        <h2 style={{ textAlign: 'center', marginTop: 0 }}>若依管理系统</h2>
        {tip && <Alert type="error" message={tip} style={{ marginBottom: 16 }} />}
        <Form form={form} onFinish={onFinish} size="large">
          <Form.Item name="username" rules={[{ required: true, message: '请输入您的账号' }]}>
            <Input prefix={<UserOutlined />} placeholder="账号" />
          </Form.Item>
          <Form.Item name="password" rules={[{ required: true, message: '请输入您的密码' }]}>
            <Input.Password prefix={<LockOutlined />} placeholder="密码" />
          </Form.Item>
          {captchaEnabled && (
            <Form.Item name="code" rules={[{ required: true, message: '请输入验证码' }]}>
              <Input
                prefix={<SafetyCertificateOutlined />}
                placeholder="验证码"
                suffix={
                  <img
                    src={codeUrl}
                    onClick={() => void getCaptcha()}
                    style={{ height: 32, cursor: 'pointer', marginRight: -24 }}
                    alt="验证码"
                  />
                }
              />
            </Form.Item>
          )}
          <Form.Item name="rememberMe" valuePropName="checked">
            <Checkbox>记住密码</Checkbox>
          </Form.Item>
          <Button type="primary" htmlType="submit" block loading={loading}>
            登 录
          </Button>
        </Form>
      </div>
    </div>
  )
}
