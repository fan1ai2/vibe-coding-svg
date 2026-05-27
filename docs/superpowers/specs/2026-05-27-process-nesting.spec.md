# Constitution-Centric 六层嵌套工作流

## 概述

将 TaskMaster + OpenSpec/Speckit + Superpower + DDD + BMAD + QA 六套工具嵌套为统一开发流程。Constitution 为圆心，各工具围绕它分阶段执行，通过标准化文件交接，每个阶段有硬门禁。

## 六层职责

| 层 | 职责 | 触发时机 |
|---|---|---|
| Superpower | brainstorm → 产出 .spec.md 设计文档 | 用户提出需求 |
| TaskMaster | 读 spec → 拆带依赖的 task 树 → tasks.md | spec.md commit 后 |
| Speckit | clarify → plan → implement，管理 task 状态 | tasks.md 产出后 |
| DDD | 战术分层约束代码结构（domain/service/infrastructure/interfaces） | implement 阶段全程 |
| BMAD | dev(Pirlo)/qa(Quinn)/sm(Bob) 三个 agent 执行写码和测试 | implement 阶段内 |
| QA | 编译+测试+覆盖率+checklist 质量门 | 所有 task 完成后 |

## 阶段与产物

```
阶段 1 — Superpower Brainstorming
  输入: 用户口述需求
  输出: docs/superpowers/specs/YYYY-MM-DD-<topic>.spec.md
  执行: Claude + brainstorming skill
  门禁: 用户 approve + git commit

阶段 2 — TaskMaster 拆任务
  输入: .spec.md
  输出: docs/speckit/<feature>/tasks.md
  执行: taskmaster generate --spec ... --output ... --format speckit
  门禁: git commit tasks.md

阶段 3 — Speckit Clarify
  输入: tasks.md
  输出: tasks.md（追加 ## Clarification 段落，不改原始定义）
  执行: Claude + speckit-clarify skill
  门禁: 无未解决 TBD 标记

阶段 4 — Speckit Plan
  输入: tasks.md + .spec.md
  输出: docs/superpowers/plans/YYYY-MM-DD-<topic>.plan.md
  执行: Claude + speckit-plan skill
  规则: 每个 task 标注目标文件路径 + DDD 分层归属
  门禁: git commit plan.md

阶段 5 — TDD 实现
  输入: plan.md + tasks.md + constitution.md
  输出: 代码 + 测试
  执行:
    5a. Quinn(qa) 先写测试 → 🔴 全红
    5b. Pirlo(dev) 写实现 → 🟢 全绿
    5c. Quinn + Pirlo 交替重构
  规则:
    - domain 层先行（entity → value object → aggregate → repository 接口）
    - service/infrastructure/interfaces 层后建
    - 每绿一个 bounded context → git commit
  门禁: 全部测试通过

阶段 6 — QA 质量门
  输入: 全部代码
  执行: bash scripts/qa.sh
  检查项:
    1. openspec validate（spec 合规）
    2. go build ./... + tsc -b（编译）
    3. go test ./... + vitest（测试）
    4. domain 层覆盖率 ≥ 80%
    5. Speckit checklist 全部打勾
  失败路径:
    - 编译失败 → 回到 5b（Pirlo 修代码）
    - 测试失败 → 回到 5b（Pirlo 修实现）
    - 覆盖率不足 → 回到 5a（Quinn 补测试）
    - spec 不合规 → 回到阶段 4（plan 与 spec 重新对齐）
```

## Constitution 结构

`.specify/memory/constitution.md` 分五个段落，每条可执行检查：

```
§1 架构原则
  - DDD 战术分层：domain/ → service/ → infrastructure/ → interfaces/
  - 依赖方向：interfaces → service → domain ← infrastructure
  - domain/ 禁止 import 数据库/Redis/S3/Gin 等外部库
  - service/ 只依赖 domain/
  - 每个 bounded context 独立 package

§2 质量门标准
  - 编译零错误
  - domain 层测试覆盖率 ≥ 80%
  - 每个 API 端点至少 1 条集成测试

§3 技术栈约束
  - 从项目实际栈取（非固定模板）

§4 工作流规则
  - 任何功能必须先有 spec 再写代码
  - 不可跳过阶段
  - 所有阶段产物 git commit

§5 禁止事项
  - 禁止跳过 spec 直接写代码
  - 禁止 domain 层 import 基础设施包
  - 禁止一个 PR 跨多个 bounded context
```

## DDD 目录模板

```
project/
├─ internal/
│   ├─ domain/<bounded-context>/     # 核心层，零外部依赖
│   │   ├─ entity.go
│   │   ├─ value_object.go
│   │   ├─ aggregate.go
│   │   └─ repository.go            # 接口定义
│   ├─ service/<bounded-context>/    # 用例编排层
│   │   ├─ command.go
│   │   └─ query.go
│   ├─ infrastructure/              # 实现层
│   │   ├─ persistence/
│   │   └─ storage/
│   └─ interfaces/                  # 对外适配
│       ├─ http/
│       └─ worker/
├─ web/src/
│   ├─ domain/                      # 前端纯逻辑
│   ├─ features/<bounded-context>/  # 按功能分包
│   └─ shared/                      # 通用组件
└─ docs/
    ├─ superpowers/specs/           # .spec.md
    ├─ superpowers/plans/           # .plan.md
    └─ speckit/<feature>/           # tasks.md
```

## BMAD Agent 定义

三个 agent，存储在 `.bmad-core/agents/`：

- **dev.md (Pirlo)**: 全栈实现者。输入 plan + tasks + constitution + QA 的测试。按 DDD 分层顺序实现。不写测试。
- **qa.md (Quinn)**: 测试工程师。输入 spec + tasks + constitution。dev 动手前先写测试。domain 层覆盖率 ≥ 80%。
- **sm.md (Bob)**: Scrum Master。调用 TaskMaster CLI 拆 task。跟踪 tasks.md 状态。不写代码。

Agent 必须适配当前项目技术栈（删除模板中的 AWS/Serverless/Cognito 引用）。

## TaskMaster CLI

```bash
taskmaster generate \
  --spec docs/superpowers/specs/<file>.spec.md \
  --output docs/speckit/<feature>/tasks.md \
  --format speckit
```

无 CLI 环境时手动按同等格式拆解。

## 新项目接入步骤

在已有项目（已有 .specify/ 和 .bmad-core/）中接入：

1. 适配 constitution.md（填入实际技术栈和 DDD 规则）
2. 适配 BMAD agent 三个文件（替换技术栈引用）
3. 修改 quality-gate.sh 中的检查命令
4. 首次使用时走完整六阶段验证流程

## 日常使用

> 用户说需求 → Superpower brainstorming → 产出 spec.md → TaskMaster 拆 task → Speckit clarify → Speckit plan → Quinn 写测试(红) → Pirlo 写实现(绿) → QA gate → 完成

每个阶段产物 git commit，任何时候中断可从上次 checkpoint 继续。
