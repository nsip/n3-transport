// n3influx_publisher.go

package n3influx

import (
	"fmt"
	"log"
	"time"

	influx "github.com/influxdata/influxdb1-client/v2"
	"github.com/nsip/n3-messages/messages/pb"
	"github.com/spf13/viper"
)

type DBClient struct {
	cl influx.Client
	ch chan *influx.Point
}

func NewPublisher() (*DBClient, error) {

	n3ic := &DBClient{
		ch: make(chan *influx.Point),
	}

	ifclient, err := influxClient()
	if err != nil {
		return nil, err
	}
	n3ic.cl = ifclient
	log.Println("starting storeage handler...")
	go n3ic.startStorageHandler()
	return n3ic, nil
}

func (n3ic *DBClient) Query(q influx.Query) (*influx.Response, error) {
	return n3ic.cl.Query(q)
}

func influxClient() (influx.Client, error) {

	influxAddr := viper.GetString("influx_addr")
	if influxAddr == "" {
		// if not in config try default
		influxAddr = "http://localhost:8086"
	}
	log.Println("connecting to influx on: ", influxAddr)

	c, err := influx.NewHTTPClient(influx.HTTPConfig{
		Addr: influxAddr,
		// Username: username,
		// Password: password,
	})
	if err != nil {
		log.Println("Cannot connect to influx server")
		return nil, err
	}
	return c, nil

}

//
// influx db inserts are best done as batches so this
// handler listens on the storage channel, and sends inserts to
// influx when a batch size has been reached or a timing threshold
// is passed
//
func (n3ic *DBClient) startStorageHandler() {

	var coll []*influx.Point
	batchInterval := time.Duration(time.Millisecond * 500)
	batchSize := 500
	tick := time.NewTicker(batchInterval)

	// if database already exists, this is ignored
	q := influx.NewQuery("CREATE DATABASE tuples", "", "")
	if response, err := n3ic.cl.Query(q); err == nil && response.Error() == nil {
		fmt.Println(response.Results)
	}

	for {
		timeout := false

		select {
		case pt := <-n3ic.ch:
			coll = append(coll, pt)
		case <-tick.C:
			timeout = true
		}

		if (timeout || len(coll) >= batchSize) && len(coll) > 0 {
			bp, err := influx.NewBatchPoints(influx.BatchPointsConfig{
				Database: "tuples",
				// Precision: "s",
			})
			if err != nil {
				//TODO:
				log.Println(err)
			}
			bp.AddPoints(coll)
			err = n3ic.cl.Write(bp)
			if err != nil {
				//TODO:
				log.Println(err)
			} else {
				coll = nil
			}
		}
	}
}

//
// send the tuple to influx, passes into batching storage handler
//
func (n3ic *DBClient) StoreTuple(tuple *pb.SPOTuple, contextName string) error {
	// don't publish empty objects (deletion markers): we have tombstoning for that
	if len(tuple.Object) == 0 {
		return nil
	}

	// extract data from tuple and use to construct point
	tags := map[string]string{
		"subject":   tuple.Subject,
		"predicate": tuple.Predicate,
		"object":    tuple.Object,
		"tombstone": "false",
	}
	fields := map[string]interface{}{
		"version": tuple.Version,
		// "predicate": tuple.Predicate,
		// "object":    tuple.Object,
	}

	pt, err := influx.NewPoint(contextName, tags, fields, time.Now())
	if err != nil {
		return err
	}

	n3ic.ch <- pt

	return nil
}

// "delete" the tuple: tuple is stored but tombstoned
func (n3ic *DBClient) DeleteTuple(tuple *pb.SPOTuple, contextName string) error {

	// extract data from tuple and use to construct point
	tags := map[string]string{
		"subject":   tuple.Subject,
		"predicate": tuple.Predicate,
		"object":    tuple.Object,
		"tombstone": "true",
	}
	fields := map[string]interface{}{
		"version": tuple.Version,
		// "predicate": tuple.Predicate,
		// "object":    tuple.Object,
	}

	q := influx.NewQuery(fmt.Sprintf("SELECT object, version FROM %s WHERE subject = '%s' AND predicate = '%s' ORDER BY time DESC LIMIT 1", contextName, tuple.Subject, tuple.Predicate), "tuples", "")
	if response, err := n3ic.cl.Query(q); err == nil && response.Error() == nil {
		if len(response.Results) > 0 && len(response.Results[0].Series) > 0 && len(response.Results[0].Series[0].Values) > 0 {
			if response.Results[0].Series[0].Values[0][1] != nil {
				tags["object"] = interface{}(response.Results[0].Series[0].Values[0][1]).(string)
			} else {
				tags["object"] = ""
			}
		}
	}
	pt, err := influx.NewPoint(contextName, tags, fields, time.Now())
	if err != nil {
		return err
	}

	n3ic.ch <- pt

	return nil
}
