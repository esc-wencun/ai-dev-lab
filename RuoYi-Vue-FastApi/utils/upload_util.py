"""
文件上传下载工具（对应java版FileUploadUtils / FileUtils）

规范：上传路径 {UPLOAD_PATH}/{yyyy/MM/dd}/{编码文件名}；扩展名白名单；防路径穿越
"""
import os
import re
import uuid as uuid_pkg
from datetime import datetime
from typing import List, Optional
from fastapi import UploadFile
from config.env import UploadConfig

# 允许的扩展名（对齐java版FileUploadUtils.DEFAULT_ALLOWED_EXTENSION）
DEFAULT_ALLOWED_EXTENSION = {
    # 图片
    "bmp", "gif", "jpg", "jpeg", "png",
    # word excel ppt html
    "doc", "docx", "xls", "xlsx", "ppt", "pptx", "html", "htm", "txt",
    # 压缩
    "rar", "zip", "gz", "bz2",
    # 视频
    "mp4", "avi", "rmvb",
    "pdf",
}

# 文件名最大长度（java版FileUploadUtils.DEFAULT_FILE_NAME_LENGTH=100）
MAX_FILE_NAME_LENGTH = 100


class FileUploadException(Exception):
    pass


def get_ext(file_name: str) -> str:
    """
    取扩展名（小写）
    """
    return file_name.rsplit('.', 1)[-1].lower() if '.' in file_name else ''


def validate_filename(file_name: str):
    """
    文件名长度校验（java版：上传文件名是最长100个字符）
    """
    if len(file_name) > MAX_FILE_NAME_LENGTH:
        raise FileUploadException(f'上传文件名是最长{MAX_FILE_NAME_LENGTH}个字符')


def validate_extension(file_name: str):
    """
    扩展名白名单校验（java版：上传文件扩展名是不允许的扩展名）
    """
    ext = get_ext(file_name)
    if ext not in DEFAULT_ALLOWED_EXTENSION:
        raise FileUploadException('上传文件扩展名是不允许的扩展名')


def _encode_file_name(original: str) -> str:
    """
    编码文件名（java版FilenameUtils编码语义：uuid替换原名保留扩展名，防中文/特殊字符）
    """
    ext = get_ext(original)
    date_dir = datetime.now().strftime('%Y%m%d')
    return f'{date_dir}/{uuid_pkg.uuid4().hex}.{ext}'


async def upload_one(file: UploadFile, base_path: str) -> dict:
    """
    保存单个上传文件
    :return: {fileName: /profile/yyyyMMdd/xxx.ext, newFileName, originalFilename}
    """
    original = file.filename or 'unknown'
    validate_filename(original)
    validate_extension(original)

    relative = _encode_file_name(original)
    # base_path 直接映射 /profile 前缀
    full_path = os.path.join(base_path, relative.replace('/', os.sep))
    os.makedirs(os.path.dirname(full_path), exist_ok=True)

    content = await file.read()
    with open(full_path, 'wb') as f:
        f.write(content)

    new_name = relative.rsplit('/', 1)[-1]
    return {
        'fileName': f'{UploadConfig.UPLOAD_PREFIX}/{relative}',
        'newFileName': new_name,
        'originalFilename': original,
    }


async def upload_files(files: List[UploadFile], base_path: str) -> List[dict]:
    """
    批量保存
    """
    results = []
    for f in files:
        results.append(await upload_one(f, base_path))
    return results


def is_allowed_download(file_name: str) -> bool:
    """
    下载文件名合法性（java版FileUtils.checkAllowDownload：防路径穿越+必须带合法扩展名）
    """
    if not file_name:
        return False
    if '..' in file_name or file_name.startswith('/'):
        return False
    return get_ext(file_name) in DEFAULT_ALLOWED_EXTENSION
