package main

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/ssimunic/gosensors"
)

var (
	fanspeedDesc = prometheus.NewDesc(
		"lm_fan_rpm",
		"fan rmp.",
		[]string{"fanid"},
		nil)
)

type FanCollector struct{}

func NewFanCollector() *FanCollector {
	return &FanCollector{}
}

func (c *FanCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- fanspeedDesc
}

func (c *FanCollector) Collect(ch chan<- prometheus.Metric) {
	metrics := c.getFanMetrics()

	for id, rpm := range metrics {
		ch <- prometheus.MustNewConstMetric(fanspeedDesc, prometheus.GaugeValue, rpm, id)
	}
}

func (c *FanCollector) getFanMetrics() map[string]float64 {
	sensors, err := gosensors.NewFromSystem()

	if err != nil {
		panic(err)
	}

	metrics := make(map[string]float64, 0)
	for chip := range sensors.Chips {
		// Iterate over entries
		for key, value := range sensors.Chips[chip] {
			// If CPU or GPU, print out
			if strings.Contains(key, "fan") {
				spllitted := strings.Split(value, " ")
				rpm, err := strconv.ParseFloat(spllitted[0], 64)
				if err != nil {
					log.Println(err)
					continue
				}
				metrics[key] = rpm
			}
		}
	}

	return metrics
}

func main() {
	fanCollector := NewFanCollector()
	prometheus.MustRegister(fanCollector)
	http.Handle("/metrics", promhttp.Handler())

	err := http.ListenAndServe(":9225", nil)

	if err != nil {
		log.Println(err)
	}
}
