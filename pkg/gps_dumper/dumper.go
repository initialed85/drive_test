package gps_dumper

import (
	"fmt"
	"github.com/stratoberry/go-gpsd"
	"log"
	"time"
)

type Dumper struct {
	gps      *gpsd.Session
	callback func(Output) error
}

type Output struct {
	Timestamp time.Time       `json:"timestamp"`
	Report    *gpsd.TPVReport `json:"report"`
}

func New(host string, port int, callback func(Output) error) (Dumper, error) {
	gps, err := gpsd.Dial(fmt.Sprintf("%v:%v", host, port))
	if err != nil {
		return Dumper{}, err
	}

	d := Dumper{
		gps:      gps,
		callback: callback,
	}

	gps.AddFilter("TPV", d.tpvFilter)

	return d, nil
}

func (d *Dumper) tpvFilter(r interface{}) {
	report := r.(*gpsd.TPVReport)

	output := Output{
		time.Now(),
		report,
	}

	err := d.callback(output)
	if err != nil {
		log.Fatal(err)
	}
}

func (d *Dumper) Watch() (done chan bool) {
	return d.gps.Watch()
}
