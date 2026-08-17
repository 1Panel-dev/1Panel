# Docker 端口防护需求边界

## 1. 目标

在 Docker 使用 `iptables`、`iptables-nft` 或原生 `nftables` 作为防火墙后端时，为 Docker bridge 网络发布到宿主机的端口提供来源地址拒绝控制。

iptables 后端使用 Docker 提供的 `DOCKER-USER` 入口；原生 nftables 后端使用 1Panel 独立表中的 `forward` base chain，不修改 Docker 自有的 `docker-bridges` 表。

Docker 发布端点默认保持开放。用户可以禁止所有来源、禁止指定的 IPv4/IPv6 地址或 CIDR，或者仅允许指定来源访问；未配置策略的端点继续交给 Docker。

## 2. 页面位置

在“主机 / 防火墙”下增加独立的“容器端口防护”Tab：

```text
防火墙
├── 端口 / IP 规则
├── 容器端口防护
└── 端口转发
```

不把 Docker 防护规则混入现有“端口 / IP 规则”，原因是两者作用的流量路径不同：

- 端口 / IP 规则：管理宿主机 `INPUT` 流量。
- 容器端口防护：管理 Docker DNAT 后进入 `FORWARD` 的流量。

现有规则页可以识别并提示 Docker 占用的端口，但只负责展示和跳转，不在该页面修改 Docker 防护策略。

## 3. 支持范围

第一版支持：

- Docker Engine 的 iptables backend。
- Docker Engine 的 iptables-nft backend。
- Docker Engine 的原生 nftables backend。
- firewalld、UFW、iptables、nftables 四种宿主机防火墙 provider。
- Docker bridge 网络通过 `ports` 或 `-p` 发布的 TCP、UDP 端口。
- 单个宿主机端口和 Docker 展开的端口范围。
- IPv4；Docker 已启用 IPv6 且存在对应 IPv6 `DOCKER-USER` 链或 `ip6 docker-bridges` 表时支持 IPv6。
- 按来源 IP 或 CIDR 设置黑名单或白名单。
- 新增、修改和删除防护策略。
- 展示当前发布端口所属的容器，以及 Compose/应用信息（能够获取时）。

## 4. 不在第一版支持

- Swarm ingress/overlay 发布端口。
- macvlan、ipvlan 网络的直接访问控制。
- Kubernetes 或其他容器运行时。
- 限速、连接数限制、地域规则、时间规则和日志审计。
- 自动识别公网网卡或自动生成可信网段。
- 修改 Docker 的 `daemon.json`。
- 设置 `iptables=false` 后由 1Panel 接管 Docker 的全部网络规则。
- 管理 Docker 自己创建的 `DOCKER`、`DOCKER-FORWARD`、`DOCKER-BRIDGE` 等链。
- 将所有已发布端口默认改为拒绝访问。

`network_mode: host` 的容器不属于本功能。它与宿主机共用网络命名空间，继续由现有“端口 / IP 规则”管理。

## 5. 用户可见行为

首次进入页面且尚未初始化时，展示“初始化容器端口防护”。初始化只创建并绑定 1Panel 自有链，不改变任何 Docker 发布端口的访问行为。

页面以 Docker 发布端口为主体，而不是以底层 iptables 规则为主体。每条记录至少展示：

- 容器名称。
- 宿主机绑定地址。
- 宿主机发布端口和协议。
- 容器端口。
- 当前是否对外暴露。
- 防护状态。
- 被禁止的来源地址。

防护策略以发布端点为目标：

```text
HostIP + HostPort + Protocol + IP Family
```

例如：

```text
0.0.0.0:8080/tcp -> nginx:80/tcp
禁止来源：192.168.1.0/24、203.0.113.10/32
```

启用后：

- 已建立连接和关联连接不受影响。
- 黑名单来源访问该发布端点时被丢弃，其他来源继续交给 Docker 原有规则判断。
- 白名单来源访问该发布端点时返回 Docker 原有规则，其他来源被丢弃。
- 没有配置防护策略的发布端口保持 Docker 默认行为。

访问策略提供三种模式：

```text
禁止所有访问
禁止指定来源 IP/CIDR
仅允许指定来源 IP/CIDR
```

“禁止指定来源”模式下来源列表不能为空。“仅允许指定来源”的白名单可以为空，空白名单等价于禁止所有访问。不提供 Allow/Deny 规则自由排序。

## 6. 底层规则边界

iptables/iptables-nft 后端中，1Panel 创建并只管理自己的链：

```text
FORWARD
└── DOCKER-USER
    └── 1PANEL_DOCKER
```

基本要求：

- `1PANEL_DOCKER` 从 `DOCKER-USER` 跳入。
- 1Panel 只增删 `1PANEL_DOCKER` 内的规则及指向它的跳转规则。
- 不修改、清空或删除 Docker 管理的链。
- 不修改用户在 `DOCKER-USER` 中已有的其他规则。
- 解绑时只移除 `DOCKER-USER` 跳转并保留 1Panel 自有链，不影响其他规则。
- IPv4 和 IPv6 使用各自规则空间中的同名链。

原生 nftables 后端没有 `DOCKER-USER` 链。1Panel 分别在 `ip`、`ip6` family 中创建自有表和链：

```text
table nft_1panel_docker
├── NFT_1PANEL_DOCKER_FORWARD  // type filter hook forward priority filter - 1
└── NFT_1PANEL_DOCKER
```

- `NFT_1PANEL_DOCKER_FORWARD` 只保留到 `NFT_1PANEL_DOCKER` 的入口跳转。
- 入口链优先级为 `filter - 1`，保证规则先于 Docker priority `filter` 的 forward 链执行。
- 只通过 `ip docker-bridges`、`ip6 docker-bridges` 判断 Docker 对应 family 是否可用，不向其中写入规则。
- 初始化、绑定、解绑和清理只修改 `nft_1panel_docker` 表。

链内逻辑保持简单：

```text
ESTABLISHED,RELATED                         -> RETURN
匹配发布端点且来源在黑名单                 -> DROP
匹配发布端点且来源在白名单                 -> RETURN
启用白名单但来源未命中                     -> DROP
禁止所有访问的发布端点                     -> DROP
其他流量                                   -> RETURN
```

1Panel 自有链中不生成 `ACCEPT` 规则。未被拒绝的流量使用 `RETURN`，返回后仍由 Docker 判断端口是否真实发布以及是否允许转发，避免绕过 Docker 自己的网络隔离规则。

进入自有 forward 入口链时 DNAT 已经完成。匹配宿主机发布端口时使用 conntrack 原始目标信息，而不是直接用当前数据包的目标端口：

- 原始目标地址用于区分绑定到不同宿主机 IP 的相同端口。
- 原始目标端口对应用户看到的宿主机发布端口。
- 当前目标端口通常已经是容器端口。

第一版不自动推断流量属于公网还是内网。“禁止指定来源”只拒绝用户填写的 IP/CIDR；“仅允许指定来源”拒绝未填写的其他来源；“禁止所有访问”和空白白名单拒绝所有经过自有链且命中该发布端点的新连接。

## 7. 最小数据模型

策略单独存储，不复用宿主机 `FirewallRule` 表。最少包含：

```text
UUID
Family
HostIP
HostPort
Protocol
Mode            // deny_sources、allow_sources 或 deny_all
Sources         // 被禁止或允许的 IP/CIDR 列表
Description
CreatedAt
UpdatedAt
```

第一版不把容器 ID 作为策略身份。容器、Compose 和应用归属从 Docker 当前运行状态动态关联，仅用于展示。

这样可以避免 Compose 重建容器后策略失效。相同发布端点之后被其他容器复用时，原有策略仍继续保护该端点。

同一个发布端点只允许存在一条有效策略，避免规则顺序产生歧义。

## 8. 状态与生命周期

数据库中的策略是期望状态，iptables 或 nftables 中的规则是运行状态。

### 8.1 初始化

容器端口防护有独立的初始化状态，不复用现有 `1PANEL_BASIC` 或 `1PANEL_FORWARD` 的初始化状态。

初始化前置条件：

- Docker 正在运行。
- Docker 使用 iptables、iptables-nft 或原生 nftables backend。
- iptables 后端要求 IPv4 的 `iptables` 和 `DOCKER-USER` 链可用。
- nftables 后端要求 `nft` 命令和 `ip docker-bridges` 表可用。

初始化执行：

1. 按当前 Docker backend 创建 IPv4 `1PANEL_DOCKER` 链，或 `nft_1panel_docker` 表及其两个链；已存在时直接复用。
2. 确保 `DOCKER-USER` 或 nftables base chain 第一条存在到自有规则链的唯一跳转。
3. 在链末尾保留 `RETURN`。
4. Docker IPv6 可用且存在 IPv6 `DOCKER-USER` 或 `ip6 docker-bridges` 表时，同步初始化 IPv6 链；IPv6 不可用不阻止 IPv4 初始化。
5. 初始化及绑定状态直接通过当前 backend 的自有链和入口跳转判断，不保存额外状态。

初始化本身不创建 `DROP` 规则，因此不会改变当前容器端口的访问状态。只有用户保存具体防护策略后，才向 `1PANEL_DOCKER` 写入拒绝规则。

未初始化时可以保存策略，策略显示为未生效；初始化后自动下发。Docker 未运行、backend 不受支持或当前 backend 的 Docker 规则入口不存在时，初始化失败并直接提示原因，不尝试接管其他 `FORWARD` 规则。

解绑容器端口防护时：

1. 删除 `DOCKER-USER` 或 nftables base chain 中指向自有规则链的跳转。
2. 保留自有规则链和数据库中的策略，并显示为未生效。
3. 再次绑定后恢复已有策略。

### 8.2 恢复与同步

以下场景执行一次简单的 reconcile：

- Agent 启动且检测到 `1PANEL_DOCKER` 链已初始化。
- 用户新增、修改或删除策略。
- 1Panel 执行 Docker 启动或重启后。
- 1Panel 执行防火墙启动、重启或 reload 后。
- 用户在容器端口防护页面显式执行“同步”操作时。

reconcile 只需要：

1. 检查当前 backend 的 Docker 规则入口是否存在。
2. 1Panel 对应的自有链不存在时不自动初始化。
3. 按已有策略通过 `iptables-restore --noflush` 或单次 `nft -f -` 事务重建已存在的 1Panel 自有链，不自动绑定。
4. 回读自有链及对应入口跳转并返回初始化、绑定和生效状态。

Docker 未运行或当前 backend 的 Docker 规则入口不存在时保留期望状态，页面显示“未生效”，不创建替代的 `FORWARD` 接管逻辑。

第一版不监听所有 Docker 事件。由于策略绑定的是发布端点而不是容器 ID，容器重建不要求同步修改策略；页面刷新时重新关联容器信息即可。

列表查询只读取 Docker API、数据库策略和 iptables 运行状态并拼装响应，不执行 reconcile，也不修改系统规则。页面提供独立的“同步”操作，用于按需重新下发数据库中的期望策略。

## 9. 最小校验和错误处理

只做保证规则能够安全下发的必要校验：

- 端口为 `1-65535`。
- 协议为 TCP 或 UDP。
- HostIP、被禁止的来源 IP/CIDR 与 Family 一致。
- 同一发布端点不存在重复有效策略。
- 下发前确认命令和目标链属于 1Panel 允许的固定范围。

不增加复杂的网络可达性判断、网段归属推断或冲突预测。

规则修改串行执行。下发失败时返回错误并保留数据库期望状态，页面显示运行状态与期望状态不一致，后续可再次 reconcile。第一版不要求跨数据库和防火墙运行时的完整分布式事务。

## 10. 与现有页面的联动

现有“端口 / IP 规则”页补充 Docker 运行态识别：

- 宿主机进程：当前 INPUT 规则可以生效。
- Docker bridge 发布端口：提示当前 INPUT 规则不能直接保护该端口，并提供跳转。
- Docker host 网络端口：提示由当前 INPUT 规则管理。
- 已配置 Docker 防护时展示其防护状态。

Docker 占用是运行态信息，不给宿主机防火墙规则增加永久的 `isDocker` 字段。

## 11. 第一版验收条件

- 在 firewalld、UFW、iptables、nftables provider 下，只要 Docker 使用受支持的 iptables、iptables-nft 或 nftables backend，防护行为一致。
- 可以独立初始化和关闭容器端口防护链，初始化本身不影响现有端口访问。
- 能发现并展示 Docker bridge 发布端口。
- 能为单个发布端点配置禁止所有访问、禁止指定来源或仅允许指定来源。
- 命中黑名单的来源被丢弃；白名单模式仅放行配置来源，空白白名单拒绝所有来源。
- 未配置策略的 Docker 发布端口不受影响。
- Docker 重启后可以恢复 1Panel 自有链和规则。
- firewalld/UFW reload 后可以通过 reconcile 恢复规则。
- 删除策略后，该端点恢复 Docker 默认访问行为。
- 用户已有的 `DOCKER-USER` 规则和 Docker 自有链不被修改。
- 宿主机 INPUT 规则与 Docker 防护规则在页面上有清晰区分。
