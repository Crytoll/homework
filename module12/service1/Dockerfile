# 使用 multi-stage build
FROM golang:1.17-alpine as builder

# 环境变量
ENV CGO_ENABLED=0 GO111MODULE=on GOPROXY=https://goproxy.cn,direct

# copy 源码
COPY . /app

# 切换工作目录
WORKDIR /app

# 编译
RUN GOOS=linux go build -installsuffix cgo -o httpserver main.go

# 工作镜像, 推荐小镜像
FROM alpine:3.15

# 切换工作目录
WORKDIR /app

# 拷贝上面编译出来的可执行文件
COPY --from=builder /app/httpserver /app/

USER 1000

# 启动程序
CMD ["/app/httpserver"]
