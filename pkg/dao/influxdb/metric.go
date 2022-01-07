package influxdb

import (
	"encoding/json"
	"errors"

	client "github.com/influxdata/influxdb1-client/v2"
)

type metric struct{}

var Metric = new(metric)

// 查询数据
func (m *metric) Query(cmd, database, retentionPolicy string, chunk int, cli client.Client) ([]float64, error) {
	var result = make([]float64, 0)

	if cli == nil {
		return result, errors.New("influxdb does not initialize")
	}

	query := client.Query{
		Command:         cmd,
		Database:        database,
		RetentionPolicy: retentionPolicy,
		ChunkSize:       chunk,
	}

	if response, err := cli.Query(query); err == nil {
		if len(response.Results) > 0 {
			series := response.Results[0].Series
			if len(series) > 0 {
				if len(series[0].Values) > 0 {
					for _, v := range series[0].Values {
						value, _ := v[1].(json.Number).Float64()
						result = append(result, value)
					}
				}
			}
		}
		return result, nil
	} else {
		return result, err
	}
}
