# [项目名称]

## 工作流程

本项目使用 **以章程为中心的六层工作流**：

```
需求 → Superpower(头脑风暴→规格) → TaskMaster(任务) → Speckit(澄清→计划→实现) → DDD(代码) → BMAD(开发+测试) → QA(关卡)
```

完整约束请阅读 `.specify/memory/constitution.md`。

## 快速参考

| 阶段 | 工具 | 输出物 |
|-------|------|--------|
| 设计 | Superpower 头脑风暴 | `docs/superpowers/specs/*.spec.md` |
| 拆解 | TaskMaster CLI | `docs/speckit/<功能>/tasks.md` |
| 澄清 | Speckit clarify | tasks.md（追加内容） |
| 计划 | Speckit plan | `docs/superpowers/plans/*.plan.md` |
| 实现 | BMAD agents + DDD | `internal/` + `web/src/` 下的代码 |
| 验证 | QA 关卡 | `bash scripts/qa.sh` |

## 技术栈

<!-- 请填写实际技术栈 -->

## 常用命令

```bash
bash scripts/qa.sh                              # 运行质量关卡
bash scripts/ddd-scaffold.sh <限界上下文名称>      # 脚手架创建新的限界上下文
taskmaster generate --spec <spec> --output <out> --format speckit  # 将规格拆解为任务
```

## 规则

- 禁止在规格之前编写代码（章程 §4）
- `domain/` 层零外部依赖（章程 §1）
- QA 先写测试，Dev 后实现（TDD 红→绿）
- 每个阶段的产物都要 git 提交 —— 可从任意检查点恢复
