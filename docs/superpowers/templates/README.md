# 六层工作流 — 快速参考

## 一句话总结

> 需求 → 头脑风暴 → 规格说明 → TaskMaster → 任务拆解 → 澄清 → 计划 → TDD(红→绿) → QA 关卡 → 完成

## 每个功能会产生的文件

```
docs/
├── superpowers/specs/YYYY-MM-DD-<话题>.spec.md   # 第一阶段输出
├── superpowers/plans/YYYY-MM-DD-<话题>.plan.md    # 第四阶段输出
└── speckit/<功能>/
    └── tasks.md                                     # 第二~三阶段输出

internal/
├── domain/<限界上下文>/        # 第五阶段：实体、值对象、聚合、仓库接口
├── service/<限界上下文>/       # 第五阶段：命令处理器、查询
├── infrastructure/                  # 第五阶段：仓库实现、适配器
└── interfaces/                      # 第五阶段：HTTP 处理器、Worker 处理器
```

## 阶段检查清单

- [ ] **阶段 1**：超级能力头脑风暴 → 用户批准规格 → 提交 `.spec.md`
- [ ] **阶段 2**：TaskMaster 任务拆解 → 提交 `tasks.md`
- [ ] **阶段 3**：Speckit 澄清 → 追加验收标准 → 无残留 TBD
- [ ] **阶段 4**：Speckit 计划 → 文件路径 + 分层分配 → 提交 `.plan.md`
- [ ] **阶段 5**：Quinn 编写测试（红灯）→ Pirlo 实现（绿灯）→ 每个限界上下文提交一次
- [ ] **阶段 6**：`bash scripts/qa.sh` → 所有关卡通过 → 完成

## Agent 角色

| Agent | 角色 | 负责 | 不负责 |
|-------|------|------|----------|
| Bob (sm) | Scrum Master | TaskMaster CLI，跟踪 tasks.md 状态 | 编写代码或测试 |
| Quinn (qa) | QA 工程师 | 在实现之前编写测试 | 编写实现代码 |
| Pirlo (dev) | 开发者 | 根据 QA 的测试实现代码，DDD 分层 | 编写测试或管理任务 |

## 章程 §1 速查表

```
interfaces → service → domain ← infrastructure
```

- `domain/`：零外部导入。仅使用接口。
- `service/`：仅依赖 `domain/`
- `infrastructure/`：实现 `domain/` 接口
- `interfaces/`：HTTP/Worker 适配器，依赖 `service/`
- 依赖注入仅在 `cmd/server/main.go` 中进行
