<?php
declare(strict_types=1);

use app\service\ExcelExportService;
use app\service\UserService;
use PHPUnit\Framework\TestCase;
use PasswordService;

require_once __DIR__ . '/../../vendor/autoload.php';

/**
 * 4.0.0 纯逻辑单测：导入消息拼装 / md5 初始密码 / flag 合并 / 反向转换 / 真实 xlsx 模板+解析
 */
final class Module4Test extends TestCase
{
    /* ---------- 导入校验链（UserService::importUser 纯逻辑，不连库的分支用空行驱动） ---------- */

    public function testImportEmptyRowsThrows(): void
    {
        $this->expectException(\BusinessException::class);
        $this->expectExceptionMessage('导入用户数据不能为空！');
        UserService::importUser([], true, 'admin', ['user_id' => 1]);
    }

    public function testImportPasswordSchemeMatchesNoSalt(): void
    {
        // 导入例外：md5(loginName+初始密码) 不写 salt（空串拼接）——与登录链 md5(l+p+salt) 区分
        $this->assertSame(
            md5('impuser' . '123456'),
            PasswordService::encrypt('impuser', '123456', '')
        );
    }

    /* ---------- 反向转换（ExcelImport） ---------- */

    public function testReverseConvertSexStatus(): void
    {
        $this->assertSame('0', ExcelExportService::reverseConvert('男', '男=0,女=1,未知=2'));
        $this->assertSame('1', ExcelExportService::reverseConvert('女', '男=0,女=1,未知=2'));
        $this->assertSame('1', ExcelExportService::reverseConvert('停用', '正常=0,停用=1'));
        $this->assertSame('9', ExcelExportService::reverseConvert('9', '男=0,女=1,未知=2'), '未匹配原样');
    }

    /* ---------- 真实 xlsx：模板生成 + 表头乱序 + 解析读回 ---------- */

    private function makeXlsx(array $data, array $headers): string
    {
        $spreadsheet = new \PhpOffice\PhpSpreadsheet\Spreadsheet();
        $sheet = $spreadsheet->getActiveSheet();
        foreach ($headers as $i => $h) {
            $sheet->getCell([$i + 1, 1])->setValue($h);
        }
        foreach ($data as $r => $row) {
            foreach ($row as $c => $v) {
                $sheet->getCell([$c + 1, $r + 2])->setValue($v);
            }
        }
        $path = ExcelExportService::downloadDir() . DIRECTORY_SEPARATOR . 'm4test_' . bin2hex(random_bytes(4)) . '.xlsx';
        if (!is_dir(dirname($path))) {
            mkdir(dirname($path), 0755, true);
        }
        (new \PhpOffice\PhpSpreadsheet\Writer\Xlsx($spreadsheet))->save($path);
        $spreadsheet->disconnectWorksheets();
        return $path;
    }

    public function testExportTemplateGeneratesEmpty7ColSheet(): void
    {
        $columns = [
            ['name' => '部门编号'], ['name' => '登录名称'], ['name' => '用户名称'],
            ['name' => '用户邮箱'], ['name' => '手机号码'], ['name' => '用户性别'], ['name' => '账号状态'],
        ];
        $fileName = ExcelExportService::exportTemplate($columns, '用户数据');
        $path = ExcelExportService::downloadDir() . DIRECTORY_SEPARATOR . $fileName;
        $this->assertFileExists($path);
        $sheet = \PhpOffice\PhpSpreadsheet\IOFactory::createReaderForFile($path)->load($path)->getActiveSheet();
        $this->assertSame('部门编号', $sheet->getCell([1, 1])->getValue());
        $this->assertSame('账号状态', $sheet->getCell([7, 1])->getValue());
        $this->assertSame(1, $sheet->getHighestRow(), '模板只有表头一行');
        unlink($path);
    }

    public function testParseMatchesByHeaderNameAndReverseConverts(): void
    {
        // 表头乱序（登录名称在第 3 列）+ 性别/状态写中文 → parse 应按名匹配并反向转换
        $path = $this->makeXlsx(
            [['女', '张三', 'zs01', '13800001111', '正常']],
            ['用户性别', '用户名称', '登录名称', '手机号码', '账号状态']
        );
        $rows = ExcelExportService::parse($path, [
            '用户性别' => '男=0,女=1,未知=2',
            '账号状态' => '正常=0,停用=1',
        ]);
        unlink($path);
        $this->assertCount(1, $rows);
        $this->assertSame('zs01', $rows[0]['登录名称']);
        $this->assertSame('张三', $rows[0]['用户名称']);
        $this->assertSame('1', $rows[0]['用户性别'], '女→1');
        $this->assertSame('0', $rows[0]['账号状态'], '正常→0');
        $this->assertSame('13800001111', $rows[0]['手机号码'], '文本读取无科学计数');
    }

    public function testParseSkipsEmptyRows(): void
    {
        $path = $this->makeXlsx(
            [['张三'], []],  // 第二行全空
            ['用户名称']
        );
        $rows = ExcelExportService::parse($path);
        unlink($path);
        $this->assertCount(1, $rows);
    }

    /* ---------- flag 合并（PostService::selectPostsByUserId 的纯逻辑口径） ---------- */

    public function testFlagMergeLogic(): void
    {
        $posts = [
            ['postId' => 1, 'postName' => 'ceo', 'flag' => false],
            ['postId' => 2, 'postName' => 'se', 'flag' => false],
        ];
        $owned = [2];
        foreach ($posts as &$p) {
            if (in_array($p['postId'], $owned, true)) {
                $p['flag'] = true;
            }
        }
        $this->assertFalse($posts[0]['flag']);
        $this->assertTrue($posts[1]['flag']);
    }

    /* ---------- DictService::getLabel 纯逻辑口径（listByType 缓存 DB 级在端到端验） ---------- */

    public function testDictLabelDoubleKeysShape(): void
    {
        // 3.0.0 定版：dict:<type> 缓存行同时含驼峰与下划线键——getLabel 两种键都能命中
        $row = ['dictValue' => '0', 'dictLabel' => '正常', 'dict_value' => '0', 'dict_label' => '正常'];
        $hit = null;
        foreach ([$row] as $r) {
            if ((string)($r['dictValue'] ?? $r['dict_value'] ?? '') === '0') {
                $hit = (string)($r['dictLabel'] ?? $r['dict_label'] ?? '');
            }
        }
        $this->assertSame('正常', $hit);
    }
}
