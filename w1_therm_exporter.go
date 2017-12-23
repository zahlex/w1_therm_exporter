package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"flag"
)

var (
	tempVec = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "w1",
			Subsystem: "therm",
			Name:      "temperature_celsius",
			Help:      "Temperatures in Celsius read via w1_therm linux kernel module",
		},
		[]string{"sensor_id"},
	)

	currentSensors = make(map[string]int64)
)

func init() {
	// Metrics have to be registered to be exposed:
	prometheus.MustRegister(tempVec)
}

func main() {

	go func() {
		for {
			// Fetch list of sensors
			sensorsList, err := ioutil.ReadDir("/sys/bus/w1/devices/")
			if err != nil {
				log.Fatal(err)
			}

			// Iterate over list of sensors
			for _, sensor := range sensorsList {

				// Fetch sensors current value as file
				file, err := ioutil.ReadFile("/sys/bus/w1/devices/" + sensor.Name() + "/w1_slave")
				if err != nil {
					continue
				}

				str := string(file)

				// Check for correct checksum
				if !strings.Contains(str, "YES") {
					tempVec.DeleteLabelValues(sensor.Name())
					continue
				}

				// Extract temperature value (Celsius) from file
				slc := strings.SplitAfter(str, "t=")

				// Check if there is an valid string in file
				if len(slc) < 2 {
					continue
				}
				str = slc[1]

				// Remove unnecessary characters
				str = strings.Replace(str, "\n", "", -1)
				str = strings.Replace(str, "\r", "", -1)
				str = strings.Replace(str, "\t", "", -1)

				// Convert string to float
				tempInt, err := strconv.Atoi(str)
				if err != nil {
					continue
				}

				// Correct formatting
				tempInt = tempInt / 100
				tempFloat := float64(tempInt) / 10

				// Print to stdout and add to / update gauge vector
				fmt.Println(sensor.Name(), tempFloat)
				tempVec.WithLabelValues(sensor.Name()).Set(tempFloat)

				// Refresh timeout
				currentSensors[sensor.Name()] = time.Now().UTC().Unix()
			}

			// Remove timed out sensors from gauge vector
			for name, timestamp := range currentSensors {

				// Check if last value is older than 5 min
				if (timestamp + 300) < time.Now().UTC().Unix() {
					tempVec.DeleteLabelValues(name)
					delete(currentSensors, name)
				}
			}

			// Give 1w bus time for some housekeeping
			time.Sleep(10 * time.Second)
		}
	}()

	httpAddr := flag.String("httpAddr", "0.0.0.0:8080", "HTTP Address")
	flag.Parse()

	// The Handler function provides a default handler to expose metrics
	// via an HTTP server. "/metrics" is the usual endpoint for that.
	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(*httpAddr, nil))
}
