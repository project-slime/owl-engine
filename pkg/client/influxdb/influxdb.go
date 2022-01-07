package influxdb

import (
	"owl-engine/pkg/xlogs"

	"owl-engine/pkg/config"

	client "github.com/influxdata/influxdb1-client/v2"
)

var InfluxDBClient *client.Client

func Setup(conf *config.InfluxDBOptions) {
	cli, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     conf.Address,
		Username: conf.Username,
		Password: conf.Password,
	})

	if err != nil {
		xlogs.Fatalf("InfluxDB client initialize error, %s", err.Error())
	}

	InfluxDBClient = &cli
	xlogs.Info("InfluxDB client initialize success")
}
