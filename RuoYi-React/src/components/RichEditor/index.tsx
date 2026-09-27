// RichEditor —— 对位基准 components/Editor（Quill 2 直封，HTML 存储）
// 图片上传/粘贴上传走 /common/upload（Bearer、file 字段、插入 VITE_APP_BASE_API + fileName）

import { useEffect, useRef } from 'react'
import Quill from 'quill'
import 'quill/dist/quill.snow.css'
import axios from 'axios'
import { message } from 'antd'
import { getToken } from '@/utils/auth'

interface RichEditorProps {
  value?: string
  onChange?: (html: string) => void
  height?: number
  minHeight?: number
  readOnly?: boolean
  fileSize?: number // MB
  placeholder?: string
}

// 工具栏对齐基准 options 清单
const TOOLBAR = [
  [{ header: [1, 2, 3, 4, 5, 6, false] }, { font: [] }],
  [{ list: 'ordered' }, { list: 'bullet' }],
  ['bold', 'italic', 'underline', 'strike'],
  ['blockquote', 'code-block'],
  [{ color: [] }, { background: [] }],
  [{ align: [] }, { indent: '-1' }, { indent: '+1' }],
  ['link', 'image', 'video'],
  ['clean'],
]

// 图片上传（multipart /common/upload → 返回 fileName → 插入 baseUrl+fileName）
async function uploadImage(file: File, fileSize: number): Promise<string> {
  if (!['image/jpeg', 'image/png', 'image/svg+xml'].includes(file.type)) {
    throw new Error('图片格式错误，请上传 jpeg/jpg/png/svg 格式图片')
  }
  if (file.size > fileSize * 1024 * 1024) {
    throw new Error(`上传图片大小不能超过 ${fileSize}MB!`)
  }
  const form = new FormData()
  form.append('file', file)
  const res = await axios.post(import.meta.env.VITE_APP_BASE_API + '/common/upload', form, {
    headers: { Authorization: 'Bearer ' + getToken() },
  })
  const rsp = res.data
  if (rsp.code === 200) return import.meta.env.VITE_APP_BASE_API + rsp.fileName
  throw new Error(rsp.msg || '上传失败')
}

export default function RichEditor({ value, onChange, height = 260, minHeight, readOnly = false, fileSize = 5, placeholder }: RichEditorProps) {
  const hostRef = useRef<HTMLDivElement>(null)
  const quillRef = useRef<Quill | null>(null)
  // onChange 触发中标志（ setContent 会再触发 text-change，防回环）
  const silentRef = useRef(false)

  useEffect(() => {
    if (!hostRef.current || quillRef.current) return
    const quill = new Quill(hostRef.current, {
      theme: 'snow',
      placeholder,
      readOnly,
      modules: { toolbar: { container: TOOLBAR, handlers: { image: imageHandler } } },
    })
    quillRef.current = quill

    function imageHandler(this: { quill: Quill }) {
      const input = document.createElement('input')
      input.type = 'file'
      input.accept = 'image/*'
      input.click()
      input.onchange = async () => {
        const file = input.files?.[0]
        if (!file) return
        const hide = message.loading('上传中...', 0)
        try {
          const url = await uploadImage(file, fileSize)
          const range = quill.getSelection(true)
          quill.insertEmbed(range.index, 'image', url, 'user')
          quill.setSelection(range.index + 1)
        } catch (e) {
          message.error((e as Error).message)
        } finally {
          hide()
        }
      }
    }

    quill.on('text-change', () => {
      if (silentRef.current) return
      onChange?.(quill.root.innerHTML)
    })

    // 粘贴图片上传（对位基准 handlePasteCapture）
    quill.root.addEventListener('paste', onPaste)
    function onPaste(e: ClipboardEvent) {
      const items = e.clipboardData?.items
      if (!items) return
      for (const item of items) {
        if (item.type.indexOf('image') === 0) {
          const file = item.getAsFile()
          if (!file) continue
          e.preventDefault()
          void uploadImage(file, fileSize)
            .then((url) => {
              const range = quill.getSelection(true)
              quill.insertEmbed(range.index, 'image', url, 'user')
              quill.setSelection(range.index + 1)
            })
            .catch((err) => message.error(err.message))
        }
      }
    }
    return () => { quill.off('text-change'); quill.root.removeEventListener('paste', onPaste) }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  // 外部 value 回显（立即同步 + 空值兜底 <p></p>——对位基准 watch modelValue immediate）
  useEffect(() => {
    const quill = quillRef.current
    if (!quill) return
    const html = value || '<p></p>'
    if (quill.root.innerHTML !== html) {
      silentRef.current = true
      quill.root.innerHTML = html
      silentRef.current = false
    }
  }, [value])

  return (
    <div style={{ height, minHeight: minHeight ?? height }}>
      <div ref={hostRef} style={{ height: '100%' }} />
    </div>
  )
}
