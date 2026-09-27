// 锁屏解锁页 —— 对位基准 views/lock.vue（粒子背景简化，功能等价：时钟/密码/解锁回跳/退出重登）

import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router'
import { Button, Input } from 'antd'
import { unlockScreen as unlockApi } from '@/api/login'
import { unlockScreen } from '@/store/modules/lock'
import { logOut } from '@/store/modules/user'
import { useAppDispatch, useAppSelector } from '@/store/hooks'

export default function Lock() {
  const dispatch = useAppDispatch()
  const navigate = useNavigate()
  const lockPath = useAppSelector((s) => s.lock.lockPath)
  const nickName = useAppSelector((s) => s.user.nickName || s.user.name)
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const [time, setTime] = useState(new Date())

  useEffect(() => {
    const timer = setInterval(() => setTime(new Date()), 1000)
    return () => clearInterval(timer)
  }, [])

  const handleUnlock = async () => {
    try {
      await unlockApi(password)
      dispatch(unlockScreen())
      navigate(lockPath, { replace: true })
    } catch {
      // 后端 msg 已由 request.ts 弹出；输入框抖动由 error 状态标记
      setError('unlock-failed')
      setPassword('')
    }
  }

  const handleRelogin = () => {
    dispatch(unlockScreen())
    void dispatch(logOut())
    navigate('/login', { replace: true })
  }

  return (
    <div
      style={{
        display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center',
        minHeight: '100vh', background: 'linear-gradient(135deg, #1f2d3d 0%, #2d3a4b 100%)', color: '#fff',
      }}
    >
      <div style={{ fontSize: 56, fontWeight: 300, letterSpacing: 4 }}>
        {time.toLocaleTimeString('zh-CN', { hour12: false })}
      </div>
      <div style={{ fontSize: 16, margin: '16px 0 32px', opacity: 0.8 }}>
        {time.toLocaleDateString('zh-CN', { year: 'numeric', month: 'long', day: 'numeric', weekday: 'long' })}
      </div>
      <div className={error ? 'lock-shake' : ''} style={{ display: 'flex', gap: 8 }}>
        <Input.Password
          placeholder={`${nickName}，请输入密码解锁`}
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          onPressEnter={handleUnlock}
          style={{ width: 280 }}
        />
        <Button type="primary" onClick={handleUnlock}>
          解锁
        </Button>
      </div>
      <Button type="link" onClick={handleRelogin} style={{ marginTop: 24, color: 'rgba(255,255,255,0.65)' }}>
        退出重新登录
      </Button>
      {error && (
        <style>{`.lock-shake { animation: lockShake 0.4s; } @keyframes lockShake { 0%,100%{transform:translateX(0)} 25%{transform:translateX(-8px)} 75%{transform:translateX(8px)} }`}</style>
      )}
    </div>
  )
}
