<?php
declare(strict_types=1);

use app\service\ExcelExportService;
use PHPUnit\Framework\TestCase;

require_once __DIR__ . '/../../vendor/autoload.php';

/**
 * 3.0.0 纯逻辑单测：Excel 导出转换 / 文件名解析 / Ztree 组装 / ids 解析
 * （ancestors 级联是 SQL 层 REPLACE 逻辑，端到端 DB 核对覆盖；这里测纯 PHP 部分）
 */
final class Module3Test extends TestCase
{
    /* ---------- ExcelExportService ---------- */

    public function testConvertByExpStatus(): void
    {
        $this->assertSame('正常', ExcelExportService::convertByExp('0', '0=正常,1=停用'));
        $this->assertSame('停用', ExcelExportService::convertByExp('1', '0=正常,1=停用'));
        $this->assertSame('9', ExcelExportService::convertByExp('9', '0=正常,1=停用'), '未匹配值原样返回');
    }

    public function testResolveDownloadNameStripsUuid(): void
    {
        // <uuid>_岗位数据.xlsx → <timestamp>岗位数据.xlsx
        $real = ExcelExportService::resolveDownloadName('a1b2c3d4e5f60718_岗位数据.xlsx');
        $this->assertMatchesRegularExpression('/^\d+岗位数据\.xlsx$/u', $real);
        $this->assertStringNotContainsString('a1b2c3d4e5f60718', $real);
    }

    public function testResolveDownloadNameWithoutUnderscore(): void
    {
        $real = ExcelExportService::resolveDownloadName('plain.xlsx');
        $this->assertMatchesRegularExpression('/^\d+plain\.xlsx$/u', $real);
    }

    /* ---------- 导出文件真实生成（临时目录，不污染共享库） ---------- */

    public function testExportGeneratesRealXlsx(): void
    {
        $rows = [
            ['post_id' => 1, 'post_code' => 'ceo', 'post_name' => '董事长', 'post_sort' => 1, 'status' => '0'],
            ['post_id' => 2, 'post_code' => 'se', 'post_name' => '项目经理', 'post_sort' => 2, 'status' => '1'],
        ];
        $columns = [
            ['name' => '岗位序号', 'field' => 'post_id', 'numeric' => true],
            ['name' => '岗位编码', 'field' => 'post_code'],
            ['name' => '岗位名称', 'field' => 'post_name'],
            ['name' => '岗位排序', 'field' => 'post_sort', 'numeric' => true],
            ['name' => '状态', 'field' => 'status', 'convert' => '0=正常,1=停用'],
        ];
        $fileName = ExcelExportService::export($rows, $columns, '岗位数据');
        $path = ExcelExportService::downloadDir() . DIRECTORY_SEPARATOR . $fileName;
        $this->assertFileExists($path);
        $this->assertSame('PK', substr((string)file_get_contents($path, false, null, 0, 2), 0, 2), 'xlsx 是 zip 容器（PK 魔数）');

        // 读回验证内容
        $reader = \PhpOffice\PhpSpreadsheet\IOFactory::createReaderForFile($path);
        $sheet = $reader->load($path)->getActiveSheet();
        $this->assertSame('岗位数据', $sheet->getTitle());
        $this->assertSame('岗位序号', $sheet->getCell([1, 1])->getValue());
        $this->assertSame('岗位编码', $sheet->getCell([2, 1])->getValue());
        $this->assertSame('正常', $sheet->getCell([5, 2])->getValue(), '状态 0 → 正常');
        $this->assertSame('停用', $sheet->getCell([5, 3])->getValue(), '状态 1 → 停用');
        unlink($path);
    }

    /* ---------- Ztree 结构（DeptService::selectDeptTreeData 的组装口径——excludeId 过滤逻辑等价复现） ---------- */

    public function testZtreeExcludeLogic(): void
    {
        // 模拟 excludeId 过滤口径：自身 + ancestors 含 excludeId 的后代都排除
        $rows = [
            ['dept_id' => 100, 'parent_id' => 0, 'ancestors' => '0', 'dept_name' => '总公司', 'order_num' => 0],
            ['dept_id' => 10, 'parent_id' => 100, 'ancestors' => '0,100', 'dept_name' => '研发部门', 'order_num' => 1],
            ['dept_id' => 101, 'parent_id' => 10, 'ancestors' => '0,100,10', 'dept_name' => '研发一组', 'order_num' => 1],
        ];
        $excludeId = 10;
        $kept = [];
        foreach ($rows as $row) {
            $id = (int)$row['dept_id'];
            if ($excludeId > 0 && ($id === $excludeId || str_contains(',' . $row['ancestors'] . ',', ',' . $excludeId . ','))) {
                continue;
            }
            $kept[] = $id;
        }
        $this->assertSame([100], $kept, '排除 10 自身与后代 101，仅剩根');
    }

    public function testZtreeAncestorsBoundaryNoFalseMatch(): void
    {
        // "0,10" 不得误匹配 ancestors 里的 "0,101"（spec 加固点）
        $ancestors = '0,101';
        $excludeId = 10;
        $hit = str_contains(',' . $ancestors . ',', ',' . $excludeId . ',');
        $this->assertFalse($hit, '"0,101" 不应命中 excludeId=10');
        $hit2 = str_contains(',' . '0,10,101' . ',', ',10,');
        $this->assertTrue($hit2, '真实链 0,10,101 应命中 excludeId=10');
    }

    /* ---------- ids 解析（PostService 删除入参口径） ---------- */

    public function testIdsParse(): void
    {
        $ids = array_filter(explode(',', '1, 2,,3'), fn($v) => $v !== '');
        $this->assertSame(['1', ' 2', '3'], array_values($ids));
        $this->assertSame(3, count(array_map('intval', $ids)));
    }
}
