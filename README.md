# CISSP Quiz Trainer

Windows 11 x64 本地 CISSP 学习工具：JSON 导入 → 学习/考试 → 解析 → 错题/收藏 → 统计 → 完整备份恢复。

## 使用

本地构建产物在 `build/bin/CISSPQuizTrainer.exe` 和 `build/bin/CISSPQuizTrainer-win-x64.zip`。解压 ZIP，双击 EXE，在“题库与导入”选择同目录 `demo-questions.json`，确认导入即可练习。演示题为原创功能样例，不是官方真题。

- [中文使用说明与人工验收](docs/user-guide.md)
- [实施计划及业务规则](docs/implementation-plan.md)
- [执行记录与验证证据](docs/implementation-progress.md)
- [需求基线 v0.3](docs/requirements-v0.3.md)
- [知识学习模块需求规范 v0.4（目标 v0.2，尚未实现）](docs/requirements-v0.4.md)
- [知识学习开发 GOAL 与验收矩阵](docs/knowledge-learning-v0.4-goals.md)
- [Lesson JSON Schema v1.0](docs/schemas/lesson-v1.schema.json)
- [Week 01 Day 01 Lesson 示例 JSON](samples/cissp-week01-day01-lesson.json)
- [开发约束](AGENTS.md)

默认数据目录 `%LOCALAPPDATA%\CISSPQuizTrainer\data`，不写入 EXE 目录；通过 `CISSP_QUIZ_DATA_DIR` 可指定隔离目录。完全离线、无账号、无遥测。最终用户只需要 Windows 11 x64 和 WebView2 Runtime。

## 知识学习模块（v0.2 开发计划，当前版本尚未实现）

已在 `feature/knowledge-learning-v0.2` 分支定义独立“知识学习”模块：Lesson JSON → 阅读八个基础知识点及综合案例 → 理解自检与持久化学习进度 → 关联现有题库刷题。**这些功能目前只是需求和开发任务，不能在 v0.1.2 中直接使用。**开发与验收按照 v0.4 需求、Lesson Schema、GOAL 清单执行。现有 Question JSON v1.0 契约不变。

## 开发环境

Go 1.27.x、Wails CLI v2.15.0、Node.js 22.x 基线、npm、WebView2。前端 Vue 3 + TypeScript + Pinia，数据库为 database/sql + modernc.org/sqlite（纯 Go）。

### Version declarations

`go.mod` 声明 `go 1.25.0`，与实际 Wails v2.15.0 依赖最低版本一致；这是模块最低语言/工具要求，开发和 CI 仍选择 Go 1.27.x。初始骨架的 `go 1.23.0` 在解析实际依赖后由 Go 工具纠正，不能据此声称整个依赖图兼容 Go 1.23。

本机实际验证 Go 1.27.1、Wails 2.15.0、Node 24.18.0 / npm 12.0.1；2026-10-01 远端 Windows CI 的 Go 1.27.x、Wails 2.15.0、Node 22 全套检查和打包通过（[运行记录](https://github.com/yitau/cissp-quiz-trainer/actions/runs/36867447204)）。真实窗口已验证学习、续答、考试交卷、结果筛选、错题重练及恢复数据核对；原生文件对话框等验收边界见执行记录。

### 干净检出首次构建

```powershell
go install github.com/wailsapp/wails/v2/cmd/wails@v2.15.0
go mod download
npm --prefix frontend ci
wails build -clean
```

首次必须先通过 Wails 建立嵌入目录、生成 `frontend/wailsjs` 并编译前端。直接从没有 bindings 的检出执行前端类型检查会失败；无需也不得手工生成替代 bindings。

普通开发运行 `wails dev`。后续修改的必需检查：

```powershell
# 若更改 Go 绑定表面，先 wails build -clean 更新生成文件。
npm --prefix frontend run type-check
npm --prefix frontend run build
gofmt -w .
go vet ./...
go test ./...
wails build -clean
./scripts/package.ps1
```

Go 不在 PATH 时先检查已有的 `%LOCALAPPDATA%\Programs\go\bin` 和 `%USERPROFILE%\go\bin`，不要重复安装。本机沙箱对工具目录和 esbuild 上级路径有限制，受影响命令需在有相应权限的终端运行。

## 实现与数据规则

Vue → frontend services → Wails 薄适配器 → service → repository 接口 → SQLite。SQL 只在 repository/database；migration 管理数据库版本。题目不可变 JSON 与独立会话/答案表分开保存；导入和交卷事务写入；恢复先验证并安全备份，再事务替换。

学习提交后显示解析，考试交卷前服务端不返回解答；草稿即时保存、重启可继续，重复提交幂等。累计题次只包含已计分记录（含漏答），正确率显示样本量。错题历史在掌握后仍保留。

v0.1.1 支持无练习历史题集的确认删除，有历史题集可归档和恢复显示；未完成会话也受删除保护。数据库自动迁移至版本 2，旧版 0.1.0 完整备份仍可恢复。

v0.1.2 修正考试已选进度；继续时定位下一道未完成题，答题卡区分草稿与提交并支持未答/待提交/标记检查；结果可筛选本次错题/漏答并创建独立学习复习。数据库仍为版本 2，兼容 0.1.0/0.1.1 完整备份。

不实现题目编辑、单题/批量删除、回收站或合并、Markdown/文本导入、高级组卷、复杂诊断、账号/云/AI/付费或自动更新。应用不进行网络访问；开发依赖下载不属于运行期功能。

## 发布边界

`./scripts/package.ps1` 只在本地产出包含 EXE、演示 JSON、中文说明的 ZIP 及 SHA-256 清单。构建产物、bindings、数据库、日志和本地备份均不提交。CI 对 main、feature/mvp-v0.1 的 push 和面向 main 的 PR 执行检查并构建 artifact，不自动发布 GitHub Release。阶段 7 用户授权验证后提交推送功能分支并核验 Actions；不合并、打标签或发布 Release。
