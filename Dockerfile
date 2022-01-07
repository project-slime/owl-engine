# Author: Stalker-lee
# Date: 2021-12-16 13:37:00
# Vesion: 1.0.0
# Description:
#       通过 docker 编译运行 owl-engine 镜像
#           (1) builder 为 构建部分
#           (2) prod 为 运行部分, 默认以 apollo 方式载入配置
#           (3) 说明: 应尽量满足如下特点: 体积小、构建快、够安全
#
# Docker Check:
#       文档: https://github.com/hadolint/hadolint
#       安装: brew install hadolint
#       运行: hadolint --config hadolint.yaml Dockerfile
#
# Build Command:
#       部署环境分为: dev, test, prod; 请依据实际环境设置 MODE 的值
#
#       docker build --build-arg MODE=dev -t owl-engine:v1.0.0 .
#       docker image ls
#       docker save -o owl-engine:v1.0.0.tar owl-engine:v1.0.0
#       docker load < owl-engine:v1.0.0.tar
#       docker run -d --name owl-engine owl-engine:v1.0.0

# 打包依赖阶段使用 golang 作为基础镜像
FROM golang:1.14.4 as builder

# 设置环境变量
ENV GO111MODULE="on" \
    GOPROXY="https://goproxy.cn,direct" \
    GOPATH="/go"

# owl-engine 存储路径
WORKDIR ${GOPATH}/src/owl-engine

# 复制当前根目录下的目标项目文件
COPY . .

RUN go mod download  \
    && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o ${GOPATH}/bin/owl-engine

# 镜像运行
FROM alpine:3.9 as prod

# 添加标签
LABEL version="1.0.0"
LABEL description="Rule calculation engine image by docker builder"
LABEL vendor="Open Source Incorporated"

# 设置构建时的命令行参数
ARG MODE=dev

# 设置环境变量
ENV GOPATH=/go \
    DEPLOYMODE=${MODE} \
    PROJECTDIR=/data/owl-engine \
    AGOLLO_CONF=${PROJECTDIR}/bin/app.properties

# 工作目录
WORKDIR ${PROJECTDIR}

# 注意: 在 Linux 环境下, docker 调用 /bin/sh -c 的运行命令, 不能使用 连接命令创建多个目录
# 注意: 需要将 app.properties 放在和 owl-engine 可执行程序同级目录
RUN echo "https://mirror.tuna.tsinghua.edu.cn/alpine/v3.8/main" > /etc/apk/repositories \
    && apk add jq \
    && rm -rf /var/cache/apk/* \
    && if [ "${DEPLOYMODE}" = "dev" ]; then \
           echo '{"appId":"owl-engine","cluster":"default","namespaceName":"application","ip":"http://xxxx:80"}' | jq . > ${AGOLLO_CONF} ; \
       elif [ "${DEPLOYMODE}" = "test" ]; then \
           echo '{"appId":"owl-engine","cluster":"default","namespaceName":"application","ip":"http://xxxx:80"}' | jq . > ${AGOLLO_CONF} ; \
       elif [ "${DEPLOYMODE}" = "prod" ]; then \
           echo '{"appId":"owl-engine","cluster":"default","namespaceName":"application","ip":"http://xxxx:80"}' | jq . > ${AGOLLO_CONF} ; \
       fi \
    && mkdir -p ./bin \
    && mkdir -p ./data \
    && mkdir -p ./conf \
    && mkdir -p ./logs \
    && adduser -s /bin/bash devuser -D -H \
    && chown -R devuser:devuser ${PROJECTDIR}

COPY --from=builder /usr/share/zoneinfo/Asia/Shanghai /etc/localtime
COPY --from=builder ${GOPATH}/bin/owl-engine ${PROJECTDIR}/bin

# 设置普通用户操作
USER devuser

# 启动程序
ENTRYPOINT ["./bin/owl-engine", "server", "-t", "apollo"]