FROM golang:1.25.3-alpine3.22 AS builder

# 设置工作目录
WORKDIR /app

# 复制go模块依赖
COPY go.mod go.sum ./

# 下载并缓存所有依赖
RUN go mod download

# 复制项目源码
COPY . .

# 编译Go程序
RUN go build -o /myapp .

# 运行镜像
FROM alpine:latest

WORKDIR /app

COPY --from=builder /myapp ./

ADD config ./config

# 时区配置
RUN apk update && apk add --no-cache tzdata
ENV TZ=Asia/Shanghai

# 暴露容器端口
#EXPOSE 8080

# 设置容器启动时执行的命令
CMD ["/app/myapp"]