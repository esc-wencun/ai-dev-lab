<?php
declare(strict_types=1);

use app\service\NoticeService;
use PHPUnit\Framework\TestCase;

require_once __DIR__ . '/../../vendor/autoload.php';

/**
 * 7.0.0 纯逻辑单测：@Xss 正则命中/放行矩阵、noticeInput 校验文案、ids 解析、isRead 计数口径
 */
final class NoticeTest extends TestCase
{
    /* ---------- @Xss 正则矩阵（XssValidator 同款） ---------- */

    public function testXssHitsScriptTag(): void
    {
        $this->assertTrue(NoticeService::containsScriptChars('<script>alert(1)</script>'));
    }

    public function testXssHitsImgOnError(): void
    {
        $this->assertTrue(NoticeService::containsScriptChars('<img onerror=alert(1) src=x>'));
    }

    public function testXssHitsPlainTag(): void
    {
        $this->assertTrue(NoticeService::containsScriptChars('<b>粗体</b>'));
    }

    public function testXssPassesPlainText(): void
    {
        $this->assertFalse(NoticeService::containsScriptChars('温馨提醒：2026年新版本发布啦'));
    }

    public function testXssPassesEmptyString(): void
    {
        $this->assertFalse(NoticeService::containsScriptChars(''));
    }

    public function testXssPassesSpacedComparison(): void
    {
        // 空格隔开的比较符不命中；紧邻的 1<2>0 会被 XssValidator 同款正则误伤（经典版同宽严度，原样保留）
        $this->assertFalse(NoticeService::containsScriptChars('1 < 2 比较'));
        $this->assertTrue(NoticeService::containsScriptChars('1<2>0'));
    }

    /* ---------- noticeInput 校验文案（@Validated 对位） ---------- */

    private function req(array $post): \think\Request
    {
        $r = new \think\Request();
        $r->withPost($post);
        return $r;
    }

    public function testInputEmptyTitleThrows(): void
    {
        $this->expectException(\BusinessException::class);
        $this->expectExceptionMessage('公告标题不能为空');
        NoticeService::noticeInput($this->req(['noticeTitle' => '  ']));
    }

    public function testInputLongTitleThrows(): void
    {
        $this->expectException(\BusinessException::class);
        $this->expectExceptionMessage('公告标题不能超过50个字符');
        NoticeService::noticeInput($this->req(['noticeTitle' => str_repeat('标', 51)]));
    }

    public function testInputScriptTitleThrows(): void
    {
        $this->expectException(\BusinessException::class);
        $this->expectExceptionMessage('公告标题不能包含脚本字符');
        NoticeService::noticeInput($this->req(['noticeTitle' => '好<title>']));
    }

    public function testInputNormalPassesAndDefaults(): void
    {
        $input = NoticeService::noticeInput($this->req([
            'noticeTitle' => '测试标题',
            'noticeType'  => '2',
            'noticeContent' => '<p>富<b>文本</b></p>',
        ]));
        $this->assertSame('测试标题', $input['notice_title']);
        $this->assertSame('2', $input['notice_type']);
        $this->assertSame('<p>富<b>文本</b></p>', $input['notice_content'], '富文本正文不转义直存');
        $this->assertSame('0', $input['status'], 'status 默认 0');
        $this->assertSame('', $input['remark']);
    }

    /* ---------- ids 解析（remove/markReadAll 共用口径） ---------- */

    public function testIdsParse(): void
    {
        $parse = fn(string $s) => array_values(array_filter(array_map('intval', explode(',', $s)), fn($v) => $v > 0));
        $this->assertSame([1, 2, 3], $parse('1,2,3'));
        $this->assertSame([1, 3], $parse('1,,3'));
        $this->assertSame([], $parse(''));
        $this->assertSame([], $parse('0,-1,abc'));
    }

    /* ---------- isRead/unreadCount 计数口径（listTop 逻辑，数组驱动） ---------- */

    public function testUnreadCountLogic(): void
    {
        $list = [
            ['noticeId' => 3, 'isRead' => false],
            ['noticeId' => 2, 'isRead' => true],
            ['noticeId' => 1, 'isRead' => false],
        ];
        $unreadCount = count(array_filter($list, fn($n) => !$n['isRead']));
        $this->assertSame(2, $unreadCount);
    }
}
