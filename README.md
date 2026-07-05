# Docker Gateway Backend

Docker Gateway 是一个远程 Docker 管理后台。它以当前 Go 服务作为控制端，通过部署在远程机器上的 Agent 管理 Docker 主机，实现容器编排配置编辑、日志查看、性能监控、镜像拉取、容器重启更新和业务程序远程部署。

## 项目定位

本项目后台负责做“控制面”，不建议直接暴露远程 Docker TCP API。推荐每台远程 Docker 机器安装一个轻量 Agent，由 Agent 连接本服务端并执行受控 Docker 操作。

核心目标：

- 统一管理多台远程 Docker 主机。
- 在线编辑和版本化保存容器 YAML 配置。
- 查看容器实时日志和历史日志。
- 监控主机、Docker Daemon、镜像、容器和 Compose 应用状态。
- 远程拉取镜像、创建容器、更新容器、重启容器、停止容器。
- 通过配置更新或镜像更新完成业务程序远程部署。
- 保留关键操作审计，避免远程执行不可追踪。

暂不作为第一阶段目标：

- 完整 Kubernetes 替代品。
- 任意 Shell 运维平台。
- 多租户计费系统。
- 直接把 Docker Socket 暴露到公网。

## 当前工程现状

当前后台位于 `docker-go` 目录：

- 语言：Go
- Web 框架：Gin，来自 `nav-common-go-lib` 初始化体系
- 默认数据库：SQLite
- 配置文件：`config.yaml`
- 已有模块：运行时设置 `settings`

当前代码仍是新工程雏形，后续需要优先补齐：

- 模块名残留检查。
- 设置路由中未实现接口的清理或补全。
- 网络、Volume 和 Compose 项目资源同步。
- 权限细化、操作审批和镜像仓库凭据。
- 指标存储保留策略和部署模板。

## 推荐架构

```text
Browser / Frontend
        |
        | HTTPS + JWT
        v
NAV Docker Gateway Server
        |
        | WebSocket / gRPC stream / HTTPS polling
        v
Remote Docker Agent
        |
        | Docker Engine API / docker compose
        v
Remote Docker Host
```

### Server

服务端负责：

- 用户认证和权限控制。
- 主机、Agent、容器、镜像、Compose 应用元数据管理。
- YAML 配置存储、校验、版本记录和发布。
- 任务创建、排队、下发、重试和审计。
- 日志流转发。
- 指标采集结果入库和查询。
- API 提供给前端使用。

### Agent

Agent 部署在每台远程 Docker 主机上，负责：

- 主动连接服务端，避免远程主机暴露管理端口。
- 上报主机信息、Docker 版本、容器列表、镜像列表。
- 执行服务端下发的 Docker 操作任务。
- 读取容器日志并按需实时推送。
- 采集 CPU、内存、磁盘、网络和容器资源使用指标。
- 在本地调用 Docker Engine API 或 `docker compose`。

### 为什么优先使用 Agent

- 安全性更好：远程 Docker Socket 不需要暴露公网。
- 网络适应性更强：Agent 主动连服务端，适合 NAT、内网机器。
- 权限可控：Agent 只实现有限命令，不提供任意 Shell。
- 可观测性更好：所有操作通过任务系统记录状态和审计。
- 后续扩展方便：可以支持边缘节点、离线重连和任务重试。

## 核心模块规划

### 1. 主机与 Agent 管理

管理远程机器和 Agent 生命周期。

能力：

- 创建主机记录。
- 生成 Agent 注册 Token。
- Agent 首次注册。
- Agent 心跳和在线状态。
- 上报主机基本信息。
- 标记 Agent 版本和 Docker 版本。
- 禁用、删除、重新生成 Token。

建议实体：

- `hosts`：远程主机。
- `agents`：Agent 连接实例。
- `agent_sessions`：当前在线连接。

### 2. Docker 资源管理

统一展示远程 Docker 资源。

能力：

- 容器列表、详情、状态同步。
- 镜像列表、镜像拉取、镜像删除。
- 网络列表。
- Volume 列表。
- Compose 项目列表。

建议实体：

- `docker_containers`
- `docker_images`
- `docker_networks`
- `docker_volumes`
- `compose_projects`

### 3. YAML 配置中心

用于在线编辑容器或 Compose 配置。

能力：

- 创建 YAML 配置。
- 在线编辑。
- YAML 语法校验。
- Docker Compose 配置校验。
- 保存版本。
- 对比版本差异。
- 发布到指定主机。
- 回滚到历史版本。

建议实体：

- `deploy_configs`：配置主表。
- `deploy_config_versions`：配置版本。
- `deploy_releases`：发布记录。

### 4. 任务系统

远程操作全部通过任务系统执行。

任务类型：

- `agent.ping`
- `docker.container.restart`
- `docker.container.stop`
- `docker.container.start`
- `docker.container.remove`
- `docker.image.pull`
- `docker.compose.up`
- `docker.compose.down`
- `docker.compose.restart`
- `docker.compose.pull`
- `docker.config.validate`
- `docker.config.deploy`

任务状态：

- `pending`
- `dispatched`
- `running`
- `success`
- `failed`
- `timeout`
- `cancelled`

建议实体：

- `tasks`
- `task_events`
- `operation_audits`

### 5. 日志中心

查看远程容器日志。

能力：

- 拉取容器最近 N 行日志。
- WebSocket 实时日志流。
- 按容器、主机、时间过滤。
- 保存关键任务日志。
- 支持前端停止订阅。

第一阶段可以不长期保存全部容器日志，只做按需读取和实时转发。

### 6. 监控中心

采集和查询性能指标。

主机指标：

- CPU 使用率。
- 内存使用率。
- 磁盘使用率。
- 网络 IO。
- Docker Daemon 状态。

容器指标：

- CPU 使用率。
- 内存使用量和限制。
- 网络 IO。
- Block IO。
- 重启次数。
- 健康检查状态。

建议实体：

- `host_metrics`
- `container_metrics`

第一阶段可以保存短周期数据，例如最近 24 小时或 7 天，后续再接 Prometheus、VictoriaMetrics 或 TimescaleDB。

### 7. 部署发布

通过 YAML 或镜像版本完成业务更新。

典型流程：

1. 用户编辑 Compose YAML。
2. 服务端保存为新版本。
3. 服务端创建校验任务。
4. Agent 在目标机器执行 `docker compose config`。
5. 校验通过后创建发布任务。
6. Agent 拉取镜像。
7. Agent 执行 `docker compose up -d`。
8. Agent 回传容器状态、日志和结果。
9. 服务端记录发布审计。

## 后台 API 规划

统一前缀沿用当前配置：`/api`

### 主机与 Agent

```http
GET    /api/hosts
POST   /api/hosts
GET    /api/hosts/:guid
PUT    /api/hosts/:guid
DELETE /api/hosts/:guid

POST   /api/hosts/:guid/agent-token
GET    /api/agents
GET    /api/agents/:guid
POST   /api/agents/register
POST   /api/agents/heartbeat
```

### Docker 资源

```http
GET    /api/hosts/:hostGuid/containers
GET    /api/hosts/:hostGuid/containers/:containerId
POST   /api/hosts/:hostGuid/containers/:containerId/start
POST   /api/hosts/:hostGuid/containers/:containerId/stop
POST   /api/hosts/:hostGuid/containers/:containerId/restart
DELETE /api/hosts/:hostGuid/containers/:containerId

GET    /api/hosts/:hostGuid/images
POST   /api/hosts/:hostGuid/images/pull
DELETE /api/hosts/:hostGuid/images/:imageId

GET    /api/hosts/:hostGuid/networks
GET    /api/hosts/:hostGuid/volumes
GET    /api/hosts/:hostGuid/compose-projects
```

### YAML 配置

```http
GET    /api/deploy-configs
POST   /api/deploy-configs
GET    /api/deploy-configs/:guid
PUT    /api/deploy-configs/:guid
DELETE /api/deploy-configs/:guid

GET    /api/deploy-configs/:guid/versions
POST   /api/deploy-configs/:guid/validate
POST   /api/deploy-configs/:guid/deploy
POST   /api/deploy-configs/:guid/rollback
```

### 日志

```http
GET    /api/hosts/:hostGuid/containers/:containerId/logs
GET    /api/ws/hosts/:hostGuid/containers/:containerId/logs
GET    /api/tasks/:taskGuid/logs
```

### 监控

```http
GET    /api/hosts/:hostGuid/metrics
GET    /api/hosts/:hostGuid/containers/:containerId/metrics
GET    /api/hosts/:hostGuid/overview
```

### 任务

```http
GET    /api/tasks
GET    /api/tasks/:guid
POST   /api/tasks/:guid/cancel
GET    /api/tasks/:guid/events
```

## Agent 通信协议规划

第一阶段推荐 WebSocket 长连接：

- Agent 主动连接：`GET /api/agent/ws?token=xxx`
- 服务端通过连接下发任务。
- Agent 执行后回传任务事件和最终结果。
- 断线后 Agent 自动重连。

消息结构建议：

```json
{
  "type": "task.dispatch",
  "taskGuid": "task_xxx",
  "hostGuid": "host_xxx",
  "payload": {
    "action": "docker.image.pull",
    "image": "nginx:1.27"
  },
  "timestamp": 1720000000
}
```

Agent 回传：

```json
{
  "type": "task.event",
  "taskGuid": "task_xxx",
  "status": "running",
  "message": "pulling image nginx:1.27",
  "data": {},
  "timestamp": 1720000001
}
```

后续如果任务量、日志流、指标流变大，可以升级为 gRPC stream 或 NATS。

## 数据库表草案

### hosts

| 字段 | 说明 |
| --- | --- |
| guid | 主机 GUID |
| name | 主机名称 |
| address | 展示用地址 |
| description | 备注 |
| status | online/offline/disabled |
| last_heartbeat_at | 最近心跳时间 |

### agents

| 字段 | 说明 |
| --- | --- |
| guid | Agent GUID |
| host_guid | 所属主机 |
| token_hash | 注册 Token Hash |
| version | Agent 版本 |
| docker_version | Docker 版本 |
| os | 操作系统 |
| arch | CPU 架构 |
| status | online/offline/disabled |

### deploy_configs

| 字段 | 说明 |
| --- | --- |
| guid | 配置 GUID |
| name | 配置名称 |
| type | compose/container |
| current_version_guid | 当前版本 |
| description | 备注 |

### deploy_config_versions

| 字段 | 说明 |
| --- | --- |
| guid | 版本 GUID |
| config_guid | 配置 GUID |
| version_no | 版本号 |
| content | YAML 内容 |
| checksum | 内容摘要 |
| created_by | 创建人 |

### tasks

| 字段 | 说明 |
| --- | --- |
| guid | 任务 GUID |
| host_guid | 目标主机 |
| type | 任务类型 |
| status | 任务状态 |
| payload | JSON 参数 |
| result | JSON 结果 |
| error_message | 错误信息 |
| timeout_seconds | 超时时间 |

## 推荐目录结构

```text
docker-go/
  apis/
    host_api.go
    agent_api.go
    container_api.go
    image_api.go
    deploy_config_api.go
    task_api.go
    log_api.go
    metric_api.go
  domains/
    host.go
    agent.go
    docker_resource.go
    deploy_config.go
    task.go
    metric.go
  routers/
    host_router.go
    agent_router.go
    docker_router.go
    deploy_router.go
    task_router.go
  services/
    host_service.go
    agent_service.go
    task_service.go
    deploy_service.go
    docker_query_service.go
  agent/
    hub.go
    session.go
    protocol.go
    dispatcher.go
  docker/
    command.go
    compose.go
    models.go
  utils/
```

## 安全设计

必须优先保证：

- 服务端 API 使用 JWT 鉴权。
- Agent 使用一次性注册 Token 或长期 Token Hash。
- Agent Token 只保存 Hash，不保存明文。
- 任务 payload 白名单校验。
- 禁止第一阶段提供任意 Shell 命令执行。
- 敏感字段脱敏记录。
- 所有远程修改类操作写入审计日志。
- YAML 发布前必须校验。
- Agent 只能执行绑定主机上的任务。

## 配置规划

后续可以在 `config.yaml` 增加：

```yaml
docker-gateway:
  agent:
    heartbeat-timeout: 60s
    task-timeout: 10m
    log-stream-timeout: 30m
  metrics:
    collect-interval: 10s
    retention-days: 7
  deploy:
    workspace-path: ./data/deploy
    max-config-size: 2MiB
```

## 开发里程碑

### M1：后台基础骨架

- 修复当前模块名和路由占位问题。
- 新增 Host、Agent、Task 基础表。
- 实现 Host CRUD。
- 实现 Agent 注册、心跳、在线状态。
- 实现任务创建、查询和事件记录。

### M2：Agent 长连接

- 服务端实现 Agent WebSocket Hub。
- Agent 可以注册并保持在线。
- 服务端可以下发 ping 任务。
- Agent 回传任务状态。
- 支持断线重连和任务超时。

### M3：Docker 基础操作

- Agent 调用 Docker Engine API。
- 同步容器列表、镜像列表。
- 支持容器启动、停止、重启。
- 支持镜像拉取。
- 支持基础任务审计。

### M4：日志和监控

- 容器日志最近 N 行查询。
- 容器日志 WebSocket 实时流。
- 主机和容器指标采集。
- 首页概览接口。

### M5：YAML 配置和部署

- YAML 配置 CRUD。
- 配置版本管理。
- Compose 配置校验。
- 发布、回滚、部署记录。
- 执行 `docker compose pull/up/down/restart`。

### M6：增强能力

- 权限细化到主机和操作。
- 指标存储优化。
- 镜像仓库凭据管理。
- 部署模板。
- 操作审批。
- Agent 自动升级。

## 第一阶段最小可用闭环

建议先实现这个闭环：

1. 后台创建主机。
2. 后台生成 Agent Token。
3. 远程机器启动 Agent 并注册。
4. 后台显示 Agent 在线。
5. 后台创建一个 `docker.image.pull` 任务。
6. Agent 拉取镜像并回传结果。
7. 后台记录任务事件。
8. 后台可以查询任务状态。

这个闭环跑通后，再扩展容器重启、日志、监控和 Compose 部署会稳很多。

## 本地启动

启动服务端：

```bash
go mod tidy
go run .
```

默认配置：

- 服务端口：`8888`
- API 前缀：`/api`
- 数据库：SQLite
- 数据目录：`./data`

启动 Agent：

```bash
go run ./cmd/nav-docker-agent \
  -server http://127.0.0.1:8888/api \
  -token nav_agent_xxx
```

如果已经知道 Agent GUID，也可以显式传入：

```bash
go run ./cmd/nav-docker-agent \
  -server http://127.0.0.1:8888/api \
  -agent-guid agent_xxx \
  -token nav_agent_xxx
```

Agent 默认通过 Docker SDK 读取本机 Docker 配置，也可以指定 Docker Host：

```bash
go run ./cmd/nav-docker-agent \
  -server http://127.0.0.1:8888/api \
  -token nav_agent_xxx \
  -docker-host unix:///var/run/docker.sock
```

Compose 部署会在 Agent 本地工作目录写入 `compose.yaml` 后执行 `docker compose`，默认工作目录为 `~/.nav-docker/workspaces`，也可以指定：

```bash
go run ./cmd/nav-docker-agent \
  -server http://127.0.0.1:8888/api \
  -token nav_agent_xxx \
  -workspace /data/nav-docker/workspaces
```

M3 已实现的 Docker 基础接口：

```http
GET    /api/hosts/:guid/containers
POST   /api/hosts/:guid/containers/sync
POST   /api/hosts/:guid/containers/:containerId/start
POST   /api/hosts/:guid/containers/:containerId/stop
POST   /api/hosts/:guid/containers/:containerId/restart
DELETE /api/hosts/:guid/containers/:containerId

GET    /api/hosts/:guid/images
POST   /api/hosts/:guid/images/sync
POST   /api/hosts/:guid/images/pull
```

M4 已实现的日志和监控接口：

```http
GET    /api/hosts/:guid/containers/:containerId/logs?tail=200
GET    /api/ws/hosts/:guid/containers/:containerId/logs?tail=200&timeoutSeconds=3600

POST   /api/hosts/:guid/metrics/sync
GET    /api/hosts/:guid/metrics
GET    /api/hosts/:guid/metrics/containers
GET    /api/hosts/:guid/overview
```

说明：

- 最近日志接口会创建 `docker.container.logs` 任务，日志结果在任务 `result` 中。
- 实时日志接口会创建 `docker.container.logs.stream` 任务，并把 Agent 回传的 `task.event` 通过 WebSocket 转发给前端。
- 指标同步接口会创建 `docker.metrics.snapshot` 任务，Agent 采集主机指标和运行中容器 stats 后回传并入库。

M5 已实现的 YAML 配置和部署接口：

```http
GET    /api/deploy-configs/list
POST   /api/deploy-configs
GET    /api/deploy-configs/:guid
PUT    /api/deploy-configs/:guid
DELETE /api/deploy-configs/:guid

GET    /api/deploy-configs/:guid/versions
POST   /api/deploy-configs/:guid/versions
GET    /api/deploy-configs/:guid/versions/:versionGuid

POST   /api/deploy-configs/:guid/validate
POST   /api/deploy-configs/:guid/deploy
POST   /api/deploy-configs/:guid/rollback

GET    /api/deploy-releases/list
GET    /api/deploy-releases/:guid
```

说明：

- 配置类型当前支持 `compose`，服务端会做 YAML 基础校验和 `services` 结构校验。
- 保存配置会生成版本；内容变化时 `PUT /api/deploy-configs/:guid` 会自动生成新版本。
- 远端校验会创建 `docker.config.validate` 任务，由 Agent 执行 `docker compose config`。
- 发布会创建 `docker.config.deploy` 任务和 `deploy_releases` 记录，支持 `up/down/restart/pull` 动作。
- 回滚会选择历史版本并以 `docker compose up -d` 发布，成功后更新配置当前版本。
- Agent 不执行任意 Shell，只通过固定参数调用 `docker compose`。

M6 已实现的增强能力接口：

```http
GET    /api/agent-upgrade-packages/list
POST   /api/agent-upgrade-packages
GET    /api/agent-upgrade-packages/:guid
PUT    /api/agent-upgrade-packages/:guid
DELETE /api/agent-upgrade-packages/:guid
POST   /api/agents/:guid/upgrade

GET    /api/registry-credentials/list
POST   /api/registry-credentials
GET    /api/registry-credentials/:guid
PUT    /api/registry-credentials/:guid
DELETE /api/registry-credentials/:guid

GET    /api/deploy-templates/list
POST   /api/deploy-templates
GET    /api/deploy-templates/:guid
PUT    /api/deploy-templates/:guid
DELETE /api/deploy-templates/:guid
POST   /api/deploy-templates/:guid/render
POST   /api/deploy-templates/:guid/create-config

GET    /api/operation-policies/list
POST   /api/operation-policies
GET    /api/operation-policies/:guid
PUT    /api/operation-policies/:guid
DELETE /api/operation-policies/:guid

GET    /api/operation-approvals/list
POST   /api/operation-approvals
GET    /api/operation-approvals/:guid
POST   /api/operation-approvals/:guid/approve
POST   /api/operation-approvals/:guid/reject
POST   /api/operation-approvals/:guid/cancel

GET    /api/operation-audits/list
GET    /api/operation-audits/:guid
```

说明：

- 镜像拉取接口支持 `registryCredentialGuid`，服务端会生成 Docker `registryAuth` 下发给 Agent，接口响应和审计中不会返回明文密码。
- 部署模板支持 `{{name}}` 和 `${name}` 两种占位符，渲染后会做 Compose YAML 基础校验。
- 操作策略可按 `hostGuid + action + resourceType` 配置是否必须审批；没有策略时保持原有直接执行行为。
- 需要审批时，先创建并 approve `operation-approvals`，再在部署、镜像拉取或容器操作请求中传入 `approvalGuid`。
- 所有通过任务系统创建的远程操作都会写入 `operation-audits`，任务状态变化会同步审计状态、结果和错误信息。
- Agent 自动升级通过 `agent.upgrade` 任务执行，Agent 会下载新二进制、校验 SHA256、可选校验 Ed25519 签名、备份旧二进制、替换当前可执行文件，然后延迟退出。
- Agent 自动升级需要 systemd、supervisor、Docker restart policy 或其它进程管理器负责重启；Agent 自身不会 fork 新进程守护自己。

Agent 自动升级启动参数：

```bash
go run ./cmd/nav-docker-agent \
  -server http://127.0.0.1:8888/api \
  -agent-guid agent_xxx \
  -token nav_agent_xxx \
  -upgrade-public-key BASE64_ED25519_PUBLIC_KEY \
  -require-upgrade-signature=true
```

升级包元数据示例：

```json
{
  "version": "v0.2.0",
  "os": "linux",
  "arch": "amd64",
  "downloadUrl": "https://example.com/nav-docker-agent-linux-amd64",
  "sha256": "64位hex编码",
  "signature": "base64-ed25519-signature",
  "status": "enabled",
  "releaseNotes": "support compose deploy and metrics"
}
```

下发升级任务示例：

```json
{
  "packageGuid": "pkg_xxx",
  "force": false,
  "restartDelaySeconds": 3,
  "approvalGuid": "approval_xxx",
  "operator": "admin"
}
```

## 后续优先事项

1. 补齐 Docker network、volume 和 Compose project 同步。
2. 增加部署版本 diff 和模板变量校验规则。
3. 增加真正的加密密钥管理或对接外部 Secret Manager。
4. 优化指标保留策略和监控趋势查询。
5. 设计 Agent 自动升级的签名校验、灰度和回滚机制。
