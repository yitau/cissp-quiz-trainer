# v0.1 执行记录

时区：Asia/Shanghai（UTC+08:00）。目标状态：进行中。当前阶段：0。

| 阶段 | 状态 | 关联提交 |
| --- | --- | --- |
| 0 基线与业务规则 | 进行中 | — |
| 1 导入与学习 | 未开始 | — |
| 2 考试与历史 | 未开始 | — |
| 3 复习与统计 | 未开始 | — |
| 4 备份与日常使用 | 未开始 | — |
| 5 最终验收与交付 | 未开始 | — |

## 阶段 0 — 基线与业务规则

- 2026-10-01（Asia/Shanghai）：读取 AGENTS.md、README.md、requirements-v0.3.md、实际骨架和 CI。初始 main@9fb84ed，`git status --short` 为空，无用户改动。
- 建立持续 Goal，写入 implementation-plan.md。本仓库仅初始 App/HealthCheck 与 Vue 占位页；业务包只有 doc.go。
- `node --version` = v24.18.0，`npm --version` = 12.0.1；仍需实际兼容性验证。
- PATH 无 Go/Wails；找到 `%USERPROFILE%\\go\\bin\\wails.exe`，`version` = v2.15.0；WebView2 目录有 153/154 安装版本。Go 继续定位。
- `git switch -c feature/mvp-v0.1` 沙箱内被 .git 写权限阻止；提权执行遇跨用户 ownership 检查，将只为本次命令指定当前仓库 safe.directory。
- 部分安装目录只读检查遇沙箱拒绝，申请只读访问；未重复安装工具。
- 下一步：找到 Go，下载依赖、骨架检查，记录基线后提交阶段 0；继续阶段 1。
- 2026-10-01 16:37 +08:00：Go 已在 `%LOCALAPPDATA%\Programs\go\bin` 找到（go1.27.1）；Wails doctor 确认 Windows 11 amd64 / WebView2 154。未安装重复工具。
- `npm install --cache ../tmp/npm-cache` 成功；type-check 通过。普通构建受 esbuild 上级目录读取权限阻碍，提权 `npm run build` 成功（Vite 7.3.6）。Node 24.18.0 本次骨架兼容性已验证。
- 初次 Go vet/test 因缺 dist 与完整 go.sum 失败；随后 Wails 构建补齐模块图与前端产物，`gofmt -w .`、`go vet ./...`、`go test ./...` 通过（骨架无测试），`wails build -clean` 成功产出 EXE。尚未运行界面。
- Wails 构建解析实际依赖后将 go.mod 最低版本升至 1.25.0；保留工具实际计算的最低版本并同步版本说明，开发工具链仍为 1.27.x，不通过降依赖伪装 1.23 兼容。
- 阶段 0 已完成：计划/规则、锁文件、构建忽略规则。下一步阶段 1 进行中：实现 importer/domain、迁移/repository、学习服务与真实 UI。
