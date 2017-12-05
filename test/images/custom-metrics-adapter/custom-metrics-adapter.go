/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"strconv"
	"time"
)

var (
	port = flag.Int("port", 80, "Port number.")

	metrics = []metric{
		// order is significant; queries must be first.
		{"queries", time.Second, 0},
		{"packets", time.Second, 0},
	}
)

type metric struct {
	Name  string
	Rate  time.Duration // XXX remove
	Value int64
}

func (m metric) String() string {
	return fmt.Sprintf("%s %v", m.Name, m.Value)
}

func heartbeat(interval time.Duration) {
	for {
		dumpMetrics(metrics)
		time.Sleep(interval)
	}
}

func logRequest(req *http.Request) {
	if dreq, err := httputil.DumpRequest(req, true); err != nil {
		log.Println(err)
	} else {
		log.Println(string(dreq))
	}
}

func writeText(w http.ResponseWriter, s string) {
	log.Println(s)
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(s))
	w.Write([]byte("\r\n"))
}

func dumpMetrics(metrics []metric) {
	for _, m := range metrics {
		log.Println(m)
	}
}

func updateMetric(value string, m *metric) error {
	if value == "" {
		return nil
	}
	n, err := strconv.ParseInt(value, 10, 0)
	if err != nil {
		return err
	}
	m.Value = n
	return nil
}

func serveMetrics(w http.ResponseWriter, req *http.Request) {
	metrics[0].Value += 1

	logRequest(req)

	w.Header().Set("Content-Type", "text/plain")

	for _, m := range metrics {
		w.Write([]byte(fmt.Sprintf("%s\n", m)))
	}
}

func serveRoot(w http.ResponseWriter, req *http.Request) {
	metrics[0].Value += 1

	logRequest(req)

	if req.Method != "POST" {
		return
	}

	if err := req.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	for i := range metrics {
		if err := updateMetric(req.FormValue(metrics[i].Name), &metrics[i]); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	dumpMetrics(metrics)
}

func main() {
	log.SetFlags(log.Flags() | log.Lmicroseconds)
	flag.Parse()
	http.HandleFunc("/metrics", serveMetrics)
	http.HandleFunc("/", serveRoot)
	go heartbeat(5 * time.Second)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), nil))
}
