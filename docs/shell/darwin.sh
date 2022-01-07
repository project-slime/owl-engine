#!/bin/bash -
#####################################################################
#            FILE: darwin.sh
#           USAGE: bash darwin.sh [artifact name] [project root path] [store artifact dir] [store generate scripts dir]
#
#     DESCRIPTION:
#               build darwin binary artifact for x64 platform
#         OPTIONS:
#
#    REQUIREMENTS:
#
#            BUGS:
#          AUTHOR: Stalker-lee  1318895540@qq.com
#    ORGANIZATION:
#         CREATED: 2022-01-06 18:00:10
#       REVERSION: 1.0.0
#####################################################################

set -o nounset             # Treat unset variables as an error
set -o errexit
set -o pipefail

# 终端文本高亮显示
function show_color_text_by_echo()
{
    local content="$1"
    local choose="red"

    if [[ $# -eq 2 ]]; then
        local choose="$2"
    fi

    case "${choose}" in
        "black")
            color="30m"
            ;;
        "red")
            color="31m"
            ;;
        "green")
            color="32m"
            ;;
        "yellow")
            color="33m"
            ;;
        "blue")
            color="34m"
            ;;
        "purple")
            color="35m"
            ;;
        "white")
            color="37m"
            ;;
        *)
            color="31m"
            ;;
    esac

    echo -e "\033[${color}${content}\033[0m"
}

# 检查命令是否存在
function command_exist()
{
    command -v "$@" >/dev/null 2>&1
}

if [[ $# -ne 4 ]]; then
    show_color_text_by_echo "Usage: bash darwin.sh [artifact name] [project root path] [store artifact dir] [store generate scripts dir]"
    exit 1
fi

readonly ARTIFACT_NAME="$1"
readonly PROJECT_ROOT_PATH="$2"
readonly STORE_ARTIFACT_DIR="$3"
readonly STORE_SCRIPTS_DIR="$4"

# 判断 go 编译器是否已安装
if ! command_exist go
then
    show_color_text_by_echo "[*] go interpreter does not installed" "red"
    exit 1
fi

# override build target
readonly ARTIFACT_TARGET="${ARTIFACT_NAME}"

# create build dir
[[ -d ${STORE_ARTIFACT_DIR} ]] && rm -rf ${STORE_ARTIFACT_DIR}
mkdir -p ${STORE_ARTIFACT_DIR}/{bin,logs,conf,data}

show_color_text_by_echo "[*] create store artifact dir ${STORE_ARTIFACT_DIR}/{bin,logs,conf,data} success" "green"

cd "${PROJECT_ROOT_PATH}"

# 添加快点发布平台管理脚本
if [[ -f ${STORE_SCRIPTS_DIR}/control.sh ]]; then
    cp -a ${STORE_SCRIPTS_DIR}/control.sh ${STORE_ARTIFACT_DIR}
    show_color_text_by_echo "[*] copy kuaidian manager script to ${STORE_ARTIFACT_DIR} success" "green"
else
    show_color_text_by_echo "[*] copy kuaidian manager script to ${STORE_ARTIFACT_DIR} failed" "red"
    exit 1
fi

# copy apollo config and self config file
if [[ -f app.properties ]]; then
    cp -a app.properties ${STORE_ARTIFACT_DIR}/bin

    # 如果存在配置文件, 则复制该配置文件到 conf 目录下
    if [[ -f conf/config.yaml ]]; then
        cp -a conf/config.yaml ${STORE_ARTIFACT_DIR}/conf
    fi
    if [[ -f conf/config-prod.yaml ]]; then
        cp -a conf/config-prod.yaml ${STORE_ARTIFACT_DIR}/conf
    fi

    show_color_text_by_echo "[*] copy apollo config file app.properties and config.yaml to ${STORE_ARTIFACT_DIR}/conf success" "green"
else
    show_color_text_by_echo "[*] apollo config file app.properties not found" "red"
    exit 1
fi

# copy manager script
if [[ -f ${STORE_SCRIPTS_DIR}/manager_owl-engine.sh ]]; then
    cp -a ${STORE_SCRIPTS_DIR}/manager_owl-engine.sh ${STORE_ARTIFACT_DIR}/bin/
    show_color_text_by_echo "[*] copy ${STORE_SCRIPTS_DIR}/manager_owl-engine.sh to ${STORE_ARTIFACT_DIR}/bin success" "green"
else
    show_color_text_by_echo "[*] copy ${STORE_SCRIPTS_DIR}/manager_owl-engine.sh to ${STORE_ARTIFACT_DIR}/bin failed" "red"
    exit 1
fi

show_color_text_by_echo "[*] start building ${ARTIFACT_TARGET} ..." "green"

CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -a -installsuffix cgo -o "${STORE_ARTIFACT_DIR}/bin/${ARTIFACT_TARGET}"

if [[ $? -eq 0 ]] && \
    [[ -f ${STORE_ARTIFACT_DIR}/bin/${ARTIFACT_TARGET} ]]; then
    chmod +x ${STORE_ARTIFACT_DIR}/bin/${ARTIFACT_TARGET}

    show_color_text_by_echo "[*] build ${ARTIFACT_TARGET} on darwin for x86 platform success" "green"

    cd "${PROJECT_ROOT_PATH}"
    tar -zcf $(basename ${STORE_ARTIFACT_DIR}).tar.gz $(basename ${STORE_ARTIFACT_DIR})
    rm -rf $(basename ${STORE_ARTIFACT_DIR})

    show_color_text_by_echo "[*] archive ${ARTIFACT_TARGET} to ${STORE_ARTIFACT_DIR}.tar.gz success" "green"
else
     show_color_text_by_echo "[*] build ${ARTIFACT_TARGET} on darwin for x86 platform fail" "red"
     exit 1
fi

exit 0
