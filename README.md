# Prometheus SQL Exporter

[源项目地址](https://github.com/justwatchcom/sql_exporter)

集成支持国产数据库，目前支持达梦，人大金仓数据库

---

* 在**job_config.yml**中自定义配置数据库监测指标
*note:需要更改*配置信息里的数据库连接信息：
```yaml
# config.yml 默认配置了达梦与金仓作为示例
    connections:
      - '{DATABASE_DSN}://{USERNAME}:{PASSWORD}!@{HOST}:{PORT}'
```
* 在**config.yml**中自定义端口与日志级别

```yaml
# 日志级别：
INFO
DEBUG
WARN
ERROR
```

## 使用

* docker运行：

```shell
# 进入项目目录

chmod +x gorun.sh

./gorun.sh

```

* 直接运行（需要golang1.19依赖）

```shell
# 进入项目目录
go mod download
go build
./sql_zh_exporter

# 查看可配置参数：
./sql_zh_exporter --help
```

* [下载](https://github.com/NoahAmethyst/sql_zh_exporter/releases)对应操作系统与架构版本：

```shell
# 直接运行
./sql_zh_exporter

# 查看可配置参数：
./sql_zh_exporter --help

```