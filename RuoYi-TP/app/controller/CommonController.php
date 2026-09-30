<?php
declare(strict_types=1);

namespace app\controller;

use app\service\ExcelExportService;
use AjaxResult;
use think\Response;

/**
 * 通用控制器（对位经典版 CommonController 的 fileDownload + FileUploadUtils.upload）
 */
class CommonController extends \app\BaseController
{
    /** 下载目录（ExcelExportService 输出位置） */
    private const DOWNLOAD_DIR = 'download';

    /** 上传目录（对位 RuoYiConfig.getUploadPath() = runtime/upload） */
    private const UPLOAD_DIR = 'upload';

    /** 下载扩展名白名单 */
    private const ALLOW_EXTENSIONS = ['xlsx', 'xls', 'csv', 'pdf', 'zip', 'png', 'jpg', 'jpeg', 'gif'];

    /** 上传扩展名白名单（对位 MimeTypeUtils.DEFAULT_ALLOWED_EXTENSION 逐项抄） */
    private const UPLOAD_EXTENSIONS = [
        // 图片
        'bmp', 'gif', 'jpg', 'jpeg', 'png',
        // word excel powerpoint
        'doc', 'docx', 'xls', 'xlsx', 'ppt', 'pptx', 'txt',
        // 压缩文件
        'rar', 'zip', 'gz', 'bz2',
        // 视频格式
        'mp4', 'avi', 'rmvb',
        // pdf
        'pdf',
    ];

    /** 上传扩展名 → 响应 Content-Type（GET upload/:fileName 服务用） */
    private const MIME_TYPES = [
        'bmp' => 'image/bmp', 'gif' => 'image/gif', 'jpg' => 'image/jpeg',
        'jpeg' => 'image/jpeg', 'png' => 'image/png', 'pdf' => 'application/pdf',
        'txt' => 'text/plain', 'mp4' => 'video/mp4',
    ];

    /** GET /common/download?fileName=<uuid>_岗位数据.xlsx&delete=true */
    public function download(\think\Request $request): Response
    {
        $fileName = (string)$request->get('fileName', '');
        $delete = $request->get('delete', '') === 'true';

        if ($fileName === '' || str_contains($fileName, '..')) {
            throw new \BusinessException('非法文件名');
        }
        $ext = strtolower(pathinfo($fileName, PATHINFO_EXTENSION));
        if (!in_array($ext, self::ALLOW_EXTENSIONS, true)) {
            throw new \BusinessException('不支持的文件类型');
        }

        $path = ExcelExportService::downloadDir() . DIRECTORY_SEPARATOR . $fileName;
        if (!is_file($path)) {
            throw new \BusinessException('文件不存在');
        }

        $realName = ExcelExportService::resolveDownloadName($fileName);
        $content = (string)file_get_contents($path);
        if ($delete) {
            @unlink($path);
        }

        return Response::create($content)
            ->header([
                'Content-Type'        => 'application/octet-stream',
                'Content-Disposition' => 'attachment; filename="' . rawurlencode($realName) . '"',
                'Content-Length'      => (string)strlen($content),
                'Cache-Control'       => 'no-store',
            ]);
    }

    /** POST /common/upload：summernote 图片上传等（multipart 字段 file；对位 CommonController.uploadFile） */
    public function upload(\think\Request $request): Response
    {
        $file = $request->file('file');
        if ($file === null) {
            return AjaxResult::error('上传文件不能为空');
        }
        $originalName = (string)$file->getOriginalName();
        $ext = strtolower(pathinfo($originalName, PATHINFO_EXTENSION));
        if (!in_array($ext, self::UPLOAD_EXTENSIONS, true)) {
            return AjaxResult::error('上传文件格式不允许：' . $ext);
        }
        // 防目录穿越（对位 FileUploadUtils.checkAllowDownload 同款 ".." 检查）
        if (str_contains($originalName, '..')) {
            return AjaxResult::error('非法文件名');
        }

        $dir = runtime_path() . self::UPLOAD_DIR;
        if (!is_dir($dir)) {
            mkdir($dir, 0755, true);
        }
        // 文件名 <uuid>_<原始名>（对位 FileUploadUtils 编码策略，防中文/重名冲突）
        $newFileName = bin2hex(random_bytes(16)) . '_' . $originalName;
        $file->move($dir, $newFileName);

        $fileName = '/' . self::UPLOAD_DIR . '/' . $newFileName;
        return AjaxResult::success('操作成功', [
            'url'              => $request->domain() . $fileName,
            'fileName'         => $fileName,
            'newFileName'      => $newFileName,
            'originalFilename' => $originalName,
        ]);
    }

    /** GET /upload/{fileName}：上传文件静态服务（对位 Spring ResourcesConfig /profile/** 映射） */
    public function serveUpload(\think\Request $request, string $fileName): Response
    {
        if ($fileName === '' || str_contains($fileName, '..')) {
            throw new \BusinessException('非法文件名');
        }
        $path = runtime_path() . self::UPLOAD_DIR . DIRECTORY_SEPARATOR . $fileName;
        if (!is_file($path)) {
            throw new \BusinessException('文件不存在');
        }
        $ext = strtolower(pathinfo($fileName, PATHINFO_EXTENSION));
        $mime = self::MIME_TYPES[$ext] ?? 'application/octet-stream';
        $content = (string)file_get_contents($path);
        return Response::create($content)
            ->header([
                'Content-Type'   => $mime,
                'Content-Length' => (string)strlen($content),
                'Cache-Control'  => 'max-age=86400',
            ]);
    }
}
