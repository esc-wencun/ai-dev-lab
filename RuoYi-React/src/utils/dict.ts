import { useEffect, useState } from 'react'
import { useDispatch, useSelector } from 'react-redux'
import { getDicts } from '@/api/system/dict/data'
import { setDict } from '@/store/modules/dict'
import type { DictDataOption } from '@/store/modules/dict'
import type { RootState } from '@/store'

/**
 * 获取字典数据 —— 对位 RuoYi-Vue3 src/utils/dict.js useDict
 * 先查 store 缓存，未命中 GET /system/dict/data/type/{dictType} 后写缓存。
 * 注：基准版对同一类型的并发调用没有去重（各自请求），此处保持一致。
 */
export function useDict(...args: string[]): Record<string, DictDataOption[]> {
  const dispatch = useDispatch()
  const dictMap = useSelector((state: RootState) => state.dict.dict)
  const [fetched, setFetched] = useState<Record<string, DictDataOption[]>>({})

  useEffect(() => {
    args.forEach((dictType) => {
      const cached = dictMap[dictType]
      if (cached) return
      getDicts(dictType).then((resp: { data: Record<string, unknown>[] }) => {
        const options = resp.data.map((p) => ({
          label: p.dictLabel as string,
          value: p.dictValue as string,
          elTagType: p.listClass as string,
          elTagClass: p.cssClass as string,
        }))
        // 写 store 缓存 + 本地状态（未挂 Provider 的场景兜底不崩）
        try {
          dispatch(setDict({ key: dictType, value: options }))
        } catch {
          setFetched((prev) => ({ ...prev, [dictType]: options }))
        }
      })
    })
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [args.join(',')])

  const result: Record<string, DictDataOption[]> = {}
  args.forEach((dictType) => {
    result[dictType] = dictMap[dictType] || fetched[dictType] || []
  })
  return result
}
