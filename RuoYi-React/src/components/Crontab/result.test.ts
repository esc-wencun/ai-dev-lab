// result.ts（最近5次运行时间，逐行移植自基准 result.vue）单测：
// 用固定 now 保证确定性；期望值由 cron 语义人工推得，覆盖通配/工作日/月末/第N周/最后星期X/指定周几/年域等主要分支
import { describe, expect, it } from 'vitest'
import { getNextRuns } from '@/components/Crontab/result'
describe('getNextRuns 算法 sanity（移植自 result.vue）', () => {
  it('每分钟第 0 秒：0 * * * * ?（2026-09-28 08:00:30 → 前 5 个整分）', () => {
    const now = new Date(2026, 8, 28, 8, 0, 30)
    const runs = getNextRuns('0 * * * * ?', now)
    expect(runs).toEqual([
      '2026-09-28 08:01:00',
      '2026-09-28 08:02:00',
      '2026-09-28 08:03:00',
      '2026-09-28 08:04:00',
      '2026-09-28 08:05:00',
    ])
  })

  it('每天 8 点整：0 0 8 * * ?', () => {
    const now = new Date(2026, 8, 28, 9, 0, 0)
    const runs = getNextRuns('0 0 8 * * ?', now)
    expect(runs[0]).toBe('2026-09-29 08:00:00')
    expect(runs[4]).toBe('2026-10-03 08:00:00')
  })

  it('工作日 15 号：0 0 0 15W * ?（2026-09-15 是周二 → 9/15 本身；now=9/16 → 10/15）', () => {
    const now = new Date(2026, 8, 16, 0, 0, 0)
    const runs = getNextRuns('0 0 0 15W * ?', now)
    expect(runs[0]).toBe('2026-10-15 00:00:00')
  })

  it('每月最后一天：0 0 0 L * ?', () => {
    const now = new Date(2026, 8, 20, 0, 0, 0)
    const runs = getNextRuns('0 0 0 L * ?', now)
    expect(runs[0]).toBe('2026-09-30 00:00:00')
    expect(runs[1]).toBe('2026-10-31 00:00:00')
  })

  it('第 2 周星期一：0 0 0 ? * 2#2（9/28 是 9 月第 4 个周一 → 10 月第 2 个周一 10/12）', () => {
    const now = new Date(2026, 8, 28, 0, 0, 0)
    const runs = getNextRuns('0 0 0 ? * 2#2', now)
    expect(runs[0]).toBe('2026-10-12 00:00:00')
  })

  it('本月最后一个星期五：0 0 0 ? * 6L（2026-09-25 是最后一个周五）', () => {
    const now = new Date(2026, 8, 20, 0, 0, 0)
    const runs = getNextRuns('0 0 0 ? * 6L', now)
    expect(runs[0]).toBe('2026-09-25 00:00:00')
    expect(runs[1]).toBe('2026-10-30 00:00:00')
  })

  it('指定周几集合：0 0 0 ? * 2,4,6（周一/三/五）', () => {
    const now = new Date(2026, 8, 28, 0, 0, 0) // 周一
    const runs = getNextRuns('0 0 0 ? * 2,4,6', now)
    expect(runs[0]).toBe('2026-09-28 00:00:00') // 当下即周一，第一项就是当前时刻
    expect(runs[1]).toBe('2026-09-30 00:00:00') // 周三
    expect(runs[2]).toBe('2026-10-02 00:00:00') // 周五
  })

  it('带年份 + 步长秒：0/30 0 0 1 * ? 2027', () => {
    const now = new Date(2026, 8, 28, 0, 0, 0)
    const runs = getNextRuns('0/30 0 0 1 * ? 2027', now)
    expect(runs[0]).toBe('2027-01-01 00:00:00')
    expect(runs[1]).toBe('2027-01-01 00:00:30')
    expect(runs[2]).toBe('2027-02-01 00:00:00')
  })
})
