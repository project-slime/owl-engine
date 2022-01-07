package config

import (
	"github.com/spf13/pflag"
	"owl-engine/pkg/util/reflectutils"
	"strings"
)

type RedisOptions struct {
	Mode            string `json:"mode,omitempty" yaml:"mode"`
	Address         string `json:"address" yaml:"address"`
	DB              int    `json:"db,omitempty" yaml:"db"`
	Password        string `json:"password,omitempty" yaml:"password"`
	MaxActive       int    `json:"max_active" yaml:"maxActive"`
	MaxIdle         int    `json:"max_idle" yaml:"maxIdle"`
	Wait            bool   `json:"wait" yaml:"wait"`
	IdleTimeout     int    `json:"idle_timeout" yaml:"idleTimeout"`
	MaxConnLifeTime int    `json:"max_conn_life_time" yaml:"maxConnLifetime"`
}

func NewRedisOptions() *RedisOptions {
	return &RedisOptions{
		Mode:            "alone", // 默认是单实例模式
		Address:         "",
		DB:              0,
		Password:        "",
		MaxActive:       0,
		MaxIdle:         0,
		Wait:            false,
		IdleTimeout:     0,
		MaxConnLifeTime: 0,
	}
}

// Validate 校验配置
func (r *RedisOptions) Validate() []error {
	errors := make([]error, 0)

	return errors
}

func (r *RedisOptions) ApplyTo(options *RedisOptions) {
	if strings.Compare(options.Address, "") != 0 {
		reflectutils.Override(options, r)
	}
}

func (r *RedisOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&r.Mode, "redis-mode", "alone", "Redis mode, eg: alone, cluster or sentinel")
	fs.StringVar(&r.Address, "redis-host", "", "Redis connection URL. If left blank, means redis is unnecessary")
	fs.IntVar(&r.DB, "redis-db", 0, "Redis db, if not specify, default 0")
	fs.StringVar(&r.Password, "redis-password", "", "Redis password, if not specify, default empty")
}
