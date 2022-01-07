package redis

import (
	"strings"
	"time"

	"owl-engine/pkg/config"
	"owl-engine/pkg/xlogs"

	"github.com/gomodule/redigo/redis"
	"github.com/letsfire/redigo/v2"
	"github.com/letsfire/redigo/v2/mode/alone"
	"github.com/letsfire/redigo/v2/mode/cluster"
	"github.com/letsfire/redigo/v2/mode/sentinel"
)

var RedisClient *redigo.Client

func Setup(conf *config.RedisOptions) {
	switch conf.Mode {
	case "alone":
		var aloneMode = alone.New(
			alone.Addr(conf.Address),
			alone.PoolOpts(
				redigo.MaxActive(conf.MaxActive),                                        // 最大连接数，默认0无限制
				redigo.MaxIdle(conf.MaxIdle),                                            // 最多保持空闲连接数，默认2*runtime.GOMAXPROCS(0)
				redigo.Wait(conf.Wait),                                                  // 连接耗尽时是否等待，默认false
				redigo.IdleTimeout(time.Duration(conf.IdleTimeout)*time.Minute),         // 空闲连接超时时间，默认0不超时
				redigo.MaxConnLifetime(time.Duration(conf.MaxConnLifeTime)*time.Minute), // 连接的生命周期，默认0不失效
				redigo.TestOnBorrow(nil),                                                // 空间连接取出后检测是否健康，默认nil
			),
			alone.DialOpts(
				redis.DialReadTimeout(time.Second),    // 读取超时，默认time.Second
				redis.DialWriteTimeout(time.Second),   // 写入超时，默认time.Second
				redis.DialConnectTimeout(time.Second), // 连接超时，默认500*time.Millisecond
				redis.DialPassword(conf.Password),     // 鉴权密码，默认空
				redis.DialDatabase(conf.DB),           // 数据库号，默认0
				redis.DialKeepAlive(time.Minute*5),    // 默认5*time.Minute
				redis.DialNetDial(nil),                // 自定义dial，默认nil
				redis.DialUseTLS(false),               // 是否用TLS，默认false
				redis.DialTLSSkipVerify(false),        // 服务器证书校验，默认false
				redis.DialTLSConfig(nil),              // 默认nil，详见tls.Config
			),
		)
		RedisClient = redigo.New(aloneMode)
	case "sentinel":
		var sentinelMode = sentinel.New(
			sentinel.Addrs(strings.Split(conf.Address, ",")),
			sentinel.PoolOpts(
				redigo.MaxActive(conf.MaxActive),                                        // 最大连接数，默认0无限制
				redigo.MaxIdle(conf.MaxIdle),                                            // 最多保持空闲连接数，默认2*runtime.GOMAXPROCS(0)
				redigo.Wait(conf.Wait),                                                  // 连接耗尽时是否等待，默认false
				redigo.IdleTimeout(time.Duration(conf.IdleTimeout)*time.Minute),         // 空闲连接超时时间，默认0不超时
				redigo.MaxConnLifetime(time.Duration(conf.MaxConnLifeTime)*time.Minute), // 连接的生命周期，默认0不失效
				redigo.TestOnBorrow(nil),                                                // 空间连接取出后检测是否健康，默认nil
			),
			sentinel.DialOpts(
				redis.DialReadTimeout(time.Second),    // 读取超时，默认time.Second
				redis.DialWriteTimeout(time.Second),   // 写入超时，默认time.Second
				redis.DialConnectTimeout(time.Second), // 连接超时，默认500*time.Millisecond
				redis.DialPassword(conf.Password),     // 鉴权密码，默认空
				redis.DialDatabase(conf.DB),           // 数据库号，默认0
				redis.DialKeepAlive(time.Minute*5),    // 默认5*time.Minute
				redis.DialNetDial(nil),                // 自定义dial，默认nil
				redis.DialUseTLS(false),               // 是否用TLS，默认false
				redis.DialTLSSkipVerify(false),        // 服务器证书校验，默认false
				redis.DialTLSConfig(nil),              // 默认nil，详见tls.Config
			),
		)
		RedisClient = redigo.New(sentinelMode)
	case "cluster":
		var clusterMode = cluster.New(
			cluster.Nodes(strings.Split(conf.Address, ",")),
			cluster.PoolOpts(
				redigo.MaxActive(conf.MaxActive),                                        // 最大连接数，默认0无限制
				redigo.MaxIdle(conf.MaxIdle),                                            // 最多保持空闲连接数，默认2*runtime.GOMAXPROCS(0)
				redigo.Wait(conf.Wait),                                                  // 连接耗尽时是否等待，默认false
				redigo.IdleTimeout(time.Duration(conf.IdleTimeout)*time.Minute),         // 空闲连接超时时间，默认0不超时
				redigo.MaxConnLifetime(time.Duration(conf.MaxConnLifeTime)*time.Minute), // 连接的生命周期，默认0不失效
				redigo.TestOnBorrow(nil),                                                // 空间连接取出后检测是否健康，默认nil
			),
			cluster.DialOpts(redis.DialReadTimeout(time.Second), // 读取超时，默认time.Second
				redis.DialWriteTimeout(time.Second),   // 写入超时，默认time.Second
				redis.DialConnectTimeout(time.Second), // 连接超时，默认500*time.Millisecond
				redis.DialPassword(conf.Password),     // 鉴权密码，默认空
				redis.DialDatabase(conf.DB),           // 数据库号，默认0
				redis.DialKeepAlive(time.Minute*5),    // 默认5*time.Minute
				redis.DialNetDial(nil),                // 自定义dial，默认nil
				redis.DialUseTLS(false),               // 是否用TLS，默认false
				redis.DialTLSSkipVerify(false),        // 服务器证书校验，默认false
				redis.DialTLSConfig(nil),              // 默认nil，详见tls.Config
			))
		RedisClient = redigo.New(clusterMode)
	default:
		xlogs.Fatalf("the redis mode is not recognized. eg: alone,cluster,sentinel")
	}

	var echoStr = "hello world"
	pong, err := RedisClient.String(func(c redis.Conn) (res interface{}, err error) {
		return c.Do("ECHO", echoStr)
	})

	if err != nil {
		xlogs.Fatalf("connect redis fail error, %v", err.Error())
	} else if strings.Compare(pong, echoStr) != 0 {
		xlogs.Fatalf("redis send ping command unexpected result, expect = %s, but = %s", echoStr, pong)
	} else {
		xlogs.Info("redis client initialize success")
	}
}
