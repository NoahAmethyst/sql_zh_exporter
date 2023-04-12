# Prometheus SQL Exporter

[源项目地址](https://github.com/justwatchcom/sql_exporter)

集成支持国产数据库，目前支持达梦，人大金仓数据库

----

* 在**job_config.yml**中自定义配置数据库监测指标

* docker运行：
```
# 进入项目目录

chmod +x gorun.sh

./gorun.sh

```

* 直接运行（需要golang1.19依赖）
```
# 进入项目目录
go mod download
go build
./sql_zh_exporter
```