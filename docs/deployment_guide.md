# PtLPOJ Deployment & Operations Guide

本文档专为**项目管理员、教师及系统维护人员**编写，详细说明了如何在局域网或内网公有服务器上（如实验室主机、校园自建轻量服务器）从零部署并运行 PtLPOJ 系统。

---

## 1. 系统要求与环境准备 (Prerequisites)

PtLPOJ 依赖极少，服务端的核心执行逻辑完全由编译好的无依赖二进制程序以及容器构成。

*   **操作系统**: Linux (推荐 Ubuntu 20.04/22.04 LTS 或 Debian 11+)。*注: 不建议在 Windows Server / macOS 上直接部署沙盒评测端，因为其重度依赖 Linux 原生的 Docker cgroups 资源限制。*
*   **Go 环境**: `Go 1.21` 或更高版本 (用于构建后端)。
*   **Docker 运行时**: 必须安装 `Docker Engine` 并确保当前运行服务端程序的用户已加入 `docker` 用户组（即可以执行无 `sudo` 的 `docker ps`）。

### 1.1 安装依赖示例 (Ubuntu)

```bash
# 1. 安装 Docker Engine
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh

# 2. 将当前用户加入 docker 组 (注销并重新登录以生效)
sudo usermod -aG docker $USER

# 3. 安装 Go 环境 (如系统还未安装)
wget https://go.dev/dl/go1.21.5.linux-amd64.tar.gz
sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin
```

---

## 2. 编译与运行 (Build & Run)

### 2.1 拉取代码与初始化沙盒镜像
第一步，系统需要提前拉取用于评测学生代码的基础沙盒环境。我们使用的是官方的极简 `alpine` 镜像以实现**毫秒级启动**。

```bash
git clone https://github.com/LynPtl/PtLPOJ.git
cd PtLPOJ/server

# 拉取极其轻量的 Alpine + Python 3 运行时作为评测沙盒 (约 15MB)
docker pull python:3.11-alpine
```

### 2.2 编译后端架构程序
```bash
# 下载 Go Module 依赖并编译为单个二进制可执行文件
go mod download
go build -o ptlpoj-server main.go
```

### 2.3 启动服务
你可以直接在宿主机前台运行，或使用 `systemd` / `tmux` / `nohup` 让其保持在后台运行。

```bash
# 配置环境变量并启动 (以下为假定的环境变量配置示范，可根据需要调整)
export PTLPOJ_PORT=8080
export PTLPOJ_DB_DSN="ptlpoj_dev.db"
export JWT_SECRET="your-super-strong-jwt-secret"

./ptlpoj-server
```

启动成功后，控制台会输出如下字样，表示 API Server 已经就绪，内置的 Judge Worker 也开始监听任务列队：
```text
2026/02/24 14:00:00 SQLite database connected successfully with WAL mode.
2026/02/24 14:00:00 API Router configured successfully
2026/02/24 14:00:00 Starting PtLPOJ Server on :8080
2026/02/24 14:00:00 Sandbox Manager initialized using [python:3.11-alpine]
```

---

## 3. 运维及数据库管理 (Operations & Storage)

### 3.1 SQLite WAL 设计与快照
PtLPOJ 系统没有外部数据库集群依赖。所有的用户、评测元数据存放在配置的 SQLite 数据库文件（如 `ptlpoj_dev.db`）中。

为了满足 Judge Workers 的高并发回写判定状态以及处理长轮询请求，底层已经强制开启了 `_journal_mode=WAL`。因此在您的部署目录下，您会看到三个核心文件：
1.  `ptlpoj_dev.db` (主数据)
2.  `ptlpoj_dev.db-wal` (高频预写日志缓冲)
3.  `ptlpoj_dev.db-shm` (共享内存映射)

**【备份策略】**：复制文件前，请确保在复制的瞬间使用 SQLite 自带的安全备份命令，或者暂短关闭 Server 后直接拷贝上述三个文件。

### 3.2 自定义用户白名单
在当前的系统安全版本下，只有处于白名单的用户邮箱才可以登录本系统获取 OTP 验证码。如果想要开放新同学的注册，可以在 SQLite 中执行 `INSERT` 语句或者后期通过 Admin API 导入。
*(注意：在初级开发版中，OTP 是直接打印在控制台标准输出的，实际部署推荐在代码内嵌企业 SMTP 或 SES 邮件发送钩子)*

---

## 4. 故障排查 (Troubleshooting)

| 症状 / 报错信息 | 常见原因 | 解决办法 |
| :--- | :--- | :--- |
| **"Cannot connect to the Docker daemon at unix:///var/run/docker.sock."** | 运行 Server 的用户没有 Docker 操作权限，或 Docker Engine 尚未启动。 | 检查 `systemctl status docker`，并使用 `sudo usermod -aG docker $USER` 提权。 |
| **"bind: address already in use"** | 设定的端口 (默认 8080) 已被其他进程占用。 | 使用 `lsof -i:8080` 杀死占用进程，或修改 `PTLPOJ_PORT` 换一个端口。 |
| **客户端评测一直处于 "PENDING" 卡死** | Server 遇到恐慌崩溃，导致 Channel 队列丢弃。或 Docker 守护进程宕机导致沙盒无限超时。 | 重新启动 `./ptlpoj-server`。系统带有**奔溃恢复(Crash Recovery)**逻辑，系统重启时会自动从数据库中拉起遗留为 `PENDING`/`RUNNING` 的孤儿判题。 |

## 5. (即将推出) 数据同步与教师管理台
在系统升级至 `Phase 8` 后，所有从基于文集的题库 (Markdown 配置) 向 SQLite 的迁移工作都将被内部自动化脚本接管；新题目的上传将统一开放 REST APIs。
