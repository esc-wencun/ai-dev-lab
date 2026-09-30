<?php
declare(strict_types=1);

namespace app\service;

/**
 * 服务监控采集（对位经典版 Server/oshi；三档降级：可得/条件可得/「—」——deviations #11 定稿方案）
 *
 * 零依赖：只内建函数，不引 exec；每项 try-catch 兜底 null（页面显「—」）。
 */
class ServerInfoService
{
    public static function collect(): array
    {
        return [
            'cpu'   => self::cpu(),
            'mem'   => self::memory(),
            'sys'   => self::sysInfo(),
            'php'   => self::phpInfo(),
            'disks' => self::disks(),
        ];
    }

    private static function cpu(): array
    {
        $count = null;
        try {
            if (PHP_OS_FAMILY === 'Windows') {
                $count = (int)getenv('NUMBER_OF_PROCESSORS') ?: null;
            } else {
                $cpuinfo = @file_get_contents('/proc/cpuinfo');
                if ($cpuinfo !== false) {
                    $count = preg_match_all('/^processor\s*:/m', $cpuinfo) ?: null;
                }
            }
        } catch (\Throwable) {
        }
        return ['count' => $count, 'user' => null, 'sys' => null, 'idle' => null];
    }

    private static function memory(): array
    {
        $total = $used = $free = $usageRate = null;
        try {
            if (PHP_OS_FAMILY === 'Windows') {
                // wmic/COM 不可依赖——Windows 物理内存总量从环境变量不可得，显示「—」；
                // PHP 进程内存可得
            } else {
                $meminfo = @file_get_contents('/proc/meminfo');
                if ($meminfo !== false && preg_match('/MemTotal:\s+(\d+) kB/', $meminfo, $m1) && preg_match('/MemAvailable:\s+(\d+) kB/', $meminfo, $m2)) {
                    $total = (int)$m1[1] * 1024;
                    $free = (int)$m2[1] * 1024;
                    $used = $total - $free;
                    $usageRate = $total > 0 ? round($used / $total * 100, 1) . '%' : null;
                }
            }
        } catch (\Throwable) {
        }
        return [
            'total'     => $total,
            'used'      => $used,
            'free'      => $free,
            'usageRate' => $usageRate,
            'phpUsage'  => memory_get_usage(true),
            'phpLimit'  => self::iniBytes((string)get_cfg_var('memory_limit')),
        ];
    }

    private static function sysInfo(): array
    {
        return [
            'name' => php_uname('n'),
            'ip'   => request()?->ip() ?? '—',
            'os'   => PHP_OS_FAMILY . ' / ' . php_uname('s') . ' ' . php_uname('r'),
            'arch' => php_uname('m'),
        ];
    }

    private static function phpInfo(): array
    {
        return [
            'version'      => PHP_VERSION,
            'sapi'         => PHP_SAPI,
            'memoryUsage'  => memory_get_usage(true),
            'memoryLimit'  => self::iniBytes((string)get_cfg_var('memory_limit')),
            'startTime'    => null, // built-in server 单进程起时取不上——「—」
        ];
    }

    /** Windows 逐盘符（C~Z 存在即采）；Linux 挂载点根分区 */
    private static function disks(): array
    {
        $disks = [];
        try {
            if (PHP_OS_FAMILY === 'Windows') {
                foreach (range('C', 'Z') as $letter) {
                    $dir = $letter . ':\\';
                    $total = @disk_total_space($dir);
                    if ($total === false || $total <= 0) {
                        continue;
                    }
                    $free = @disk_free_space($dir) ?: 0;
                    $used = $total - $free;
                    $disks[] = self::diskRow($letter . ':', 'NTFS', '本地磁盘', $total, $free, $used);
                }
            } else {
                $total = @disk_total_space('/');
                if ($total !== false && $total > 0) {
                    $free = @disk_free_space('/') ?: 0;
                    $used = $total - $free;
                    $disks[] = self::diskRow('/', 'ext4', '根分区', $total, $free, $used);
                }
            }
        } catch (\Throwable) {
        }
        return $disks;
    }

    private static function diskRow(string $dir, string $type, string $typeName, float $total, float $free, float $used): array
    {
        return [
            'dir'       => $dir,
            'type'      => $type,
            'typeName'  => $typeName,
            'total'     => self::formatBytes($total),
            'free'      => self::formatBytes($free),
            'used'      => self::formatBytes($used),
            'usageRate' => $total > 0 ? round($used / $total * 100, 1) : 0,
        ];
    }

    private static function iniBytes(string $v): ?int
    {
        if ($v === '' || $v === '-1') {
            return null;
        }
        $unit = strtolower(substr($v, -1));
        $n = (int)$v;
        return match ($unit) {
            'g' => $n * 1024 ** 3,
            'm' => $n * 1024 ** 2,
            'k' => $n * 1024,
            default => (int)$v,
        };
    }

    private static function formatBytes(float $bytes): string
    {
        $units = ['B', 'KB', 'MB', 'GB', 'TB'];
        $i = 0;
        while ($bytes >= 1024 && $i < count($units) - 1) {
            $bytes /= 1024;
            $i++;
        }
        return round($bytes, 1) . ' ' . $units[$i];
    }
}
