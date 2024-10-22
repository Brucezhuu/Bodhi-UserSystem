# 使用官方的Go语言镜像作为基础镜像
FROM golang:1.23-alpine AS builder

# 设置环境变量
ENV GO111MODULE=on

# 创建并设置工作目录
WORKDIR /app

# 复制 go.mod 和 go.sum 文件，以便首先安装依赖项
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 将项目文件复制到容器，包括 .env 文件
COPY . .

# 构建Go应用
RUN go build -o app .

# 创建最终的运行阶段
FROM alpine:latest

# 设置工作目录
WORKDIR /root/

# 将构建好的二进制文件复制到最终镜像
COPY --from=builder /app/app .

# 复制 .env 文件到容器
COPY --from=builder /app/.env .

# 暴露服务端口
EXPOSE 8080 5432

# 运行Go应用
CMD ["./app"]
