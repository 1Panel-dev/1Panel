# 批量清理未使用的容器网络

网络页使用原接口 `POST /api/v2/containers/prune`，请求体为：

```json
{"taskID":"客户端生成的 UUID","pruneType":"network","withTagAll":false}
```

`withTagAll=true` 时只清理创建至少 24 小时的网络，创建时间未知的网络保留。
Prune 是异步接口，提交成功后使用请求中的 `taskID` 查看日志。实际结果在任务中心和任务日志查看，HTTP 请求结束不会取消任务。

规则：

- 永远保留 `none`、`host`、`bridge`、`1panel-network`。
- 保留所有容器（包括停止容器）引用的网络，删除前再通过 NetworkInspect 检查连接。
- 不检查 Compose 文件或应用配置引用；仅被这些配置引用、没有容器连接的普通网络也会被删除。
- 当前仅清理本地普通网络，跳过 Swarm、ingress 和 config-only 网络。
- Docker 删除时发现新连接会拒绝删除；不会强制断开容器。

任务类型为 Container，操作为 TaskClean。网络页“清理网络”按钮调用 Prune 接口，提交成功后使用请求中的 taskID 打开任务日志，并刷新任务中心。
每个网络处理完成后写入名称、ID 和结果；有容器连接或 Docker 返回占用冲突时，明确记录“未删除”。最后记录已删除、跳过和失败数量。
跳过在用网络属于正常结果；检查/删除失败会将任务标记为失败，已删除的网络不回滚。

日志调用链：

- `agent/app/service/container.go`：Prune 创建和执行任务；`container_network_cleanup.go` 调用清理辅助函数并通过 t.Log 写逐项日志。
- `agent/app/task/task.go`：任务持久化，维护状态，写入开始、结束标记；日志文件为 `<global.Dir.TaskDir>/Container/<taskID>.log`。
- `frontend/src/views/container/network/index.vue`：请求成功后调用 openTaskLog(taskID)。
- `frontend/src/components/log/task/index.vue`：openWithTaskID 打开日志弹窗。
- `frontend/src/components/log/file/index.vue`：通过 `/logs/tasks/read` 按行读取任务日志。

范围：不检查 Compose 文件或应用配置引用；仅清理本地普通网络。原 `/containers/prune` 的 network 分支已使用上述保护逻辑。
测试使用假 Docker 客户端和临时任务数据库，不执行真实网络删除。
