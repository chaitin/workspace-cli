# SafeLine CLI

SafeLine WAF Skyview API 命令行工具。

## 安装

```bash
go build -o safeline-cli
```

## 配置

支持通过命令行参数、环境变量或项目根目录 `.env` 配置：

| 参数 | 环境变量 | 说明 |
|------|----------|------|
| `--url` | `SAFELINE_URL` | Skyview API 地址（必填） |
| `--api-key` | `SAFELINE_API_KEY` | API Token |
| `--indent` | - | 格式化 JSON 输出 |
| `--insecure` | - | 跳过 TLS 证书验证 |

## 命令概览

```
safeline [全局参数] <命令> [子命令] [参数]
```

### 全局参数

```bash
--url string        Skyview API 地址
--api-key string    API Token
--indent            格式化 JSON 输出
--insecure          跳过 TLS 证书验证
```

---

## 命令详情

### stats - 统计概览

查看 SafeLine 统计数据。

```bash
# 查看 24 小时统计
safeline stats overview --duration h

# 查看 30 天统计
safeline stats overview --duration d
```

---

### site - 站点管理

管理 SafeLine 站点（网站）。

#### 列出站点

```bash
safeline site list
```

#### 查看站点详情

```bash
safeline site get <id>
```

#### 更新站点策略组

```bash
safeline site update <id> --policy-group <group-id>
safeline site update <id> --policy-group 0  # 取消关联
```

#### 启用/禁用站点

```bash
safeline site enable <id>
safeline site disable <id>
```

---

### ip-group - IP 组管理

管理 IP 组（别名：`ipgroup`）。

#### 列出 IP 组

```bash
safeline ip-group list
safeline ip-group list --name "office"  # 按名称过滤
safeline ip-group list --count 50       # 限制数量
```

#### 查看 IP 组详情

```bash
safeline ip-group get <id>
```

#### 创建 IP 组

```bash
safeline ip-group create --name "Office" --ips "192.168.1.0/24,10.0.0.1"
safeline ip-group create --name "DC" --ips "172.16.0.0/16" --comment "数据中心"
```

#### 删除 IP 组

```bash
safeline ip-group delete <id>
safeline ip-group delete 1 2 3  # 批量删除
```

#### 添加/移除 IP

```bash
safeline ip-group add-ip <id> --ips "192.168.3.0/24,10.0.1.0/24"
safeline ip-group remove-ip <id> --ips "192.168.3.0/24"
```

---

### acl - ACL 管理

管理 ACL（访问控制/限速）规则。

#### acl template - ACL 模板管理

##### 列出模板

```bash
safeline acl template list
safeline acl template list --name "limit"  # 按名称过滤
```

##### 查看模板详情

```bash
safeline acl template get <id>
```

##### 创建模板

**手动规则（指定 IP）：**

```bash
safeline acl template create --name "Block IPs" \
  --template-type manual \
  --target-type cidr \
  --action forbid \
  --targets "192.168.1.100,10.0.0.50"
```

**手动规则（使用 IP 组）：**

```bash
safeline acl template create --name "Block Group" \
  --template-type manual \
  --target-type cidr \
  --action forbid \
  --ip-groups 1,2
```

**自动规则（自动限速）：**

```bash
safeline acl template create --name "Rate Limit" \
  --template-type auto \
  --period 60 --limit 100 \
  --action forbid
```

**限速动作：**

```bash
safeline acl template create --name "Throttle" \
  --template-type auto \
  --period 60 --limit 100 \
  --action limit_rate \
  --limit-rate-limit 10 \
  --limit-rate-period 60
```

##### 启用/禁用/删除模板

```bash
safeline acl template enable <id>
safeline acl template disable <id>
safeline acl template delete <id>
```

#### acl rule - ACL 规则管理（被封禁用户）

##### 列出规则

```bash
safeline acl rule list --template-id <id>
```

##### 删除规则（解封）

```bash
safeline acl rule delete <id>
safeline acl rule delete <id> --add-to-whitelist  # 同时加入白名单
```

##### 清除所有规则

```bash
safeline acl rule clear --template-id <id>
safeline acl rule clear --template-id <id> --add-to-whitelist
```

---

### policy-group - 策略组管理

管理策略组（检测引擎配置）。

#### 列出策略组

```bash
safeline policy-group list
```

#### 查看策略组详情

```bash
safeline policy-group get <id>
```

#### 更新模块状态

```bash
# 启用模块
safeline policy-group update <id> --module m_sqli,m_xss --state enabled

# 禁用模块
safeline policy-group update <id> --module m_cmd_injection,m_ssrf --state disabled
```

**可用模块：**

| 模块键 | 说明 |
|--------|------|
| `m_sqli` | SQL 注入检测 |
| `m_xss` | XSS 检测 |
| `m_cmd_injection` | 命令注入检测 |
| `m_file_include` | 文件包含检测 |
| `m_file_upload` | 文件上传检测 |
| `m_php_code_injection` | PHP 代码注入检测 |
| `m_php_unserialize` | PHP 反序列化检测 |
| `m_java` | Java 检测 |
| `m_java_unserialize` | Java 反序列化检测 |
| `m_ssrf` | SSRF 检测 |
| `m_ssti` | SSTI 检测 |
| `m_csrf` | CSRF 检测 |
| `m_scanner` | 扫描器检测 |
| `m_response` | 响应检测 |
| `m_rule` | 内置规则 |

---

### policy-rule - 策略规则管理

管理自定义策略规则。

#### 列出规则

```bash
safeline policy-rule list              # 全局规则（默认）
safeline policy-rule list --global=false  # 自定义规则
```

#### 查看规则详情

```bash
safeline policy-rule get <id>
```

#### 创建规则

**简单模式：**

```bash
safeline policy-rule create --comment "Block admin" \
  --target urlpath --cmp infix --value "/admin" \
  --action deny
```

**JSON 模式（复杂条件）：**

```bash
safeline policy-rule create --comment "Complex rule" \
  --pattern-json '{"$AND":[{"infix":{"urlpath":"admin"}}]}' \
  --action deny
```

**参数说明：**

| 参数 | 说明 |
|------|------|
| `--comment` | 规则描述（必填） |
| `--action` | 动作：`deny`、`dry_run`、`allow`（必填） |
| `--risk-level` | 风险等级：0=无, 1=低, 2=中, 3=高 |
| `--enabled` | 是否启用（默认 true） |
| `--expire-time` | 过期时间戳（0=永不过期） |
| `--target` | 目标字段（简单模式） |
| `--cmp` | 比较操作符（简单模式） |
| `--value` | 匹配值（简单模式） |
| `--pattern-json` | 完整的 pattern JSON（复杂模式） |

#### 查看可用的目标和操作符

```bash
# 列出所有目标
safeline policy-rule targets

# 查看特定目标的操作符
safeline policy-rule targets --cmp urlpath
```

#### 启用/禁用/删除规则

```bash
safeline policy-rule enable <id>
safeline policy-rule disable <id>
safeline policy-rule delete <id>
```

---

### monitor - 监控

监控 SafeLine 节点和系统状态。

#### monitor node - 节点监控

```bash
# 列出所有节点
safeline monitor node list

# 查看节点详情
safeline monitor node get <id>
```

#### monitor bypass - 旁路状态

```bash
safeline monitor bypass list
```

---

### system - 系统管理

系统管理命令。

#### 查看许可证

```bash
safeline system license
```

#### 获取机器 ID

```bash
safeline system machine-id
```

#### 系统日志

```bash
safeline system log list
safeline system log list --count 50 --offset 0
```

---

### log - 日志查询

查询 SafeLine 日志。

#### log detect - 检测日志（攻击日志）

```bash
# 列出检测日志
safeline log detect list
safeline log detect list --count 50

# 分页查询
safeline log detect list --current-page 1 --target-page 2 --tail-sort '[1743628800, "abc123"]'

# 查看详情
safeline log detect get --event-id "6edb4c7eb69042cd996045e3ee5526d9" --timestamp "1774857841"
```

#### log access - 访问日志

```bash
# 列出访问日志
safeline log access list
safeline log access list --count 50

# 分页查询
safeline log access list --current-page 1 --target-page 2 --tail-sort '[1743628800, "abc123"]'

# 查看详情
safeline log access get --event-id "1e1ef8e9b21d42cd996045e3ee5526d9" --req-start-time "1775117700"
```

#### log rate-limit - 限速日志（别名：`rl`）

```bash
safeline log rate-limit list
safeline log rate-limit list --count 50 --offset 20
```

---

### network - 网络配置

网络配置命令（**仅硬件模式支持**）。

> 注意：软件模式下执行这些命令会返回错误。

#### network workgroup - 工作组（别名：`wg`）

```bash
# 列出工作组
safeline network workgroup list

# 查看工作组详情
safeline network workgroup get <name>
```

#### network interface - 网络接口（别名：`if`）

```bash
# 列出网络接口
safeline network interface list

# 查看接口 IP
safeline network interface ip <name>
```

#### network gateway - 默认网关（别名：`gw`）

```bash
safeline network gateway get
```

#### network route - 静态路由（别名：`sr`）

```bash
safeline network route list
```

---

## 输出格式

默认输出表格格式，使用 `--indent` 参数输出格式化的 JSON：

```bash
# 表格输出
safeline site list

# JSON 输出
safeline site list --indent
```

---

## 示例

### 快速开始

```bash
# 设置环境变量
export SAFELINE_URL="https://your-safeline-server"
export SAFELINE_API_KEY="your-token"

# 查看统计
safeline stats overview --duration h

# 列出站点
safeline site list

# 创建 IP 组
safeline ip-group create --name "Office" --ips "192.168.1.0/24"

# 创建 ACL 规则封禁 IP
safeline acl template create --name "Block Malicious" \
  --template-type manual \
  --target-type cidr \
  --action forbid \
  --targets "10.0.0.100"
```

### 查看攻击日志

```bash
# 列出最近的攻击日志
safeline log detect list --count 20

# 查看详情
safeline log detect get --event-id "<event-id>" --timestamp "<timestamp>"
```

### 配置防护策略

```bash
# 启用 SQL 注入和 XSS 检测
safeline policy-group update 1 --module m_sqli,m_xss --state enabled

# 添加自定义规则拦截特定路径
safeline policy-rule create --comment "Block admin path" \
  --target urlpath --cmp infix --value "/admin" \
  --action deny --risk-level 3
```

---

## 开发

```bash
# 构建
go build -o safeline-cli

# 运行测试
go test ./...
```
