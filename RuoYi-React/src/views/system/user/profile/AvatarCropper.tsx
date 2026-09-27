// AvatarCropper —— 头像裁剪上传（对位基准 userAvatar.vue：react-cropper 200×200 + multipart）

import { useRef, useState } from 'react'
import { Button, Modal, Space, Typography, message } from 'antd'
import { UploadOutlined } from '@ant-design/icons'
import Cropper from 'react-cropper'
import 'cropperjs/dist/cropper.css'
import { uploadAvatar } from '@/api/system/user'

interface ReactCropperElement extends HTMLImageElement {
  cropper?: Cropper
}

interface AvatarCropperProps {
  open: boolean
  imgSrc?: string
  onClose: () => void
  onUploaded: (url: string) => void
}

export default function AvatarCropper({ open, onClose, onUploaded }: AvatarCropperProps) {
  const cropperRef = useRef<ReactCropperElement | null>(null)
  const [fileInfo, setFileInfo] = useState<{ name: string } | null>(null)
  const [preview, setPreview] = useState('')
  const [saving, setSaving] = useState(false)

  const pickFile = () => {
    const input = document.createElement('input')
    input.type = 'file'
    input.accept = 'image/*'
    input.onchange = () => {
      const file = input.files?.[0]
      if (!file) return
      if (!file.type.startsWith('image/')) {
        void message.error('请选择图片文件')
        return
      }
      setFileInfo({ name: file.name })
      const reader = new FileReader()
      reader.onload = () => setPreview(String(reader.result))
      reader.readAsDataURL(file)
    }
    input.click()
  }

  const handleUpload = async () => {
    const cropper = cropperRef.current?.cropper
    if (!cropper || !fileInfo) {
      void message.error('请先选择图片')
      return
    }
    setSaving(true)
    try {
      const canvas = cropper.getCroppedCanvas({ width: 200, height: 200 })
      const blob = await new Promise<Blob | null>((resolve) => canvas.toBlob(resolve, 'image/png'))
      if (!blob) throw new Error('裁剪失败')
      const form = new FormData()
      form.append('avatarfile', blob, fileInfo.name || 'avatar')
      const res = (await uploadAvatar(form)) as unknown as { imgUrl: string }
      onUploaded(import.meta.env.VITE_APP_BASE_API + res.imgUrl)
      message.success('头像上传成功')
      onClose()
    } catch (e) {
      void message.error((e as Error).message || '上传失败')
    } finally {
      setSaving(false)
    }
  }

  return (
    <Modal title="修改头像" open={open} width={760} onCancel={onClose} destroyOnHidden
      footer={
        <Space>
          <Button onClick={onClose}>取消</Button>
          <Button type="primary" loading={saving} onClick={() => void handleUpload()}>上传</Button>
        </Space>
      }
    >
      <div style={{ display: 'flex', gap: 16 }}>
        <div style={{ width: 380 }}>
          <Button icon={<UploadOutlined />} onClick={pickFile} style={{ marginBottom: 8 }}>选择图片</Button>
          {preview ? (
            <Cropper
              ref={cropperRef}
              src={preview}
              style={{ height: 300, width: '100%' }}
              aspectRatio={1}
              viewMode={1}
              guides={false}
              preview=".avatar-preview"
            />
          ) : (
            <div style={{ height: 300, display: 'flex', alignItems: 'center', justifyContent: 'center', background: '#fafafa' }}>
              未选择图片
            </div>
          )}
        </div>
        <div style={{ flex: 1 }}>
          <div className="avatar-preview" style={{ width: 200, height: 200, overflow: 'hidden', borderRadius: 4 }} />
          <Typography.Text type="secondary" style={{ display: 'block', marginTop: 8 }}>
            预览（200×200，裁剪后上传）
          </Typography.Text>
        </div>
      </div>
    </Modal>
  )
}
