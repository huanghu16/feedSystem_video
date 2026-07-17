# ===== 构建 =====
FROM golang:1.26-alpine AS builder

# 设置工作目录
WORKDIR /app
# 设置国内 Go 代理，加速依赖下载
ENV GOPROXY=https://goproxy.cn,direct
#先拷贝依赖文件。利用Docker缓存层
COPY go.mod go.sum ./
RUN go mod download

# 拷贝代码，编译
COPY . .
# CGO 禁用，纯静态编译
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/server ./cmd/api

# ===== 运行 =====
FROM alpine:latest

# 安装 FFmpeg（视频处理需要）和 时区数据
RUN apk add --no-cache ffmpeg tzdata ca-certificates
ENV TZ=Asia/Shanghai

WORKDIR /app
# 从构建阶段拷编译产物
COPY --from=builder /app/server .
# 拷配置文件
COPY --from=builder /app/configs ./configs

# 上传目录（挂载卷持久化）
RUN mkdir -p uploads
VOLUME /app/uploads

EXPOSE 8080
CMD ["./server"]

