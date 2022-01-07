USE
`owl`;

-- 创建 规则表
DROP TABLE IF EXISTS `engine_tbl_rules`;
CREATE TABLE `engine_tbl_rules`
(
    `id`                  bigint(11) NOT NULL AUTO_INCREMENT COMMENT '主键',
    `name`                varchar(255) DEFAULT NULL COMMENT '规则唯一名称',
    `calculate_type`      tinyint(1) NOT NULL COMMENT '计算类型:1-最大值; 2-最小值; 3-环比; 4-TopN; 5-BottomN',
    `express`             tinytext COMMENT '计算表达式',
    `metric_list`         tinytext     NOT NULL COMMENT '指标名集',
    `threshold`           float        DEFAULT '0' COMMENT '阈值, 可为零值',
    `unit`                varchar(16)  DEFAULT NULL COMMENT '单位',
    `time_window`         varchar(255) DEFAULT NULL COMMENT '时间窗口, 默认都以 分钟 作为单位',
    `duration`            int(11) DEFAULT NULL COMMENT '持续时长或次数; 如果为时长, 其单位为: 分钟',
    `origin`              varchar(64)  NOT NULL COMMENT '来源，前端-产品/业务-产品/应用-appid/组件-ip、集群名/基础-域名、ip',
    `business_type`       varchar(64)  NOT NULL COMMENT '产品名: 来源，前端-产品/业务-产品/应用-appid/组件-ip、集群名/基础-域名、ip',
    `category`            tinyint(1) DEFAULT NULL COMMENT '指标类型,1-前端监控,2-业务监控,3-应用监控,4-组件监控,5-基础监控',
    `extension_condition` varchar(255) DEFAULT NULL COMMENT '扩展条件',
    `level`               tinyint(1) NOT NULL DEFAULT '3' COMMENT '告警级别:0-Not classified; 1-Information; 2-Warning; 3-critical; 4-Disaster',
    `creator`             varchar(32)  DEFAULT NULL COMMENT '规则创建者,用户的钉钉userid',
    `updater`             varchar(32)  DEFAULT NULL COMMENT '规则更新人,用户的钉钉userid',
    `responsible_people`  varchar(255) NOT NULL COMMENT '告警事件处理人',
    `crontab`             varchar(32)  DEFAULT '* * * * *' COMMENT '每条规则的定时任务执行表达式',
    `switch`              tinyint(1) NOT NULL DEFAULT '1' COMMENT '是否启用, 1 --- on; 2 --- off',
    `inuse`               tinyint(1) NOT NULL DEFAULT '2' COMMENT '是否删除, 1 --- yes; 2 --- no',
    `group_ip`            varchar(255) NOT NULL COMMENT '告警时间接收者的组id, 多个值以 '','' 分隔',
    `web_hooks`           tinytext COMMENT '告警的 hook 地址,多个值以 '','' 分隔',
    `description`         tinytext COMMENT '规则描述',
    `created_at`          datetime(6) DEFAULT CURRENT_TIMESTAMP (6) COMMENT '记录创建时间',
    `updated_at`          datetime(6) DEFAULT NULL ON UPDATE CURRENT_TIMESTAMP (6) COMMENT '记录更新时间',
    `deleted_at`          datetime(6) DEFAULT NULL COMMENT '记录删除时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_name` (`name`) USING HASH
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='规则表'


-- 创建 告警事件表
DROP TABLE IF EXISTS `engine_tbl_alert`;
CREATE TABLE `engine_tbl_alert`
(
    `id`            bigint(20) NOT NULL AUTO_INCREMENT COMMENT '记录ID',
    `alert_id`      varchar(40)  NOT NULL COMMENT '告警事件id',
    `name`          varchar(128) NOT NULL COMMENT '告警名称, 对应规则的名称',
    `item`          varchar(128) NOT NULL COMMENT '告警项, 对应规则的表达式',
    `origin`        varchar(128) NOT NULL COMMENT '告警源，前端-产品/业务-产品/应用-appid/组件-ip、集群名/基础-域名、ip',
    `type`          varchar(128) NOT NULL COMMENT '告警子类型,前端-异常、crash/业务-业务域/应用-异常、服务、JVM/组件-db、mq、redis/基础-网络、k8s、物理机、虚拟机',
    `category`      tinyint(1) NOT NULL COMMENT '告警类型,1-前端监控,2-业务监控,3-应用监控,4-组件监控,5-基础监控',
    `value`         double                DEFAULT NULL COMMENT '告警值',
    `level`         tinyint(1) NOT NULL COMMENT '告警级别:0-Not classified; 1-Information; 2-Warning; 3-critical; 4-Disaster',
    `content`       tinytext COMMENT '告警内容',
    `rule_name`     varchar(255) NOT NULL COMMENT '规则名称',
    `group_id`      varchar(128) NOT NULL COMMENT '告警联系组id, 多个id 以 , 进行分割',
    `owner`         varchar(128)          DEFAULT NULL COMMENT '告警负责人',
    `status`        tinyint(1) DEFAULT NULL COMMENT '告警状态,1-告警中,2-恢复,3-忽略,4-静默',
    `platform`      tinyint(1) DEFAULT NULL COMMENT '告警平台,1-owl,2-zcat,3-prometheus,4-zms等',
    `platform_name` varchar(128)          DEFAULT NULL COMMENT '告警平台名称,zms/zdtp/es等',
    `aggregator_id` bigint(20) DEFAULT NULL COMMENT '告警聚合id',
    `alert_time`    timestamp    NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '告警时间',
    `creator`       varchar(128)          DEFAULT NULL COMMENT '创建人,engine/event',
    `updater`       varchar(128)          DEFAULT NULL COMMENT '更改人',
    `created_at`    timestamp    NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at`    timestamp NULL DEFAULT NULL COMMENT '更新时间',
    PRIMARY KEY (`id`),
    KEY             `idx_origin` (`origin`),
    KEY             `idx_alert_time` (`alert_time`),
    KEY             `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='告警时间记录表'

-- 创建 日志规则表
DROP TABLE IF EXISTS `engine_tbl_logger_rules`;
CREATE TABLE `engine_tbl_logger_rules`
(
    `id`            bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '自增长主键',
    `name`          varchar(255) NOT NULL COMMENT '规则名称',
    `source`        varchar(16)  NOT NULL DEFAULT 'es' COMMENT '日志数据源',
    `address`       varchar(255)          DEFAULT NULL COMMENT '对于es等数据源,会需要连接地址,多个地址以 , 分隔',
    `username`      varchar(32)           DEFAULT NULL COMMENT '对于esc等数据源，其认证的用户名',
    `password`      varchar(32)           DEFAULT NULL COMMENT '对于es等数据源, 其需要认证的密码',
    `index`         varchar(128)          DEFAULT NULL COMMENT '对于 es 等数据源的索引, 支持模糊匹配',
    `message_field` varchar(32)  NOT NULL COMMENT '告警消息的具体内容',
    `sql`           json         NOT NULL COMMENT 'es的查询语句',
    `threshold`     float(11, 0
) DEFAULT '1' COMMENT '阈值',
    `origin`             varchar(64)         NOT NULL COMMENT '来源，前端-产品/业务-产品/应用-appid/组件-ip、集群名/基础-域名、ip',
    `business_type`      varchar(64)         NOT NULL COMMENT '产品名: 来源，前端-产品/业务-产品/应用-appid/组件-ip、集群名/基础-域名、ip',
    `category`           tinyint(1)          NOT NULL COMMENT '指标类型,1-前端监控,2-业务监控,3-应用监控,4-组件监控,5-基础监控',
    `level`              tinyint(1)          NOT NULL DEFAULT '3' COMMENT '告警级别: 0 -- Not classified; 1 --- Information; 2 --- Warning; 3 --- critical; 4 --- Disaster',
    `creator`            varchar(32)         NOT NULL COMMENT '规则创建者, 用户钉钉的 userid',
    `updater`            varchar(32)                  DEFAULT NULL COMMENT '规则创建者, 用户钉钉的 userid',
    `responsible_people` varchar(255)        NOT NULL COMMENT '告警时间的处理人, 用户钉钉的 userid',
    `crontab`            varchar(32)         NOT NULL DEFAULT '* * * * *' COMMENT '每条规则的定时任务执行表达式, 默认为: "* * * * *"',
    `switch`             tinyint(1)                   DEFAULT '1' COMMENT '是否启用, 1 --- on; 2 --- off',
    `inuse`              tinyint(1)                   DEFAULT '2' COMMENT '是否删除, 1 --- yes; 2 --- no',
    `group_ip`           varchar(255)        NOT NULL COMMENT '告警时间接收者的组id, 多个值以 '','' 分隔',
    `description`        varchar(255)                 DEFAULT NULL COMMENT '描述',
    `created_at`         datetime(6)         NOT NULL COMMENT '记录插入时间',
    `updated_at`         datetime(6)                  DEFAULT NULL ON UPDATE CURRENT_TIMESTAMP(6) COMMENT '记录更新时间',
    `deleted_at`         datetime(6)                  DEFAULT NULL COMMENT '记录删除时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_name` (`name`) USING BTREE
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4 COMMENT ='日志规则记录表'