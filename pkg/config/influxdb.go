package config

import (
	"owl-engine/pkg/util/reflectutils"
	"time"

	"github.com/spf13/pflag"
)

type InfluxDBOptions struct {
	MetisUrl        string        `json:"metis_url" yaml:"metisUrl"`
	Address         string        `json:"address" yaml:"address"`
	Username        string        `json:"username" yaml:"username"`
	Password        string        `json:"password" yaml:"password"`
	Database        string        `json:"database" yaml:"database"`
	RetentionPolicy string        `json:"retention_policy" yaml:"retentionPolicy"`
	Timeout         time.Duration `json:"timeout" yaml:"timeout"`
}

func NewInfluxDBOptions() *InfluxDBOptions {
	return &InfluxDBOptions{
		MetisUrl:        "",
		Address:         "",
		Username:        "",
		Password:        "",
		Database:        "",
		RetentionPolicy: "",
		Timeout:         time.Duration(10) * time.Second,
	}
}

func (i *InfluxDBOptions) Validate() []error {
	errors := make([]error, 0)

	return errors
}

func (i *InfluxDBOptions) ApplyTo(options *InfluxDBOptions) {
	reflectutils.Override(options, i)
}

func (i *InfluxDBOptions) AddFlags(fs *pflag.FlagSet) {

}
