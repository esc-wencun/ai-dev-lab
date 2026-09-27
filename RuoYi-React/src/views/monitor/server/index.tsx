// 服务监控(server) —— features.serverMonitor 降级 + Descriptions 信息卡

import { useEffect, useState } from 'react'
import { Card, Col, Progress, Row, Typography } from 'antd'
import { getServer } from '@/api/monitor/server'
import PlatformGate from '@/components/PlatformGate'

interface ServerInfo {
  [k: string]: unknown
  cpu?: { cpuNum?: number; total?: number; sys?: number; used?: number; wait?: number; free?: number }
  mem?: { total?: number; used?: number; free?: number }
  jvm?: { total?: number; max?: number; free?: number; version?: string; home?: string; name?: string; startTime?: string; runTime?: string; used?: number; inputArgs?: string }
  sys?: { computerName?: string; osName?: string; computerIp?: string; osArch?: string }
  sysFiles?: { dirName?: string; sysTypeName?: string; typeName?: string; total?: string; free?: string; used?: string; usage?: number }[]
}

function UsageBar({ title, used, total, unit }: { title: string; used?: number; total?: number; unit: string }) {
  const percent = total ? Math.round(((used || 0) / total) * 100) : 0
  return (
    <div>
      <Typography.Text strong>{title}</Typography.Text>
      <Progress percent={percent} size="small" status={percent > 80 ? 'exception' : 'normal'} />
      <Typography.Text type="secondary">
        {(used || 0).toFixed(1)}{unit} / {(total || 0).toFixed(1)}{unit}
      </Typography.Text>
    </div>
  )
}

export default function Server() {
  const [info, setInfo] = useState<ServerInfo>({})
  const [loaded, setLoaded] = useState(false)

  useEffect(() => {
    // 基准：serverMonitor !== false 才拉数据
    getServer().then((res) => {
      const data = (res as unknown as { data: ServerInfo }).data
      setInfo(data || {})
      setLoaded(true)
    })
  }, [])

  const cpu = info.cpu || {}
  const mem = info.mem || {}
  const jvm = info.jvm || {}
  const sys = info.sys || {}

  return (
    <PlatformGate feature="serverMonitor" subtitle="服务监控在当前后端运行时未实现，暂不可用。">
      {(enabled) =>
        enabled && loaded ? (
          <div style={{ padding: 16 }}>
            <Row gutter={[16, 16]}>
              <Col xs={24} md={8}><Card title="CPU"><UsageBar title="CPU 使用率" used={cpu.used} total={cpu.total} unit="%" /></Card></Col>
              <Col xs={24} md={8}><Card title="内存"><UsageBar title="内存使用率" used={mem.used} total={mem.total} unit="GB" /></Card></Col>
              <Col xs={24} md={8}><Card title="JVM">
                <UsageBar title="JVM 使用率" used={jvm.used} total={jvm.max} unit="GB" />
                <Typography.Text type="secondary">Java 版本 {jvm.version || '-'} · 启动 {jvm.startTime || '-'}</Typography.Text>
              </Card></Col>
            </Row>
            <Row gutter={[16, 16]} style={{ marginTop: 16 }}>
              <Col xs={24} md={12}><Card title="系统信息">
                <Typography.Text>主机：{sys.computerName || '-'} · IP：{sys.computerIp || '-'}</Typography.Text><br />
                <Typography.Text>系统：{sys.osName || '-'} · 架构：{sys.osArch || '-'}</Typography.Text>
              </Card></Col>
              <Col xs={24} md={12}><Card title="磁盘状态">
                {(info.sysFiles || []).map((f, i) => (
                  <div key={i} style={{ marginBottom: 8 }}>
                    <Typography.Text>{f.dirName}（{f.sysTypeName}）</Typography.Text>
                    <Progress percent={f.usage} size="small" status={(f.usage || 0) > 80 ? 'exception' : 'normal'} />
                    <Typography.Text type="secondary">{f.used} / {f.total}（剩余 {f.free}）</Typography.Text>
                  </div>
                ))}
              </Card></Col>
            </Row>
          </div>
        ) : null
      }
    </PlatformGate>
  )
}
