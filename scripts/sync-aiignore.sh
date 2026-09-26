#!/usr/bin/env bash
# 将根目录 .aiignore（唯一权威源）同步到各 AI 编码工具：
#   1) 原样复制 → .cursorignore / .windsurfignore / .geminiignore
#   2) 转换生成 → .claude/settings.json 的 permissions.deny Read 规则
#      （保留 settings.json 中已有的其他配置，只替换本清单产生的 Read(...) 条目）
# 用法：bash scripts/sync-aiignore.sh   （在任意目录执行均可，自动定位仓库根）
# 清单语法：gitignore。行尾带 / 视为目录（生成 **/ 前缀匹配 + /** 后缀）；
# 不带 / 的裸文件名按 gitignore 语义匹配任意层级；不支持取反（!）行。
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
SRC="$ROOT/.aiignore"
[ -f "$SRC" ] || { echo "错误：找不到 $SRC" >&2; exit 1; }

# ---- 1) gitignore 语法的工具副本 ----
for f in .cursorignore .windsurfignore .geminiignore; do
  cp "$SRC" "$ROOT/$f"
  echo "已更新 $f"
done

# ---- 2) Claude Code deny 规则 ----
PY=python
command -v "$PY" >/dev/null 2>&1 || PY=python3

"$PY" - "$SRC" "$ROOT/.claude/settings.json" <<'PY'
import json
import os
import sys

src, dst = sys.argv[1], sys.argv[2]
patterns = []
for raw in open(src, encoding="utf-8"):
    line = raw.strip()
    if not line or line.startswith("#"):
        continue
    if line.startswith("!"):
        print(f"警告：取反行不被支持，已跳过：{line}", file=sys.stderr)
        continue
    is_dir = line.endswith("/")
    p = line.rstrip("/")
    if p.startswith("**/"):
        body = p
    elif p.startswith("/"):
        body = p.lstrip("/")
    elif "/" in p:
        body = p
    else:
        body = "**/" + p
    patterns.append(f"Read(./{body}/**)" if is_dir else f"Read(./{body})")

os.makedirs(os.path.dirname(dst), exist_ok=True)
try:
    with open(dst, encoding="utf-8") as f:
        settings = json.load(f)
except FileNotFoundError:
    settings = {}

deny = settings.setdefault("permissions", {}).setdefault("deny", [])
deny[:] = [r for r in deny if not (isinstance(r, str) and r.startswith("Read("))]
deny.extend(patterns)

with open(dst, "w", encoding="utf-8") as f:
    json.dump(settings, f, ensure_ascii=False, indent=2)
    f.write("\n")
print(f"已更新 .claude/settings.json（{len(patterns)} 条 Read deny 规则）")
PY
