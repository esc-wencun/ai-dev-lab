// Crontab 最近 5 次运行时间纯算法 —— 对位基准 RuoYi-Vue3/src/components/Crontab/result.vue 逐行移植
// 基准在 Vue 组件内用 ref 存中间状态（dateArr / dayRule / dayRuleSup），此处改为函数内局部变量，
// 算法分支、循环标签、拼接格式、结果文案（「没有达到条件的结果！」「最近100年内只有上面N条结果！」）与基准逐字等价。
// 与基准的差异：getNextRuns 增加可选 now 参数（默认 new Date()），仅为单测可确定性；基准只在组件挂载/表达式变化时同步计算，
// isShow/「计算结果中...」占位在 Vue 同一 tick 内实际不会渲染，React 侧同步 useMemo 计算后行为一致（见 index.tsx 注释）。

export function getNextRuns(ex: string, now: Date = new Date()): string[] {
  // 获取规则数组[0秒、1分、2时、3日、4月、5星期、6年]
  const ruleArr = ex.split(' ')
  // 用于记录进入循环的次数
  let nums = 0
  // 用于暂时存符号时间规则结果的数组
  const resultArr: string[] = []
  // 获取当前时间精确至[年、月、日、时、分、秒]
  let nYear = now.getFullYear()
  let nMonth = now.getMonth() + 1
  let nDay = now.getDate()
  let nHour = now.getHours()
  let nMin = now.getMinutes()
  let nSecond = now.getSeconds()
  // 规则拆出的候选值数组[0秒、1分、2时、3日、4月、5年]（对位基准 dateArr）
  const dateArr: number[][] = [[], [], [], [], [], []]
  // 日期附加规则（对位基准 dayRule / dayRuleSup ref）
  // 初始化即断言为完整联合类型：赋值发生在下方嵌套函数内，TS 控制流分析看不到，
  // 若以字面量 '' 起始会把变量窄化成 '' 导致后续比较全部报错
  let dayRule = '' as '' | 'workDay' | 'lastDay' | 'weekDay' | 'assWeek' | 'lastWeek'
  let dayRuleSup = '' as string | number | number[]

  // 根据规则获取到近100年可能年数组、月数组等等
  getSecondArr(ruleArr[0] ?? '')
  getMinArr(ruleArr[1] ?? '')
  getHourArr(ruleArr[2] ?? '')
  getDayArr(ruleArr[3] ?? '')
  getMonthArr(ruleArr[4] ?? '')
  getWeekArr(ruleArr[5] ?? '')
  getYearArr(ruleArr[6], nYear)
  // 将获取到的数组赋值-方便使用
  const sDate = dateArr[0]
  const mDate = dateArr[1]
  const hDate = dateArr[2]
  const DDate = dateArr[3]
  const MDate = dateArr[4]
  const YDate = dateArr[5]
  // 获取当前时间在数组中的索引
  let sIdx = getIndex(sDate, nSecond)
  let mIdx = getIndex(mDate, nMin)
  let hIdx = getIndex(hDate, nHour)
  let DIdx = getIndex(DDate, nDay)
  let MIdx = getIndex(MDate, nMonth)
  let YIdx = getIndex(YDate, nYear)
  // 重置月日时分秒的函数(后面用的比较多)
  const resetSecond = function () {
    sIdx = 0
    nSecond = sDate[sIdx]
  }
  const resetMin = function () {
    mIdx = 0
    nMin = mDate[mIdx]
    resetSecond()
  }
  const resetHour = function () {
    hIdx = 0
    nHour = hDate[hIdx]
    resetMin()
  }
  const resetDay = function () {
    DIdx = 0
    nDay = DDate[DIdx]
    resetHour()
  }
  const resetMonth = function () {
    MIdx = 0
    nMonth = MDate[MIdx]
    resetDay()
  }
  // 如果当前年份不为数组中当前值
  if (nYear !== YDate[YIdx]) {
    resetMonth()
  }
  // 如果当前月份不为数组中当前值
  if (nMonth !== MDate[MIdx]) {
    resetDay()
  }
  // 如果当前"日"不为数组中当前值
  if (nDay !== DDate[DIdx]) {
    resetHour()
  }
  // 如果当前"时"不为数组中当前值
  if (nHour !== hDate[hIdx]) {
    resetMin()
  }
  // 如果当前"分"不为数组中当前值
  if (nMin !== mDate[mIdx]) {
    resetSecond()
  }
  // 循环年份数组
  goYear: for (let Yi = YIdx; Yi < YDate.length; Yi++) {
    const YY = YDate[Yi]
    // 如果到达最大值时
    if (nMonth > MDate[MDate.length - 1]) {
      resetMonth()
      continue
    }
    // 循环月份数组
    goMonth: for (let Mi = MIdx; Mi < MDate.length; Mi++) {
      // 赋值、方便后面运算
      let MM: number | string = MDate[Mi]
      MM = MM < 10 ? '0' + MM : MM
      // 如果到达最大值时
      if (nDay > DDate[DDate.length - 1]) {
        resetDay()
        if (Mi === MDate.length - 1) {
          resetMonth()
          continue goYear
        }
        continue
      }
      // 循环日期数组
      goDay: for (let Di = DIdx; Di < DDate.length; Di++) {
        // 赋值、方便后面运算（DD 保持数值参与运算，thisDD 为补零展示值——对位基准同名变量）
        let DD: number = DDate[Di]
        let thisDD: number | string = DD < 10 ? '0' + DD : DD
        // 如果到达最大值时
        if (nHour > hDate[hDate.length - 1]) {
          resetHour()
          if (Di === DDate.length - 1) {
            resetDay()
            if (Mi === MDate.length - 1) {
              resetMonth()
              continue goYear
            }
            continue goMonth
          }
          continue
        }
        // 判断日期的合法性，不合法的话也是跳出当前循环
        if (
          checkDate(YY + '-' + MM + '-' + thisDD + ' 00:00:00') !== true &&
          dayRule !== 'workDay' &&
          dayRule !== 'lastWeek' &&
          dayRule !== 'lastDay'
        ) {
          resetDay()
          continue goMonth
        }
        // 如果日期规则中有值时
        if (dayRule === 'lastDay') {
          // 如果不是合法日期则需要将前将日期调到合法日期即月末最后一天
          if (checkDate(YY + '-' + MM + '-' + thisDD + ' 00:00:00') !== true) {
            while (DD > 0 && checkDate(YY + '-' + MM + '-' + thisDD + ' 00:00:00') !== true) {
              DD--
              thisDD = DD < 10 ? '0' + DD : DD
            }
          }
        } else if (dayRule === 'workDay') {
          // 校验并调整如果是2月30号这种日期传进来时需调整至正常月底
          if (checkDate(YY + '-' + MM + '-' + thisDD + ' 00:00:00') !== true) {
            while (DD > 0 && checkDate(YY + '-' + MM + '-' + thisDD + ' 00:00:00') !== true) {
              DD--
              thisDD = DD < 10 ? '0' + DD : DD
            }
          }
          // 获取达到条件的日期是星期X
          const thisWeek = formatDate(new Date(YY + '-' + MM + '-' + thisDD + ' 00:00:00'), 'week')
          // 当星期日时
          if (thisWeek === 1) {
            // 先找下一个日，并判断是否为月底
            DD++
            thisDD = DD < 10 ? '0' + DD : DD
            // 判断下一日已经不是合法日期
            if (checkDate(YY + '-' + MM + '-' + thisDD + ' 00:00:00') !== true) {
              DD -= 3
            }
          } else if (thisWeek === 7) {
            // 当星期6时只需判断不是1号就可进行操作
            if (dayRuleSup !== 1) {
              DD--
            } else {
              DD += 2
            }
          }
        } else if (dayRule === 'weekDay') {
          // 如果指定了是星期几
          // 获取当前日期是属于星期几
          const thisWeek = formatDate(new Date(YY + '-' + MM + '-' + DD + ' 00:00:00'), 'week')
          // 校验当前星期是否在星期池（dayRuleSup）中
          if ((dayRuleSup as number[]).indexOf(thisWeek) < 0) {
            // 如果到达最大值时
            if (Di === DDate.length - 1) {
              resetDay()
              if (Mi === MDate.length - 1) {
                resetMonth()
                continue goYear
              }
              continue goMonth
            }
            continue
          }
        } else if (dayRule === 'assWeek') {
          // 如果指定了是第几周的星期几
          // 获取每月1号是属于星期几
          const thisWeek = formatDate(new Date(YY + '-' + MM + '-' + DD + ' 00:00:00'), 'week')
          if ((dayRuleSup as number[])[1] >= thisWeek) {
            DD = ((dayRuleSup as number[])[0] - 1) * 7 + (dayRuleSup as number[])[1] - thisWeek + 1
          } else {
            DD = (dayRuleSup as number[])[0] * 7 + (dayRuleSup as number[])[1] - thisWeek + 1
          }
        } else if (dayRule === 'lastWeek') {
          // 如果指定了每月最后一个星期几
          // 校验并调整如果是2月30号这种日期传进来时需调整至正常月底
          if (checkDate(YY + '-' + MM + '-' + thisDD + ' 00:00:00') !== true) {
            while (DD > 0 && checkDate(YY + '-' + MM + '-' + thisDD + ' 00:00:00') !== true) {
              DD--
              thisDD = DD < 10 ? '0' + DD : DD
            }
          }
          // 获取月末最后一天是星期几
          const thisWeek = formatDate(new Date(YY + '-' + MM + '-' + thisDD + ' 00:00:00'), 'week')
          // 找到要求中最近的那个星期几
          if ((dayRuleSup as number) < thisWeek) {
            DD -= thisWeek - (dayRuleSup as number)
          } else if ((dayRuleSup as number) > thisWeek) {
            DD -= 7 - ((dayRuleSup as number) - thisWeek)
          }
        }
        // 判断时间值是否小于10置换成"05"这种格式
        if (DD < 10) {
          thisDD = '0' + DD
        } else {
          thisDD = DD
        }
        // 循环"时"数组
        goHour: for (let hi = hIdx; hi < hDate.length; hi++) {
          const hh = hDate[hi] < 10 ? '0' + hDate[hi] : hDate[hi]
          // 如果到达最大值时
          if (nMin > mDate[mDate.length - 1]) {
            resetMin()
            if (hi === hDate.length - 1) {
              resetHour()
              if (Di === DDate.length - 1) {
                resetDay()
                if (Mi === MDate.length - 1) {
                  resetMonth()
                  continue goYear
                }
                continue goMonth
              }
              continue goDay
            }
            continue
          }
          // 循环"分"数组
          goMin: for (let mi = mIdx; mi < mDate.length; mi++) {
            const mm = mDate[mi] < 10 ? '0' + mDate[mi] : mDate[mi]
            // 如果到达最大值时
            if (nSecond > sDate[sDate.length - 1]) {
              resetSecond()
              if (mi === mDate.length - 1) {
                resetMin()
                if (hi === hDate.length - 1) {
                  resetHour()
                  if (Di === DDate.length - 1) {
                    resetDay()
                    if (Mi === MDate.length - 1) {
                      resetMonth()
                      continue goYear
                    }
                    continue goMonth
                  }
                  continue goDay
                }
                continue goHour
              }
              continue
            }
            // 循环"秒"数组
            goSecond: for (let si = sIdx; si <= sDate.length - 1; si++) {
              const ss = sDate[si] < 10 ? '0' + sDate[si] : sDate[si]
              // 添加当前时间（时间合法性在日期循环时已经判断）
              if (MM !== '00' && thisDD !== '00') {
                resultArr.push(YY + '-' + MM + '-' + thisDD + ' ' + hh + ':' + mm + ':' + ss)
                nums++
              }
              // 如果条数满了就退出循环
              if (nums === 5) break goYear
              // 如果到达最大值时
              if (si === sDate.length - 1) {
                resetSecond()
                if (mi === mDate.length - 1) {
                  resetMin()
                  if (hi === hDate.length - 1) {
                    resetHour()
                    if (Di === DDate.length - 1) {
                      resetDay()
                      if (Mi === MDate.length - 1) {
                        resetMonth()
                        continue goYear
                      }
                      continue goMonth
                    }
                    continue goDay
                  }
                  continue goHour
                }
                continue goMin
              }
            } //goSecond
          } //goMin
        } //goHour
      } //goDay
    } //goMonth
  }
  // 判断100年内的结果条数
  if (resultArr.length === 0) {
    return ['没有达到条件的结果！']
  }
  if (resultArr.length !== 5) {
    resultArr.push('最近100年内只有上面' + resultArr.length + '条结果！')
  }
  return resultArr

  // 用于计算某位数字在数组中的索引
  function getIndex(arr: number[], value: number): number {
    if (value <= arr[0] || value > arr[arr.length - 1]) {
      return 0
    }
    for (let i = 0; i < arr.length - 1; i++) {
      if (value > arr[i] && value <= arr[i + 1]) {
        return i + 1
      }
    }
    return 0
  }
  // 获取"年"数组
  function getYearArr(rule: string | undefined, year: number) {
    dateArr[5] = getOrderArr(year, year + 100)
    if (rule !== undefined) {
      if (rule.indexOf('-') >= 0) {
        dateArr[5] = getCycleArr(rule, year + 100, false)
      } else if (rule.indexOf('/') >= 0) {
        dateArr[5] = getAverageArr(rule, year + 100)
      } else if (rule !== '*') {
        dateArr[5] = getAssignArr(rule)
      }
    }
  }
  // 获取"月"数组
  function getMonthArr(rule: string) {
    dateArr[4] = getOrderArr(1, 12)
    if (rule.indexOf('-') >= 0) {
      dateArr[4] = getCycleArr(rule, 12, false)
    } else if (rule.indexOf('/') >= 0) {
      dateArr[4] = getAverageArr(rule, 12)
    } else if (rule !== '*') {
      dateArr[4] = getAssignArr(rule)
    }
  }
  // 获取"日"数组-主要为日期规则
  function getWeekArr(rule: string) {
    // 只有当日期规则的两个值均为""时则表达日期是有选项的
    if (dayRule === '' && dayRuleSup === '') {
      if (rule.indexOf('-') >= 0) {
        dayRule = 'weekDay'
        dayRuleSup = getCycleArr(rule, 7, false)
      } else if (rule.indexOf('#') >= 0) {
        dayRule = 'assWeek'
        const matchRule = rule.match(/[0-9]{1}/g)!
        dayRuleSup = [Number(matchRule[1]), Number(matchRule[0])]
        dateArr[3] = [1]
        if ((dayRuleSup as number[])[1] === 7) {
          ;(dayRuleSup as number[])[1] = 0
        }
      } else if (rule.indexOf('L') >= 0) {
        dayRule = 'lastWeek'
        dayRuleSup = Number(rule.match(/[0-9]{1,2}/g)![0])
        dateArr[3] = [31]
        if (dayRuleSup === 7) {
          dayRuleSup = 0
        }
      } else if (rule !== '*' && rule !== '?') {
        dayRule = 'weekDay'
        dayRuleSup = getAssignArr(rule)
      }
    }
  }
  // 获取"日"数组-少量为日期规则
  function getDayArr(rule: string) {
    dateArr[3] = getOrderArr(1, 31)
    dayRule = ''
    dayRuleSup = ''
    if (rule.indexOf('-') >= 0) {
      dateArr[3] = getCycleArr(rule, 31, false)
      dayRuleSup = 'null'
    } else if (rule.indexOf('/') >= 0) {
      dateArr[3] = getAverageArr(rule, 31)
      dayRuleSup = 'null'
    } else if (rule.indexOf('W') >= 0) {
      dayRule = 'workDay'
      dayRuleSup = Number(rule.match(/[0-9]{1,2}/g)![0])
      dateArr[3] = [dayRuleSup as number]
    } else if (rule.indexOf('L') >= 0) {
      dayRule = 'lastDay'
      dayRuleSup = 'null'
      dateArr[3] = [31]
    } else if (rule !== '*' && rule !== '?') {
      dateArr[3] = getAssignArr(rule)
      dayRuleSup = 'null'
    } else if (rule === '*') {
      dayRuleSup = 'null'
    }
  }
  // 获取"时"数组
  function getHourArr(rule: string) {
    dateArr[2] = getOrderArr(0, 23)
    if (rule.indexOf('-') >= 0) {
      dateArr[2] = getCycleArr(rule, 24, true)
    } else if (rule.indexOf('/') >= 0) {
      dateArr[2] = getAverageArr(rule, 23)
    } else if (rule !== '*') {
      dateArr[2] = getAssignArr(rule)
    }
  }
  // 获取"分"数组
  function getMinArr(rule: string) {
    dateArr[1] = getOrderArr(0, 59)
    if (rule.indexOf('-') >= 0) {
      dateArr[1] = getCycleArr(rule, 60, true)
    } else if (rule.indexOf('/') >= 0) {
      dateArr[1] = getAverageArr(rule, 59)
    } else if (rule !== '*') {
      dateArr[1] = getAssignArr(rule)
    }
  }
  // 获取"秒"数组
  function getSecondArr(rule: string) {
    dateArr[0] = getOrderArr(0, 59)
    if (rule.indexOf('-') >= 0) {
      dateArr[0] = getCycleArr(rule, 60, true)
    } else if (rule.indexOf('/') >= 0) {
      dateArr[0] = getAverageArr(rule, 59)
    } else if (rule !== '*') {
      dateArr[0] = getAssignArr(rule)
    }
  }
  // 根据传进来的min-max返回一个顺序的数组
  function getOrderArr(min: number, max: number): number[] {
    const arr: number[] = []
    for (let i = min; i <= max; i++) {
      arr.push(i)
    }
    return arr
  }
  // 根据规则中指定的零散值返回一个数组
  function getAssignArr(rule: string): number[] {
    const arr: number[] = []
    const assiginArr = rule.split(',')
    for (let i = 0; i < assiginArr.length; i++) {
      arr[i] = Number(assiginArr[i])
    }
    arr.sort(compare)
    return arr
  }
  // 根据一定算术规则计算返回一个数组
  function getAverageArr(rule: string, limit: number): number[] {
    const arr: number[] = []
    const agArr = rule.split('/')
    let min = Number(agArr[0])
    const step = Number(agArr[1])
    while (min <= limit) {
      arr.push(min)
      min += step
    }
    return arr
  }
  // 根据规则返回一个具有周期性的数组
  function getCycleArr(rule: string, limit: number, status: boolean): number[] {
    // status--表示是否从0开始（则从1开始）
    const arr: number[] = []
    const cycleArr = rule.split('-')
    const min = Number(cycleArr[0])
    let max = Number(cycleArr[1])
    if (min > max) {
      max += limit
    }
    for (let i = min; i <= max; i++) {
      let add = 0
      if (status === false && i % limit === 0) {
        add = limit
      }
      arr.push(Math.round((i % limit) + add))
    }
    arr.sort(compare)
    return arr
  }
  // 比较数字大小（用于Array.sort）
  function compare(value1: number, value2: number): number {
    if (value2 - value1 > 0) {
      return -1
    } else {
      return 1
    }
  }
}

// 格式化日期格式如：2017-9-19 18:04:33（对位基准 formatDate；type='week' 时返回 Quartz 语义星期：1=星期日）
export function formatDate(value: Date): string
export function formatDate(value: Date, type: 'week'): number
export function formatDate(value: Date, type?: string): number | string {
  // 计算日期相关值
  const Y = value.getFullYear()
  const M = value.getMonth() + 1
  const D = value.getDate()
  const h = value.getHours()
  const m = value.getMinutes()
  const s = value.getSeconds()
  const week = value.getDay()
  // 如果传递了type的话
  if (type === undefined) {
    return (
      Y +
      '-' +
      (M < 10 ? '0' + M : M) +
      '-' +
      (D < 10 ? '0' + D : D) +
      ' ' +
      (h < 10 ? '0' + h : h) +
      ':' +
      (m < 10 ? '0' + m : m) +
      ':' +
      (s < 10 ? '0' + s : s)
    )
  } else if (type === 'week') {
    // 在quartz中 1为星期日
    return week + 1
  }
  return ''
}

// 检查日期是否存在（对位基准 checkDate：解析后再格式化，与原串一致才合法）
export function checkDate(value: string): boolean {
  const time = new Date(value)
  const format = formatDate(time)
  return value === format
}
