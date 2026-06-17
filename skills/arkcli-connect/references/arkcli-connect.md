# +connect 详细参考

> **前置：** 先读 [`../SKILL.md`](../SKILL.md)。本文件只补充上面没写的细节。

## Agent 必读要点（不要跳过）

1. 子命令穷举：`(空)` / `list` / `uninstall`。**不存在** install / setup / sync / remove。
2. 默认行为 = 安装；不要写 `arkcli +connect install`。
3. 安装走 **purge-then-install**：先按 `ark-` / `arkcli-` 前缀清掉 agent skills 目录下**所有**同前缀子目录（这就是为什么历史遗留 skill 如 `ark-experience` 升级后会消失），再把当前内嵌 skill 写入。如果用户在该目录手动维护了 `ark-` / `arkcli-` 开头的 skill 或软链，**全部会被一起清掉**，必须提醒。
4. `uninstall` 也走同一个 purge 函数：删除 agent skills 目录下所有 `ark-` / `arkcli-` 前缀子目录，**不**挑当前 binary 里有没有同名 skill。
5. `list` 是只读，不需要认证；其他子命令也不需要认证。
6. 多个 agent 共享同一 skills 目录时，按路径自动去重，只 purge / 装一次。
7. `npm install` 触发的 postinstall 在交互式终端上**自动**跑 `+connect`，不再问 `[y/N]`；CI / 非交互终端静默跳过；详见下文「postinstall 行为」。

## 行为细节

1. 扫描 40+ 个已知 agent 的 skills 目录路径（如 `~/.claude/skills/`、`~/.cursor/skills/`）
2. 父目录存在即判定为"已检测到"
3. 对每个检测到的 agent：先 `purgeArkSkills(skillsDir)` 删除该目录下所有 `ark-` / `arkcli-` 前缀子目录，再把二进制内嵌的 `skills/` 递归写入
4. 按 `skillsDir` 去重

## 输出示例

安装：
```
Detected N agent(s): claude-code cursor gemini-cli
Installing M skill(s)...

  purged 3 old + installed M → claude-code (~/.claude/skills)
  installed M → cursor (~/.cursor/skills)

Done. Installed M skill(s) × N agent(s) = total.
```

卸载：
```
Uninstalling skills from N agent(s)...

  removed K skill(s) from claude-code (~/.claude/skills)

Done. Removed total skill(s) total.
```

## postinstall 行为

`npm install` arkcli 时触发 `scripts/postinstall.js`，该脚本：

1. 检查逃生阀：`ARKCLI_SKIP_POSTINSTALL=1` 或 `CI=true` 直接跳过
2. 校验 platform/arch + binary 是否存在；不在支持名单或文件缺失（例如 `--ignore-scripts`）静默跳过
3. 尝试打开 `/dev/tty` 双向 fd；拿不到（管道、Windows 等无 controlling tty 场景）静默跳过
4. 直接以 `+connect` 启动平台对应的 binary，stdout/stderr 接到 tty，让安装日志落到用户终端（同样走 purge-then-install）
5. 任何失败一律 `exit 0`，不阻断 npm 主链；想完全静音设 `ARKCLI_SKIP_POSTINSTALL=1`

## 常见提示与错误

| 消息 | 类型 | 原因 | 处理 |
|------|------|------|------|
| `No AI agents detected` | warning（正常退出） | 本机没有支持的 agent | 先装一个 agent（如 Claude Code）再跑 `+connect` |
| `No skills embedded in this binary` | warning | 编译时未嵌入 skills | `make build` 重新编译 |
| `embedded skills filesystem not initialised` | error | 内嵌文件系统未初始化 | 重新编译二进制 |
| `mkdir ... permission denied` | error | skills 目录无写权限 | 检查目标目录权限 |

## 参考

- [arkcli-connect](../SKILL.md) -- skill 入口
- [arkcli-shared](../../arkcli-shared/SKILL.md) -- 共享认证与全局参数
