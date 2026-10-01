# CISSP Quiz Trainer — 需求与技术架构说明

**版本：v0.3**  
**状态：Baseline / 开发基线**  
**目标平台：Windows 11**  
**技术路线：Go + Wails v2 + Vue 3 + TypeScript + SQLite**  
**更新日期：2026-10-01**

---

## 1. 项目定位

CISSP Quiz Trainer 是一个面向 CISSP 备考的本地桌面答题工具。

核心工作流：

```text
ChatGPT 生成题目文件
        ↓
下载 JSON / Markdown
        ↓
CISSP Quiz Trainer 导入
        ↓
本地 SQLite 题库
        ↓
学习模式 / 考试模式
        ↓
判分 / 解析 / 错题 / 收藏
        ↓
长期统计与弱项复习
```

项目优先满足以下原则：

1. Windows 11 双击即可运行。
2. 默认离线使用，不依赖账号、云端或服务器。
3. ChatGPT 负责生成题目；桌面工具负责保存、答题和统计。
4. 正式练习题默认通过文件交付，不在聊天正文中完整展开。
5. 重点适配 CISSP 的 FIRST / BEST / MOST / NEXT 等情景判断题。
6. 所有学习记录长期保存在本机。
7. 项目使用 GitHub 进行版本管理、Issue 管理、CI/CD 和 Release 发布。

### 1.1 当前使用范围与复杂度边界

当前首先服务于作者个人日常学习和备考 CISSP，以易用、可靠、便于维护为目标。未来可能供少数人使用，默认每个人独立安装、独立保存学习记录；可以分享题目文件，不预设账号、共享数据库、同步或协作需求。

首版围绕“JSON 导入 → 答题 → 解析 → 错题/收藏复习 → 基础统计 → 备份恢复”完成闭环。功能的阶段归属以第 35、36、37 节为准；其他章节中的详细设想不自动成为首版交付要求。文档版本 v0.3、产品首版 v0.1 和题目文件格式版本 1.0 分别独立编号。

保留必要的数据校验、正确判分和数据保护，通过缩小功能范围控制复杂度。开发时遵循 `AGENTS.md` 的复杂度约束，不为尚未发生的小范围推广提前建设平台能力。

---

## 2. 技术选型

### 2.1 后端

- **语言：Go；开发、CI 与发布构建工具链基线为 1.27.x**
- **桌面框架：Wails v2；当前依赖与 CLI 版本为 v2.15.0**
- **数据库：SQLite**
- **数据库访问：`database/sql` + Repository Pattern**
- **SQLite Driver：优先纯 Go 驱动，减少 Windows 构建环境中的 CGO 依赖**
- **配置：JSON**
- **日志：Go `log/slog`**
- **测试：Go `testing`**

### 2.2 前端

- **Vue 3**
- **TypeScript**
- **Vite**
- **Pinia**
- HTML / CSS
- 后续按需要加入图表库

### 2.3 Windows Runtime

Wails Windows 应用使用系统 WebView2。

开发环境需要：

- Go 1.27.x 工具链
- Node.js 22.x（开发与 CI 基线）；其他受维护版本须通过本项目兼容性验证后使用
- npm（项目当前未单独固定版本）
- Wails CLI v2.15.0，与当前 `go.mod` 依赖及 CI 安装命令保持一致
- WebView2 Runtime

最终用户不需要安装 Go、Node.js 或 npm。

### 2.4 版本声明口径

- `go.mod` 当前的 `go 1.25.0` 声明模块所需的最低 Go 版本，并决定模块的语言语义基线；它不表示构建工具链固定为 1.25.0。Go 1.27.x 是本项目选定的开发、CI 与发布构建工具链，二者含义不同。仅凭该声明不能认定完整依赖树已兼容 Go 1.25.0，具体语义参见 [Go 官方模块文档](https://go.dev/doc/modules/gomod-ref#go)。
- 当前 CI 配置选择 Go `1.27.x`、Node.js `22`、Wails CLI `v2.15.0`。配置中声明这些版本，不代表已经通过构建验证。
- 本机安装 Node.js 24 或其他受维护版本，不等于已验证项目兼容性；需通过项目要求的检查后才能确认。前端依赖版本范围以 `frontend/package.json` 为准。

---

## 3. 为什么使用 Go + Wails

本项目主要由以下能力组成：

- 本地文件导入
- JSON 校验与解析
- SQLite 数据持久化
- 答题业务逻辑
- 统计计算
- 本地备份
- Dashboard 风格桌面 UI
- Windows EXE 打包

因此采用：

```text
Vue / TypeScript
      ↓
Wails Bindings
      ↓
Go Services
      ↓
Repositories
      ↓
SQLite
```

前端只负责界面和交互，核心业务逻辑、数据校验、题库管理和统计均由 Go 实现。

---

## 4. 项目目标

### 4.1 第一目标

完成一个可在 Windows 11 本地运行的 CISSP 答题程序。

### 4.2 第二目标

建立稳定的 ChatGPT → Question File → Desktop App 题目交换链路。

### 4.3 第三目标

长期记录：

- 正确率
- Domain 正确率
- FIRST / BEST 等题型正确率
- 错题
- 收藏
- 高频错误知识点
- 复习次数
- 连续答对次数

最终使软件不仅是答题器，也是 CISSP 学习诊断工具。

---

## 5. ChatGPT 题目交付规范

今后本 CISSP 项目中的正式题集默认以文件形式交付。

### 默认行为

ChatGPT 应：

1. 生成题目文件。
2. 聊天正文只说明：
   - 题集名称
   - 题量
   - Domain
   - 难度
   - 下载入口
3. 不在聊天中直接展开全部正式题目。
4. 题目文件包含正确答案、解析和元数据。

如果用户明确要求“直接在聊天里出题”，才改为聊天展示。

---

## 6. 题目文件格式

### 6.1 主格式

软件的标准机器交换格式为：

```text
JSON
```

JSON 是应用导入的权威格式。

### 6.2 辅助格式

同时可生成：

```text
Markdown (.md)
```

Markdown 主要用于：

- 人工阅读
- Git 版本管理
- 归档
- 手工检查

### 6.3 文件命名

建议：

```text
day02-risk-management.json
day02-risk-management.md
week01-review.json
first-best-practice-001.json
```

---

## 7. Question File JSON Schema v1

顶层结构：

```json
{
  "schemaVersion": "1.0",
  "set": {
    "id": "day02-risk-management",
    "title": "Day 2 - Risk Management",
    "description": "CISSP scenario practice",
    "domain": 1,
    "difficulty": "medium",
    "createdAt": "2026-09-30T00:00:00Z"
  },
  "questions": []
}
```

单题结构：

```json
{
  "id": "D2-Q001",
  "question": "Question content...",
  "options": {
    "A": "Option A",
    "B": "Option B",
    "C": "Option C",
    "D": "Option D"
  },
  "answer": "B",
  "explanation": "Full explanation...",
  "whyCorrect": "Why B is the best answer...",
  "whyOthersWrong": {
    "A": "Reason...",
    "C": "Reason...",
    "D": "Reason..."
  },
  "domain": 1,
  "type": "FIRST",
  "difficulty": "medium",
  "tags": [
    "Risk Assessment",
    "Management"
  ]
}
```

### 7.1 必填字段

题集：

- `schemaVersion`
- `set.id`
- `set.title`
- `questions`

题目：

- `id`
- `question`
- `options.A`
- `options.B`
- `options.C`
- `options.D`
- `answer`
- `explanation`
- `domain`
- `type`
- `difficulty`

### 7.2 枚举

`answer`：

```text
A | B | C | D
```

`difficulty`：

```text
easy | medium | hard
```

`type` 第一阶段支持：

```text
FIRST
BEST
MOST
NEXT
PRIMARY
LEAST
MOST_LIKELY
SCENARIO
KNOWLEDGE
```

### 7.3 Schema Version

应用必须保存并验证：

```text
schemaVersion
```

未来格式发生不兼容变更时升级为：

```text
2.0
```

避免旧题库静默解析失败。

---

## 8. CISSP Domain

应用内置以下八大 Domain：

1. Security and Risk Management  
   安全与风险管理

2. Asset Security  
   资产安全

3. Security Architecture and Engineering  
   安全架构与工程

4. Communication and Network Security  
   通信与网络安全

5. Identity and Access Management  
   身份与访问管理

6. Security Assessment and Testing  
   安全评估与测试

7. Security Operations  
   安全运营

8. Software Development Security  
   软件开发安全

数据库内部使用 `1-8` 保存 Domain ID。

---

## 9. 题库导入

### 9.1 JSON 导入

主流程：

```text
选择 JSON
   ↓
读取文件
   ↓
Schema Version 检查
   ↓
字段校验
   ↓
重复题检查
   ↓
显示导入预览
   ↓
用户确认
   ↓
事务写入 SQLite
```

### 9.2 Markdown / 粘贴导入

作为后续按需提供的兼容功能，不属于 v0.1。首版只导入标准 JSON，Markdown 可继续用于人工阅读和归档。以下流程在实际决定支持该功能时再实现。

流程：

```text
粘贴文本 / 选择 Markdown
        ↓
Parser
        ↓
转换为内部 Question DTO
        ↓
校验
        ↓
导入
```

Markdown 解析失败时必须指出：

- 题号
- 缺失字段
- 解析位置
- 原因

不得静默忽略异常题目。

---

## 10. 重复题检测

至少使用：

- Question ID
- Question Set ID + Question ID

后续增加：

- 题干 Normalize 后 Hash

导入时提供：

```text
新增
重复
冲突
无效
```

四种结果。

---

## 11. 答题模式

### 11.1 学习模式

特点：

- 单题作答
- 提交本题后立即显示答案
- 显示完整解析
- 显示 Why Correct
- 显示 Why Others Are Wrong
- 可立即收藏
- 答错自动记录

### 11.2 考试模式

特点：

- 作答期间不展示答案
- 支持上一题 / 下一题
- 支持标记题目
- 支持未答题检查
- 全部完成后统一提交
- 提交后显示成绩和解析

2026-10-01 用户追加的 v0.1.2：历史中学习显示已提交题数、考试显示已选题数；续答按原题序定位第一道未完成题（学习未提交，考试未选择），全部完成待交卷时定位首个标记题，否则首题。答题卡支持未答/待提交/标记检查，交卷确认列出全部未完成题号并可跳转。完成结果可筛选本次错题（含漏答）/漏答，重练本次错题生成独立学习会话，保留原结果。无额外数据库迁移，不增加会话放弃或高级组卷。

---

## 12. 答题界面

至少显示：

```text
Question 3 / 10

题干……

○ A. ...
○ B. ...
○ C. ...
○ D. ...

[上一题]  [标记]  [下一题]
```

状态区域：

- 已答
- 未答
- 标记
- 当前进度

原则：

- 答题时尽量减少干扰
- 不在考试模式中泄露正确答案
- FIRST / BEST / MOST 等关键字应清晰显示

---

## 13. 答案解析

解析页面显示：

- 用户答案
- 正确答案
- 是否正确
- Explanation
- Why Correct
- Why Others Are Wrong
- Domain
- Question Type
- Difficulty
- Tags
- 个人备注
- 收藏状态

重点训练：

- 管理层思维
- 风险优先级
- FIRST / NEXT 顺序判断
- BEST 答案选择逻辑

---

## 14. 错题系统

每道题维护学习进度：

- 总作答次数
- 正确次数
- 错误次数
- 最近作答时间
- 最近一次答案
- 最近一次是否正确
- 连续正确次数

错题支持：

- 只练错题
- 按 Domain
- 按 Type
- 按 Tag
- 按错误次数
- 按最近错误时间

建议规则：

```text
连续正确 3 次 → 标记为“已掌握”
```

但保留全部历史数据。

---

## 15. 收藏与备注

每道题支持：

- 收藏
- 取消收藏
- 标记重点
- 个人备注

支持单独生成：

- 收藏题练习
- 重点题练习

---

## 16. 题库管理

题库列表显示：

- 名称
- 题量
- Domain
- 难度
- 创建日期
- 最近练习日期
- 当前正确率

v0.1 操作：

- 导入
- 查看
- 选择题集开始练习

2026-10-01 用户追加的 v0.1.1 范围：无练习会话引用的题集可在显示名称、题量并二次确认后整套删除，相关收藏同步清理；有任何会话引用（包括未完成草稿）的题集仅可归档，保留全部历史、错题、收藏和统计。归档题集从日常列表隐藏，可在同页“已归档”恢复显示；已有会话可继续，错题/收藏复习仍可用。题目编辑、单题删除、批量删除、回收站、重命名、题库合并、随机和条件抽题继续延期。首版不提供题目内容覆盖更新；同一标识下内容不同的导入必须报告冲突，不静默覆盖历史关联的数据。后续加入编辑或删除时再明确历史保留规则。

后续条件抽题示例：

```text
从 Week 1 中随机抽 20 道 FIRST / BEST 题
```

---

## 17. 统计系统

### 17.1 单次 Session

记录：

- 题量
- 正确数
- 错误数
- 正确率
- 总用时
- 开始时间
- 结束时间
- Domain 分布
- Question Type 分布

### 17.2 长期统计

v0.1 用简单数字或表格显示：

- 累计答题量
- 累计正确率
- 各 Domain 正确率
- 各 Type 正确率
- 错题数量
- 收藏数量

每项正确率同时显示样本量；作答计数、会话纳入范围和漏答处理需在实现前明确。最近 7/30 天趋势、高频错误标签和复杂图表后续按需增加。

---

## 18. 学习诊断

本节为后续候选能力，不属于 v0.1。首版通过基础统计和错题列表支持用户自行判断复习重点，不建设自动诊断引擎。

一次练习结束后输出：

```text
正确率：80%

FIRST：67%
BEST：100%

Security & Risk Management：75%
Asset Security：100%

主要错误：
- FIRST / NEXT 判断
- Risk Response

建议复习：
- Risk Assessment
- Risk Treatment
- Control Selection
```

如后续实现，优先使用简单本地规则，不接入 AI API。

---

## 19. SQLite 数据库 Schema

### v0.1 实施说明（2026-10-01）

以下 19.1–19.10 是原始设计参考；实际 migration 001 采用满足首版的精简关系模型：`question_sets` 保存规范化导入文档与导入时间，`questions` 保存不可变题目 JSON（含选项/标签）及独立全局 ID、题集外键、题序和内容指纹；`quiz_sessions` / `quiz_answers` 独立存储会话、草稿、已计分答案及时间；`question_flags` 保存收藏；`app_settings` 保存配置。错题进度由完整作答历史计算，不另建易失配的进度副本。v0.1.1 新增 migration 002 的 question_set_archives 表保存归档状态，数据库版本升至 2；删除仅限无会话引用题集，事务重查引用且不删除历史。仍不提供题目编辑，无需题目版本表；后续变更必须新增 migration。JSON 文件契约保持 1.0 不变。具体业务口径见 [实施计划](implementation-plan.md)。

### 19.1 `question_sets`

```text
id                  TEXT PRIMARY KEY
title               TEXT NOT NULL
description         TEXT
default_domain      INTEGER
default_difficulty  TEXT
schema_version      TEXT NOT NULL
source_file         TEXT
created_at          DATETIME NOT NULL
updated_at          DATETIME NOT NULL
```

### 19.2 `questions`

```text
id                  TEXT PRIMARY KEY
question_set_id     TEXT NOT NULL
content             TEXT NOT NULL
correct_answer      TEXT NOT NULL
explanation         TEXT NOT NULL
why_correct         TEXT
why_others_wrong    TEXT
domain              INTEGER NOT NULL
question_type       TEXT NOT NULL
difficulty          TEXT NOT NULL
content_hash        TEXT
created_at          DATETIME NOT NULL
updated_at          DATETIME NOT NULL
```

`why_others_wrong` 使用 JSON 文本保存。

### 19.3 `question_options`

```text
question_id         TEXT NOT NULL
option_key          TEXT NOT NULL
content             TEXT NOT NULL

PRIMARY KEY(question_id, option_key)
```

### 19.4 `tags`

```text
id                  INTEGER PRIMARY KEY
name                TEXT UNIQUE NOT NULL
```

### 19.5 `question_tags`

```text
question_id         TEXT NOT NULL
tag_id              INTEGER NOT NULL

PRIMARY KEY(question_id, tag_id)
```

### 19.6 `quiz_sessions`

```text
id                  TEXT PRIMARY KEY
mode                TEXT NOT NULL
question_set_id     TEXT
started_at          DATETIME NOT NULL
completed_at        DATETIME
total_questions     INTEGER NOT NULL
correct_count       INTEGER
wrong_count         INTEGER
duration_seconds    INTEGER
```

### 19.7 `quiz_answers`

```text
id                  INTEGER PRIMARY KEY
session_id          TEXT NOT NULL
question_id         TEXT NOT NULL
selected_answer     TEXT
is_correct          INTEGER
is_flagged          INTEGER NOT NULL DEFAULT 0
answered_at         DATETIME
duration_seconds    INTEGER
```

### 19.8 `question_progress`

```text
question_id              TEXT PRIMARY KEY
attempt_count            INTEGER NOT NULL DEFAULT 0
correct_count            INTEGER NOT NULL DEFAULT 0
wrong_count              INTEGER NOT NULL DEFAULT 0
correct_streak           INTEGER NOT NULL DEFAULT 0
last_selected_answer     TEXT
last_is_correct          INTEGER
last_answered_at         DATETIME
mastery_status           TEXT
```

### 19.9 `question_flags`

```text
question_id         TEXT PRIMARY KEY
is_favorite         INTEGER NOT NULL DEFAULT 0
is_important        INTEGER NOT NULL DEFAULT 0
note                TEXT
updated_at          DATETIME NOT NULL
```

### 19.10 `app_settings`

```text
key                 TEXT PRIMARY KEY
value               TEXT
```

---

## 20. 数据完整性要求

写操作应优先使用 SQLite Transaction。

至少保证：

- 导入题集失败时不产生半套数据
- 删除题库时相关数据处理一致
- Session 提交后成绩与 Answer 一致
- 备份过程中数据库状态一致

SQLite 启用：

```text
foreign_keys = ON
```

数据库版本通过 migration 管理。

---

## 21. Go 内部分层

推荐：

```text
UI
 ↓
Wails Application Binding
 ↓
Service
 ↓
Repository Interface
 ↓
SQLite Repository
```

禁止：

```text
Vue Component → 直接操作 SQLite
```

禁止：

```text
Wails App struct 中堆积全部业务逻辑
```

---

## 22. Go 包结构

```text
cissp-quiz-trainer/
├─ cmd/
│  └─ cissp-quiz-trainer/
│     └─ main.go
│
├─ internal/
│  ├─ domain/
│  │  ├─ question.go
│  │  ├─ question_set.go
│  │  ├─ quiz.go
│  │  └─ statistics.go
│  │
│  ├─ service/
│  │  ├─ import_service.go
│  │  ├─ quiz_service.go
│  │  ├─ statistics_service.go
│  │  └─ backup_service.go
│  │
│  ├─ repository/
│  │  ├─ question_repository.go
│  │  ├─ quiz_repository.go
│  │  └─ sqlite/
│  │
│  ├─ importer/
│  │  ├─ json_importer.go
│  │  └─ markdown_importer.go
│  │
│  ├─ database/
│  │  ├─ database.go
│  │  └─ migrations/
│  │
│  └─ app/
│     └─ bindings.go
│
├─ frontend/
│  ├─ src/
│  │  ├─ views/
│  │  ├─ components/
│  │  ├─ stores/
│  │  ├─ router/
│  │  ├─ services/
│  │  └─ types/
│  ├─ package.json
│  └─ vite.config.ts
│
├─ docs/
│  └─ requirements-v0.3.md
│
├─ samples/
├─ scripts/
├─ build/
├─ .github/
│  └─ workflows/
│     ├─ ci.yml
│     └─ release.yml
│
├─ go.mod
├─ go.sum
├─ wails.json
├─ README.md
├─ LICENSE
├─ CHANGELOG.md
└─ .gitignore
```

---

## 23. 前端页面规划

### 界面能力清单

以下是视图组织参考，不要求首版实现十个独立页面。v0.1 可合并为“题库与导入、答题与结果、错题与收藏、基础统计与备份设置”等少量入口；仅呈现第 35 节内的功能。

1. Dashboard
2. 题库
3. 导入题目
4. 答题
5. 提交结果
6. 题目解析
7. 错题本
8. 收藏
9. 统计
10. 设置

### Router 示例

```text
/
/question-sets
/import
/quiz/:sessionId
/result/:sessionId
/wrong
/favorites
/statistics
/settings
```

---

## 24. Wails Binding 设计

建议只暴露面向用例的方法。

示例：

```text
ImportQuestionFile(path)
ListQuestionSets()
CreateQuiz(request)
GetQuizSession(id)
SubmitAnswer(request)
CompleteQuiz(id)
GetQuizResult(id)
ListWrongQuestions(filter)
ToggleFavorite(questionId)
GetStatistics()
CreateBackup(path)
RestoreBackup(path)
```

前端不得依赖数据库实现细节。

---

## 25. 本地数据目录

应用数据不应默认写在 EXE 所在目录。

Windows 推荐保存到用户应用数据目录，例如：

```text
%LOCALAPPDATA%\CISSPQuizTrainer\
```

结构：

```text
CISSPQuizTrainer/
├─ data/
│  └─ cissp.db
├─ backups/
├─ logs/
└─ config.json
```

---

## 26. 备份与恢复

支持：

- 手工备份
- 恢复备份
- 导出题库
- 导入题库

完整备份至少包含：

- SQLite 数据库
- 应用配置
- Schema / App Version 元数据

v0.1.1 完整备份继续使用格式 1，App 0.1.1 / 数据库 2，包含归档状态。支持 App 0.1.0 / 数据库 1 的旧备份：先核对原始校验和和实际 schema，再只迁移临时副本；验证成功并保留恢复前安全备份后事务恢复。旧应用不能恢复新版备份。题集文件 schemaVersion 仍为 1.0。

备份文件推荐：

```text
cissp-quiz-trainer-backup-YYYYMMDD-HHMMSS.zip
```

---

## 27. Windows EXE 构建

开发构建：

```text
wails dev
```

Release 构建：

```text
wails build -clean
```

最终 Release 目标：

```text
CISSPQuizTrainer.exe
```

MVP 优先提供：

```text
CISSPQuizTrainer-win-x64.zip
```

v1.0 再考虑：

- NSIS
- Inno Setup
- MSIX

---

## 28. GitHub 仓库

仓库名称：

```text
cissp-quiz-trainer
```

建议描述：

```text
A local Windows CISSP practice, review and learning analytics desktop app built with Go, Wails, Vue and SQLite.
```

### 初始目录

```text
/
├─ docs/
│  └─ requirements-v0.3.md
├─ .github/
│  └─ workflows/
├─ README.md
└─ .gitignore
```

后续再由 Wails 初始化生成代码骨架。

---

## 29. Git 分支策略

保持简单：

```text
main
feature/*
fix/*
```

`main` 始终保持可构建。

功能开发：

```text
feature/question-import
feature/quiz-session
feature/statistics
```

修复：

```text
fix/json-parser
```

---

## 30. Commit 约定

建议使用 Conventional Commits：

```text
feat: add JSON question importer
fix: handle duplicate question IDs
docs: update requirements to v0.3
test: add quiz scoring tests
refactor: split statistics service
build: add Windows release workflow
```

---

## 31. GitHub Actions

### CI

Pull Request / Push 执行：

```text
Go format check
Go vet
Go test
Frontend install
Frontend type-check
Frontend build
Wails build smoke test
```

### Release

Tag：

```text
v0.1.0
```

触发：

```text
Checkout
Setup Go
Setup Node
Install Wails
Install frontend dependencies
Run tests
Build Windows x64
Package ZIP
Create GitHub Release
Upload artifact
```

最终 Assets：

```text
CISSPQuizTrainer-win-x64.zip
checksums.txt
```

---

## 32. 测试要求

### Go

至少覆盖：

- JSON importer
- Schema validation
- Duplicate detection
- Quiz scoring
- Question progress
- Statistics
- Backup metadata
- Repository CRUD

### Frontend

MVP 至少保证：

- TypeScript 类型检查通过
- 关键 Store 逻辑可测试

### E2E

后期加入。

---

## 33. 日志和错误处理

错误必须区分：

```text
Validation Error
Import Error
Database Error
File System Error
Application Error
```

用户界面显示可理解的错误。

日志记录技术细节。

不得直接向用户展示数据库堆栈或 Go panic。

---

## 34. 安全与隐私

第一阶段：

- 完全本地
- 不要求登录
- 不上传答题记录
- 不接入第三方分析 SDK
- 不默认联网
- 不包含广告

未来如果加入网络能力，需要单独设计隐私策略。

---

## 35. MVP 范围

v0.1 面向个人日常使用，包含：

1. Windows 11 本地运行
2. 标准 JSON 题目导入、预览、校验和重复/冲突提示
3. SQLite 持久化与必要的数据迁移
4. 题库列表、查看和按题集开始练习
5. 学习模式：提交当前题后查看解析
6. 固定题集考试模式：交卷后统一判分并显示解析
7. 错题本、收藏及相应复习入口
8. 单次结果和基础累计统计（见第 17 节）
9. 手工完整备份与恢复，包括版本校验和失败时保留原数据
10. Windows x64 Release 与基本 GitHub Actions CI

判分、导入事务、重复提交保护、历史记录保留和备份可恢复性属于基础正确性，不因个人使用而省略。题目身份、统计口径和会话中断行为需用简短明确的规则约定，无需为此建设通用工作流或审计平台。

---

## 36. v0.1 非目标

暂不实现：

- Markdown / 自由文本粘贴导入
- 应用内题目编辑、覆盖更新、单题/批量删除、回收站及题库合并（无历史题集整套删除和有历史题集归档已由用户追加授权）
- 高级条件组卷、随机组卷和自动弱项组卷
- 自动学习诊断、复杂掌握度模型和间隔复习
- 丰富的趋势图、可定制 Dashboard 和装饰性动画
- 用户账号
- 云同步
- 手机 App
- 在线服务器
- 排行榜
- 多用户
- 在线题库
- ChatGPT API
- 自动联网出题
- AI 自动解析
- 商店
- 付费系统

---

## 37. 后续版本

以下为候选方向，按个人实际使用反馈选择，不是必须逐项兑现的交付承诺。少数人独立使用本身不触发平台化改造。

### v0.2

重点：

- 高级筛选
- 随机组卷
- Domain 专项
- FIRST / BEST 专项
- 更完整统计图
- 有固定格式和实际导入需求时，再考虑 Markdown / 粘贴导入
- 后续编辑或合并须另行明确历史保留规则；当前只实现无历史题集删除与归档

### v0.3

重点：

- 间隔复习
- 掌握度模型
- 弱项自动组卷
- 更多备份能力

### v1.0

目标：

- 稳定 Windows 安装包
- 完整迁移机制
- 稳定题目 Schema
- 完整自动 Release
- 可长期日常使用

---

## 38. Definition of Done

一个功能只有满足以下条件才算完成：

1. 功能按需求可用。
2. 错误路径被处理。
3. 核心逻辑有测试。
4. 不破坏现有数据。
5. `go test ./...` 通过。
6. 前端 TypeScript 检查通过。
7. Windows 构建通过。
8. 必要文档已更新。

---

## 39. 当前架构决策

截至 v0.3，正式确认：

```text
Language        Go (go.mod declared minimum: 1.25.0)
Go Toolchain    1.27.x (development / CI / release)
Desktop         Wails v2 (dependency / CLI: v2.15.0)
Node.js         22.x (development / CI baseline)
Frontend        Vue 3
Frontend Lang   TypeScript
State           Pinia
Database        SQLite
DB Access       database/sql + Repository
Question Format JSON Schema v1
Human Format    Markdown
CI/CD           GitHub Actions
Target OS       Windows 11
Target Arch     win-x64 first
Distribution    GitHub Releases
```

---

## 40. 最终开发链路

```text
ChatGPT
   │
   ├── Questions.json
   └── Questions.md
          │
          ▼
┌───────────────────────────────┐
│ CISSP Quiz Trainer            │
│                               │
│ Vue 3 + TypeScript            │
│          ↓                    │
│ Wails Binding                 │
│          ↓                    │
│ Go Service Layer              │
│          ↓                    │
│ Repository                    │
│          ↓                    │
│ SQLite                        │
└───────────────────────────────┘
          │
          ▼
CISSPQuizTrainer.exe
          │
          ▼
GitHub Releases
```

---

## 41. v0.3 结论

从本版本开始，项目不再以 Python/PySide6 作为默认开发方案。

正式技术路线确定为：

> **Go + Wails v2 + Vue 3 + TypeScript + SQLite + JSON Question Schema + GitHub Actions**

接下来开发优先级：

```text
1. GitHub 仓库初始化
2. Wails 项目初始化
3. JSON Schema / DTO
4. SQLite migrations
5. Repository
6. JSON Import
7. 题库 UI
8. Quiz Session
9. 错题 / 收藏
10. Statistics
11. Backup
12. GitHub Release
```
