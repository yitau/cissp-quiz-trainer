# v0.2.0 知识学习验收记录

日期：2026-10-08（Asia/Shanghai）。分支：`feature/knowledge-learning-v0.2`。

G1～G5 的 P0 实现及下表验收已完成。应用通过真实 Wails bindings、Go Service 和 SQLite 持久化运行，没有演示用 Mock 后端。P1/P2 未实现。

## 环境与数据边界

- Windows 11 家庭版中文版，10.0.26200，x64；实际本地命令版本：Go 1.27.1、Wails CLI 2.15.0、Node 24.18.0、npm 11.16.0。Node 22 仍是 CI 基线，本地 Node 24 已通过本次项目检查。
- `go test` 全部使用 `t.TempDir()`；应用适配器测试使用 `t.Setenv("CISSP_QUIZ_DATA_DIR", ...)`。
- 原生窗口使用 `%TEMP%\cissp-v020-ui-20261008`。没有打开或改写默认真实用户数据库。
- Day 1 Lesson 来自仓库样例。关联实测使用本机已有 30 题文件，核对 `set.id = w01d01-review-20261008-v1` 后通过原生对话框导入隔离库；该私有题集、数据库、日志和备份不提交，也不包含在发行包中。

## AC-01～AC-21

| ID | 结果 | 实际证据 |
| --- | --- | --- |
| AC-01 | PASS | `TestLessonSample`、`TestLessonWorkflowRestartBackup`；原生文件选择 → Week 1 / Day 1 / Domain 1 / 2 章节 / 8 知识点预览 → 确认导入 → 唯一课程卡片。 |
| AC-02 | PASS | 样例解析覆盖八概念及案例全部字段；阅读器按 `sections[].conceptIds` 生成目录，字段安全文本渲染；原生窗口显示完整性讲解及五步案例，见 `resume.png`、`case-keyboard.png`。 |
| AC-03 | PASS | 原生机密性答案初始隐藏，点击后出现，切换完整性后隐藏；完整性答案可用空格键隐藏；`TestLessonWorkflowRestartBackup` 断言统计与会话数量不变。 |
| AC-04 | PASS | `TestLessonWorkflowRestartBackup` 遍历三种持久状态、首次打开与无记录状态；`TestLearningForeignKeyAndStateConstraints` 拒绝落库 not_started；原生目录显示未学习/学习中/已理解/需复习。 |
| AC-05 | PASS | `TestLessonWorkflowRestartBackup` 断言 1/8、全部 understood 时 8/8 完成、改 needs_review 后 7/8 未完成；案例不增加分母。 |
| AC-06 | PASS | `TestLessonWorkflowRestartBackup` 关闭数据库重开并 DeepEqual；原生 Alt+F4 退出、重新启动后继续定位 Integrity，需复习和机密性已理解保持，见 `resume.png`。 |
| AC-07 | PASS | `TestLessonLinkedQuiz` 测试准确 ID 的学习/考试调用；原生导入实际 30 题后，从课程分别创建原有学习和考试会话，均显示第 1/30 题。 |
| AC-08 | PASS | `TestLessonLinkedQuiz` 拒绝缺失/归档题集启动，恢复显示后可开始；原生未导入时显示具体 ID 与导入指引，没有开始按钮，见 `recall-keyboard.png`。 |
| AC-09 | PASS | `TestLessonWorkflowRestartBackup` 对改变 JSON 属性顺序/缩进的相同内容判重复，进度 DeepEqual；同 ID 修改正文判冲突并拒绝覆盖。 |
| AC-10 | PASS | `TestLessonValidation` 表驱动覆盖格式、版本、空白、缺字段、null、类型、超长文本、超 16 MiB；错误包含字段路径。 |
| AC-11 | PASS | 同一表驱动测试覆盖重复 section/concept ID、未知 sectionId、重复/遗漏/不存在引用、错误章节归属。 |
| AC-12 | PASS | 表驱动测试覆盖重复 JSON 属性、尾随 JSON、未知字段、Domain/Week/Day 范围、空集合；`TestLessonOptionalAndCanonicalNumbers` 覆盖可选字段及 JSON Schema 整数拼写。 |
| AC-13 | PASS | `TestLessonCancelInvalidAndTransaction` 断言预览不写库、取消/非法预览使 token 失效、AFTER INSERT 注入失败整体回滚且无幽灵进度。 |
| AC-14 | PASS | `TestLessonV2MigrationAndRestore` 用真正 v2 结构保存作答/收藏/归档，升级后统计与归档一致；`TestMigrationV1AndLegacyBackup` 保留 v1 路径；`TestMigration003FailureRollsBack` 验证失败不留下部分表或升版本。 |
| AC-15 | PASS | `TestLessonWorkflowRestartBackup` 对课程内容、两种理解状态、最近位置 DeepEqual 往返；`TestFullWorkflowBackupRestore` 与 `TestArchivePreservesHistoryReviewsAndBackup` 保留旧数据/归档；安全备份可再次恢复。 |
| AC-16 | PASS | 真正 v1 / App 0.1.0，以及 v2 / App 0.1.1、0.1.2 备份恢复成功，原 ZIP 字节不变；恢复后课程为空；恢复前安全备份可找回当前课程。 |
| AC-17 | PASS | `TestLessonCorruptBackupAndRollback` 覆盖坏正文、指纹、概念引用、状态、时间及恢复写入失败；`TestInvalidBackupPreservesCurrentData` 覆盖校验和/版本/配置/SQLite/安全备份失败；原库保持。 |
| AC-18 | PASS | `TestExamRestartAndIdempotency`、`TestStudyDraftIsNotSubmission`、`TestImportAndStudy`；原生课程启动考试后仍显示交卷前隐藏答案，未出现解析。 |
| AC-19 | PASS | 既有 `TestReviewAndStatistics`、`TestProgressAndResumeAfterRestart`、`TestSessionMistakesReviewPreservesResults` 及删除/归档/回滚测试全部通过。 |
| AC-20 | PASS | 下列必需命令均 exit 0；ZIP 核对含 EXE、原 demo、Day 1 Lesson、中文使用说明四项。 |
| AC-21 | PASS | 真实 Win11 原生窗口已执行课程导入、阅读/自检、状态、退出重启、30 题学习/考试关联；约 760 像素窄窗、Shift+Tab/Enter 与空格操作成功；文件选择框通过真实文件名输入完成。系统 DPI 125%/150% 等额外边界明确见下文。 |

## 实际执行的命令

最终代码检查（全部 PASS / exit 0）：

```powershell
gofmt -w .
go vet ./...
go test ./... -count=1
npm --prefix frontend run type-check
npm --prefix frontend run build
wails build -clean
./scripts/package.ps1
git diff --check
```

首次已检查干净工作区，执行 `git fetch origin`，通过 `git switch --track origin/feature/knowledge-learning-v0.2` 建立跟踪分支。本地依赖已存在；Wails 实际生成 bindings 并完成前后端构建，没有手写 bindings。未单独重新执行 `npm ci` 或 `go mod download`（NOT RUN：不是缺依赖的干净检出）。

首轮 Go 回归测试发现旧夹具把当前 schema 当作 v2：修正为删除 v3 新表、设置 user_version=2 并更新校验和的真实 v2 快照后通过；没有放宽备份兼容规则。导入校验还明确接受 JSON Schema 中等价的整数拼写 `1.0` / `1e0`。

## 数据与实现

- migration 003：`learning_units` 保存不可变规范课程 JSON/指纹/UTC 导入时间；`learning_progress` 联合主键、课程外键、三种持久状态、最近打开/更新时间。未学习无记录，服务计算已理解数和续学位置。
- 课程导入预览与确认分开；确认事务写入，服务再次检查重复/冲突。文件严格遵守 Lesson v1 契约及引用完整性。现有 Question v1 不变。
- 打开概念只更新阅读时间，不覆盖自主理解状态；主动标记只更新状态，不错误移动阅读位置。最近位置比较解析后的时间，时钟回拨时保持本课阅读顺序。
- 备份格式仍为 1，App 0.2.0 / DB 3。DB v1/v2 先校验实际结构，再迁移提取的受保护临时副本；业务校验包括课程指纹、引用、状态和时间。验证后保存恢复前备份，再事务替换。
- 导航在操作繁忙时禁用，避免页面挂载时丢失正在进行的刷新。知识阅读和正式考试完全分开，关联复用 StartQuiz，不创建新判分引擎。

## 原生证据

这些截图只包含仓库课程内容和隔离学习状态，不包含私有题目正文：

- [重启后定位 Integrity，保留需复习及 1/8](evidence/v0.2.0/resume.png)
- [约 760 像素单列目录](evidence/v0.2.0/narrow.png)
- [窄窗自检按钮键盘焦点、缺失关联题集引导](evidence/v0.2.0/recall-keyboard.png)
- [Shift+Tab / Enter 打开综合案例](evidence/v0.2.0/case-keyboard.png)

## 未执行及限制

- NOT RUN：Windows 系统 DPI 125%/150% 的独立矩阵；已实际验证窗口收窄至约 760 像素，不能等同于多 DPI 验证。
- NOT RUN：本轮原生保存备份/恢复确认对话框完整操作；数据层完整备份、旧版恢复、安全备份及失败回滚已实际自动化验证。原生课程和题库文件选择已实际成功。
- NOT RUN：本轮用 UI 逐个阅读全部八概念、逐个标满 8/8；八概念结构/全部状态与完成计算由自动化测试覆盖，UI 实测机密性、完整性及综合案例。
- NOT RUN：新增独立前端浏览器测试框架；本项目继续使用 TypeScript/build 和真实原生交互验证。
- 可执行文件未代码签名；Windows 属性 API 在此环境未读出 FileVersion/ProductVersion，应用界面、备份元数据和 Wails 配置的版本为 0.2.0。未将空属性读数声称为资源验证通过。
- 只交付本地 EXE/ZIP 和功能分支提交推送；未合并 main、打标签或创建 Release。
