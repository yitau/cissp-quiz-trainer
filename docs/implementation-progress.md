# v0.1 执行记录

时区：Asia/Shanghai（UTC+08:00）。目标状态：v0.1.2 基础练习体验迭代进行中。当前阶段：7。

| 阶段 | 状态 | 关联提交 |
| --- | --- | --- |
| 0 基线与业务规则 | 已完成 | a58623d |
| 1 导入与学习 | 已完成 | 24881f6 |
| 2 考试与历史 | 已完成 | ca47ef9 |
| 3 复习与统计 | 已完成 | bfc64a5 |
| 4 备份与日常使用 | 已完成 | 8bd2d1a |
| 5 最终验收与交付 | 已完成 | 1190faf |
| 6 题集删除与归档 | 已完成 | 49b21d8 |
| 7 基础练习体验与 CI | 进行中 | 待验证提交 |

## 阶段 7 — 基础练习体验与 CI

- 2026-10-01 19:32 +08:00：开始，重新读取 AGENTS、README、需求、计划/记录、实际 UI/服务/仓储与 CI。基线 feature/mvp-v0.1，工作区干净，origin=https://github.com/yitau/cissp-quiz-trainer.git。用户授权验证后提交推送并确认 Actions 成功。
- 已确认：历史考试进度错误使用 scored；继续总回首题；交卷仅显示未完成数量；缺本次错题重练。CI 仅监听 main，需要追加功能分支触发。
- gh auth status 确认 yitau 已登录；提权 shell 的 git ls-remote 未定位仓库，将改用明确工作目录及命令级 safe.directory 重试。
- 业务决策及验收顺序见计划阶段 7。下一步实现服务输出、UI 检查与重练及回归测试；本阶段尚未执行验证或推送。

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

## 阶段 1 — 导入与学习

- 2026-10-01 16:48 +08:00：完成 domain/models、importer/json、database/migrations/001、repository 接口及 SQLite 实现、service/trainer、App bindings、Vue/Pinia/API 模块和 4 题原创演示文件。
- SQLite 外键、迁移拒绝未知已有库、全局 ID/题序约束；题目 JSON 不可变，答案独立保存。规范化比较重复，冲突拒绝，可能重复内容提示；预览持有已校验数据，确认不重新读取变化的文件。
- 学习提交后后端才返回当前题答案；未提交题裁剪答案/解析/正确性；提交/结束幂等。题集一览、学习和历史接真实 Wails 服务，没有 mock。
- `gofmt -w .`、`go mod tidy`、`go test ./...`、`go vet ./...`、`wails build -clean` 均通过；Wails 实际生成 bindings。独立 `npm run type-check` / `npm run build` 通过。
- 测试覆盖非法 JSON/重复属性/未知版本/缺失字段/枚举/重复 ID，重复/冲突导入、强制数据库约束失败后的整套回滚、学习显示权限、重复提交、漏答和迁移保护。数据库均使用 t.TempDir。
- 原生界面操作尚未验证。下一步阶段 2：考试草稿、标记、交卷权限与进程重启持久化测试。

## 阶段 2 — 考试与历史

- 2026-10-01 16:53 +08:00：实现固定题集考试、选择/标记自动保存、题号导航、交卷确认及历史模式区分。学习草稿也保存，但未提交即结束仍按漏答计错。
- 关键文件：service/trainer.go、repository/sqlite/trainer.go、app.go、frontend/src/App.vue、services/api.ts、service/exam_test.go。
- 新增测试：考试开始/保存/恢复均隐藏答案；考试不能走学习评分接口；真实关闭数据库后重新打开保留选择/标记；8 个并发重复交卷只记一次；已结束会话禁止改写；草稿不计统计。
- `gofmt -w .`、`go test ./...`、`go vet ./...`、`wails build -clean`、`npm --prefix frontend run type-check`、`npm --prefix frontend run build` 全部通过。当前验证为服务/SQLite 层重启，不等于原生窗口端到端操作。
- 下一步阶段 3：错题、收藏、复习入口及统一的累计/单次/Domain/Type 统计。

## 阶段 3 — 错题、收藏与统计

- 2026-10-01 17:00 +08:00：增加错题历史/收藏页与各自复习入口、题内收藏、累计及单次 Domain/Type 表格、正确率样本量/漏答/用时。连续正确 3 次标记已掌握，曾错题继续保留。
- 关键文件：domain/statistics.go、service/statistics.go、repository/sqlite/statistics.go、service/statistics_test.go、frontend/src/components/StatisticsTable.vue、App.vue、Pinia store。
- 学习进度直接从独立作答历史推导，收藏独立存储；评分与统计口径在 Go 服务统一计算，界面只显示。用时为开始至交卷的墙钟时间（含暂停）。
- `gofmt -w .`、`go test ./...`、`go vet ./...`、`wails build -clean`、前端独立 type-check/build 全通过。测试验证已掌握后保留错题、收藏幂等/取消、空复习、考试草稿排除、漏答计数及 Domain/Type 样本量加总一致。
- 下一步阶段 4：安全完整备份恢复、严格版本与内容校验、损坏和事务失败后的原数据保护，以及日常错误/空状态。

## 阶段 4 — 完整备份恢复与日常使用

- 2026-10-01 17:03 +08:00：完整 ZIP 含 SQLite 一致快照、config.json 和版本/校验和 metadata.json；同名文件不覆盖。恢复严格验证归档条目、大小、格式/App/数据库版本、完整性、外键、实际 schema、题集/题目内容、会话评分及配置一致性。
- 恢复前自动在数据目录 backups 内保留完整安全备份，再用单连接 ATTACH + 事务替换业务表，配置同事务恢复。失败不改原库；安全备份失败也取消恢复。恢复成功清空界面旧会话及导入预览。
- 关键文件：domain/backup.go、repository/sqlite/backup.go、service/backup.go、service/backup_test.go、App 原生文件选择和恢复确认、Vue 统计/备份区；Wails 添加单实例设置与文件/数据库操作期间关闭保护。
- `go test ./internal/service -run 'TestFullWorkflowBackupRestore|TestInvalidBackup|TestRestoreTransaction' -count=1 -v` 通过。覆盖真实数据库完整闭环、恢复后重启、安全备份可回退、已有备份不覆盖、损坏 ZIP、格式/数据库版本、校验和/路径/配置/评分篡改、损坏 SQLite、安全备份创建失败、写入中途事务失败回滚。
- 随后 gofmt、全量 go test、go vet、Wails Windows build、前端 type-check/build 全通过。全部使用隔离目录，没有访问真实学习库。原生文件对话框、关闭保护的窗口行为待人工验证。
- 下一步阶段 5：自查、修正干净检出构建顺序、最终全套验收、实际 EXE 启动检查、ZIP 与使用/人工验收说明。

## 阶段 5 — 最终验收与交付

- 2026-10-01 17:08 +08:00：自查修正干净检出构建顺序。CI 改为 npm ci → Wails 生成 bindings/构建 → 独立前端和 Go 检查 → 本地打包；没有添加远程发布。同步 README/AGENTS 的首次构建说明。
- 阅读实际 Wails v2.15.0 Windows dialog.go 发现 MessageDialog 忽略自定义中文按钮且返回 Yes/No；修正恢复确认比较及默认 No，避免无法确认恢复。增加 App 层本地错误日志、友好数据库错误和适配器重启测试。
- 移除已核对路径的生成目录 frontend/dist、frontend/wailsjs 后执行 `npm --prefix frontend ci`、`wails build -clean`、前端 type-check/build、`go vet ./...`、`go test ./... -count=1 -v`、gofmt 检查，全部通过。未手工伪造 bindings。
- `go test ./... -count=1 -json` 保存于忽略的 tmp/final-tests.jsonl：10 个顶层测试 + 17 个子用例通过，失败 0。完整使用流程测试覆盖导入→学习错误及解析→收藏→考试→统计→重启→备份→新数据库恢复核对→安全备份回退。
- 2026-10-01 18:36 +08:00：实际 EXE 在 `tmp/ui-source` 隔离目录启动，生成 65536 字节 cissp.db、空 app.log；Windows Computer Use 的 list_windows 返回 CISSP Quiz Trainer 窗口。首次隐藏启动没有可操作窗口，重启为可见窗口后找到；get_window/get_window_state 在应用访问审批步骤报 `Computer Use app approval timed out`，没有获得截图、没有执行 UI 点击或输入。未访问默认学习数据目录；测试进程已清理。
- 验证边界：Go 单元/SQLite 集成、App 适配器重启、Windows 构建与进程/窗口创建成功；原生界面渲染、文件对话框、实际按钮流程、缩放/键盘及关闭保护未验收，不宣称全部端到端通过。按用户允许的工具限制路径保留 docs/user-guide.md 人工步骤。
- `scripts/package.ps1` 已产出 EXE/ZIP/checksums。读取 PE 头确认为 0x8664（x64）；检查 ZIP 恰有 EXE、demo-questions.json、使用说明.md，无数据库或私有数据。
- 自查补充单次结果的显式错误数量和损坏原库时在新数据目录恢复的说明；正在进行受影响的最终重建与重新打包，完成后保存本地提交。
- 2026-10-01 18:39:02 +08:00（Asia/Shanghai）：结果错误数补充后的最终 gofmt 检查、go vet、go test -count=1、Wails Windows 构建、独立前端 type-check/build 全通过；重新生成正式 bindings 与 EXE/ZIP。最终测试结果仍为 10 个顶层 + 17 个子用例、失败 0。
- 最终 EXE 再以 tmp/final-startup 隔离启动：进程存活、数据库初始化成功、错误日志为空，检查后清理进程。原生窗口操作审批超时未解决，仍作为人工验收限制保留，不宣称全部 E2E 通过。
- 本地交付：build/bin/CISSPQuizTrainer.exe（16,053,248 字节）、build/bin/CISSPQuizTrainer-win-x64.zip（6,758,247 字节）。SHA-256 见 build/bin/checksums.txt；EXE = 8ecee0c89b4c1b0c78024711af4d097314403498d4548ac17898c0cf211f8f32；ZIP = e98bfc196c44f17a5a060d48ae8200f9439837a8e04557e0f2e2e7103fd290cf。
- 自查：未新增运行期网络、账号或复杂框架；SQL/服务/绑定边界保留；无真实数据验证；git diff --check 通过；未跟踪 EXE、数据库、生成 bindings/dist、node_modules。没有推送、合并、打标签或发布。
- 已知限制：原生文件对话框/按钮全流程/显示缩放与键盘/关闭保护待人工；Node 22 与远程 GitHub Actions 未实际运行；发行文件未签名；仅支持同 App/数据库/备份版本恢复，损坏原库须保留原目录并用新的空数据目录恢复。下一步是按 user-guide.md 人工验收步骤核验原生界面，无未实现的本次业务功能。

## 本地交付收据

- 分支：feature/mvp-v0.1。阶段提交：a58623d（基线）、24881f6（导入学习）、ca47ef9（考试历史）、bfc64a5（错题收藏统计）、8bd2d1a（备份恢复）、1190faf（验收打包说明）。本段作为最后的记录提交补入阶段关联号。
- 阶段 5 实现提交后 `git status --short` 为空；本次仅进一步提交这份收据。忽略目录中保留本地依赖、构建产物与隔离验证记录，没有提交真实题库或数据库。原生 UI 人工验收限制仍有效。

## 阶段 6 — 题集删除与归档

- 状态：已完成（代码/测试/构建；原生 UI 验收待人工）。开始时间：2026-10-01 19:15 +08:00（Asia/Shanghai）。用户明确授权追加最小删除/归档功能；不做单题删除、批量删除和回收站。
- 基线：feature/mvp-v0.1，工作区干净。已重新读取 AGENTS、README、计划/记录、实际数据库/服务/界面/备份调用链。
- 决策：任何会话引用均阻止物理删除（含未完成草稿），归档保留全部关联记录；同页查看已归档及恢复显示。用新增归档表避免改写现有题目；保留旧版完整备份恢复能力。
- 下一步：实现 migration 002、事务删除与归档服务、确认界面和回归测试；验证后更新产物与提交。
- 2026-10-01 19:25:08 +08:00（Asia/Shanghai）：完成 migration 002、新增归档表、事务删除及二次检查、服务/绑定/API、日常/已归档筛选和恢复显示、删除确认 dialog（名称/题量/不可撤销/收藏说明，默认取消及 Escape）。产品 0.1.1，数据库 2；JSON 契约仍为 1.0。
- 关键文件：internal/database/{database.go,migrations/002_set_archives.sql}、repository/sqlite/{set_management.go,trainer.go,backup.go}、service/{set_management.go,trainer.go,backup.go,set_management_test.go}、domain/{models.go,backup.go}、app.go、frontend/src/App.vue、components/DeleteSetDialog.vue、services/api.ts；更新 README、需求、使用说明、计划与本记录。
- 旧版 App 0.1.0 / DB 1 备份可恢复：校验原始校验和/实际版本/schema，仅迁移解包的临时数据库，然后完整校验、安全备份、事务恢复。兼容 CRLF/LF schema；旧备份原文件未修改；新备份完整保留归档状态。
- 新增定向命令 go test ./internal/service -run 'TestDelete|TestArchive|TestMigrationV1' -count=1 -v 通过。包括：无历史删除后重新导入、过期预览不重建删除题集、空会话/考试草稿/学习已计分/完成考试/收藏复习的删除保护、删除中途失败回滚、归档与取消归档幂等、归档后会话/统计/复习保留、重启及新旧备份恢复、伪造版本拒绝且原数据不变。
- 全量 gofmt -w . / gofmt -l .、go vet ./...、go test ./... -count=1 -json、npm --prefix frontend run type-check、npm --prefix frontend run build、wails build -clean、scripts/package.ps1 全通过。测试证据 tmp/iteration6-tests.jsonl：15 顶层测试 + 22 子用例通过，失败 0。全部为临时隔离数据；没有运行真实数据删除。
- 本地 EXE/ZIP 已重新打包（build/bin）：EXE 16,070,656 字节，ZIP 6,765,919 字节；SHA-256 分别为 de691673505d21e540dbe81251c3452e6d461c955af33aa807d8a2734e261724 / cb41355ee17572af83d305b754cb081f3740a36f6fb2bf5d835aebd49a421396。
- git diff --check 通过。原生 UI 本轮没有操作验证，仍按 user-guide.md 删除/归档人工验收步骤核对对话框取消/确认、切换列表与键盘；不将类型检查或构建称为 UI 验收通过。下一步：使用更新 ZIP 进行人工界面验收。本地提交后继续保留不推送边界。
- 阶段 6 关联功能提交：49b21d8。提交后 git status --short 为空；本次仅补入该关联号。分支仍为 feature/mvp-v0.1，没有推送。

### 阶段 7 实现与首轮验证（2026-10-01 19:42:11 +08:00，Asia/Shanghai）

- 已实现进度区分、基于持久化答案的下一未完成题续答、题号筛选与交卷前未完成列表、本次错题/漏答筛选及独立学习重练；数据库仍为 2，产品 0.1.2，兼容 0.1.0/0.1.1 备份。更新使用说明与需求范围。
- 关键文件：internal/service/session_review.go 与测试、domain/models.go、repository/sqlite/trainer.go、service/backup.go、app.go、frontend/src/App.vue、services/api.ts、.github/workflows/ci.yml。未增加依赖或迁移。
- 定向测试 TestProgressAndResume / TestSessionMistakes / TestVersion011 通过，覆盖真实 DB 关闭重启、草稿保密/不计分、选择清空、全答完标记回退、原结果不变、归档题复习、全对/未交卷拒绝及 0.1.1 备份恢复。
- gofmt -w . / gofmt -l .、npm --prefix frontend run type-check / build、go vet ./...、go test ./... -count=1 -json 均通过。tmp/iteration7-tests.jsonl：18 顶层 + 26 子用例、失败 0。git diff --check 通过。
- wails build -clean 两次因正在运行的旧版 EXE（PID 51716）无法覆盖而失败。普通沙箱 Get-Process 的空结果不能证明进程已退出；提权只读查询确认窗口仍在。已请用户关闭，未终止用户进程。
- 改用 wails build -o CISSPQuizTrainer-v0.1.2.exe 完成实际 Windows 生产构建，生成正式 bindings。原生窗口读取返回 minimized，尚未执行新版界面操作。
- 远端只存在 main@9fb84ed，GitHub yitau 已登录；复用当前 feature/mvp-v0.1。下一步提交推送本次功能，跟踪 Actions 干净构建，旧窗口关闭后更新默认 EXE/ZIP 并验收新窗口。

### 首轮远端 CI 与修复（2026-10-01 19:48:30 +08:00，Asia/Shanghai）

- 功能提交 4bdc02b 已普通推送，Actions run 36857039509：Go/Node 22 准备、依赖、Wails 干净 Windows 生产构建、前端 type-check/build 通过；Go formatting 失败，后续 vet/test/打包未执行。
- gh run view --log-failed 显示全部 Go 文件需要格式化；原因是 Windows 默认 autocrlf 检出，没有仓库级 LF 约束。新增 .gitattributes 的 *.go text eol=lf，保留原 gofmt 检查，未关闭检查或自动格式化掩盖错误。
- 独立本地产物 build/bin/CISSPQuizTrainer-v0.1.2.exe（16076288 字节）和 CISSPQuizTrainer-v0.1.2-win-x64.zip 已生成；ZIP 已核对只含 EXE、原创演示 JSON 和使用说明。旧版窗口仍占用默认 EXE，尚未做新版 UI 验收。
- 下一步：提交换行修复，验证全新检出格式，继续跟踪修复后的完整 Actions。
