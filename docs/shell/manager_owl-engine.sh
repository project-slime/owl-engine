#!/bin/bash -
#####################################################################
#            FILE: manager_owl-engine.sh
#           USAGE: export OWL_ENGINE=dev|test|prod ; bash manager_owl-engine.sh owl-engine status|stop|start
#
#     DESCRIPTION:
#           (1) 脚本自动化管理进程
#           (2) dev 环境下, 其启动是使用本地的配置文件 config.yaml 进行项目配置
#               test 环境下, 其启动是使用 Apollo 的 DEV 进行项目配置
#               prod 环境下, 其启动是使用 Apollo 的 PRO 进行项目配置
#
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

# 进程管理用户和用户组
readonly USERNAME="baseuser"
readonly GROUPNAME="baseuser"

# 环境变量
MODE=${OWL_ENGINE_MODE:-dev}

# 项目部署家目录
readonly PROJECT_PATH=/data/owl-engine-github

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

function command_exist()
{
    command -v "$@" > /dev/null 2>&1
}

# 验证是否以超级用户进行管理
if [[ ${UID} -ne 0 ]]
then
    show_color_text_by_echo "You should switch to root user" "red"
    exit 1
fi

# 创建用户
function create_username_and_groupname()
{
    if [[ -f /etc/group ]]; then
        local is_exist=$(cat /etc/group | sed -n "/${GROUPNAME}/p")
        if [[ -z "${is_exist}" ]]; then
            groupadd ${GROUPNAME}

            show_color_text_by_echo "[*] create system groupname ${GROUPNAME} success" "green"
        fi
    fi

    if [[ -f /etc/passwd ]]; then
        local is_exist=$(cat /etc/passwd | sed -n "/${USERNAME}/p")
        if [[ -z "${is_exist}" ]]; then
            useradd -M -s /sbin/nologin ${USERNAME} -g ${GROUPNAME}

            show_color_text_by_echo "[*] create system username ${USERNAME} success" "green"
        fi
    fi

    chown -R ${USERNAME}:${GROUPNAME} ${PROJECT_PATH}
    return 0
}

# 帮助文档
function help_doc() {
    local doc=$(cat << AOF
Usage: export OWL_ENGINE=dev|test|prod ; bash manager_owl-engine.sh owl-engine status|stop|start
AOF
)
    show_color_text_by_echo "${doc}" "red"
    return 0
}

function get_pid()
{
    # 应用名称
    local app_name="$1"

    if command_exist ps; then
        local pid=$(ps aux | grep -Ev "grep|bash|sh" | grep -E "${app_name}" | awk '{print $2}')
        echo ${pid}
    else
        show_color_text_by_echo "ps command does not found" "red"
        return 1
    fi

    return 0
}

# 启动
function start()
{
    local app_name="$1"
    local pid=$(get_pid ${app_name})
    if [[ -n ${pid} ]]; then
        show_color_text_by_echo "${app_name} is running [${pid}]" "green"
    else
        # 进入到项目的根目录
        cd "${PROJECT_PATH}"

        if [[ "${MODE}" == "dev" ]]; then
            # 测试环境
            cat > ./bin/app.properties << AOF
{
    "appId": "owl-engine",
    "cluster":"default",
    "namespaceName": "application",
    "ip": "http://xxxx:80"
}
AOF
            su ${USERNAME} -c "umask=022 ; cd ./bin && nohup ./${app_name} -t apollo &> /dev/null &"
        elif [[ "${MODE}" == "test" ]]; then
            # 测试环境
            cat > ./bin/app.properties << AOF
{
    "appId": "owl-engine",
    "cluster":"default",
    "namespaceName": "application",
    "ip": "http://xxxx:80"
}
AOF
            su ${USERNAME} -c "umask=022 ; cd ./bin && nohup ./${app_name} -t apollo &> /dev/null &"
        elif [[ "${MODE}" == "prod" ]]; then
            # 生产环境
            cat > ./bin/app.properties << AOF
{
    "appId": "owl-engine",
    "cluster":"default",
    "namespaceName": "application",
    "ip": "http://xxxx:80"
}
AOF
            su ${USERNAME} -c "umask=022 ; cd ./bin && nohup ./${app_name} -t apollo &> /dev/null &"
        else
            show_color_text_by_echo "can not specify mode" "red"
            exit 1
        fi

        sleep 1

        pid=$(get_pid ${app_name})
        if [[ -n ${pid} ]]; then
            show_color_text_by_echo "${app_name} start success [${pid}]" "green"
        else
            show_color_text_by_echo "${app_name} start failed" "red"
        fi
    fi

    return 0
}

function stop()
{
    local app_name="$1"
    local current_datetime=$(date +%s)

    while true
    do
        pid=$(get_pid ${app_name})
        if [[ -n ${pid} ]]; then
            kill ${pid}
        else
            show_color_text_by_echo "${app_name} has been stopped" "green"
            break
        fi

        sleep 1
        if [[ $(expr $(date +%s) - ${current_datetime}) -gt 10 ]]; then
            show_color_text_by_echo "stop ${app_name} operation failed" "red"
            break
        fi
    done

    return 0
}

function status()
{
    local app_name="$1"
    local pid=$(get_pid ${app_name})

    if [[ -n ${pid} ]]; then
        show_color_text_by_echo "${app_name} is running [${pid}]" "green"
    else
        show_color_text_by_echo "${app_name} is not running" "green"
    fi

    return 0
}

function  main()
{
    if [[ $# -ne 2 ]]; then
        help_doc
        return 1
    fi

    local app_name="$1"
    local action="$2"

    if [[ ! -d ${PROJECT_PATH} ]]; then
        show_color_text_by_echo "${PROJECT_PATH} does not exist" "red"
        return 1
    fi

    if [[ ! -f ${PROJECT_PATH}/bin/${app_name} ]]; then
        show_color_text_by_echo "${PROJECT_PATH}/bin/${app_name} does not found" "red"
        return 1
    fi

    # 检测用户是否存在, 若不存在, 则创建
    create_username_and_groupname

    case ${action} in
        start)
            start ${app_name}
            ;;
        stop)
            stop ${app_name}
            ;;
        restart)
            restart ${app_name}
            ;;
        status)
            status ${app_name}
            ;;
        *)
           help_doc
           return 1
           ;;
    esac

    return 0
}

main "$@"
exit 0
