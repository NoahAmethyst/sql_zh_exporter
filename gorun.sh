#！/bin/bash
docker container stop sql_zh_exporter
docker container rm sql_zh_exporter
docker image rm sql_zh_exporter:latest
docker build -t sql_zh_exporter:latest .
docker run --name sql_zh_exporter --restart always -d -p 9237:9237 sql_zh_exporter:latest
docker logs -f sql_zh_exporter
