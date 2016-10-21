package client

import (
	"fmt"

	log "github.com/Sirupsen/logrus"
	client "github.com/influxdata/influxdb/client/v2"
	"github.com/influxdata/influxdb/models"
	"github.com/serialx/hashring"
)

type Client struct {
	config        Config                   //represents configuration file
	downstream    []client.Client          // represents alive nodes' url
	downstreamMap map[string]client.Client //represents a mapping relationship between nodes' url and actual client
	ring          interface {
		GetNode(string) (string, bool)
	}
}

func NewClient(config Config) (*Client, error) {
	c := &Client{config: config}

	addrs := config.Addrs

	c.downstreamMap = make(map[string]client.Client, len(addrs))
	for _, addr := range addrs {
		influxClient, err := client.NewHTTPClient(client.HTTPConfig{
			Addr: addr,
		})
		if err != nil {
			return nil, fmt.Errorf("new http client: %v", err)
		}
		c.downstream = append(c.downstream, influxClient)
		c.downstreamMap[addr] = influxClient
	}

	c.ring = hashring.New(addrs)
	return c, nil
}

//makeBatchPoints will convert points to BatchPoints which client can write
//into influxDB
func (c *Client) makeBatchPointsMap(database string, consistency string, precision string, points []models.Point) map[string]client.BatchPoints {
	batchMap := map[string]client.BatchPoints{}
	for _, point := range points {
		measurement := point.Name()
		if node, ok := c.ring.GetNode(measurement); ok {
			var batch client.BatchPoints
			var ok bool
			if batch, ok = batchMap[node]; !ok {
				var err error
				batch, err = client.NewBatchPoints(client.BatchPointsConfig{
					Precision:        precision,
					Database:         database,
					WriteConsistency: consistency,
				})
				if err != nil {
					log.Errorln("failed to create new batch point: %v", err)
					continue
				}
				batchMap[node] = batch
			}
			clientPoint, err := client.NewPoint(
				point.Name(),
				point.Tags().Map(),
				point.Fields(),
				point.Time(),
			)
			if err != nil {
				log.Errorln("failed to create new point: %v", err)
				continue
			}
			batch.AddPoint(clientPoint)
		}
	}
	return batchMap
}

//Write will write batch points to nodes. Note: Except node leavs or joins cluster, Write will write data with same measurement into same group of nodes.
func (c *Client) Write(database string, consistency string, precision string, points []models.Point) error {
	batchMap := c.makeBatchPointsMap(database, consistency, precision, points)
	for node, batch := range batchMap {
		downstream, ok := c.downstreamMap[node]
		if !ok {
			return fmt.Errorf("client %s does not exists", downstream)
		}
		if err := downstream.Write(batch); err != nil {
			return fmt.Errorf("failed to write batch: %v", err)
		}
	}
	return nil
}
