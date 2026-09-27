// 请求层 —— 逐条复刻 RuoYi-Vue3 src/utils/request.js 行为契约（三版后端兼容的生命线）
// 任何分支顺序、文案、存储键名不得"顺手优化"；可疑处按原版语义复刻并注明。

import axios from 'axios'
import { Modal, message, notification } from 'antd'
import type { AxiosError, AxiosRequestConfig } from 'axios'
import { getToken } from '@/utils/auth'
import errorCode from '@/utils/errorCode'
import { tansParams, blobValidate } from '@/utils/ruoyi'
import cache from '@/plugins/cache'
import { saveAs } from 'file-saver'

// 是否显示重新登录（401 防重入标志，守卫层也消费）
export const isRelogin = { show: false }

// 创建 axios 实例
const service = axios.create({
  // axios中请求配置有baseURL选项，表示请求URL公共部分
  baseURL: import.meta.env.VITE_APP_BASE_API,
  // 超时
  timeout: 10000,
})

service.defaults.headers['Content-Type'] = 'application/json;charset=utf-8'

// request 拦截器
service.interceptors.request.use(
  (config) => {
    // 是否需要设置 token（headers.isToken === false 跳过）
    const isToken = (config.headers || {}).isToken === false
    // 是否需要防止数据重复提交（headers.repeatSubmit === false 跳过）
    const isRepeatSubmit = (config.headers?.repeatSubmit as boolean | undefined) === false
    // 间隔时间(ms)，小于此时间视为重复提交（headers.interval 自定义）
    const interval = ((config.headers?.interval as number) || 1000) as number
    if (getToken() && !isToken) {
      // 让每个请求携带自定义token —— Bearer 前缀不能少（裸 token 一律 401 的坑）
      config.headers.set('Authorization', 'Bearer ' + getToken())
    }
    // get请求映射params参数
    if (config.method === 'get' && config.params) {
      let url = config.url + '?' + tansParams(config.params)
      url = url.slice(0, -1)
      config.params = {}
      config.url = url
    }
    // 防重复提交（仅 post/put）
    if (!isRepeatSubmit && (config.method === 'post' || config.method === 'put')) {
      const requestObj = {
        url: config.url as string,
        data: typeof config.data === 'object' ? JSON.stringify(config.data) : config.data,
        time: new Date().getTime(),
      }
      const requestSize = Object.keys(JSON.stringify(requestObj)).length // 请求数据大小
      const limitSize = 5 * 1024 * 1024 // 限制存放数据5M
      if (requestSize >= limitSize) {
        console.warn(`[${config.url}]: ` + '请求数据大小超出允许的5M限制，无法进行防重复提交验证。')
        return config
      }
      const sessionObj = cache.session.getJSON<{ url: string; data: string; time: number } | string>('sessionObj') as { url: string; data: string; time: number } | null
      if (sessionObj === undefined || sessionObj === null || (sessionObj as unknown as string) === '') {
        cache.session.setJSON('sessionObj', requestObj)
      } else {
        const s_url = sessionObj.url // 请求地址
        const s_data = sessionObj.data // 请求数据
        const s_time = sessionObj.time // 请求时间
        if (s_data === requestObj.data && requestObj.time - s_time < interval && s_url === requestObj.url) {
          const message = '数据正在处理，请勿重复提交'
          console.warn(`[${s_url}]: ` + message)
          return Promise.reject(new Error(message))
        } else {
          cache.session.setJSON('sessionObj', requestObj)
        }
      }
    }
    return config
  },
  (error) => {
    console.log(error)
    Promise.reject(error)
  },
)

// 响应拦截器（按 body code 分支，不按 HTTP 状态码）
service.interceptors.response.use(
  (res) => {
    // 未设置状态码则默认成功状态
    const code = res.data.code || 200
    // 获取错误信息
    const msg = errorCode[code] || res.data.msg || errorCode['default']
    // 二进制数据则直接返回
    if (res.request.responseType === 'blob' || res.request.responseType === 'arraybuffer') {
      return res.data
    }
    if (code === 401) {
      if (!isRelogin.show) {
        isRelogin.show = true
        // 文案与基准逐字一致；确认 → logOut + 回 /index；取消 → 复位标志
        Modal.confirm({
          title: '系统提示',
          content: '登录状态已过期，您可以继续留在该页面，或者重新登录',
          okText: '重新登录',
          cancelText: '取消',
          onOk: () => {
            isRelogin.show = false
            // 动态引用避免 request ↔ store 循环依赖；dispatch 全局 store（main.tsx 注入）
            import('@/store').then(({ store }) => {
              import('@/store/modules/user').then(({ logOut }) => {
                store.dispatch(logOut()).then(() => {
                  window.location.href = '/index'
                })
              })
            })
          },
          onCancel: () => {
            isRelogin.show = false
          },
        })
      }
      return Promise.reject('无效的会话，或者会话已过期，请重新登录。')
    } else if (code === 500) {
      message.error(msg)
      return Promise.reject(new Error(msg))
    } else if (code === 601) {
      message.warning(msg)
      return Promise.reject(new Error(msg))
    } else if (code !== 200) {
      notification.error({ message: msg })
      return Promise.reject('error')
    } else {
      // 注意：返回整个 body（含 code/msg/data/rows/total），不是 data 字段
      return Promise.resolve(res.data)
    }
  },
  (error: AxiosError) => {
    console.log('err' + error)
    let msg = error.message
    if (msg == 'Network Error') {
      msg = '后端接口连接异常'
    } else if (msg.includes('timeout')) {
      msg = '系统接口请求超时'
    } else if (msg.includes('Request failed with status code')) {
      msg = '系统接口' + msg.slice(-3) + '异常'
    }
    message.error(msg)
    return Promise.reject(error)
  },
)

// 通用下载方法（POST + tansParams 表单编码 + blob，错误藏在 JSON blob 里二次解析）
export function download(
  url: string,
  params?: Record<string, unknown>,
  filename?: string,
  config?: AxiosRequestConfig,
) {
  const hide = message.loading('正在下载数据，请稍候', 0)
  return service
    .post(url, params, {
      transformRequest: [
        (p: unknown) => tansParams((p ?? {}) as Record<string, unknown>),
      ],
      headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
      responseType: 'blob',
      ...config,
    })
    .then(async (data: unknown) => {
      const blobData = data as Blob
      const isBlob = blobValidate(blobData)
      if (isBlob) {
        const blob = new Blob([blobData])
        saveAs(blob, filename)
      } else {
        const resText = await blobData.text()
        const rspObj = JSON.parse(resText)
        const errMsg = errorCode[rspObj.code] || rspObj.msg || errorCode['default']
        message.error(errMsg)
      }
      hide()
    })
    .catch((r) => {
      console.error(r)
      message.error('下载文件出现错误，请联系管理员！')
      hide()
    })
}

export default service
