FROM golang:alpine as builder
# WORKDIR：工作目录路径
WORKDIR /go/src/github.com/lihao20110/simple-douyin/server
# COPY All things inside the project and build
# COPY . . ：将我们项目中的所有内容复制到工作目录中。
COPY . .

RUN go env -w GO111MODULE=on \
    && go env -w GOPROXY=https://goproxy.cn,direct \
    && go env -w CGO_ENABLED=0 \
    && go env \
    && go mod tidy \
    && go build -o server .

FROM alpine:latest

#MAINTAINER 维护者信息   simple-douyin server
LABEL MAINTAINER="lihao@simple-douyin@github.com"

WORKDIR /go/src/github.comlihao20110/simple-douyin/server
# Dockerfile多阶段构建，--from=0参数，从前边的阶段中拷贝文件到当前阶段中，多个FROM语句时，0代表第一个阶段
COPY --from=0 /go/src/github.com/lihao20110/simple-douyin/server/server ./
COPY --from=0 /go/src/github.com/lihao20110/simple-douyin/server/config.docker.yaml ./

# EXPOSE 8080：告诉 docker 我们的代码将暴露端口8080。
EXPOSE 8080

#当我们运行这个镜像的容器时，它将从我们构建的可执行文件开始执行。
ENTRYPOINT ./server -c config.docker.yaml


# 运行docker命令创建镜像           镜像标签   .命令是当前文件夹下的Dockerfile
# docker build -f Dockerfile -t douyin-docker:latest .

# 运行容器命令
# docker run -d -p宿主机端口∶容器端口 --name 容器名称 镜像的标识镜像名称[:tag]
# -d∶ 代表后台运行容器
#-p 宿主机端口∶容器端口∶ 为了映射当前Linux的端口和容器的端口
#--name 容器名称∶指定容器的名称
