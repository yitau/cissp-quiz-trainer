# CISSP Quiz Trainer — 知识学习模块需求规范 v0.4

> 文档状态：v0.2.0 实施基线（P0 已实现，验收见 knowledge-learning-v0.2-acceptance.md）
> 日期：2026-10-08  
> 代码基线：`main` 已合并的 v0.1.2  
> 目标产品版本：v0.2.0（功能分支构建，未发布 Release）
> 目标平台：Windows 11 x64，本地离线  
> 主要关联文档：[既有需求基线 v0.3](requirements-v0.3.md)、[实施目标及验收](knowledge-learning-v0.4-goals.md)、[机器可读 Lesson Schema](schemas/lesson-v1.schema.json)、[Day 1 学习包](../samples/cissp-week01-day01-lesson.json)

## 1. 版本关系与约束

- `requirements-v0.3.md` 保留为已实现刷题业务与架构的历史基线；本文在其基础上**增量定义**独立的知识学习模块。若本模块范围内表述冲突，以本文为准；现有刷题/考试、数据安全和架构规则仍以 v0.3 为准。
- **需求文档版本** v0.4、**目标应用版本** v0.2.0、**Lesson 交换格式版本** 1.0、既有 **Question 交换格式版本** 1.0 是四个独立版本号。
- 不重命名已有“学习模式”：该模式仍指“逐题提交后看解析”。新增顶部主导航“**知识学习**”，其功能为结构化阅读和理解自检。
- v0.2 仅实现本文 P0。P1/P2 是后续候选需求，未经明确授权不得提前开发。
- 沿用 Go 1.27.x + Wails v2.15.0 + Vue 3 + TypeScript + Pinia + SQLite + 现有 Repository/Service 架构，遵守根目录 `AGENTS.md`；不迁移技术栈。

## 2. 背景、问题、用户与目标

使用场景：学习者完成 Week 01 · Day 01 的 CIA Triad 与 Asset/Threat/Vulnerability/Risk/Control 等基础知识后，间隔数周需重新建立概念记忆，再使用现有刷题器验证。现有 v0.1.2 只有答案解析，没有课程目录、系统性的知识讲解、阅读位置与知识点理解状态。

**GOAL（完成定义）**：学习者能够在完全离线的 Windows 应用中导入 Day 1 Lesson JSON，按照既定顺序阅读 **2 个章节、8 个知识点及 1 个综合案例**，查看术语/类比/易错点/理解自检，标记学习状态，关闭重启后从最后阅读位置继续，并跳转到正确关联的 **30 道复习题集**（该题集已导入时）。恢复/升级不会损坏既有题库与历史。

功能目标（可验收）：
1. 用户不需要刷题即可开始知识学习；阅读和理解自检不产生答题成绩。
2. 课程由 ChatGPT 离线生成的结构化 JSON 文件提供；不需要联网或 AI API。
3. 相同的学习结构可用于以后 Week/Day 的课程，而无需重新开发页面。
4. 原有题库、考试答案隐藏、统计、错题、归档和备份规则继续正常工作。

## 3. 功能范围及非目标

| 分类 | 优先级 | 功能 |
| --- | --- | --- |
| 课程导入 | P0 | 单个 Lesson JSON 的选择、校验、预览、确认导入、重复/冲突识别 |
| 课程目录 | P0 | 主导航“知识学习”；按 Week/Day 排列课程，展示章节数、知识点数和进度 |
| 阅读器 | P0 | 章节目录、前后知识点导航、术语中英对照、讲解/类比/案例/易错点/考试提示 |
| 理解自检 | P0 | 先展示自检问题，用户点击才揭示参考答案；不判分、不算正式题次 |
| 进度记录 | P0 | 逐知识点状态 + 最近打开位置；关闭重启后继续，主动标记“已理解/需复习” |
| 综合案例 | P0 | 课程末显示跨知识点案例，按内容顺序阅读；不作为单独已理解计数项 |
| 关联题库 | P0 | 用已有 `set.id` 定位 Question Set；已存在可进入原学习/考试模式，缺失给出导入指引 |
| 可靠性 | P0 | 新数据库 migration；完整备份/恢复包含课程与进度，兼容既有旧备份 |
| 交付 | P0 | Day 1 示例学习包、自动测试、Windows 构建打包、使用说明和验收记录 |
| 错题映射 | P1 | 单道正式题关联一个或多个 concept ID，错题跳回准确知识点 |
| 检索与聚合 | P1 | 术语搜索、薄弱知识汇总、学习与题目成绩综合展示 |
| 智能学习 | P2 | AI 在线讲解/出题、SM-2 间隔重复、音视频课程、云账号同步 |

**非目标**：Markdown/HTML 自由导入、课程正文编辑或覆盖、课程合并、题集结构变更、复杂课件渲染器、远程联网、知识测验独立判分引擎。尤其不得因新增课程而改变现有 Question Schema v1.0。

## 4. 用户流程

```text
主导航“知识学习”
  → 导入 Lesson JSON → 验证 → 预览（名称/Week/Day/Domain/2 章节/8 知识点）
  → 确认导入 → 课程列表
  → 打开课程 → 左侧章节目录 / 右侧知识点正文
  → 阅读术语、定义、原理、类比、案例、常见误区、考试提示
  → 自检：点击“查看参考答案” → 按需隐藏
  → 主动标记“已理解”或“需复习” → 下一个知识点
  → 综合案例 → 关联题库入口
    ├─ 已存在且未归档：调用现有 StartQuiz（学习/考试）
    ├─ 尚未导入：提示关联题集 ID，进入“题库与导入”
    └─ 已归档：提示先在题库中恢复显示，不绕过既有归档限制
  → 重新打开课程：从最近打开的知识点继续，原状态保留
```

课程阅读和正式做题在不同页面、不同持久化记录中进行；学习页不得自动提交答题，也不得改变原有答题分数。

## 5. 可测试的功能需求（P0）

### FR-01 课程导入和预览

- 新增仅接受 `cissp-lesson` 格式的导入入口；通过现有 Wails 原生文件选择方式读取本地 UTF-8 `.json`。
- 导入前先进行格式/字段/跨引用/重复校验；预览显示课程名、Week、Day、Domain、章节名和知识点计数、关联题集 ID，不需展开所有正文。
- 用户确认后**一次事务**写入课程；取消、格式错误、导入失败均不得写入半套数据或创建进度记录。
- 严格区分课程与题集导入，错误类型应明确，例如“文件是题库 JSON，请在题库与导入中打开”。

### FR-02 列表、阅读和导航

- 在现有主导航增加“知识学习”，不替换“题库与导入”的“开始学习”按钮。
- 按 week、day、lesson.id 排序；卡片至少显示标题、Domain、知识点数量、`已理解 / 总数` 和“继续学习”。
- 课程章节和知识点顺序以 `sections[].conceptIds` 为权威顺序，不按 ID 字典序排序；左侧/上方显示目录，右侧/下方显示当前概念。
- 单知识点正文按稳定顺序显示：英文术语及中文译名 → 学习目标 → 定义 → 原理 → 类比 → 业务场景 → 相关控制措施（如有）→ 常见误区 → CISSP 提示 → 理解自检。
- 首尾知识点的“上一条/下一条”应正常禁用；可使用键盘和鼠标，窄窗口不得水平溢出。不得渲染导入内容为未经净化的 HTML（默认安全文本）。
- 显示可选的综合案例 `caseStudy`；Day 1 必须包含资产、威胁、脆弱性、风险和控制措施的综合逻辑。

### FR-03 理解自检

- `recall.prompt` 始终显示；`recall.answer` 仅在用户显式点击“查看参考答案”后显示，允许再次隐藏。
- 这是非判分的记忆回想题，不创建 quiz_session/quiz_answer、不计入正式题统计、不影响既有考试模式的答案披露规则。
- 不需要持久化自检答案是否展开。切换知识点后默认隐藏。

### FR-04 进度与续学

- 概念状态：`not_started`（无记录，派生）/ `in_progress` / `understood` / `needs_review`。
- 第一次打开概念：写入 `in_progress` 和最近打开时间。再次打开只更新最近打开时间，**不得**将已理解/需复习自动降为学习中。
- 用户可以显式标记 `understood`（已理解）或 `needs_review`（需复习）；允许显式回到 `in_progress`。操作应持久化且重复点击不产生重复的状态记录。
- “已理解数”只计 `understood`，课程完成条件是全部概念为 `understood`。`needs_review` 不视为完成；`caseStudy` 不占 8 个知识点计数。
- `最近打开位置` 从本课最近一次 `last_opened_at` 非空的知识点推导；没有进度从第一知识点开始。重启继续不改变用户的主动理解标记。
- “已理解”是用户自评，不等于 CISSP 考试已经掌握；不得与现有“连续正确 3 次已掌握”的错题规则混淆。

### FR-05 关联题库与边界

- `lesson.linkedQuestionSetId` 对应现有 Question Set 的 **set.id**，而不是 JSON 文件名、题集显示名称或单题 id。
- 允许先导入课程、后导入题集。关联题集不存在时不失败、不新建空题集；显示具体 ID 和“前往题库与导入”。
- 已存在且未归档：点击“开始复习题（学习）”或“考试模式”复用现有会话 API 和答题体验。已归档：按既有归档规则引导恢复，不绕过限制。
- 关联只是一条字符串引用；**P0 不修改 Question Schema，也不提供单题 → concept 的精确映射**。

### FR-06 重复、冲突及错误语义

- 课程 `lesson.id` 是稳定的全局身份；相同 ID 与规范化内容完全相同提示“重复”，不新增、不覆盖、不清除进度；相同 ID 但内容不同提示“冲突”，拒绝覆盖。
- 课程内部 `section.id` 和 `concept.id` 在各自集合内唯一。每个概念必须且只能被 `sections[].conceptIds` 引用一次；不能指向不存在的概念、不能重复引用或遗漏。
- 按 JSON Schema 验证类型、枚举、最大最小值、字段；此外须拒绝重复 JSON 属性、仅由空格组成的必填文本、过大文件、未知格式版本。
- 文件上限 **16 MiB**，和当前题库上限一致；至少 1 个章节和 1 个知识点。报错包括字段路径/知识点 ID，导入失败不留下半套课程。
- 规范化相等可通过解析后确定性序列化（排除仅空白和 JSON 属性顺序差异）得到的哈希判定；概念与章节数组的顺序是业务内容，不可打乱。

### FR-07 SQLite、业务层与安全

- 新增 migration 003，保留既有版本 1 → 2 路径。建议新增 `learning_units`（课程规范 JSON、指纹、导入时间）、`learning_progress`（课程 ID + 概念 ID 联合键、状态、最近打开时间、更新时间）两表。
- `learning_progress` 通过 `unit_id` 外键引用课程；概念属于 JSON 文档，更新进度时在服务层验证概念 ID 必须存在。
- `not_started` 不落库，无记录时派生；记录中状态只可为 `in_progress`、`understood`、`needs_review`；默认不用第三张“阅读位置”表，按最近打开时间查询即可。
- 继续遵守 `Vue → frontend services → Wails thin adapter → service → repository → SQLite`，业务验证与事务在 Go 层实现。不得在 Vue 中直接读写数据库。
- 时间持久化使用 UTC；UI 显示本地时间。离线运行，不引入账号、遥测、AI API 或额外后端服务。

### FR-08 完整备份恢复和旧数据兼容

- 更新完整备份与恢复：导出和恢复课程 JSON 与 `learning_progress`，在恢复后的数据库中保留全量学习与答题数据。
- 新版能读取和正确迁移旧 App 0.1.0/0.1.1/0.1.2 的合法备份（DB v1/v2），在受保护的临时副本执行必要 migration 后再事务恢复。旧备份无课程应显示空课程而非错误；原始旧备份不可被修改。
- 新版备份反向兼容旧应用**不要求支持**。只允许经现有版本、完整性、业务校验后替换；保留恢复前安全备份，失败时现用数据库不变。
- 题库、考试会话、答题历史、归档/收藏、已计分统计继续保持原语义。数据库版本预计升为 3，应用版本预计升为 0.2.0，实现时与迁移测试结果一致。

### FR-09 样例内容及交付

- 仓库提交 `samples/cissp-week01-day01-lesson.json`：**两个章节（安全三要素、风险基础）、8 个概念、一段综合案例**；每个概念的六大讲解字段及自检字段均非空。
- 示例的 `linkedQuestionSetId` 为 `w01d01-review-20261008-v1`，对应 2026-10-08 已生成的 30 道 Day 1 复习题集的 `set.id`。这套题集不是官方真题，可能不在仓库内；关联目标未导入时必须优雅降级。
- 更新 README、`docs/user-guide.md`、`docs/implementation-progress.md` 说明知识学习入口、导入样例、升级与恢复规则；Windows ZIP 包应增加该 Lesson 示例而不移除现有 `demo-questions.json`。
- 有能力运行的测试必须实际执行；不可把文档清单或编译通过声称为 Windows 原生交互完整验收。

## 6. Lesson JSON v1.0 — 权威交换契约

机器可读 Schema：`docs/schemas/lesson-v1.schema.json`（JSON Schema Draft 2020-12）。此 `schemaVersion` 只表示 Lesson 格式，不等于 Question 格式，导入时须同时检验 `format`。

顶层结构（此处为结构摘要；字段约束以机器 Schema 与本节跨字段规则共同定义）：

```json
{
  "format": "cissp-lesson",
  "schemaVersion": "1.0",
  "lesson": {
    "id": "cissp-w01-d01",
    "title": "Week 01 · Day 01 — 安全基础",
    "description": "CIA Triad 与风险基础概念",
    "week": 1,
    "day": 1,
    "domain": 1,
    "createdAt": "2026-10-08T00:00:00Z",
    "linkedQuestionSetId": "w01d01-review-20261008-v1"
  },
  "sections": [
    {
      "id": "cia",
      "title": "安全三要素",
      "conceptIds": ["w01d01-confidentiality"]
    }
  ],
  "concepts": [
    {
      "id": "w01d01-confidentiality",
      "sectionId": "cia",
      "term": {"en": "Confidentiality", "zh": "机密性"},
      "objective": "理解未授权披露",
      "definition": "信息仅供获得授权的主体访问",
      "explanation": "机密性关注信息是否向未经授权的人披露。",
      "analogy": "病历档案只向授权人员开放。",
      "scenario": "医院员工未经授权读取患者病历。",
      "controls": ["Encryption（加密）", "Access Control（访问控制）"],
      "commonMistake": "不要把仅发生的未授权读取判定为完整性受损。",
      "examTip": "先识别业务目标和安全属性，再选择合适措施。",
      "recall": {
        "prompt": "只发生未授权读取时主要破坏什么？",
        "answer": "Confidentiality（机密性）。"
      }
    }
  ],
  "caseStudy": {
    "title": "数据库风险综合案例",
    "background": "业务数据库保存客户隐私数据。",
    "steps": [
      {"label": "Asset", "detail": "客户信息数据库"}
    ],
    "takeaway": "先分析业务风险，再选择与风险匹配的控制措施。"
  }
}
```

上例只是字段示意，不满足完整 `sections[].conceptIds` 与 `concepts[]` 全量覆盖；用于实际导入请使用 `samples/` 下的完整 Day 1 文件。

**必填与约束摘要**：
- 根：`format` = `cissp-lesson`、`schemaVersion` = `1.0`、`lesson`、`sections`（非空）、`concepts`（非空）；可选 `caseStudy`。
- Lesson：`id`、`title`、`week`（>=1）、`day`（1..7）、`domain`（1..8）、`createdAt`（带时区，示例使用 Z）必填；`description` 可选；`linkedQuestionSetId` 可为字符串或 null，也可缺省，推荐显式给出。
- Section：`id`、`title`、`conceptIds`（非空、唯一）必填。
- Concept：`id`、`sectionId`、`term.en`、`term.zh`、`objective`、`definition`、`explanation`、`analogy`、`scenario`、`commonMistake`、`examTip`、`recall.prompt`、`recall.answer` 必填；`controls` 为可选非空字符串数组。
- `caseStudy`（若存在）：`title`、`background`、非空 `steps[]`（`label` / `detail`）及 `takeaway`。
- 纯文本字段不得只有空格；JSON Schema 对空白的 `minLength` 不足以替代 Go 代码的 trim 校验。
- 不接受未声明属性；不得通过删减内容、静默忽略无效知识点或允许缺少引用来“容错”。

## 7. UI/服务契约边界

设计服务能力即可，不强制接口命名；至少提供：预览导入、确认导入、列课程、取课程内容及进度、打开知识点/记录最近访问、更新理解状态、查询关联题集、导航进入既有 quiz。前端可复用 Pinia、Wails Bindings 和现有文件选择逻辑。

- 切换知识点后自检默认为隐藏；关联题集按钮只有可用时才能启动会话。
- 未导入任何课程时应显示明确空状态与“导入 Lesson JSON”入口。
- 进度标签与数据保持一致，不采用仅存内存的完成状态。
- 自检不视为考试，不能污染 Domain/Type 正确率及练习历史。
- 若课程 JSON 含外部链接、脚本、HTML 字符串，一律按普通文本显示；不引入富文本执行通道。

## 8. 验收与实施顺序

权威目标矩阵与执行任务见 [`knowledge-learning-v0.4-goals.md`](knowledge-learning-v0.4-goals.md)。实施建议：
1. 先基于样例及 Schema 写校验和导入冲突测试；
2. migration 003、事务导入、进度读写与备份恢复回归；
3. 课程目录/阅读器与自检交互；
4. 用既有 `set.id` 打通关联题库和考试，不绕过归档检查；
5. 执行 Go、Vue、Windows Wails 构建和打包，并记录原生 UI 验收。

交付状态：本文保持需求基线；P0 实现与实际验收证据见 [v0.2.0 验收记录](knowledge-learning-v0.2-acceptance.md)，未执行项在该记录单独列明。
