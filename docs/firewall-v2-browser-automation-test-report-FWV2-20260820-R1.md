# 防火墙 V2 浏览器自动化回归报告

- 测试轮次：`FWV2-20260820-R1`
- 日期：`2026-08-20`
- 被测分支：`refactor/firewall-v2-foundation`
- 提交 SHA：`b23b0a3956a26de9b449c2632b3a6a7c01b20577`
- 面板版本：`v2.2.4`
- 测试入口：`http://172.16.10.111:9999/hosts/firewall/rules`
- 浏览器：当前已登录的 Google Chrome
- 状态：已完成（报告已追加至 `FWV2-20260821-R3`；存在普通 FAIL 与明确 BLOCKED 项）

## 环境记录

| 节点 | 后端 | 后端版本 | 初始状态 | 清理状态 |
|---|---|---|---|---|
| 主节点（页面显示 `sloooop`） | iptables | 1.8.4 | 已启用、已绑定；初始规则 8 条 | 已清理并复核：页面 8 条、测试端口命中 0、无 `E2E-FW-` |
| 110（`172.16.10.110`） | UFW | 0.36 | active；系统初始 numbered 11 条 | 已清理并复核：页面/系统均 11 条，默认策略和 before/after 计数未变化，E2E 系统/元数据/profile 命中 0 |
| 98（`172.16.10.98`） | firewalld | 0.5.3 | running；default public；active docker/public；页面稳定基线 22 条 | 已清理并复核：runtime/permanent 恢复基线，测试 zone/service/rich rule/ports/元数据命中 0 |
| 47.120（`47.120.41.13`） | nftables | 1.1.3 | 已启用但未初始化；页面 0 条；原有 4 张兼容表、49 个 handle | 已完成初始化/绑定并清理：页面保留 3 条环境规则；原有 4 张表仍为 49 个 handle；E2E 系统/持久化/元数据命中 0 |

## 用例结果

| ID | 优先级 | 节点/后端 | 结果 | 证据/备注 |
|---|---|---|---|---|
| NAV-01 | P0 | 主节点 / iptables | PASS | `/hosts/firewall/rules` 默认选中“主机防火墙”，页面正常加载。 |
| NAV-02 | P0 | 主节点 / iptables | PASS | 依次切换容器端口防护、端口转发、设置，URL 分别为 `/docker`、`/forward`、`/setting`，页面控件与选中项匹配。 |
| STATUS-01 | P0 | 主节点 / iptables | PASS | 页面显示 `iptables`、`已启用`、`版本: 1.8.4`。 |
| SET-01 | P0 | 主节点 / iptables | PASS | 设置页显示主机防火墙、容器端口防护、端口转发三个独立选择器。 |
| SET-02 | P1 | 主节点 / iptables | PASS | 主机后端下拉顺序为 iptables、nftables、firewalld、ufw。 |
| SET-03 | P1 | 主节点 / iptables | PASS | iptables 显示已启动、IPv4 已初始化、IPv6 未初始化；其余三个后端显示未安装。 |
| SET-04 | P0 | 主节点 / iptables | PASS | nftables、firewalld、ufw 选项均为 disabled，无法选择。 |
| RULE-LIST-01 | P0 | 主节点 / iptables | PASS | 表头显示动作、优先级、状态、协议、IP、端口、已使用、描述、操作；共 8 条。 |
| RULE-LIST-10 | P1 | 主节点 / iptables | PASS | 系统实例被移除后，旧页面将对应管理记录显示为“异常：1Panel 管理记录与系统实际规则不一致”；该行复选框禁用，且无编辑、删除入口。刷新后的新页面恢复真实基线。 |
| RULE-FORM-01 | P0 | 主节点 / iptables | FAIL | IP、端口均为空时点击“检查”，按钮结束操作但抽屉无可见校验提示；DOM 中无表单错误或 alert，立即截图亦未显示预期“至少填写一个 IP 地址或端口”。 |
| RULE-FORM-02 | P0 | 主节点 / iptables | FAIL | 分别输入非法 IPv4 `999.1.1.1`、非法 IPv6 `2001:db8:::1`、非法 CIDR `192.0.2.0/99`，点击“检查”后均停留在表单且无可见错误提示。 |
| RULE-FORM-03 | P0 | 主节点 / iptables | FAIL | 端口 `65536` 与倒序范围 `55120-55110` 均返回明确参数错误；端口 `0` 点击检查后无任何提示，故整体失败。 |
| RULE-FORM-12 | P1 | 主节点 / iptables | PASS | 切换 ALL、ICMP、ICMPV6 时端口输入框均禁用；预览端口显示 `*`。 |
| RULE-CHECK-01 | P0 | 主节点 / iptables | PASS | TCP/任意 IPv4/55101 的检查结果显示“可创建 · 1”“检查通过”，提交按钮可用。 |
| RULE-CHECK-02 | P0 | 主节点 / iptables | PASS | 对完全相同的面板规则再次检查，显示“已存在 · 1”“本次将自动跳过”；提交后列表仍为 9 条，未重复写入。 |
| RULE-CHECK-04 | P1 | 主节点 / iptables | PASS | 创建 ICMP 子集时识别到已有同动作 ALL 覆盖规则，显示“现有规则已经覆盖本次配置”警告，允许继续提交。 |
| RULE-CHECK-06 | P1 | 主节点 / iptables | PASS | 创建 ALL/IPv4 时识别到与现有拒绝规则部分重叠，显示顺序影响警告且允许提交。 |
| RULE-CHECK-11 | P1 | 主节点 / iptables | PASS | IPv4/IPv6 批量提交时，失败框准确标识 `#3` IPv6 规则失败及 `#4` 未执行；前两条 IPv4 已成功写入并由 iptables 回读确认。 |
| RULE-CREATE-01 | P0 | 主节点 / iptables | PASS | 已创建允许 TCP、`0.0.0.0/0`、端口 55101、描述 `E2E-FW-HOST-V4-TCP`；页面从 8 条增至 9 条。Web 终端确认 `1PANEL_BASIC` 中存在 `--dport 55101 ... -j ACCEPT`，并带 `1panel-rule` 注释。 |
| RULE-CREATE-02 | P0 | 主节点 / iptables | PASS | 已创建拒绝 UDP、`198.51.100.10/32`、端口 55102；页面增至 10 条，iptables 实际规则为 `DROP`。 |
| RULE-CREATE-03 | P0 | 主节点 / iptables | PASS | TCP/UDP、`198.51.100.0/24`、55103 预览展开 2 条，页面分别显示 TCP 与 UDP 两行，iptables 存在两条 ACCEPT。 |
| RULE-CREATE-04 | P0 | 主节点 / iptables | FAIL | ALL、`0.0.0.0/0`、无端口规则成功写入 iptables，但列表显示为“系统保护”、无编辑/删除入口，无法按通用断言从页面删除。唯一系统 UUID 为 `0ba6ee3f-35c2-4dbf-85ef-b4e6960910a5`；已完成精确系统规则与孤儿元数据恢复。 |
| RULE-CREATE-05 | P0 | 主节点 / iptables | PASS | ICMP、`198.51.100.0/24`、无端口规则创建成功，页面显示协议 ICMP、端口 `*`、面板创建状态。 |
| RULE-CREATE-06 | P0 | 主节点 / iptables | FAIL | ICMPV6、`2001:db8::/64` 预检查通过，但提交弹出“操作失败：ip6tables-restore: line 2 failed”；未写入页面或系统。 |
| RULE-CREATE-07 | P0 | 主节点 / iptables | FAIL | TCP、`2001:db8::10/128`、55104 预检查通过，提交同样失败于 `ip6tables-restore: line 2 failed`。 |
| RULE-CREATE-08 | P0 | 主节点 / iptables | FAIL | 拒绝 UDP、`2001:db8::/64`、55105-55107 预检查通过，提交失败于 `ip6tables-restore: line 2 failed`。 |
| RULE-CREATE-09 | P0 | 主节点 / iptables | FAIL | `198.51.100.11,2001:db8::11` × `55108,55109` 正确展开 4 条；提交后 2 条 IPv4 成功，首条 IPv6 失败、后一条标记未执行，未满足全部创建成功断言。 |
| RULE-CREATE-10 | P0 | 主节点 / iptables | PASS | TCP/UDP × 两个 IPv4 地址 × 两个端口正确展开 8 条，页面总数增至 24；Web 终端统计 55110/55111 的 iptables 规则为 8 条。 |
| RULE-SEARCH-01 | P0 | 主节点 / iptables | PASS | 搜索框输入完整端口 55101，未按 Enter、未失焦即过滤为唯一匹配规则（共 1 条）。 |
| RULE-SEARCH-03 | P0 | 主节点 / iptables | FAIL | 清空搜索框后等待约 700 ms，列表仍显示“共 1 条”，未立即恢复全部规则；刷新页面后恢复 9 条。 |
| RULE-PERSIST-01 | P0 | 主节点 / iptables | PASS | 刷新页面后 `E2E-FW-HOST-V4-TCP` 仍存在，列表为 9 条。 |
| RULE-FORM-11 | P1 | 主节点 / iptables | PASS | 描述 `E2E-FW-DESC 中文 空格 !@#` 在预览中原样保留并成功创建，页面总数从 24 增至 25。搜索与导出断言待分别记录。 |
| EDIT-01 | P0 | 主节点 / iptables | FAIL | 尝试把 55101 规则完整修改为拒绝、UDP、`198.51.100.21/32`、55113 并修改描述；两次点击“检查”后均停留在编辑表单，未进入确认页，且无可见错误。 |
| EDIT-02 | P0 | 主节点 / iptables | PASS | 取消上述完整编辑后，原规则仍为允许、TCP、`0.0.0.0/0`、55101，页面与系统未发生变更。 |
| EDIT-03 | P1 | 主节点 / iptables | PASS | 仅修改描述时正常进入二次确认，确认框仅列出变更字段“描述”；提交后页面显示新描述，iptables 中 55101 实例仍为 1，未产生重复规则。 |
| EXT-01 | P0 | 主节点 / iptables | PASS | Web 终端插入无 `1panel-rule` 标记的 TCP/`198.51.100.10/32`/55131 规则；页面显示“外部规则”、原始描述与字段均正确。 |
| EXT-02 | P0 | 主节点 / iptables | PASS | 搜索 55131 无需回车即命中唯一外部规则。 |
| EXT-03 | P0 | 主节点 / iptables | PASS | 搜索 `E2E-FW-EXTERNAL-DESC` 即时命中唯一外部规则。 |
| EXT-04 | P0 | 主节点 / iptables | PASS | 外部规则选择框禁用，操作列只有“纳管”，无编辑和删除。 |
| EXT-05 | P0 | 主节点 / iptables | PASS | 纳管确认框点击取消后仍显示“外部规则”，系统实例保持 1 条。 |
| EXT-06 | P0 | 主节点 / iptables | PASS | 确认纳管后状态变为“外部纳管”并出现编辑/删除；按源地址与端口统计系统实例仍为 1，未重复写入。 |
| EXT-07 | P0 | 主节点 / iptables | PASS | 删除确认先取消，规则保持；再次确认后页面为 0 条，Web 终端按源地址和 55131 统计为 0。 |
| DELETE-01 | P0 | 主节点 / iptables | PASS | 对外部纳管规则点击删除后取消，页面和系统规则均保持。 |
| DELETE-02 | P0 | 主节点 / iptables | PASS | 确认删除外部纳管规则后，页面记录和 iptables 实例同步消失。 |
| DELETE-03 | P1 | 主节点 / iptables | PASS | 批量选择 16 条可操作测试规则，确认框准确显示“将删除 16 条规则”；确认后 16 条均从页面和系统移除，系统保护的 ALL 规则未被误选。 |
| DG-STATUS-03 | P0 | 主节点 / iptables-docker | PASS | 容器端口防护页显示后端 `iptables-docker`、状态“已启用”、`IPv4: 已生效`，说明已初始化并绑定。 |
| DG-STATUS-08 | P1 | 主节点 / iptables-docker | PASS | Docker IPv6 未启用时，页面明确显示 `IPv6: 未启用`，同时 IPv4 仍为已生效。 |
| DG-LIST-01 | P0 | 主节点 / iptables-docker | PASS | 页面按容器聚合显示名称、Compose/应用与发布端口；实测 docker-registry、halo、mysql、sftpgo 均正常展示。 |
| DG-LIST-03 | P0 | 主节点 / iptables-docker | PASS | 点击 docker-registry 的“详情”正常打开端口详情抽屉。 |
| DG-LIST-04 | P1 | 主节点 / iptables-docker | PASS | 详情抽屉准确显示 `127.0.0.1:5000 → 5000/tcp`、防护状态“未防护”和描述 `-`。 |

## 失败与阻塞

- `RULE-FORM-01`：空 IP、空端口检查没有显示预期校验提示；普通失败，继续执行后续用例。
- `RULE-FORM-02`：非法 IPv4、IPv6、CIDR 均未显示校验提示；普通失败，继续执行。
- `RULE-FORM-03`：端口 0 未显示校验提示（65536 和倒序范围提示正常）；普通失败，继续执行。
- `RULE-SEARCH-03`：清空搜索框后列表未立即恢复；刷新页面后恢复，普通失败，继续执行。
- `RULE-CREATE-04`：ALL 规则创建后错误显示为系统保护，页面不可删除；可通过唯一 UUID 精确恢复，继续执行。
- `RULE-CREATE-06`、`RULE-CREATE-07`、`RULE-CREATE-08`：IPv6 链未初始化，预检查通过但实际提交均失败于 `ip6tables-restore: line 2 failed`；普通失败，继续执行。
- `RULE-CREATE-09`：IPv4/IPv6 混合批量仅 IPv4 部分成功；失败和未执行项展示正确，继续执行。
- `EDIT-01`：多字段编辑两次检查均无响应且无错误提示；取消后数据未变化，后续以最小描述编辑验证编辑链路可用，普通失败，继续执行。
- Docker 列表搜索矩阵在 Chrome 控制调用超过 300 秒后中断；未据此标记产品用例结果，页面无写入，后续分项重试。
- 测试开始前 Chrome 控制连接曾阻塞；重装 Chrome 插件并完全重启 Chrome 后恢复。未执行任何防火墙写入，不计入产品用例失败。

## 测试数据与清理日志

- 主节点描述测试规则创建后共 25 条（基线 8 条）。已创建 55101、55102、两条 55103、ALL、ICMP、两条 55108/55109、八条 55110/55111、55112；IPv6 提交失败，无 IPv6 测试规则残留。
- 页面批量删除了 16 条可操作测试规则；ALL 测试规则被错误识别为系统保护，因此先按唯一 UUID `0ba6ee3f-35c2-4dbf-85ef-b4e6960910a5` 精确删除其 iptables 实例。系统规则删除后页面留下同 UUID 漂移记录，受保护状态导致 UI/API 均拒绝删除。
- 对孤儿元数据执行精确恢复前，已将完整数据库行备份到 `/opt/1panel/tmp/FWV2-20260820-R1-orphan-0ba6ee3f-35c2-4dbf-85ef-b4e6960910a5.json`；随后在事务内按 UUID 与描述双重断言删除，输出 `DELETED=1`、`REMAINING=0`。
- 主节点最终清理复核：页面恢复基线 `共 8 条`，无 `E2E-FW-`；Web 终端检查 55101–55112 与 `E2E-FW` 命中数为 `0`。
- 后续接管未刷新的旧标签时短暂看到 9 条及上述异常行；在当前 Chrome 中新开同一 URL 后为 `共 8 条` 且无测试描述，确认这是旧页面缓存状态，并非孤儿数据重新生成。
- 外部规则 55131 已完成创建、取消纳管、确认纳管、取消删除、确认删除全流程；页面与系统均已归零，无残留。

## 修复后定向回归：`FWV2-20260821-R1-FIX`

- 日期：`2026-08-21`
- 被测分支：`refactor/firewall-v2-foundation`
- 版本标记：`b23b0a3956a26de9b449c2632b3a6a7c01b20577 + working tree firewall fixes`
- 面板版本：`v2.2.4`
- 测试入口：`http://localhost:4004/hosts/firewall/rules`
- 浏览器：当前已登录的 Google Chrome
- 节点/后端：主节点（`sloooop`）/ iptables 1.8.4
- 结果：`PASS（IPv6 缺链恢复场景未在在线主节点破坏性构造）`

| 回归项 | 结果 | 证据/备注 |
|---|---|---|
| `RULE-FORM-01` 空 IP、空端口 | PASS | 点击检查后显示“请至少填写一个 IP 地址或端口”。 |
| `RULE-FORM-02` 非法 IP/CIDR | PASS | `999.1.1.1`、`2001:db8:::1`、`192.0.2.0/99` 均显示“请输入正确的 IP 地址”。 |
| `RULE-FORM-03` 非法端口 | PASS | `0`、`65536`、`55120-55110` 均显示“请输入正确的端口,1-65535”。 |
| `RULE-CREATE-04` ALL 规则保护状态 | PASS | 创建允许 ALL、`0.0.0.0/0`、无端口规则后，列表显示普通面板规则，编辑和删除入口均存在；随后已通过页面删除。 |
| `RULE-CREATE-06` IPv6 创建 | PASS | 创建 ICMPV6、`2001:db8::/64`、描述 `FWV2-REG-20260821-IPV6` 成功，列表由 9 条增至 10 条；随后已通过页面删除。 |
| `EDIT-01` 多字段编辑 | PASS | 新建 TCP/允许/`198.51.100.21/32`/55121 后，一次修改为 UDP/拒绝/`198.51.100.22/32`/55122 并修改描述；确认框准确列出协议、源 IP、目标端口、动作、描述，提交回读全部一致。 |
| 动作相反重叠保护 | PASS | 在 ALL/允许规则存在时创建范围内拒绝规则，预检显示“存在范围重叠但动作相反的规则，本次操作已停止”，提交按钮禁用；属预期安全行为。 |
| 状态栏精简 | PASS | 规则页保留后端、启用状态和版本；不再显示 `IPv4:`、`IPv6:` 状态标签。当前 IPv6 链健康，未显示恢复按钮。 |
| IPv6 专用初始化入口 | 部分覆盖 | 当前在线 IPv6 链健康，按设计不显示按钮。未删除在线链来强制构造缺链状态，避免影响面板访问；专用 `init-ipv6-base` 路径及缺链判断由后端单元测试覆盖。 |
| 浏览器错误日志 | PASS | 定向回归结束时页面错误日志为空。 |
| 数据清理 | PASS | 删除本轮 `FWV2-REG-20260821-*` 全部临时规则，描述前缀命中 0，页面恢复 `共 9 条`。 |

## 完整回归续测：`FWV2-20260821-R2`

- 日期：`2026-08-21`
- 被测分支：`refactor/firewall-v2-foundation`
- 版本标记：`b23b0a3956a26de9b449c2632b3a6a7c01b20577 + working tree firewall fixes`
- 面板版本：`v2.2.4`
- 测试入口：`http://localhost:4004/hosts/firewall/rules`
- 浏览器：当前已登录的 Google Chrome，沿用现有 1Panel 标签页；系统命令仅通过新开的 1Panel Web 终端执行
- 节点顺序：主节点 iptables -> 110 UFW -> 98 firewalld -> 47.120 nftables
- 状态：`BLOCKED（主节点本轮范围已执行并清理；三个远端节点因版本不一致无法切换）`

### 主节点执行结果

| 范围 | 结果 | 证据/备注 |
|---|---|---|
| 导航与状态 | PASS | `NAV-03` 反复切换主机、Docker、转发、设置并刷新，无白屏、重复弹错或登录丢失；状态栏显示 iptables、已启用、1.8.4。 |
| 规则列表、分页与筛选 | PASS | 覆盖 IPv4/IPv6、允许/拒绝、无端口 `*`、管理/纳管/外部/保护/异常状态；页大小 5 刷新保持，末页切换筛选自动回第 1 页；链筛选至少保留一项。 |
| 本地搜索 | PASS | 覆盖描述、中文、协议、动作、状态、IPv4/IPv6 CIDR、端口范围、大小写及首尾空格；点击清除按钮恢复全部。搜索由前端本地计算完成，不触发规则清单重复请求。 |
| 创建表单 | PASS | 覆盖中英文分隔符、多 IP、多端口、去重、回车新增输入、删空输入、256 条上限与 257 条阻止；8 条 IPv4 批量和 2 条 IPv6 规则创建成功。 |
| 冲突检查 | PASS / BLOCKED | `RULE-CHECK-05`、`RULE-CHECK-10`、`RULE-CHECK-12`、保护链冲突均通过；`RULE-CHECK-13` 在“检查后终端插入同规则”场景因 Chrome 提交按钮交互超时未完成，系统临时规则已清理。 |
| 编辑 | FAIL | `EDIT-04`、`EDIT-06` 通过。`EDIT-05` 失败：编辑表单输入可拆成多条的多个 IP 时仅检查第一条，多个端口时停留在表单且无可见说明，未明确阻止“一次编辑变成多条”。 |
| 删除与使用情况 | PASS / 部分覆盖 | 删除 21 显示 pure-ftpd 风险，删除 3306 显示 docker-proxy 与 `Docker: 1Panel-mysql`；进程详情和 `+3` 使用项展开通过。Docker 使用项跳转未得到确定结果。 |
| 外部、保护、原生和异常规则 | PASS | `EXT-01` 至 `EXT-10`、`EXT-13` 至 `EXT-15` 完成：外部识别、搜索、取消/确认纳管、重复外部实例单条纳管、BEFORE 保护、opaque 原始详情、管理实例删除/动作漂移异常检测均正确；每次均精确恢复并复核。 |
| 主机规则导出 | PASS / 部分覆盖 | 全量导出 19 条、选中导出 2 条；JSON 无 UUID，字段与中文描述正确。导出时未同时保留外部/保护临时规则，`RULE-IO-04` 未完整覆盖。 |
| 主机规则导入 | BLOCKED | `RULE-IO-05` 至 `RULE-IO-13`：Chrome 文件选择器超时并重置浏览器控制内核，未把控制阻塞误记为产品失败。 |
| 端口转发 | FAIL | 创建 IPv4/IPv6、TCP/UDP、空目标默认值、等长范围、指定网卡、搜索、编辑和批量删除均通过；`FWD-CREATE-09` 中 65536 与倒序范围被拒绝，但端口 `0` 无可见错误提示。测试 8 条转发均已删除，恢复原有 3 条。 |
| Docker 容器端口防护 | FAIL | 列表搜索/组合筛选、三种模式、来源拆分去重、批量设置/清除、同步及既有策略保持均通过。`DG-POLICY-07` 地址族不匹配、`DG-POLICY-08` deny_sources 为空均阻止保存，但页面没有预期可见错误。文件导入用例受文件选择器阻塞；外部客户端访问效果未执行。 |
| 禁 Ping | PASS / 部分覆盖 | 页面从原始“停用”切到“启用”，刷新后仍启用，再恢复“停用”并复查。未从外部客户端验证 ICMP 实际通断，故 `PING-01/02` 的访问效果部分未完成。 |
| 端口白名单 | PASS / 部分覆盖 | `WL-01` 至 `WL-11`：新增并刷新回读 IPv4 TCP 55120、IPv6 UDP 55121-55123；编辑取消恢复；0、65536、倒序、非数字均提示错误；完全重复和范围重叠均提示；IPv4/IPv6 与 TCP/UDP 相同端口允许；空白行忽略。4 条临时项全部删除并恢复原有 IPv4 TCP 80。`WL-12` 未执行初始化/启动高影响操作。 |

### 本轮新增普通失败

1. `EDIT-05`（P1）：多数据编辑没有明确、可见地阻止；多个 IP 会静默只取第一条进行检查。
2. `FWD-CREATE-09`（P0）：源端口 `0` 被阻止写入，但无可见错误提示。
3. `DG-POLICY-07`（P0）：IPv4 端点填写 IPv6 来源时阻止保存，但无可见错误提示。
4. `DG-POLICY-08`（P1）：禁止指定来源但来源为空时阻止保存，但无可见错误提示。

### 浏览器控制阻塞

- 主机规则导入与 Docker 策略导入：Chrome 文件选择器超时；导出文件已成功落盘并完成 JSON 校验。
- `RULE-CHECK-13`：检查完成后通过 Web 终端插入同规则，随后 Chrome 对提交按钮交互超时；临时系统规则已清理，不计产品失败。
- Chrome 控制产生的 Statsig/扩展网络超时不属于 1Panel 页面错误，未计入产品结果。

### 跨节点阻塞

按规定顺序分别尝试切换：

| 节点 | 目标后端 | 结果 | 可见证据 |
|---|---|---|---|
| 172.16.10.110 | UFW | BLOCKED | 点击节点后提示“节点版本与主节点不一致，暂不支持切换，请在节点管理中升级后重试”。 |
| 172.16.10.98 | firewalld | BLOCKED | 同上，仍停留主节点 iptables。 |
| 47.120.41.13 | nftables | BLOCKED | 同上，仍停留主节点 iptables。 |

升级远端 Agent 可能造成节点短暂离线，符合“可能中断访问时停下来询问”的停止条件，因此未擅自升级。三个节点未切入、未写入任何测试数据，后端版本仍待读取。

### 清理与基线复核

- 主节点主机规则：删除本轮 10 条 `E2E-FW-` 测试规则，页面恢复 `共 9 条`，前缀残留 0。
- 主节点端口转发：删除 8 条 552xx 测试转发，恢复原有 `共 3 条`。
- 主节点 Docker：清除全部 `E2E-FW-DOCKER*` 测试策略；保留原有 sftp 2022 来源策略。
- 主节点白名单：删除 4 条临时规则，恢复原有 `IPv4 / TCP / 80`，设置页显示 1 条。
- 主节点禁 Ping：恢复原始“停用”。
- 外部、保护、opaque 和异常恢复测试所建系统规则均已通过 Web 终端精确删除或恢复。
- 110、98、47.120：由于版本不一致无法切入，本轮未写入测试数据；尚不能完成各节点内部残留复核。
- 下载产物：`/Users/songliu/Downloads/1panel-firewall-rules-20260821122754.json`、`/Users/songliu/Downloads/1panel-firewall-rules-20260821122938.json`。受 Chrome 文件选择器控制问题影响未完成导回；未擅自删除用户 Downloads 下文件。

### `FWV2-20260821-R2` 缺陷修复后定向回归

- 版本标记：`b23b0a3956a26de9b449c2632b3a6a7c01b20577 + working tree firewall fixes`
- 节点/后端：主节点（`sloooop`）/ iptables 1.8.4
- 入口/浏览器：`http://localhost:4004/hosts/firewall/rules` / 当前已登录的 Google Chrome

| 原失败项 | 修复 | 回归结果 |
|---|---|---|
| `EDIT-05` 多数据编辑静默取第一条 | 编辑模式在表单验证和标准化两层拒绝多个源 IP 或多个目标端口，并显示具体字段错误。 | PASS：多个 IP 显示“IP: 不支持的当前操作”；多个端口显示“目标端口: 不支持的当前操作”；均不进入检查预览、未调用更新。 |
| `FWD-CREATE-09` 端口 0 无可见错误 | 提交前复用端口范围校验，并增加可见错误消息兜底。 | PASS：源端口 `0` 显示“请输入正确的端口,1-65535”和表单端口错误，抽屉保持打开，规则总数仍为 3。 |
| `DG-POLICY-07` 地址族不匹配无可见错误 | Docker 策略提交前按每个目标端点的地址族验证全部来源。 | PASS：IPv4 `127.0.0.1:5000/tcp` 填写 `2001:db8::1` 时显示“请输入正确的 IP 地址”，策略对话框保持打开。 |
| `DG-POLICY-08` deny_sources 为空无可见错误 | deny_sources 提交前要求至少一个来源；保留 allow_sources 为空即拒绝全部的既有设计。 | PASS：空来源显示“请填写必填项”，策略对话框保持打开。 |

验证结果：目标 3 个 Vue 文件 ESLint 通过；`npm run build:dev` 通过。全量 `npm run type-check` 仍被仓库已有的 enterprise VM ISO、AppMain、容器网络、进程详情、终端等无关类型错误阻断，本次修改文件未产生新的类型错误。定向回归全程只提交非法表单并取消，没有创建或修改主机规则、转发规则及 Docker 防护策略。

## 跨后端回归续测：`FWV2-20260821-R3`

- 日期：`2026-08-21`
- 被测分支：`refactor/firewall-v2-foundation`
- 提交 SHA：`b23b0a3956a26de9b449c2632b3a6a7c01b20577`
- 版本标记：`b23b0a3956a26de9b449c2632b3a6a7c01b20577 + dirty working tree`
- Working tree：`dirty`；沿用本轮开始前已有的防火墙后端、前端、配置、测试与文档修改，本轮未修复产品代码，仅更新本报告
- 面板版本：`v2.2.4`
- 测试入口：`http://localhost:4004/hosts/firewall/rules`
- 浏览器：当前已登录的 Google Chrome，沿用现有 1Panel 标签页
- 远端命令约束：全部系统读取、注入、恢复和清理命令均在对应节点新开的 1Panel Web 终端执行；未用本地终端修改远端防火墙
- 执行顺序：`172.16.10.110 / UFW` -> `172.16.10.98 / firewalld` -> `47.120.41.13 / nftables`
- 总体状态：`COMPLETED WITH FAILURES / BLOCKED ITEMS`；三个节点均完成测试数据清理和零残留复核
- 复核校准：UFW 描述编辑未复现规则丢失或数据漂移；三个后端库存复核正常；UFW 页面重新启用失败时会明确展示 `ip6tables` 错误，不属于静默失败。

### 节点、实际后端与前后基线

| 节点 | 页面/系统确认的实际后端 | 测试前基线 | 测试后基线与清理结果 |
|---|---|---|---|
| `172.16.10.110` | UFW `0.36` | 页面显示已启用，库存与 Web 终端 `Status: active`、numbered 规则 11 条一致；默认 INPUT/FORWARD 为 DROP、OUTPUT 为 ACCEPT；`before_input=12`、`after_input=7`、`before6_input=33`、`after6_input=6`；application profile 0 个 | 页面与系统均为 11 条；`Status: active`；默认策略及 before/after 计数均未变化；`E2E_SYSTEM=0`、`E2E_DB=0`、测试 application profile 不存在 |
| `172.16.10.98` | firewalld `0.5.3` | 页面库存与系统状态一致；firewalld running，default zone 为 public；active zones 为 docker（3 个 bridge/docker 接口）和 public（`ens192 ens32`）；public runtime/permanent 均为 12 个 ports、`ssh dhcpv6-client`、8 条 rich rule | 页面为 22 条；状态 running、default zone 仍为 public，active zones/interface 未变化；public runtime/permanent ports、services 和 8 条 rich rule 完全回到基线；`E2E_SYSTEM=0`、`E2E_DB=0`，测试 zone/service 均不存在 |
| `47.120.41.13` | nftables `1.1.3` | 页面显示已启用但未初始化，规则 0 条；系统有 4 张 iptables-nft 兼容表、21 条 chain、49 个 handle；`nftables.service` 为 disabled/inactive；管理元数据 0；`/etc/nftables.conf` SHA-256 为 `60dac93ffe0ea440fc4a8941a080b6fb8d2c8655d47baf856e97182a0d1ca29a` | 完成 1Panel IPv4/IPv6 管理表初始化与绑定后保留该状态；页面为 3 条环境规则（TCP 443、TCP 80、UDP 443），无 E2E；系统共 6 张表，其中原有 4 张表仍为 49 个 handle；IPv4/IPv6 输入链各保留 3 个管理跳转；配置哈希未变化；`E2E_SYSTEM=0`、`E2E_PERSIST=0`、`E2E_DB=0`、全部测试端口命中 0 |

### 按后端统计

以下数量只统计本轮表格中明确列出的实际执行项；`SKIPPED` 和 `BLOCKED` 均未计入 PASS。

| 后端 | PASS | FAIL | BLOCKED | SKIPPED |
|---|---:|---:|---:|---:|
| UFW | 20 | 3 | 0 | 2 |
| firewalld | 17 | 4 | 1 | 3 |
| nftables | 19 | 3 | 1 | 0 |

### UFW 结果

| 执行项 | 结果 | 证据/备注 |
|---|---|---|
| 状态、版本、初始化/绑定和初始库存 | PASS | 页面与 Web 终端确认 UFW 0.36、active，页面库存与系统实际 11 条一致。 |
| IPv4 TCP 单端口允许创建与回读 | PASS | 创建 55101，页面、`ufw status numbered` 和元数据可对应。 |
| IPv6 UDP 端口范围拒绝创建与回读 | PASS | 创建 `2001:db8::/64`、55105-55107、UDP deny，IPv6 系统回读正确。 |
| 中文描述创建 | PASS | `E2E-FW-UFW-中文-创建` 创建成功。 |
| 中文描述搜索 | FAIL | 输入中文后仍显示完整库存，未过滤。 |
| 中文描述编辑 | PASS | 复核未出现系统规则丢失或系统/元数据漂移，撤销原“已确认缺陷”结论。 |
| 中文规则独立页面删除 | SKIPPED | 未继续覆盖该独立页面删除步骤；最终使用唯一系统对象和 UUID 精确恢复。 |
| 完全重复规则处理 | PASS | 预检显示“已存在”，未新增实例。 |
| 相反动作冲突检查 | PASS | 同范围相反动作被阻止，提交不可用。 |
| 外部规则识别 | PASS | 55131 被识别为外部规则，字段和来源地址正确。 |
| 取消纳管 | PASS | 取消后系统规则保持，页面仍为外部规则。 |
| 确认纳管与持久化 | PASS | 纳管元数据持久化，刷新后可回读。 |
| 纳管后的即时列表刷新 | FAIL | 确认成功后当前列表仍显示“纳管”，需重新加载。 |
| 外部规则独立页面删除 | SKIPPED | 未单独覆盖页面删除；最终通过精确系统和元数据清理归零。 |
| 页面与系统命令一致性 | PASS | 字段、编号、系统规则和管理元数据回读一致。 |
| 原生不支持规则只读/opaque | PASS | `ufw limit` 55133 仅显示详情，无编辑、纳管和删除入口。 |
| UFW Application 规则 | PASS | 测试 profile 与 application 规则以 APP/详情形式展示。 |
| numbered 顺序与删除 | PASS | reload 前后 numbered 顺序稳定，测试编号规则按精确对象删除。 |
| before/after 非默认范围保护 | PASS | 页面只读保护，测试前后 before/after 计数完全一致。 |
| active/inactive 状态展示 | PASS | 页面关闭后选择稍后手动重启，页面与终端均正确显示 inactive；通过 Web 终端恢复 active。 |
| 页面重新启用 UFW | FAIL | 点击页面“开启”并确认后 UFW 未启用；页面明确返回 `stderr: ERROR: problem running ip6tables , err: exit status 1`，不是静默失败。该结果证明 IPv6 后端命令执行失败，尚不能仅凭页面现象归因为前端或库存问题。 |
| reload 后持久化 | PASS | Web 终端 reload 后测试规则、numbered 顺序和 active 状态保持。 |
| 默认策略和 application 配置保护 | PASS | 默认 INPUT/FORWARD/OUTPUT 策略未变化，原有 application 配置未修改。 |
| 页面库存可靠性 | PASS | 复核未发现规则短暂归零或陈旧库存；`ip6tables`/xtables 命令错误应作为独立系统命令失败处理，不作为库存不可靠的证据。 |
| 完整清理 | PASS | 删除全部页面/外部/opaque/application 测试对象；精确元数据备份为 `/opt/1panel/tmp/FWV2-20260821-R3-ufw-cleanup.json`，`DELETED=3 REMAINING=0`。 |

### firewalld 结果

| 执行项 | 结果 | 证据/备注 |
|---|---|---|
| 状态、版本、初始 zone/库存 | PASS | 页面与终端确认 firewalld 0.5.3、running、default public、active docker/public。 |
| IPv4 TCP 单端口允许 | PASS | 55101 创建成功，runtime/permanent 均存在。 |
| IPv6 UDP 范围拒绝 | PASS | `2001:db8::/64`、55105-55107 创建成功，runtime/permanent 均存在。 |
| 中文描述创建 | PASS | 中文描述规则创建并在页面展示。 |
| 中文描述搜索 | FAIL | 输入中文后总数保持 `共 30 条`，未过滤。 |
| 中文描述编辑 | FAIL | 完成检查和确认后页面仍保持旧描述，未完成期望更新。 |
| 完全重复规则 | PASS | 显示“已存在”，系统数量未增加。 |
| 部分重叠冲突告警 | PASS | 与原有 drop rich rule 部分重叠时显示顺序/动作冲突警告。 |
| 完全相反动作冲突 | SKIPPED | 表单动作状态重置后未完成该独立提交，不记 PASS。 |
| 外部规则识别 | PASS | public rich rule 55131 被识别为外部规则。 |
| 取消纳管 | PASS | 取消后系统 runtime/permanent 规则保持。 |
| 确认纳管 | FAIL | 页面操作成功但数据库无 55131 元数据，刷新后仍回到外部规则。 |
| 外部规则独立页面删除 | SKIPPED | 纳管失败，无法覆盖“纳管后页面删除”；最终通过 Web 终端精确删除。 |
| 页面与系统命令一致性 | PASS | port/service/rich rule、runtime/permanent 和漂移状态与命令回读一致。 |
| 复杂 rich rule opaque 展示 | FAIL | 含 log/limit 的 55136 原生 rich rule 未出现在页面库存。 |
| active/default zone | PASS | public 与 docker 活动 zone、接口和默认 public 显示/回读正确。 |
| 非默认 zone 的 port/service/rich rule | PASS | `e2e-fw-zone` 的 55134、http、55135 在 reload 后 runtime/permanent 一致，且未成为 active zone。 |
| 自定义 service | PASS | `E2E-FW-FIREWALLD-SERVICE` 定义及 55137/tcp 关联可回读，页面以 SERVICE/详情展示。 |
| rich rule priority | BLOCKED | 远端 firewalld 0.5.3 低于 priority 所需版本；页面不提供 priority 字段，未升级 Agent 或系统组件。 |
| runtime_only/permanent_only 漂移 | PASS | 页面显示 ports 漂移；reload 后 runtime-only 55132 消失，permanent-only 55133 收敛到 runtime/permanent。 |
| reload 后持久化 | PASS | 页面规则、55131 和复杂 rich rule 均在 runtime/permanent 中保持；活动/默认 zone 未变化。 |
| 默认 zone、其他活动 zone 和已有 service 保护 | PASS | default public、docker/public 活动 zone、既有 ssh/dhcpv6-client 和 8 条 rich rule 最终全部回到基线。 |
| firewalld 停止/启动 | SKIPPED | UFW 已覆盖状态切换；为减少高影响重复操作，本后端未停止服务。 |
| 页面库存可靠性 | PASS | 复核页面库存与 runtime/permanent 回读一致，未复现短暂归零或陈旧库存；撤销原库存缺陷结论。 |
| 完整清理 | PASS | 精确删除 4 条测试 rich rule、55133、测试 service 和 zone；4 条元数据备份为 `/opt/1panel/tmp/FWV2-20260821-R3-firewalld-cleanup.json`，`DELETED=4 REMAINING=0`。 |

### nftables 结果

| 执行项 | 结果 | 证据/备注 |
|---|---|---|
| 实际后端、版本、状态和初始库存 | PASS | 页面确认 nftables 1.1.3、已启用但未初始化、0 条；系统初始为 4 张兼容表、21 条 chain、49 个 handle。 |
| 1Panel 管理链初始化与绑定 | PASS | 仅新增 `ip`/`ip6 nft_1panel_filter`；输入链按 BEFORE/BASIC/AFTER 各 3 个 jump 绑定，9999/22 白名单和 policy accept 存在。 |
| IPv4 TCP 单端口允许 | PASS | 55101 页面创建成功，系统 `NFT_1PANEL_BASIC` 与 UUID 注释一致。 |
| IPv6 UDP 范围拒绝 | PASS | `2001:db8::/64`、55105-55107 在 ip6 管理链回读为 drop。 |
| 中文描述创建 | PASS | `E2E-FW-NFT-中文-创建` 创建成功。 |
| 中文描述搜索 | FAIL | 输入“中文”后仍显示全部 5 条。 |
| 中文描述编辑 | PASS | 修改为 `E2E-FW-NFT-中文-已编辑` 后数据库 revision 和系统 handle 更新；当前列表未即时更新，但重新触发库存读取后显示新值。 |
| 完全重复规则 | PASS | 显示“已存在 · 1”，未提交重复实例。 |
| 相反动作冲突 | PASS | 相同范围/端口的 reject 被标记错误，提交按钮禁用。 |
| 外部规则识别 | PASS | 198.51.100.10/32、55131 原生规则显示为外部规则，原始 comment 可见。 |
| 取消纳管 | PASS | 取消后仍为外部规则，系统实例保持。 |
| 确认纳管与持久化 | PASS | 系统 comment 被替换为 `1panel-rule:<UUID>`，数据库 origin=`adopted`，持久化文件包含该规则。 |
| 纳管后的即时列表刷新 | FAIL | 操作成功后当前列表仍显示“纳管”；重新触发库存读取后才出现编辑/删除。 |
| 外部纳管规则删除 | PASS | 通过页面确认删除 55131；刷新后页面、系统、数据库和持久化文件均无该规则。 |
| 页面与系统命令一致性 | PASS | IPv4/IPv6 字段、动作、UUID comment、handle 和系统顺序均能对应。 |
| 刷新后的规则持久化 | PASS | 页面刷新及重新触发库存读取后，已创建和已编辑规则仍在；`1panel_filter.nft` 语法检查通过。 |
| set/counter/limit/ct state opaque 只读展示 | PASS | 两条复杂原生规则以禁用复选框、仅“详情”展示；详情保留 set、ct state、limit、counter、accept 和原始 comment。 |
| opaque 动作语义 | FAIL | 两条原生表达式末尾均为 accept，但列表动作错误显示为“拒绝”。 |
| handle、顺序和优先级 | PASS | 将 55101 从优先级 4 调整为 1 后，系统规则移动到链首；handle 由 31 重写为 45，后续规则 handle 依次重排，页面优先级与系统顺序一致。 |
| 原生 nftables reload 后持久化 | BLOCKED | `nftables.service` 为 disabled/inactive，`ExecReload` 会执行含 `flush ruleset` 的 `/etc/nftables.conf`；直接 reload 会清空原有 4 张 iptables-nft 表，无法可靠恢复时未执行。独立 1Panel 持久化文件已完成 `nft -c -f` 校验。 |
| 其他 table/chain 与原生配置保护 | PASS | 原有 4 张表始终保持 49 个 handle；最终配置 SHA-256 与初始一致，未修改 `/etc/nftables.conf`。 |
| 页面库存可靠性 | PASS | 复核页面库存与系统规则一致，未复现短暂归零或陈旧缓存；撤销原库存缺陷结论。 |
| 完整清理 | PASS | 页面删除 55101、55105-55107、55131；按实时 handle 删除 55136 与 set 规则并删除 `e2e_fw_ports`，原子刷新 `1panel_filter.nft`；最终页面 3 条、全部 E2E/测试端口/持久化/元数据命中 0。 |

### 本轮新增缺陷与阻塞摘要

1. 三个后端的中文描述搜索均未过滤库存；与 R2 主节点本地搜索通过的结果不一致。
2. firewalld 中文描述编辑未生效。
3. UFW 页面“开启”确认后未实际启用，页面明确返回 `stderr: ERROR: problem running ip6tables , err: exit status 1`；错误反馈正常，待继续定位 IPv6 后端命令失败原因。
4. UFW 与 nftables 纳管/编辑提交后当前列表不即时刷新；重新读取库存后实际状态正确。
5. firewalld 外部规则纳管未生成管理元数据；复杂 log/limit rich rule 未在页面展示。
6. nftables opaque accept 表达式在列表被错误标为“拒绝”，但详情中的原始表达式正确。
7. firewalld rich priority 因远端版本 0.5.3 阻塞；未自行升级。
8. nftables 原生 reload 因服务 inactive 且 `/etc/nftables.conf` 会 `flush ruleset` 而阻塞；未以破坏原有兼容表的方式强行执行。

复核撤销项：UFW 中文描述编辑未发现规则丢失或数据漂移；UFW、firewalld、nftables 库存均复核正常，系统命令执行错误不再作为库存不可靠的证据。

### 清理与保护结论

- UFW：测试前后 numbered 规则均为 11；默认策略、before/after 计数和原有 application 配置未变化。
- firewalld：public runtime/permanent 完全回到原 ports、services、8 条 rich rule；default public 与 docker/public active zones 未变化。
- nftables：保留用户要求继续后的初始化/绑定状态和 3 条环境管理规则；原有 4 张兼容表、49 个 handle 及原生配置哈希未变化。
- 三节点的 `E2E-FW-` 系统规则、页面规则、外部规则、终端原生临时规则、测试 zone/service/profile/set 和管理元数据最终均为 0。
- 未修改或删除环境原有 rule、zone、service、policy、table 或 chain；未升级任何远端 Agent；未在本轮自动修复产品代码。
