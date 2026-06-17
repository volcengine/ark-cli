# arkcli-connect 最小评估用例

目标：验证本 skill 在「该唤起 / 不该唤起 / 子命令分流 / 破坏性守卫」四个维度上行为稳定，并且不让 Agent 幻觉出不存在的 `+connect install / setup / sync / remove` 子命令。

## 1) 该唤起（Trigger）— 安装

输入（用户说法）：

- "刚装完 arkcli，想把 skills 同步到本机的 Claude Code"
- "把 arkcli 内嵌的 skills 装到我电脑上的所有 AI agent"

期望行为：

- 路由到 `arkcli-connect`
- 推荐 `arkcli +connect`（默认行为就是安装；**没有** `+connect install` 子命令）
- 建议先 `arkcli +connect list` 做预检
- 提醒同名 skill 目录会被**覆盖**（先 `rm -rf` 再复制，不合并）
- **不要**给出 `arkcli +connect uninstall`

## 2) 该唤起（Trigger）— 仅查看支持范围

输入：

- "我这台机器上 arkcli 能识别出哪些 agent？我不想现在就装东西"

期望行为：

- 路由到 `arkcli-connect`
- 只给出 `arkcli +connect list`
- 解释 list 是只读、不需要认证、不写入文件系统
- **不要**顺手 install 或 uninstall

## 3) 破坏性守卫（Guard）— 卸载

输入：

- "把之前用 +connect 装过的 skills 从所有 agent 里清理掉"

期望行为：

- 路由到 `arkcli-connect`
- 推荐 `arkcli +connect uninstall`，但**先要求用户明确确认**：哪些 agent、是否接受不可逆删除
- 说明 uninstall 会 `rm -rf` 同名 skill 目录、会从所有检测到的 agent 一起清
- 建议先 `arkcli +connect list` 看清范围
- 不要直接执行（破坏性）

## 4) 不该唤起（Anti-trigger）— 401/鉴权失败

输入：

- "我跑 arkcli 业务命令报 401，是不是 +connect 没装好？"
- "auth login 失败"

期望行为：

- 路由到 `arkcli-auth`，**不要**走 `+connect`
- 明确说明：`+connect` 是本地文件系统操作，不需要认证；与 401 无关
- 先 `arkcli auth status`，必要时 `arkcli auth login`

## 5) Agent 行为反幻觉清单（重点）

下列子命令**不存在**，任何评测里 Agent 给出都视为失分：

- `arkcli +connect install`
- `arkcli +connect install --agent <name>`
- `arkcli +connect setup`
- `arkcli +connect sync`
- `arkcli +connect remove`

允许的子命令穷举：`(空)` / `list` / `uninstall`，仅此三种。

## 6) 配套机器评测

机器评测资产位于 `tests/skills/arkcli-connect/`，复跑：

```bash
cd skill-creator
python3 -m scripts.run_arkcli_skill_benchmark \
  --skill-path ../skills/arkcli-connect \
  --workspace /tmp/arkcli-connect-bench \
  --iteration 1 \
  --runs-per-config 2 \
  --runtime claude
```
