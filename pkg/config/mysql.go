package config

import (
	"time"

	"owl-engine/pkg/util/reflectutils"

	"github.com/spf13/pflag"
)

type MySQLOptions struct {
	Host                  string        `json:"host,omitempty" yaml:"host" description:"MySQL service host address"`
	Port                  int           `json:"port,omitempty" yaml:"port" description:"MySQL service port"`
	Username              string        `json:"username,omitempty" yaml:"username"`
	Password              string        `json:"password" yaml:"password"`
	DBName                string        `json:"db_name" yaml:"db_name"`
	MaxIdleConnections    int           `json:"maxIdleConnections,omitempty" yaml:"maxIdleConnections"`
	MaxOpenConnections    int           `json:"maxOpenConnections,omitempty" yaml:"maxOpenConnections"`
	MaxConnectionLifeTime time.Duration `json:"maxConnectionLifeTime,omitempty" yaml:"maxConnectionLifeTime"`
}

func NewMySQLOptions() *MySQLOptions {
	return &MySQLOptions{
		Host:                  "",
		Port:                  3306,
		Username:              "",
		Password:              "",
		MaxIdleConnections:    100,
		MaxOpenConnections:    100,
		MaxConnectionLifeTime: time.Duration(10) * time.Second,
	}
}

// Validate 校验配置
func (m *MySQLOptions) Validate() []error {
	errors := make([]error, 0)

	return errors
}

// ApplyTo 重写配置项
func (m *MySQLOptions) ApplyTo(options *MySQLOptions) {
	reflectutils.Override(options, m)
}

// AddFlags 命令行加载配置项
func (m *MySQLOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&m.Host, "mysql-host", m.Host, ""+
		"MySQL service host address. If left blank, the following related mysql options will be ignored.")

	fs.StringVar(&m.Username, "mysql-username", m.Username, ""+
		"Username for access to mysql service.")

	fs.StringVar(&m.Password, "mysql-password", m.Password, ""+
		"Password for access to mysql, should be used pair with password.")

	fs.IntVar(&m.MaxIdleConnections, "mysql-max-idle-connections", m.MaxOpenConnections, ""+
		"Maximum idle connections allowed to connect to mysql.")

	fs.IntVar(&m.MaxOpenConnections, "mysql-max-open-connections", m.MaxOpenConnections, ""+
		"Maximum open connections allowed to connect to mysql.")

	fs.DurationVar(&m.MaxConnectionLifeTime, "mysql-max-connection-life-time", m.MaxConnectionLifeTime, ""+
		"Maximum connection life time allowed to connect to mysql.")
}
