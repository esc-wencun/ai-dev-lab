// 下载插件 —— 对位 RuoYi-Vue3 src/plugins/download.js（$download），行为一致
// 与 request.ts 的 download() 区别：走裸 axios + GET + 手动 Bearer 头

import axios from 'axios'
import { message } from 'antd'
import { saveAs } from 'file-saver'
import { getToken } from '@/utils/auth'
import errorCode from '@/utils/errorCode'
import { blobValidate } from '@/utils/ruoyi'

const baseURL = import.meta.env.VITE_APP_BASE_API

// 手动全屏 loading（对位 ElLoading.service；antd 用 message.loading + Spin 兜底即可）
function showDownloadLoading(): () => void {
  return message.loading('正在下载数据，请稍候', 0)
}

export default {
  // 通用下载（/common/download，服务器生成文件）
  name(name: string, isDelete = true) {
    const url = baseURL + '/common/download?fileName=' + encodeURIComponent(name) + '&delete=' + isDelete
    axios({
      method: 'get',
      url,
      responseType: 'blob',
      headers: { Authorization: 'Bearer ' + getToken() },
    }).then((res) => {
      const isBlob = blobValidate(res.data as Blob)
      if (isBlob) {
        const blob = new Blob([res.data as Blob])
        this.saveAs(blob, decodeURIComponent(res.headers['download-filename'] as string))
      } else {
        this.printErrMsg(res.data as Blob)
      }
    })
  },

  // 资源下载（/common/download/resource，静态资源）
  resource(resource: string) {
    const url = baseURL + '/common/download/resource?resource=' + encodeURIComponent(resource)
    axios({
      method: 'get',
      url,
      responseType: 'blob',
      headers: { Authorization: 'Bearer ' + getToken() },
    }).then((res) => {
      const isBlob = blobValidate(res.data as Blob)
      if (isBlob) {
        const blob = new Blob([res.data as Blob])
        this.saveAs(blob, decodeURIComponent(res.headers['download-filename'] as string))
      } else {
        this.printErrMsg(res.data as Blob)
      }
    })
  },

  // zip 下载（代码生成等，带 loading）
  zip(url: string, name: string) {
    const fullUrl = baseURL + url
    const hide = showDownloadLoading()
    axios({
      method: 'get',
      url: fullUrl,
      responseType: 'blob',
      headers: { Authorization: 'Bearer ' + getToken() },
    })
      .then((res) => {
        const isBlob = blobValidate(res.data as Blob)
        if (isBlob) {
          const blob = new Blob([res.data as Blob], { type: 'application/zip' })
          this.saveAs(blob, name)
        } else {
          this.printErrMsg(res.data as Blob)
        }
        hide()
      })
      .catch((r) => {
        console.error(r)
        message.error('下载文件出现错误，请联系管理员！')
        hide()
      })
  },

  saveAs(text: Blob | string, name: string, opts?: { autoBom?: boolean }) {
    // file-saver 的 FileSaverOptions 要求 autoBom 必填；转发时兜底 true（与基准默认行为一致）
    saveAs(text, name, { autoBom: true, ...opts })
  },

  // blob 里的 JSON 错误体解析
  async printErrMsg(data: Blob) {
    const resText = await data.text()
    const rspObj = JSON.parse(resText)
    const errMsg = errorCode[rspObj.code] || rspObj.msg || errorCode['default']
    message.error(errMsg)
  },
}
