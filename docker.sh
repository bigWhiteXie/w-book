docker pull prom/prometheus
docker run -itd --name=prometheus -v /etc/prometheus:/prometheus/config --restart=always -p 9091:9090 prom/prometheus --config.file=/prometheus/config/prometheus.yaml --web.enable-lifecycle

# 注意，该虚拟机启动的容器都会带有PROXY，需要取消该代理
docker pull grafana/grafana
docker run -itd --name=grafana \
--restart=always \
-p 3000:3000 \
-v /data/grafana-storage:/var/lib/grafana \
-e http_proxy="" \
-e HTTP_PROXY="" \
-e HTTPS_PROXY="" \
grafana/grafana

docker run -itd --name=otel-collector -v /etc/otel/otel-collector-config.yaml:/etc/otelcol-contrib/config.yaml -v /data/trace:/data/trace -p 4317:4317 -p 4318:4318 -p 11800:11800 -p 12800:12800 otel/opentelemetry-collector-contrib:0.114.0 


# 手动同步时间
sudo chronyc makestep

# 容器需要代理时
vi ~/.docker/config.json