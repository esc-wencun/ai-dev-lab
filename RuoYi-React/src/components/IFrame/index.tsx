// iFrame —— 对位基准 components/iFrame（全高 iframe + loading + resize 跟随）

import { useEffect, useRef, useState } from 'react'
import { Spin } from 'antd'

interface IFrameProps {
  src: string
}

export default function IFrame({ src }: IFrameProps) {
  const [loading, setLoading] = useState(true)
  const ref = useRef<HTMLIFrameElement>(null)

  const calcHeight = () => document.documentElement.clientHeight - 94.5

  useEffect(() => {
    const onResize = () => {
      if (ref.current) ref.current.height = String(calcHeight())
    }
    window.addEventListener('resize', onResize)
    return () => window.removeEventListener('resize', onResize)
  }, [])

  return (
    <div style={{ position: 'relative' }}>
      <Spin spinning={loading} tip="正在加载页面，请稍候！" size="large">
        <iframe
          ref={ref}
          src={src}
          height={calcHeight()}
          frameBorder="0"
          scrolling="auto"
          style={{ width: '100%' }}
          onLoad={() => setTimeout(() => setLoading(false), 300)}
        />
      </Spin>
    </div>
  )
}
