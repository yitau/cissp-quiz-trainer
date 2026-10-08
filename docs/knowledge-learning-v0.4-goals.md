# v0.4 → v0.2.0：知识学习模块 Implementation Goals 与验收矩阵

> 日期：2026-10-08；状态：开发执行任务输入（并非功能完成报告）  
> 需求权威：[requirements-v0.4.md](requirements-v0.4.md)  
> Lesson 契约：[schemas/lesson-v1.schema.json](schemas/lesson-v1.schema.json)  
> Day 1 样例：[../samples/cissp-week01-day01-lesson.json](../samples/cissp-week01-day01-lesson.json)  
> 起点：`main` 已实现 v0.1.2；文档工作分支 `feature/knowledge-learning-v0.2`。

## GOAL（必须最终达到）

在**不破坏 v0.1.2 任何已有功能或用户数据**的前提下，在 Windows 11 x64 的离线 CISSP Quiz Trainer 内建立完整的“JSON 课程导入 → 知识阅读/自检 → 自主标记进度 → 关闭重启续学 → 关联原题库刷题 → 完整备份恢复”闭环。

合格示例：导入 `samples/cissp-week01-day01-lesson.json` 后，课程目录有 **2 章节 / 8 概念 / 1 案例**；能依次学习八个概念，阅读自检参考答案，标记理解状态，退出软件并重启仍能恢复；存在 `set.id = w01d01-review-20261008-v1` 的已导入题集时能进入对应原学习/考试模式，不存在时给出明确提示；课程操作不改变答题成绩、错题或历史。

## 实施阶段（建议顺序）

### G1 契约与导入（先测试后实现）
- 创建独立 Lesson domain 模型及导入解析/校验（复用既有唯一 JSON 属性检测、16 MiB 限制、事务导入的适当代码，不复制脆弱实现）。
- 校验机器 Schema 结构 + Go 侧跨引用约束（section/concept 双向引用、唯一且仅引用一次）+ trim 校验。
- 提供预览/确认两阶段导入，并处理重复、冲突、非法格式、错误字段路径、取消/失败不落库。
- 课程与 Question 格式各自独立，不能以改变 Question Schema 的方式绕开新接口。

### G2 持久化与旧数据兼容
- 新建 migration 003：`learning_units` + `learning_progress`，用外键、事务、唯一约束保存不可变课程与学习状态。
- 概念最近打开时间与理解状态分开更新，重开课程的定位来自持久化最近打开记录。
- 保证数据库 v1/v2 → v3 升级不清空任何现有会话、答案、收藏、归档、统计；异常可安全失败。
- 备份/恢复同步支持 v3 课程与进度；旧 v1/v2 合法备份继续可恢复到新版，保留当前恢复前安全备份及失败回滚语义。

### G3 UI 阅读器和进度
- 主导航加“知识学习”；课程导入、空状态、Week/Day 排序、课程列表、章节目录、8 概念按引用顺序显示。
- 每个概念显示 term en/zh、objective、definition、explanation、analogy、scenario、controls（若有）、commonMistake、examTip、recall。
- 自检答案必须在点击查看后才出现，切换概念默认隐藏；不创建 Quiz Answer。
- 支持 `in_progress`、`understood`、`needs_review` 及派生 `not_started`；展示“已理解 x/8”，全部 8 个 understood 才视为完成。
- 案例页显示风险概念与 CIA 的联系，但案例页不改变知识点分母；按键盘操作，窄窗口不溢出。

### G4 关联已有刷题工具
- `linkedQuestionSetId` 使用现有 set ID，不根据文件名/标题模糊匹配。
- 存在且非归档：允许调用原有会话 API 开始学习/考试；缺失：提示导入、不能启动空会话；已归档：提示恢复显示，不旁路归档规则。
- 不改原有 FIRST/BEST 判分逻辑、考试隐藏解析、统计范围、错题保留和收藏功能。

### G5 文档、打包和验证
- 同步 README、`docs/user-guide.md`、`docs/implementation-progress.md` 及必要的架构/迁移说明；标明实际完成和未验收项。
- ZIP 中包含新 Lesson 示例与原有演示题库文件，保持 Win11 无账号、无联网依赖。
- 通过自动化、构建和可能的原生窗口验证；没有执行的项目必须列为未验证，不能“默认通过”。

## 必须验收（Definition of Done）

| ID | 测试场景 | 通过标准 |
| --- | --- | --- |
| AC-01 | 新库导入 Day 1 Lesson | 预览正确显示 Week 1 / Day 1 / Domain 1 / 2 章节 / 8 概念；确认后唯一导入 |
| AC-02 | 阅读内容 | 八概念字段均可查看；章节与概念顺序符合 `sections[].conceptIds`；综合案例可读 |
| AC-03 | 理解自检 | 答案默认不可见，点击才显示，切换题目后隐藏，且统计/练习记录不增加 |
| AC-04 | 自主状态 | 未打开为 not_started，首次打开为 in_progress；能改 understood/needs_review/恢复 in_progress |
| AC-05 | 完成计算 | only understood 计入已理解，全部 8/8 understood 才是完成；案例不计数 |
| AC-06 | 重启续学 | 最后打开位置可恢复；重开不把 understood/needs_review 强制改回 in_progress |
| AC-07 | 有关联题库 | 导入 `set.id = w01d01-review-20261008-v1` 的 Question JSON 后，可进入既有学习和考试 |
| AC-08 | 无关联题库/归档 | 不能启动空会话；缺失显示导入指引，归档时按既有规则引导恢复 |
| AC-09 | 重复/冲突 | 相同 ID 同内容报告重复且保持进度；同 ID 不同内容拒绝，数据原封不动 |
| AC-10 | 格式负例 | 错 `format`/`schemaVersion`、空白必填、缺字段、非法类型、超大文件全部拒绝 |
| AC-11 | 引用负例 | 重复 section/concept ID、未知 sectionId、重复/遗漏 conceptIds 或错误章节引用全部拒绝 |
| AC-12 | JSON 负例 | 重复 JSON 属性、未知字段、非法 Domain/Week/Day、空集合拒绝并提供定位信息 |
| AC-13 | 中断和回滚 | 取消确认、非法文件、事务写入失败均不产生半套数据或幽灵进度 |
| AC-14 | 既有数据升级 | 现有 v2 数据库迁移 v3 后，历史成绩、题集、收藏、归档和错题保持一致 |
| AC-15 | 新版完整备份往返 | 课程内容、已理解/需复习状态、最近位置均与原库一致；原有数据也一致 |
| AC-16 | 合法旧备份恢复 | v1/v2 备份可按原有安全迁移规则恢复到新版；课程为空，旧数据保持不变 |
| AC-17 | 损坏备份回滚 | 校验/迁移/恢复失败均不破坏原库，恢复前安全备份规则仍然有效 |
| AC-18 | 考试回归 | 交卷前后端不返回答案/解析；学习提交、考试交卷、漏答和重复提交计分正确 |
| AC-19 | 其他回归 | 错题重练、历史续答、题集删除/归档、累计 Domain/Type 统计均通过既有测试 |
| AC-20 | 构建和交付 | Go test/vet、前端 type-check/build、Wails Windows build、打包成功；ZIP 有旧 demo JSON 与新 lesson JSON |
| AC-21 | 原生 UI 人工验收 | Win11 上可导入、阅读、退出重开、关联刷题；缩放/键盘与原生文件对话框结果有证据或明确未验证 |

## 必须执行的检查

全量新功能自动测试应包括：校验负例表驱动测试、导入幂等/冲突/事务测试、进度状态转换和续学测试、SQLite 迁移与备份恢复往返测试、原有考试泄题与计分回归。

建议顺序（遵守当前 AGENTS.md 及真实工具链）：

```powershell
git status --short
go mod download
npm --prefix frontend ci
wails build -clean
npm --prefix frontend run type-check
npm --prefix frontend run build
gofmt -w .
go vet ./...
go test ./...
wails build -clean
./scripts/package.ps1
```

注意：干净检出时首次 `wails build -clean` 用于生成正确的 Wails bindings；不要手写生成的前端绑定文件。`gofmt -w .` 之后必须检查 diff，避免提交无关变化。Windows 原生交互测试和编译测试**不是同一种验收**，分别报告。

## 完成后的报告格式（供 Codex 输出）

1. **GOAL 达成情况**：逐条列 AC-01 至 AC-21 的 PASS / FAIL / NOT RUN（附执行证据/命令/阻塞原因）。
2. **代码/数据库变更摘要**：目录、接口、migration、兼容性和保留的既有功能。
3. **验证记录**：实际运行的命令及结果、是否真正运行了 Win11 原生 UI。
4. **数据安全确认**：使用隔离测试库，不修改用户真实数据库；确认旧备份/历史保留规则。
5. **Git 交付**：commit SHA、当前分支、远程推送状态。未授权不合并 main、不发布 Release。
6. **未完成/风险**：按 AC 编号列出，不能用“基本完成”掩盖未测或失败项目。

GOAL 判定：P0 需求已实施，AC-01～AC-20 有实际成功的证明，且 AC-21 的原生窗口验证已执行并记录成功才可以声明“全部完成”。环境缺失时只可声明“实现/自动化验证完成，原生验收待办”，不得谎报完整 GOAL。
