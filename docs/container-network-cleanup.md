# 批量清理未使用的容器网络

`POST /api/v2/containers/network/clean`，使用现有 1Panel 认证，不需要请求体。
这是异步接口，立即返回 `data.taskID`。实际结果在任务中心和任务日志查看，HTTP 请求结束不会取消任务。

规则：

- 永远保留 `none`、`host`、`bridge`、`1panel-network`。
- 保留所有容器（包括停止容器）引用的网络，删除前再通过 NetworkInspect 检查连接。
- 不检查 Compose 文件或应用配置引用；仅被这些配置引用、没有容器连接的普通网络也会被删除。
- 当前仅清理本地普通网络，跳过 Swarm、ingress 和 config-only 网络。
- Docker 删除时发现新连接会拒绝删除；不会强制断开容器。

成功响应示例：

```json
{"code":200,"data":{"taskID":"服务端生成的任务 UUID"}}
```

任务类型为 Container，操作为 TaskClean。网络页“清理网络”按钮调用此接口，拿到 taskID 后立即打开任务日志，并刷新任务中心。
每个网络处理完成后写入名称、ID 和结果；有容器连接或 Docker 返回占用冲突时，明确记录“未删除”。最后记录已删除、跳过和失败数量。
跳过在用网络属于正常结果；检查/删除失败会将任务标记为失败，已删除的网络不回滚。

日志调用链：

- `agent/app/service/container_network_cleanup.go`：NewTaskWithOps 创建任务，AddSubTask 注册清理步骤，Execute 执行任务，t.Log 写逐项日志。
- `agent/app/task/task.go`：任务持久化，维护状态，写入开始、结束标记；日志文件为 `<global.Dir.TaskDir>/Container/<taskID>.log`。
- `frontend/src/views/container/network/index.vue`：请求成功后调用 openTaskLog(taskID)。
- `frontend/src/components/log/task/index.vue`：openWithTaskID 打开日志弹窗。
- `frontend/src/components/log/file/index.vue`：通过 `/logs/tasks/read` 按行读取任务日志。

范围：不检查 Compose 文件或应用配置引用；仅清理本地普通网络。原 `/containers/prune` 接口仍保持原逻辑。
测试使用假 Docker 客户端和临时任务数据库，不执行真实网络删除。
