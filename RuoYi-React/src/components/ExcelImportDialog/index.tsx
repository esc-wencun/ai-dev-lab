// ExcelImportDialog —— 对位基准 components/ExcelImportDialog/index.vue（用户导入弹窗）
// 行为对位：
//   - 拖拽/点击上传区域（仅 xls/xlsx，maxCount 1，不自动上传，点「确定」才提交）
//   - updateSupport 勾选框（每次打开重置为 false），提交时按 `?updateSupport=0|1` 语义传参
//   - 「下载模板」链接（templateAction 传入时显示）：走 download()（POST blob），对位基准 proxy.download
//   - 上传成功：关弹窗 → Modal 弹 response.msg（导入结果）→ 触发 onSuccess → 父页刷新
//   - 未选文件/后缀不符提交：msgError 文案逐字对齐「请选择后缀为 “xls”或“xlsx”的文件。」
// 差异：el-upload 拖拽区以 antd Upload.Dragger 等价实现；上传走 importUser()（multipart + Bearer）；
//      基准经 ref.expose open() 打开，React 惯例改为父页受控 open/onClose props（语义等价）

import { useEffect, useState } from 'react'
import { Checkbox, Modal, Typography, Upload, message } from 'antd'
import { InboxOutlined } from '@ant-design/icons'
import type { UploadFile } from 'antd'
import { importUser } from '@/api/system/user'
import { download } from '@/utils/request'

interface ExcelImportDialogProps {
  /** 是否显示（父页受控，对位基准 ref.open() 的打开语义） */
  open: boolean
  /** 关闭回调（父页置回 open=false） */
  onClose: () => void
  /** 对话框标题，默认「数据导入」 */
  title?: string
  /** 上传接口（必传，对位 action） */
  action: string
  /** 模板下载接口（传入则显示「下载模板」链接） */
  templateAction?: string
  /** 模板文件名前缀，默认 template */
  templateFileName?: string
  /** 覆盖更新勾选框说明文字，默认「是否更新已经存在的数据」 */
  updateSupportLabel?: string
  /** 上传成功回调（对位 @success，父页刷新列表） */
  onSuccess?: () => void
}

export default function ExcelImportDialog({
  open,
  onClose,
  title = '数据导入',
  action,
  templateAction,
  templateFileName = 'template',
  updateSupportLabel = '是否更新已经存在的数据',
  onSuccess,
}: ExcelImportDialogProps) {
  const [fileList, setFileList] = useState<UploadFile[]>([])
  const [uploading, setUploading] = useState(false)
  const [updateSupport, setUpdateSupport] = useState(false)

  // 每次打开重置（对位基准 open()：updateSupport=false、清空已选文件）
  useEffect(() => {
    if (open) {
      setUpdateSupport(false)
      setUploading(false)
      setFileList([])
    }
  }, [open])

  // 下载模板（对位基准 proxy.download(templateAction, {}, fileName_date.xlsx)）
  const handleDownloadTemplate = () => {
    if (!templateAction) return
    download(templateAction, {}, `${templateFileName}_${Date.now()}.xlsx`)
  }

  // 提交上传（对位基准 handleSubmit：先校验后缀，再 multipart 提交）
  const handleSubmit = async () => {
    const file = fileList[0]?.originFileObj
    const name = file?.name?.toLowerCase() || ''
    if (!file || (!name.endsWith('.xls') && !name.endsWith('.xlsx'))) {
      message.error('请选择后缀为 “xls”或“xlsx”的文件。')
      return
    }
    setUploading(true)
    try {
      const res = (await importUser({ file, updateSupport: updateSupport ? 1 : 0, action })) as unknown as {
        msg?: string
      }
      onClose()
      // 导入结果确认框（对位基准 $alert(response.msg, '导入结果', dangerouslyUseHTMLString)；
      // 纯文本渲染，导入明细行仍完整展示）
      Modal.info({
        title: '导入结果',
        content: (
          <div style={{ maxHeight: '60vh', overflow: 'auto' }}>
            <Typography.Text>{res.msg || '操作成功'}</Typography.Text>
          </div>
        ),
        okText: '确定',
      })
      onSuccess?.()
    } catch {
      // 失败时 request 拦截器已弹错误提示，保持弹窗打开供重选文件（对位基准失败不关弹窗）
    } finally {
      setUploading(false)
    }
  }

  return (
    <Modal
      title={title}
      open={open}
      width={400}
      confirmLoading={uploading}
      onOk={() => void handleSubmit()}
      onCancel={onClose}
      destroyOnHidden
    >
      <Upload.Dragger
        accept=".xlsx, .xls"
        maxCount={1}
        fileList={fileList}
        beforeUpload={() => false}
        onChange={({ fileList: next }) => setFileList(next.slice(-1))}
        disabled={uploading}
      >
        <p className="ant-upload-drag-icon">
          <InboxOutlined />
        </p>
        <p className="ant-upload-text">
          将文件拖到此处，或<em>点击上传</em>
        </p>
      </Upload.Dragger>
      <div style={{ marginTop: 8, textAlign: 'center' }}>
        <div>
          <Checkbox checked={updateSupport} onChange={(e) => setUpdateSupport(e.target.checked)}>
            {updateSupportLabel}
          </Checkbox>
        </div>
        <span style={{ fontSize: 12 }}>仅允许导入xls、xlsx格式文件。</span>
        {templateAction && (
          <a onClick={handleDownloadTemplate} style={{ fontSize: 12, marginLeft: 4 }}>
            下载模板
          </a>
        )}
      </div>
    </Modal>
  )
}
