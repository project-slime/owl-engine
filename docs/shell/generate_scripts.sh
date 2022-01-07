#!/bin/bash -
#####################################################################
#            FILE: generate_scripts.sh
#           USAGE:
#               bash generate_scripts.sh [options]
#           options:
#               -p | --project-name  Specify the project name
#           example:
#               bash generate_scripts.sh --project-name=owl-engine
#
#     DESCRIPTION:
#           (1) 模版生成 Makefile 标准编译文件
#           (2) 模版生成 build-darwin MacOS x86平台下的构建脚本
#           (3) 模版生成 build-linux Linux x86平台下的构建脚本
#           (4) 模版生成 build-windows windows x86平台下的构建脚本
#			(5) 注意: 该脚本只针对于原生编译构建成可执行二进制文件，docker镜像另外提供
#
#         OPTIONS:
#
#    REQUIREMENTS:
#           (1) 该脚本必须存放在项目根目录下的 ./docs/shell 目录下
#               运行该脚本产生的编译脚本会存放在 ./docs/shell 目录下
#               产生的 Makefile 存放在项目的根目录下
#
#           (2) 其系统必须存在 getopt 命令
#               MacOS:
#                   brew 的安装: /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
#                   brew install gnu-getopt
#				MacOS <= 10.14
#               	说明: ls -aln /usr/local/opt/gnu-getopt/bin/getopt 就会存在 getopt 命令
#				MacOS >= 12.0
#               	说明: ls -aln /opt/homebrew/opt/gnu-getopt/bin/getopt 就会存在 getopt 命令
#
#               Linux: getopt 是系统自带命令
#
#               Window10:
#                   需要安装 git-bash for windows
#                   下载地址: https://github.com/git-for-windows/git/releases
#                   选择适合你环境下的二进制可执行文件进行安装
#
#                   make for x86 windows
#                   下载地址: http://www.equation.com/servlet/equation.cmd?fa=make
#                   将下载的 make.exe 放置到 C:\Program Files\Git\usr\bin 目录下
#
#            (3) go version 1.14.+
#                   下载地址: https://golang.google.cn/dl/
#
#            BUGS:
#          AUTHOR: Stalker-lee  1318895540@qq.com
#    ORGANIZATION:
#         CREATED: 2021-06-25 22:52:14
#       REVERSION: 1.5.0
#####################################################################

set -o nounset # Treat unset variables as an error
set -o errexit
set -o pipefail

# 指定环境变量
readonly PATH=/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/root/bin:/usr/local/go/bin

# 指定当前项目的家目录
# 注意: 当前该脚本必须存放在项目统一约定好的 ./docs/shell 目录下
readonly PROJECT_ROOT_PATH="$(
	cd $(dirname ${0})/../../
	pwd
)"

# 编译平台
readonly CROSS_PLATFORMS=(darwin linux windows)

# 目标主机部署目录
readonly REMOTE_DEPLOY_DIR="/data/"$(basename ${PROJECT_ROOT_PATH})

# 终端文本高亮显示
function show_color_text_by_echo() {
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

# 判断命令是否存在于系统
function command_exist() {
	command -v "$@" >/dev/null 2>&1
}

# 帮助文档
function help_doc() {
	local doc=$(
		cat <<-EOF
			Usage: bash generate_scripts [options]

			options:
			    -p | --project-name  # specify the project name

			examples:
			    bash generate_scripts.sh --project-name=owl-engine
		EOF
	)

	show_color_text_by_echo "${doc}" "red"
	return 0
}

# 不同平台下的编译脚本
function generate_cross_compile_script() {
	local project_name="$1"
	local store_dir="$2"

	# 为兼容其路径中存在空格等特殊字符, 利用 双引号 进行转义
	cd "${store_dir}"

	for os in ${CROSS_PLATFORMS[*]}; do
		if [[ -f ${os}.sh ]]; then
			read -p "$(echo -e "\033[42;31m[*] ${os}.sh already exists in current path, do you want to override it?[y/Y/n/N] \033[0m")" choose
			if [[ "${choose}" =~ ^(y|Y)$ ]]; then
				rm -rf ${os}.sh
			else
				show_color_text_by_echo "[*] Exit!" "red"
				exit 0
			fi
		fi

		cat >${os}.sh <<EOF
#!/bin/bash -
#####################################################################
#            FILE: ${os}.sh
#           USAGE: bash ${os}.sh [artifact name] [project root path] [store artifact dir] [store generate scripts dir]
#
#     DESCRIPTION:
#               build ${os} binary artifact for x64 platform
#         OPTIONS:
#
#    REQUIREMENTS:
#
#            BUGS:
#          AUTHOR: Stalker-lee  1318895540@qq.com
#    ORGANIZATION:
#         CREATED: $(date +"%Y-%m-%d %H:%M:%S")
#       REVERSION: 1.0.0
#####################################################################

set -o nounset             # Treat unset variables as an error
set -o errexit
set -o pipefail

# 终端文本高亮显示
function show_color_text_by_echo()
{
    local content="\$1"
    local choose="red"

    if [[ \$# -eq 2 ]]; then
        local choose="\$2"
    fi

    case "\${choose}" in
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

    echo -e "\033[\${color}\${content}\033[0m"
}

# 检查命令是否存在
function command_exist()
{
    command -v "\$@" >/dev/null 2>&1
}

if [[ \$# -ne 4 ]]; then
    show_color_text_by_echo "Usage: bash ${os}.sh [artifact name] [project root path] [store artifact dir] [store generate scripts dir]"
    exit 1
fi

readonly ARTIFACT_NAME="\$1"
readonly PROJECT_ROOT_PATH="\$2"
readonly STORE_ARTIFACT_DIR="\$3"
readonly STORE_SCRIPTS_DIR="\$4"

# 判断 go 编译器是否已安装
if ! command_exist go
then
    show_color_text_by_echo "[*] go interpreter does not installed" "red"
    exit 1
fi

# override build target
readonly ARTIFACT_TARGET="\${ARTIFACT_NAME}"

# create build dir
[[ -d \${STORE_ARTIFACT_DIR} ]] && rm -rf \${STORE_ARTIFACT_DIR}
mkdir -p \${STORE_ARTIFACT_DIR}/{bin,logs,conf,data}

show_color_text_by_echo "[*] create store artifact dir \${STORE_ARTIFACT_DIR}/{bin,logs,conf,data} success" "green"

cd "\${PROJECT_ROOT_PATH}"

# 添加快点发布平台管理脚本
if [[ -f \${STORE_SCRIPTS_DIR}/control.sh ]]; then
    cp -a \${STORE_SCRIPTS_DIR}/control.sh \${STORE_ARTIFACT_DIR}
    show_color_text_by_echo "[*] copy kuaidian manager script to \${STORE_ARTIFACT_DIR} success" "green"
else
    show_color_text_by_echo "[*] copy kuaidian manager script to \${STORE_ARTIFACT_DIR} failed" "red"
    exit 1
fi

# copy apollo config and self config file
if [[ -f app.properties ]]; then
    cp -a app.properties \${STORE_ARTIFACT_DIR}/bin

    # 如果存在配置文件, 则复制该配置文件到 conf 目录下
    if [[ -f conf/config.yaml ]]; then
        cp -a conf/config.yaml \${STORE_ARTIFACT_DIR}/conf
    fi
    if [[ -f conf/config-prod.yaml ]]; then
        cp -a conf/config-prod.yaml \${STORE_ARTIFACT_DIR}/conf
    fi

    show_color_text_by_echo "[*] copy apollo config file app.properties and config.yaml to \${STORE_ARTIFACT_DIR}/conf success" "green"
else
    show_color_text_by_echo "[*] apollo config file app.properties not found" "red"
    exit 1
fi

# copy manager script
if [[ -f \${STORE_SCRIPTS_DIR}/manager_${project_name}.sh ]]; then
    cp -a \${STORE_SCRIPTS_DIR}/manager_${project_name}.sh \${STORE_ARTIFACT_DIR}/bin/
    show_color_text_by_echo "[*] copy \${STORE_SCRIPTS_DIR}/manager_${project_name}.sh to \${STORE_ARTIFACT_DIR}/bin success" "green"
else
    show_color_text_by_echo "[*] copy \${STORE_SCRIPTS_DIR}/manager_${project_name}.sh to \${STORE_ARTIFACT_DIR}/bin failed" "red"
    exit 1
fi

show_color_text_by_echo "[*] start building \${ARTIFACT_TARGET} ..." "green"

CGO_ENABLED=0 GOOS=${os} GOARCH=amd64 go build -a -installsuffix cgo -o "\${STORE_ARTIFACT_DIR}/bin/\${ARTIFACT_TARGET}"

if [[ \$? -eq 0 ]] && \\
    [[ -f \${STORE_ARTIFACT_DIR}/bin/\${ARTIFACT_TARGET} ]]; then
    chmod +x \${STORE_ARTIFACT_DIR}/bin/\${ARTIFACT_TARGET}

    show_color_text_by_echo "[*] build \${ARTIFACT_TARGET} on ${os} for x86 platform success" "green"

    cd "\${PROJECT_ROOT_PATH}"
    tar -zcf \$(basename \${STORE_ARTIFACT_DIR}).tar.gz \$(basename \${STORE_ARTIFACT_DIR})
    rm -rf \$(basename \${STORE_ARTIFACT_DIR})

    show_color_text_by_echo "[*] archive \${ARTIFACT_TARGET} to \${STORE_ARTIFACT_DIR}.tar.gz success" "green"
else
     show_color_text_by_echo "[*] build \${ARTIFACT_TARGET} on ${os} for x86 platform fail" "red"
     exit 1
fi

exit 0
EOF
	done

	show_color_text_by_echo "[*] successfully generated compilation scripts for darwin/linux/windows x64 platform" "green"
	return 0
}

# 生成管理脚本
function generate_manager() {
	local project_name="$1"
	# 脚本生成路径
	local store_dir="$2"
	# 可执行程序的部署路径
	local deploy_dir="$3"

	# 为兼容其路径中存在空格等特殊字符, 利用 双引号 进行转义
	cd "${store_dir}"

	if [[ -f manager_${project_name}.sh ]]; then
		read -p "$(echo -e "\033[42;31m[*] manager_${project_name}.sh already exists in current path, do you want to override it?[y/Y/n/N] \033[0m")" choose
		if [[ "${choose}" =~ ^(y|Y)$ ]]; then
			rm -rf manager_${project_name}.sh
		else
			show_color_text_by_echo "[*] Exit!" "red"
			exit 0
		fi
	fi

	cat >manager_${project_name}.sh <<EOF
#!/bin/bash -
#####################################################################
#            FILE: manager_${project_name}.sh
#           USAGE: export $(echo ${project_name} | tr '[a-z]' '[A-Z]' | tr '-' '_')=dev|test|prod ; bash manager_${project_name}.sh ${project_name} status|stop|start
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
#         CREATED: $(date +"%Y-%m-%d %H:%M:%S")
#       REVERSION: 1.0.0
#####################################################################

set -o nounset             # Treat unset variables as an error
set -o errexit
set -o pipefail

# 进程管理用户和用户组
readonly USERNAME="baseuser"
readonly GROUPNAME="baseuser"

# 环境变量
MODE=\${$(echo ${project_name} | tr '[a-z]' '[A-Z]' | tr '-' '_')_MODE:-dev}

# 项目部署家目录
readonly PROJECT_PATH=${deploy_dir}

# 终端文本高亮显示
function show_color_text_by_echo()
{
    local content="\$1"
    local choose="red"

    if [[ \$# -eq 2 ]]; then
        local choose="\$2"
    fi

    case "\${choose}" in
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

    echo -e "\033[\${color}\${content}\033[0m"
}

function command_exist()
{
    command -v "\$@" > /dev/null 2>&1
}

# 验证是否以超级用户进行管理
if [[ \${UID} -ne 0 ]]
then
    show_color_text_by_echo "You should switch to root user" "red"
    exit 1
fi

# 创建用户
function create_username_and_groupname()
{
    if [[ -f /etc/group ]]; then
        local is_exist=\$(cat /etc/group | sed -n "/\${GROUPNAME}/p")
        if [[ -z "\${is_exist}" ]]; then
            groupadd \${GROUPNAME}

            show_color_text_by_echo "[*] create system groupname \${GROUPNAME} success" "green"
        fi
    fi

    if [[ -f /etc/passwd ]]; then
        local is_exist=\$(cat /etc/passwd | sed -n "/\${USERNAME}/p")
        if [[ -z "\${is_exist}" ]]; then
            useradd -M -s /sbin/nologin \${USERNAME} -g \${GROUPNAME}

            show_color_text_by_echo "[*] create system username \${USERNAME} success" "green"
        fi
    fi

    chown -R \${USERNAME}:\${GROUPNAME} \${PROJECT_PATH}
    return 0
}

# 帮助文档
function help_doc() {
    local doc=\$(cat << AOF
Usage: export $(echo ${project_name} | tr '[a-z]' '[A-Z]' | tr '-' '_')=dev|test|prod ; bash manager_${project_name}.sh ${project_name} status|stop|start
AOF
)
    show_color_text_by_echo "\${doc}" "red"
    return 0
}

function get_pid()
{
    # 应用名称
    local app_name="\$1"

    if command_exist ps; then
        local pid=\$(ps aux | grep -Ev "grep|bash|sh" | grep -E "\${app_name}" | awk '{print \$2}')
        echo \${pid}
    else
        show_color_text_by_echo "ps command does not found" "red"
        return 1
    fi

    return 0
}

# 启动
function start()
{
    local app_name="\$1"
    local pid=\$(get_pid \${app_name})
    if [[ -n \${pid} ]]; then
        show_color_text_by_echo "\${app_name} is running [\${pid}]" "green"
    else
        # 进入到项目的根目录
        cd "\${PROJECT_PATH}"

        if [[ "\${MODE}" == "dev" ]]; then
            # 测试环境
            cat > ./bin/app.properties << AOF
{
    "appId": "${project_name}",
    "cluster":"default",
    "namespaceName": "application",
    "ip": "http://xxxx:80"
}
AOF
            su \${USERNAME} -c "umask=022 ; cd ./bin && nohup ./\${app_name} -t apollo &> /dev/null &"
        elif [[ "\${MODE}" == "test" ]]; then
            # 测试环境
            cat > ./bin/app.properties << AOF
{
    "appId": "${project_name}",
    "cluster":"default",
    "namespaceName": "application",
    "ip": "http://xxxx:80"
}
AOF
            su \${USERNAME} -c "umask=022 ; cd ./bin && nohup ./\${app_name} -t apollo &> /dev/null &"
        elif [[ "\${MODE}" == "prod" ]]; then
            # 生产环境
            cat > ./bin/app.properties << AOF
{
    "appId": "${project_name}",
    "cluster":"default",
    "namespaceName": "application",
    "ip": "http://xxxx:80"
}
AOF
            su \${USERNAME} -c "umask=022 ; cd ./bin && nohup ./\${app_name} -t apollo &> /dev/null &"
        else
            show_color_text_by_echo "can not specify mode" "red"
            exit 1
        fi

        sleep 1

        pid=\$(get_pid \${app_name})
        if [[ -n \${pid} ]]; then
            show_color_text_by_echo "\${app_name} start success [\${pid}]" "green"
        else
            show_color_text_by_echo "\${app_name} start failed" "red"
        fi
    fi

    return 0
}

function stop()
{
    local app_name="\$1"
    local current_datetime=\$(date +%s)

    while true
    do
        pid=\$(get_pid \${app_name})
        if [[ -n \${pid} ]]; then
            kill \${pid}
        else
            show_color_text_by_echo "\${app_name} has been stopped" "green"
            break
        fi

        sleep 1
        if [[ \$(expr \$(date +%s) - \${current_datetime}) -gt 10 ]]; then
            show_color_text_by_echo "stop \${app_name} operation failed" "red"
            break
        fi
    done

    return 0
}

function status()
{
    local app_name="\$1"
    local pid=\$(get_pid \${app_name})

    if [[ -n \${pid} ]]; then
        show_color_text_by_echo "\${app_name} is running [\${pid}]" "green"
    else
        show_color_text_by_echo "\${app_name} is not running" "green"
    fi

    return 0
}

function  main()
{
    if [[ \$# -ne 2 ]]; then
        help_doc
        return 1
    fi

    local app_name="\$1"
    local action="\$2"

    if [[ ! -d \${PROJECT_PATH} ]]; then
        show_color_text_by_echo "\${PROJECT_PATH} does not exist" "red"
        return 1
    fi

    if [[ ! -f \${PROJECT_PATH}/bin/\${app_name} ]]; then
        show_color_text_by_echo "\${PROJECT_PATH}/bin/\${app_name} does not found" "red"
        return 1
    fi

    # 检测用户是否存在, 若不存在, 则创建
    create_username_and_groupname

    case \${action} in
        start)
            start \${app_name}
            ;;
        stop)
            stop \${app_name}
            ;;
        restart)
            restart \${app_name}
            ;;
        status)
            status \${app_name}
            ;;
        *)
           help_doc
           return 1
           ;;
    esac

    return 0
}

main "\$@"
exit 0
EOF

	show_color_text_by_echo "[*] Successfully generated management script" "green"
	return 0
}

# 生成 快点脚本兼容脚本
function generate_kuaidian_manager() {
	local project_name="$1"
	# 脚本生成路径
	local store_dir="$2"

	# 为兼容其路径中存在空格等特殊字符, 利用 双引号 进行转义
	cd "${store_dir}"

	if [[ -f control.sh ]]; then
		read -p "$(echo -e "\033[42;31m[*] control.sh already exists in current path, do you want to override it?[y/Y/n/N] \033[0m")" choose
		if [[ "${choose}" =~ ^(y|Y)$ ]]; then
			rm -rf control.sh
		else
			show_color_text_by_echo "[*] Exit!" "red"
			exit 0
		fi
	fi

	cat >control.sh <<EOF
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
#         CREATED: $(date +"%Y-%m-%d %H:%M:%S")
#       REVERSION: 1.0.0
#####################################################################

set -o nounset             # Treat unset variables as an error
set -o errexit
set -o pipefail

readonly DEPLOY_DIR="\$(pwd)"

# 判断监管脚本是否存在
if [[ ! -f \${DEPLOY_DIR}/bin/manager_${project_name}.sh ]]; then
    echo -e "\${DEPLOY_DIR}/bin/manager_${project_name}.sh does not exist"
    exit 1
fi

# 判断 Apollo 的配置文件是否存在
if [[ ! -f \${DEPLOY_DIR}/bin/app.properties ]]; then
    echo -e "\${DEPLOY_DIR}/bin/app.properties does not exist"
    exit 1
fi

# 获取到部署的环境
MODE="test"
if [[ -n "\$(sed -n "/xxxx/p" \${DEPLOY_DIR}/bin/app.properties)" ]]; then
    MODE="test"
elif [[ -n "\$(sed -n "/xxxx/p" \${DEPLOY_DIR}/bin/app.properties)" ]]; then
    MODE="prod"
fi

if [[ \$# -eq 1 ]]; then
    case \$1 in
    start)
        export $(echo ${project_name} | tr '[a-z]' '[A-Z]' | tr '-' '_')_MODE=\${MODE} && bash \${DEPLOY_DIR}/bin/manager_${project_name}.sh ${project_name} start > /dev/null 2>&1 &
        bash \${DEPLOY_DIR}/bin/manager_${project_name}.sh ${project_name} status
        ;;
    stop)
        bash \${DEPLOY_DIR}/bin/manager_${project_name}.sh ${project_name} stop
        ;;
    restart)
        bash \${DEPLOY_DIR}/bin/manager_${project_name}.sh ${project_name} stop
        export $(echo ${project_name} | tr '[a-z]' '[A-Z]' | tr '-' '_')_MODE=\${MODE} && bash \${DEPLOY_DIR}/bin/manager_${project_name}.sh ${project_name} start > /dev/null 2>&1 &
        bash \${DEPLOY_DIR}/bin/manager_${project_name}.sh ${project_name} status
        ;;
    status)
        bash \${DEPLOY_DIR}/bin/manager_${project_name}.sh ${project_name} status
        ;;
    esac
else
    echo -e "bash control.sh start|restart|stop"
    exit 1
fi

exit 0
EOF

	show_color_text_by_echo "[*] Successfully generated kuaidian management script" "green"
	return 0
}

# 生成 Makefile 文件
function generate_makefile() {
	local project_name="$1"
	local artifact_dir="$2"

	# 进入到项目根目录
	cd "${PROJECT_ROOT_PATH}"

	# 判断当前其 Makefile 文件是否已经存在,若存在, 则交由用户判断是否删除
	if [[ -f Makefile ]]; then
		read -p "$(echo -e "\033[42;31m[*] Makefile already exists in current path, do you want to override it?[y/Y/n/N] \033[0m")" choose
		if [[ "${choose}" =~ ^(y|Y)$ ]]; then
			rm -rf Makefile
		else
			show_color_text_by_echo "[*] Exit!" "red"
			exit 0
		fi
	fi

	local phony=""
	for os in ${CROSS_PLATFORMS[*]}; do
		temp_phony=$(
			cat <<EOF

.PHONY: fmt test build-${os}
build-${os}: ## build ${os} artifact on x86 platform
	@bash \$(STORE_GENERATE_SCRIPTS_DIR)/${os}.sh \$(NAME_OF_ARTIFACT) \$(CURDIR) \$(STORE_ARTIFACT_DIR) \$(STORE_GENERATE_SCRIPTS_DIR)
EOF
		)
		phony=$(echo -e "${phony}\n${temp_phony}")
	done

	cat >>Makefile <<EOF
# File: Makefile
# Author: Stalker-lee
# DateTime: $(date +"%Y-%m-%d %H:%M:%S")
#
# Description:  Makefile artifact from source code
# Usage:
#       make help|build-linux|build-darwin|build-windows|build-docker

# It's necessary to set this because some environments don't link sh -> bash.
SHELL := /usr/bin/env bash

# specify go env parameters
GO           ?= go
GOFMT        ?= \$(GO)fmt
FIRST_GOPATH := \$(firstword \$(subst :, ,\$(shell \$(GO) env GOPATH)))
GOOPTS       ?=
GOHOSTOS     ?= \$(shell \$(GO) env GOHOSTOS)
GOHOSTARCH   ?= \$(shell \$(GO) env GOHOSTARCH)

GO_VERSION        ?= \$(shell \$(GO) version)
GO_VERSION_NUMBER ?= \$(word 3, \$(GO_VERSION))
PRE_GO_111        ?= \$(shell echo \$(GO_VERSION_NUMBER) | grep -E 'go1\.(10|[0-9])\.')

GOVENDOR :=
GO111MODULE :=

ifeq (, \$(PRE_GO_111))
	ifneq (,\$(wildcard go.mod))
		# Enforce Go modules support just in case the directory is inside GOPATH (and for Travis CI).
		GO111MODULE := on

		ifneq (,\$(wildcard vendor))
			# Always use the local vendor/ directory to satisfy the dependencies.
			GOOPTS := \$(GOOPTS) -mod=vendor
		endif
	endif
else
	ifneq (,\$(wildcard go.mod))
		ifneq (,\$(wildcard vendor))
\$(warning This repository requires Go >= 1.11 because of Go modules)
\$(warning Some recipes may not work as expected as the current Go runtime is '\$(GO_VERSION_NUMBER)')
		endif
	else
		# This repository isn't using Go modules (yet).
		GOVENDOR := \$(FIRST_GOPATH)/bin/govendor
	endif
endif

ifeq (arm, \$(GOHOSTARCH))
	GOHOSTARM ?= \$(shell GOARM= \$(GO) env GOARM)
	GO_BUILD_PLATFORM ?= \$(GOHOSTOS)-\$(GOHOSTARCH)v\$(GOHOSTARM)
else
	GO_BUILD_PLATFORM ?= \$(GOHOSTOS)-\$(GOHOSTARCH)
endif

# 关于 docker 设置
DOCKER ?= \$(shell command -v "docker")

# The name of the binary component
NAME_OF_ARTIFACT := ${project_name}

# The directory where the artifact is stored
STORE_ARTIFACT_DIR := \$(CURDIR)/${artifact_dir}

# The storage directory of the generated script
STORE_GENERATE_SCRIPTS_DIR := \$(CURDIR)/docs/shell

.PHONY: fmt
fmt:
	@go list -f {{.Dir}} ./... | xargs gofmt -w -s -d

.PHONY: test
test: ## execute unit test for go
	@go test

${phony}

.PHONY: fmt test build-docker
build-docker: ## build docker images
	# docker 是否安装
	if [[ -n \`command -v docker\` ]]; then                 					    										\\
		is_running=\$\$(docker info | grep -E "Is the docker daemon running?") ;  											\\
		if [[ -z \$\$is_running ]]; then 									   												\\
			containerID=\$\$(docker images -q --filter reference=\$(NAME_OF_ARTIFACT) | sort | uniq) ; 						\\
			if [[ -n \$\$is_exist ]]; then																					\\
				docker rmi -f \$\$containerID ;																				\\
				echo "owl-engine image already exists. delete it now" ;														\\
			fi ;																											\\
																															\\
			docker build --build-arg MODE=\$(MODE) -t \$(NAME_OF_ARTIFACT):v1.0.0 \$(CURDIR) ;								\\
			if [[ \$\$? -eq 0 ]]; then																						\\
				cd \$(CURDIR) ;																								\\
				docker save -o \$(NAME_OF_ARTIFACT)_v1.0.0.tar \$(NAME_OF_ARTIFACT):v1.0.0 ; 								\\
				echo "\$(CURDIR)/\$(NAME_OF_ARTIFACT)_v1.0.0.tar image was successfully built, now you can distribute it!" ; \\
			fi ; 																											\\
		else																												\\
			echo "Is the docker daemon running?" ;													        				\\
		fi ;																    											\\
	fi

.PHONY: help
help: ## show make targets
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {sub("\\\\n",sprintf("\n%22c"," "), \$\$2);printf " \033[36m%-20s\033[0m  %s\n", \$\$1, \$\$2}' \$(MAKEFILE_LIST)
EOF

	show_color_text_by_echo "[*] generate Makefile success for ${project_name}" "green"
	return 0
}

# 程序主入口
function main() {
	# MacOS 系统
	if [[ "$(uname)" == "Darwin" ]]; then
		if [[ -f /usr/local/opt/gnu-getopt/bin/getopt ]]; then
			local getopts='/usr/local/opt/gnu-getopt/bin/getopt' # darwin-10.14 以下
		elif [[ -f /opt/homebrew/opt/gnu-getopt/bin/getopt ]]; then
			local getopts="/opt/homebrew/opt/gnu-getopt/bin/getopt" # darwin-10.12 以上
		else
			show_color_text_by_echo "command getopt not found, you can run [brew install gnu-getopt] to install it!" "red"
			return 1
		fi
	elif [[ "$(uname)" == "Linux" ]] ||
		[[ "$(uname)" =~ ^(MINGW) ]]; then
		local getopts='/usr/bin/getopt'
	else
		show_color_text_by_echo "can not recognize os and getopt does not install" "red"
		exit 1
	fi

	local project_name=""

	local args=$(${getopts} -o p:h -al project-name:,help -n "$0" -- "$@")
	eval set -- "${args}"

	while [[ -n "${1}" ]]; do
		case "${1}" in
		-h | --help)
			help_doc
			shift 2
			exit 0
			;;
		-p | --project-name)
			project_name=$2
			shift 2
			;;
		--)
			break
			;;
		*)
			help_doc
			exit 1
			;;
		esac
	done

	if [[ -z "${project_name}" ]]; then
		help_doc
		return 1
	fi

	# 构建结构存储目录
	local artifact_dir="${project_name}"
	# 生成的脚本的存放目录, 与 本脚本所在的目录保持一致
	local scripts_store_dir="$(cd $(dirname $0) && pwd)"

	# 生成管理脚本
	generate_manager "${project_name}" "${scripts_store_dir}" "${REMOTE_DEPLOY_DIR}"

	# 为兼容 快点发布平台, 生成 control.sh 脚本
	generate_kuaidian_manager "${project_name}" "${scripts_store_dir}"

	# 生成不同平台下的编译脚本
	generate_cross_compile_script "${project_name}" "${scripts_store_dir}"

	# 生成 Makefile
	generate_makefile "${project_name}" "${artifact_dir}"
}

main "$@"
exit 0
