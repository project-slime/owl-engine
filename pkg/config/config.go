package config

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/spf13/viper"
	"github.com/zouyx/agollo/v4"
)

var (
	// sharedConfig holds configuration across owl-engine
	sharedConfig *ServerRunOptions
)

type ServerRunOptions struct {
	ServerOptions   *ServerOptions   `mapstructure:"app"`
	MySQLOptions    *MySQLOptions    `mapstructure:"mysql"`
	RedisOptions    *RedisOptions    `mapstructure:"redis"`
	InfluxDBOptions *InfluxDBOptions `mapstructure:"influxdb"`
	LoggerOptions   *LoggerOptions   `mapstructure:"log"`
	EventOptions    *EventOptions    `mapstructure:"event"`
}

func newConfig() *ServerRunOptions {
	return &ServerRunOptions{
		ServerOptions:   NewServerOptions(),
		MySQLOptions:    NewMySQLOptions(),
		RedisOptions:    NewRedisOptions(),
		InfluxDBOptions: NewInfluxDBOptions(),
		LoggerOptions:   NewLoggerOptions(),
		EventOptions:    NewEventOptions(),
	}
}

func LoadFromApollo() error {
	// 加载配置文件
	client, err := agollo.Start()
	if err != nil {
		log.Fatalf("create client to connect apollo server error, %s", err.Error())
	}

	conf := newConfig()

	// 服务配置
	appBind := client.GetValue("engine.http.bind")
	appPort := client.GetIntValue("engine.http.port", 9530)
	appMode := client.GetValue("engine.deploy.mode")
	appSecret := client.GetValue("engine.http.secret")
	proxy := client.GetValue("engine.proxy.url")
	enableProxy := client.GetBoolValue("engine.proxy.enable", false)

	conf.ServerOptions.Bind = appBind
	conf.ServerOptions.Port = appPort
	conf.ServerOptions.Mode = appMode
	conf.ServerOptions.Secret = appSecret
	conf.ServerOptions.EnableProxy = enableProxy
	conf.ServerOptions.Proxy = proxy

	// db配置
	dbHost := client.GetValue("engine.db.host")
	dbPort := client.GetIntValue("engine.db.port", 3306)
	dbUsername := client.GetValue("engine.db.username")
	dbPassword := client.GetValue("engine.db.password")
	dbName := client.GetValue("engine.db.name")
	dbMaxOpenConn := client.GetIntValue("engine.db.maxOpenConn", 100)
	dbMaxIdleConn := client.GetIntValue("engine.db.maxIdleConn", 10)
	dbConnLifeTime := client.GetIntValue("engine.db.connLifetime", 30)

	conf.MySQLOptions.Host = dbHost
	conf.MySQLOptions.Port = dbPort
	conf.MySQLOptions.Username = dbUsername
	conf.MySQLOptions.Password = dbPassword
	conf.MySQLOptions.DBName = dbName
	conf.MySQLOptions.MaxOpenConnections = dbMaxOpenConn
	conf.MySQLOptions.MaxIdleConnections = dbMaxIdleConn
	conf.MySQLOptions.MaxConnectionLifeTime = time.Duration(dbConnLifeTime) * time.Second

	// influxdb 配置
	influxDB := client.GetValue("engine.influxdb.db")
	influxAddr := client.GetValue("engine.influxdb.addr")
	influxUsername := client.GetValue("engine.influxdb.username")
	influxPassword := client.GetValue("engine.influxdb.password")
	influxRetentionPolicy := client.GetValue("engine.influxdb.retentionPolicy")
	influxTimeout := client.GetIntValue("engine.influxdb.timeout", 30)
	metisUrl := client.GetValue("engine.metis.url")

	conf.InfluxDBOptions.Database = influxDB
	conf.InfluxDBOptions.Address = influxAddr
	conf.InfluxDBOptions.Username = influxUsername
	conf.InfluxDBOptions.Password = influxPassword
	conf.InfluxDBOptions.RetentionPolicy = influxRetentionPolicy
	conf.InfluxDBOptions.Timeout = time.Duration(influxTimeout) * time.Second
	conf.InfluxDBOptions.MetisUrl = metisUrl

	// redis配置
	redisMode := client.GetValue("engine.redis.mode")
	redisAddr := client.GetValue("engine.redis.addr")
	redisDb := client.GetIntValue("engine.redis.db", 0)
	redisPW := client.GetValue("engine.redis.password")
	redisMaxActive := client.GetIntValue("engine.redis.maxActive", 0)
	redisMaxIdle := client.GetIntValue("engine.redis.maxIdle", 0)
	redisIdleTimeout := client.GetIntValue("engine.redis.idleTimeout", 0)
	redisMaxLifeTimeout := client.GetIntValue("engine.redis.maxConnLifetime", 0)
	redisWait := client.GetBoolValue("engine.redis.wait", false)

	conf.RedisOptions.Mode = redisMode
	conf.RedisOptions.Address = redisAddr
	conf.RedisOptions.DB = redisDb
	conf.RedisOptions.Password = redisPW
	conf.RedisOptions.MaxActive = redisMaxActive
	conf.RedisOptions.MaxIdle = redisMaxIdle
	conf.RedisOptions.IdleTimeout = redisIdleTimeout
	conf.RedisOptions.MaxConnLifeTime = redisMaxLifeTimeout
	conf.RedisOptions.Wait = redisWait

	// 日志配置
	logDir := client.GetValue("engine.log.dir")
	logName := client.GetValue("engine.log.name")
	logLevel := client.GetValue("engine.log.level")
	logFormat := client.GetValue("engine.log.format")
	logAddCaller := client.GetBoolValue("engine.log.addCaller", true)
	logCallerSkip := client.GetIntValue("engine.log.callSkip", 2)
	logMaxSize := client.GetIntValue("engine.log.maxSize", 128)
	logMaxAge := client.GetIntValue("engine.log.maxAge", 7)
	logMaxBackup := client.GetIntValue("engine.log.maxBackup", 7)
	logInterval := client.GetIntValue("engine.log.interval", 24)
	logAsync := client.GetBoolValue("engine.log.async", true)
	logQueue := client.GetBoolValue("engine.log.queue", true)
	logQueueSleep := client.GetIntValue("engine.log.queueSleep", 100)
	logDebug := client.GetBoolValue("engine.log.debug", false)
	logCompress := client.GetBoolValue("engine.log.compress", true)

	conf.LoggerOptions.Dir = logDir
	conf.LoggerOptions.Name = logName
	conf.LoggerOptions.Level = logLevel
	conf.LoggerOptions.Format = logFormat
	conf.LoggerOptions.AddCaller = logAddCaller
	conf.LoggerOptions.CallerSkip = logCallerSkip
	conf.LoggerOptions.MaxSize = logMaxSize
	conf.LoggerOptions.MaxAge = logMaxAge
	conf.LoggerOptions.MaxBackup = logMaxBackup
	conf.LoggerOptions.Interval = logInterval
	conf.LoggerOptions.Async = logAsync
	conf.LoggerOptions.Queue = logQueue
	conf.LoggerOptions.QueueSleep = logQueueSleep
	conf.LoggerOptions.Debug = logDebug
	conf.LoggerOptions.Compress = logCompress

	alertHooks := client.GetValue("engine.alert.hooks")
	if strings.Compare(alertHooks, "") != 0 {
		conf.EventOptions.Hooks = append(conf.EventOptions.Hooks, strings.Split(alertHooks, ",")...)
	}

	sharedConfig = conf
	return nil
}

func LoadFromFile(configFile string) error {
	config := newConfig()

	viper.SetConfigFile(configFile)

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return fmt.Errorf("%s does not found", configFile)
		} else {
			panic(fmt.Errorf("error parsing configuration file %s", err))
		}
	}

	if err := viper.Unmarshal(config); err != nil {
		return fmt.Errorf("error unmarshal configuration %v", err)
	}

	sharedConfig = config
	return nil
}

func Get() *ServerRunOptions {
	return sharedConfig
}

func (s *ServerRunOptions) Apply(conf *ServerRunOptions) {
	if conf.RedisOptions != nil {
		conf.RedisOptions.ApplyTo(s.RedisOptions)
	}

	if conf.MySQLOptions != nil {
		conf.MySQLOptions.ApplyTo(s.MySQLOptions)
	}

	if conf.InfluxDBOptions != nil {
		conf.InfluxDBOptions.ApplyTo(s.InfluxDBOptions)
	}
}

func (s *ServerRunOptions) stripEmptyOptions() {
	if s.MySQLOptions != nil && strings.Compare(s.MySQLOptions.Host, "") == 0 {
		s.MySQLOptions = nil
	}

	if s.RedisOptions != nil && strings.Compare(s.RedisOptions.Address, "") == 0 {
		s.RedisOptions = nil
	}

	if s.InfluxDBOptions != nil && strings.Compare(s.InfluxDBOptions.Address, "") == 0 {
		s.InfluxDBOptions = nil
	}
}
