// 缓存监控主页 —— 玫瑰图（roseType）+ 仪表盘（gauge），默认主题（基准 macarons 未注册回退默认，deviations #13）

import { useEffect, useState } from 'react'
import { Card, Col, Row, Statistic } from 'antd'
import { getCache } from '@/api/monitor/cache'
import type { EChartsOption } from 'echarts'
import { useEChart } from '@/hooks/useEChart'

interface CacheInfo {
  info?: { redis_version?: string; redis_mode?: string; standalone?: string; tcp_port?: string; connected_clients?: string; expire_total_keys?: string; evicted_keys?: string; uptime_in_days?: string; used_memory_human?: string; used_cpu_user_children?: string; maxmemory_human?: string; aof_enabled?: string; rdb_last_bgsave_status?: string; rdb_last_save_time?: string; rdb_changes_since_last_save?: string }
  dbSize?: number
  commandStats?: { name: string; value: string }[]
}

export default function CacheMonitor() {
  const [info, setInfo] = useState<CacheInfo>({})

  useEffect(() => {
    getCache().then((res) => setInfo((res as unknown as CacheInfo) || {}))
  }, [])


  const commandOption = (
    info.commandStats
      ? {
          tooltip: { trigger: 'item', formatter: '{a} <br/>{b} : {c} ({d}%)' },
          series: [
            {
              name: '命令',
              type: 'pie' as const,
              roseType: 'radius' as const,
              radius: [15, 120] as [number, number],
              data: info.commandStats.map((s) => ({ name: s.name, value: s.value })),
            },
          ],
        }
      : undefined
  ) as EChartsOption | undefined
  const usedMemory = parseFloat(String(info.info?.used_memory_human || '0'))
  const gaugeOption: EChartsOption = {
    series: [
      {
        name: '内存使用',
        type: 'gauge',
        min: 0,
        max: Math.max(2, Math.ceil(usedMemory * 2)),
        detail: { formatter: `${info.info?.used_memory_human || '0'}` },
        data: [{ value: usedMemory, name: '内存消耗' }],
      },
    ],
  }

  const commandRef = useEChart(commandOption)
  const gaugeRef = useEChart(gaugeOption)

  return (
    <div style={{ padding: 16 }}>
      <Row gutter={[16, 16]}>
        <Col xs={24} md={12}>
          <Card title="Redis 信息">
            <Row gutter={[8, 8]}>
              <Col span={8}><Statistic title="Redis版本" value={info.info?.redis_version || '-'} /></Col>
              <Col span={8}><Statistic title="运行模式" value={info.info?.redis_mode || '-'} /></Col>
              <Col span={8}><Statistic title="端口" value={info.info?.tcp_port || '-'} /></Col>
              <Col span={8}><Statistic title="客户端数" value={info.info?.connected_clients || '-'} /></Col>
              <Col span={8}><Statistic title="运行时间(天)" value={info.info?.uptime_in_days || '-'} /></Col>
              <Col span={8}><Statistic title="使用内存" value={info.info?.used_memory_human || '-'} /></Col>
            </Row>
          </Card>
        </Col>
        <Col xs={24} md={12}>
          <Card title="缓存监控">
            <div ref={gaugeRef} style={{ height: 300 }} />
          </Card>
        </Col>
      </Row>
      <Row gutter={[16, 16]} style={{ marginTop: 16 }}>
        <Col xs={24}>
          <Card title="命令统计">
            <div ref={commandRef} style={{ height: 360 }} />
          </Card>
        </Col>
      </Row>
    </div>
  )
}
