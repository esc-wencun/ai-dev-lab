<?php
declare(strict_types=1);

namespace app\service;

use PhpOffice\PhpSpreadsheet\Spreadsheet;
use PhpOffice\PhpSpreadsheet\Writer\Xlsx;

/**
 * Excel 导出服务（对位经典版 ExcelUtil，列定义数组驱动，通用）
 *
 * 输出 <uuid>_<sheetName>.xlsx 到 runtime/download；返回文件名（前端随后 GET /common/download）。
 * 列定义：['name' => 列头, 'field' => 行键, 'numeric' => bool 数字格式, 'convert' => '0=正常,1=停用' 可选]
 */
final class ExcelExportService
{
    /**
     * @param array $rows 数据行（键 => 值）
     * @param array $columns 列定义（顺序即列序）
     * @param string $sheetName sheet 名（经典版=业务名，如「岗位数据」）
     * @return string 文件名 <uuid>_<sheetName>.xlsx
     */
    public static function export(array $rows, array $columns, string $sheetName): string
    {
        $spreadsheet = new Spreadsheet();
        $sheet = $spreadsheet->getActiveSheet();
        $sheet->setTitle($sheetName);

        // 列头（灰底居中，对位经典版 ExcelUtil 列头样式；实现从简仅此两样式）
        $col = 1;
        foreach ($columns as $column) {
            $cell = $sheet->getCell([$col, 1]);
            $cell->setValue($column['name']);
            $cell->getStyle()->getFont()->setBold(true);
            $cell->getStyle()->getFill()->setFillType(\PhpOffice\PhpSpreadsheet\Style\Fill::FILL_SOLID)
                ->getStartColor()->setRGB('EEEEEE');
            $col++;
        }

        // 数据行
        $rowNo = 2;
        foreach ($rows as $row) {
            $col = 1;
            foreach ($columns as $column) {
                $value = $row[$column['field']] ?? '';
                if (isset($column['convert']) && $column['convert'] !== '') {
                    $value = self::convertByExp((string)$value, $column['convert']);
                }
                $cell = $sheet->getCell([$col, $rowNo]);
                if (!empty($column['numeric']) && is_numeric($value)) {
                    $cell->setValue((float)$value);
                    $cell->getStyle()->getNumberFormat()->setFormatCode('0');
                } else {
                    $cell->setValue((string)$value);
                }
                $col++;
            }
            $rowNo++;
        }

        // 列宽自适应（简化：按列头长度）
        $col = 1;
        foreach ($columns as $column) {
            $sheet->getColumnDimensionByColumn($col)->setWidth(max(10, mb_strlen((string)$column['name']) * 3));
            $col++;
        }

        $fileName = bin2hex(random_bytes(8)) . '_' . $sheetName . '.xlsx';
        $dir = self::downloadDir();
        if (!is_dir($dir)) {
            mkdir($dir, 0755, true);
        }
        (new Xlsx($spreadsheet))->save($dir . DIRECTORY_SEPARATOR . $fileName);
        $spreadsheet->disconnectWorksheets();
        return $fileName;
    }

    /** 下载目录（兼容框架外调用：优先 runtime_path()，回退 __DIR__ 推导） */
    public static function downloadDir(): string
    {
        if (function_exists('runtime_path')) {
            return runtime_path() . 'download';
        }
        return dirname(__DIR__, 2) . DIRECTORY_SEPARATOR . 'runtime' . DIRECTORY_SEPARATOR . 'download';
    }

    /** readConverterExp 转换（"0=正常,1=停用"） */
    public static function convertByExp(string $value, string $exp): string
    {
        foreach (explode(',', $exp) as $pair) {
            [$k, $v] = array_pad(explode('=', $pair, 2), 2, '');
            if ($k === $value) {
                return $v;
            }
        }
        return $value;
    }

    /** 下载文件名解析（对位 CommonController：取第一个 "_" 之后 + timestamp 前缀） */
    public static function resolveDownloadName(string $fileName): string
    {
        $pos = strpos($fileName, '_');
        $real = $pos !== false ? substr($fileName, $pos + 1) : $fileName;
        return time() . $real;
    }

    /**
     * 导入解析（对位 ExcelUtil.importExcel）：首行表头名→列映射，按名取值；空行跳过。
     * $reverseExp：['用户性别' => '男=0,女=1,未知=2', ...] 反向转换表。
     * 返回 [ ['表头名' => 值, ...], ...]
     */
    public static function parse(string $filePath, array $reverseExp = []): array
    {
        $reader = \PhpOffice\PhpSpreadsheet\IOFactory::createReaderForFile($filePath);
        $sheet = $reader->load($filePath)->getActiveSheet();

        $highestColIdx = \PhpOffice\PhpSpreadsheet\Cell\Coordinate::columnIndexFromString($sheet->getHighestColumn());
        $highestRow = $sheet->getHighestRow();

        // 首行表头 → 列索引
        $headers = [];
        for ($c = 1; $c <= $highestColIdx; $c++) {
            $headers[$c] = trim((string)$sheet->getCell([$c, 1])->getValue());
        }

        $rows = [];
        for ($r = 2; $r <= $highestRow; $r++) {
            $row = [];
            $nonEmpty = false;
            for ($c = 1; $c <= $highestColIdx; $c++) {
                $name = $headers[$c];
                if ($name === '') {
                    continue;
                }
                // 手机号等文本列：FormattedValue 防科学计数/丢前导零
                $value = trim((string)$sheet->getCell([$c, $r])->getFormattedValue());
                if (isset($reverseExp[$name])) {
                    $value = self::reverseConvert($value, $reverseExp[$name]);
                }
                if ($value !== '') {
                    $nonEmpty = true;
                }
                $row[$name] = $value;
            }
            if ($nonEmpty) {
                $rows[] = $row;
            }
        }
        return $rows;
    }

    /** 反向转换（"男"→"0"；未匹配原样） */
    public static function reverseConvert(string $label, string $exp): string
    {
        foreach (explode(',', $exp) as $pair) {
            [$labelPart, $valuePart] = array_pad(explode('=', $pair, 2), 2, '');
            if ($labelPart === $label) {
                return $valuePart;
            }
        }
        return $label;
    }

    /** 空模板（IMPORT 列头空表；对位 importTemplateExcel），输出 <uuid>_<sheetName>.xlsx */
    public static function exportTemplate(array $columns, string $sheetName): string
    {
        return self::export([], $columns, $sheetName);
    }
}
