# 构建阶段：使用 Go 官方镜像编译代码
# 运行阶段：使用轻量级的 Alpine 镜像
FROM debian:latest

# 设置工作目录
WORKDIR /app

# 从构建阶段复制编译好的二进制文件
COPY  ./main /app/main
RUN chmod +x /app/main
# 复制配置文件
COPY etc/user.yaml ./config.yaml

# 暴露端口
EXPOSE 20880

# 运行编译好的二进制文件
CMD ["/app/main", "-f", "/app/config.yaml"]

