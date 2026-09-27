// useEChart —— ECharts 容器 hook（7.0.0 缓存监控消费）：init/setOption/resize 跟随窗口/卸载 dispose

import { useEffect, useRef } from 'react'
import * as echarts from 'echarts'

export function useEChart(option: echarts.EChartsOption | undefined) {
  const ref = useRef<HTMLDivElement>(null)
  const chartRef = useRef<echarts.ECharts | null>(null)

  useEffect(() => {
    if (!ref.current) return
    const chart = echarts.init(ref.current)
    chartRef.current = chart
    const onResize = () => chart.resize()
    window.addEventListener('resize', onResize)
    return () => {
      window.removeEventListener('resize', onResize)
      chart.dispose()
      chartRef.current = null
    }
  }, [])

  useEffect(() => {
    if (chartRef.current && option) {
      chartRef.current.setOption(option)
    }
  }, [option])

  return ref
}
