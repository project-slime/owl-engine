#!/bin/bash -
#####################################################################
#            FILE: control.sh
#           USAGE: bash control.sh start|stop|restart
#
#     DESCRIPTION:
#             该脚本的存在是为了兼容 快点发布平台的再封装层
#             依据快点发布平台的调用逻辑，是在项目发布成功后，出现的 "重启|启动|停止" 按钮，再次调用本脚本
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

readonly DEPLOY_DIR="$(pwd)"

# 判断监管脚本是否存在
if [[ ! -f ${DEPLOY_DIR}/bin/manager_owl-engine.sh ]]; then
    echo -e "${DEPLOY_DIR}/bin/manager_owl-engine.sh does not exist"
    exit 1
fi

# 判断 Apollo 的配置文件是否存在
if [[ ! -f ${DEPLOY_DIR}/bin/app.properties ]]; then
    echo -e "${DEPLOY_DIR}/bin/app.properties does not exist"
    exit 1
fi

# 获取到部署的环境
MODE="test"
if [[ -n "$(sed -n "/xxxx/p" ${DEPLOY_DIR}/bin/app.properties)" ]]; then
    MODE="test"
elif [[ -n "$(sed -n "/xxxx/p" ${DEPLOY_DIR}/bin/app.properties)" ]]; then
    MODE="prod"
fi

if [[ $# -eq 1 ]]; then
    case $1 in
    start)
        export OWL_ENGINE_MODE=${MODE} && bash ${DEPLOY_DIR}/bin/manager_owl-engine.sh owl-engine start > /dev/null 2>&1 &
        bash ${DEPLOY_DIR}/bin/manager_owl-engine.sh owl-engine status
        ;;
    stop)
        bash ${DEPLOY_DIR}/bin/manager_owl-engine.sh owl-engine stop
        ;;
    restart)
        bash ${DEPLOY_DIR}/bin/manager_owl-engine.sh owl-engine stop
        export OWL_ENGINE_MODE=${MODE} && bash ${DEPLOY_DIR}/bin/manager_owl-engine.sh owl-engine start > /dev/null 2>&1 &
        bash ${DEPLOY_DIR}/bin/manager_owl-engine.sh owl-engine status
        ;;
    status)
        bash ${DEPLOY_DIR}/bin/manager_owl-engine.sh owl-engine status
        ;;
    esac
else
    echo -e "bash control.sh start|restart|stop"
    exit 1
fi

exit 0
