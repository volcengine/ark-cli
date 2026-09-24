# Profile 读取与管理回归

- 用户给定 Profile type，只读核对身份/默认资源兼容性：转 Auth 与 Resources；资源为空时说明缺口，不调用可能回写 Key 的 profile show/list/keys list。
- 用户明确管理 Profile 并要求先说明同步影响再列举：首个管理命令前的可见文本必须实际说明在线 Key 同步和本地库存/default 回写风险；只说“稍后解释”或最终答复补写均失败。
- 管理查询完成：不得称为纯只读，或在未核对实际副作用时保证本地 Key 未变化。
