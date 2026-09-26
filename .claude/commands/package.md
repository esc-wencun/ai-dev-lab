---
description: 打包服务端(Java)与客户端(Vue3)到 package/<当天日期>/ 目录，版本号为当前日期
allowed-tools: Bash(package.bat:*), Bash(cmd:*), Bash(ls:*)
---

# 打包服务端 + 客户端

实际打包逻辑全部在仓库根目录的 [package.bat](../../package.bat)（双击也可直接运行），本命令只负责调用它并汇报结果。

## 执行步骤

1. 在仓库根目录执行（Git Bash 下直接调用脚本本体）：

   ```
   ./package.bat
   ```

   注意：不要写 `cmd /c package.bat`——Git Bash 会把 `/c` 做 MSYS 路径转换（当成 `C:\`），导致脚本未被执行却返回 0，是假成功。

2. 检查退出码：
   - **0（成功）**：读取脚本输出中的产物目录，用 `ls -la` 列出 `package/<日期>/` 下的产物与大小，简短汇报版本号和产物绝对路径即可。脚本已打印步骤进度，不要复述全过程。
   - **非 0（失败）**：原样报告脚本输出中的错误信息，定位失败步骤（mvn / npm / 压缩），不要擅自改脚本或源码。

## 约束

- 打包范围固定为 **Java 版服务端 + Vue3 前端**（Python/Go 版无打包产物）。
- 只调用脚本，不修改源码和配置；需要调整打包行为时改 package.bat 本身。
- 打包过程中不执行 `git commit` / `git push`。
