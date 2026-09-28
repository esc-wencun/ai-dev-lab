// 锁屏解锁页 —— 对位基准 views/lock.vue（视觉完整版：时钟/日期/头像/毛玻璃卡片/圆角输入+圆形按钮）
// 功能逻辑不变：unlockScreen API → 回跳 lockPath；失败显示 msg + 抖动；「退出重新登录」= unlock + logOut
// 粒子背景省略（deviations #4 允许），保留渐变底 + 时钟/卡片质感

import { useEffect, useRef, useState } from 'react'
import { useNavigate, Navigate } from 'react-router'
import { unlockScreen as unlockApi } from '@/api/login'
import { unlockScreen } from '@/store/modules/lock'
import { logOut } from '@/store/modules/user'
import { useAppDispatch, useAppSelector } from '@/store/hooks'

export default function Lock() {
  const dispatch = useAppDispatch()
  const navigate = useNavigate()
  const lockPath = useAppSelector((s) => s.lock.lockPath)
  const nickName = useAppSelector((s) => s.user.nickName || s.user.name)
  const avatar = useAppSelector((s) => s.user.avatar)
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)
  const [shaking, setShaking] = useState(false)
  const [time, setTime] = useState(new Date())
  const inputRef = useRef<HTMLInputElement>(null)

  useEffect(() => {
    const timer = setInterval(() => setTime(new Date()), 1000)
    // 挂载即聚焦密码框（对位基准 nextTick focus）
    inputRef.current?.focus()
    return () => clearInterval(timer)
  }, [])

  const showError = (msg: string) => {
    setError(msg)
    setShaking(true)
    setTimeout(() => setShaking(false), 600)
  }

  // 未锁访问 /lock 的兜底（守卫「!isLock → '/'」分支因解锁竞态移出，见 AuthGuard 注释）。
  // unlockFlow = 解锁成功后的渲染期重定向标志：任何 imperative navigate 都会与
  // redux 重渲染/自检 effect 竞态，只有渲染期 <Navigate> 是最终一致的。
  const isLock = useAppSelector((s) => s.lock.isLock)
  const [unlockFlow, setUnlockFlow] = useState(false)
  useEffect(() => {
    if (!isLock && !unlockFlow) navigate('/index', { replace: true })
  }, [isLock, unlockFlow, navigate])

  const handleUnlock = async () => {
    if (!password) {
      showError('请输入密码')
      return
    }
    setLoading(true)
    setError('')
    try {
      await unlockApi(password)
      // 进入渲染期重定向：组件重渲染（isLock 已 false）时输出 <Navigate to={lockPath}>,
      // 由 React Router 在提交阶段完成跳转——不与任何 effect 竞态
      setUnlockFlow(true)
      dispatch(unlockScreen())
    } catch (err) {
      // request.ts 已弹后端 msg；本地再显示 + 抖动（对位基准行为）
      const msg = err instanceof Error ? err.message : String(err)
      showError(msg || '解锁失败')
      setPassword('')
      inputRef.current?.focus()
    } finally {
      setLoading(false)
    }
  }

  const handleRelogin = () => {
    dispatch(unlockScreen())
    void dispatch(logOut())
    navigate('/login', { replace: true })
  }

  const pad = (n: number) => String(n).padStart(2, '0')
  const days = ['星期日', '星期一', '星期二', '星期三', '星期四', '星期五', '星期六']
  const currentTime = `${pad(time.getHours())}:${pad(time.getMinutes())}:${pad(time.getSeconds())}`
  const currentDate = `${time.getFullYear()}年${time.getMonth() + 1}月${time.getDate()}日 ${days[time.getDay()]}`

  // 解锁成功 → 渲染期重定向回锁屏前路径（最终一致，无 effect 竞态）
  if (unlockFlow) return <Navigate to={lockPath} replace />

  return (
    <div className="lock-container">
      {/* 时钟 */}
      <div className="lock-time">{currentTime}</div>
      <div className="lock-date">{currentDate}</div>

      {/* 锁屏卡片 */}
      <div className="lock-card">
        <div className="avatar-wrap">
          <img
            src={avatar || undefined}
            className="lock-avatar"
            alt=""
            onError={(e) => { (e.target as HTMLImageElement).style.visibility = 'hidden' }}
          />
          <div className="lock-icon">🔒</div>
        </div>
        <div className="lock-username">{nickName}</div>
        <div className="lock-hint">系统已锁定，请输入密码解锁</div>

        <div className={`input-wrap${shaking ? ' shake' : ''}`}>
          <input
            ref={inputRef}
            value={password}
            type="password"
            placeholder="请输入登录密码"
            className="lock-input"
            onChange={(e) => setPassword(e.target.value)}
            onKeyDown={(e) => { if (e.key === 'Enter') void handleUnlock() }}
            autoComplete="off"
          />
          <button className="unlock-btn" onClick={() => void handleUnlock()} disabled={loading}>
            {loading ? <span className="loading-dot">···</span> : <span>→</span>}
          </button>
        </div>

        {error && <div className="error-msg">{error}</div>}

        <div className="lock-footer">
          <a href="javascript:;" onClick={handleRelogin}>退出重新登录</a>
        </div>
      </div>
    </div>
  )
}
