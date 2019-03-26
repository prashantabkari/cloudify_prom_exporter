package main

import (
	"flag"
	"log"
	"net/http"
    	"github.com/prometheus/client_golang/prometheus"
    	"crypto/tls"
    	"fmt"
    	"github.com/Sirupsen/logrus"
    	"encoding/json"
	"io/ioutil"
	"bytes"
)

var (
	listeningAddress = flag.String("listen-port", ":9117", "Address on which to expose metrics.")
	metricsEndpoint  = flag.String("metrics-path", "/metrics", "Path under which to expose metrics.")
	scrapeURI        = flag.String("scrape_uri", "", "Cloudify URI endpoint. Suffix /api/v3.1/ to the HTTP URL for example http://<IP address>/api/v3.1/")
	tenant		 = flag.String("tenant", "default_tenant", "Tenant from which metrics are fetched")
	username	 = flag.String("cfy_username", "admin", "User credentials to access  cloudify cluster")
	password	 = flag.String("cfy_password","admin","Password to access cloudify cluster")
)

var apiMetricArray = [2]string{"blueprints","deployments"}


//A go structure that defines the list of metrics
//Note : the syntax is different from C++. User defined structure in C++
// is prefixed with the type of the variable but in go it is suffixed. 
type Exporter struct {
	URI    string
        client  *http.Client

	blueprints prometheus.Gauge
	deployments prometheus.Gauge
}

type  Items struct{
        Id  string   `json: "id"`
}


type Pagination struct{
        Total string `json: "total"`
        Offset string `json: "offset"`
        SizeData string `json: "size"`
}

type  Metadata struct{
        PaginationData Pagination `json: "pagination"`
}

type  CfyResponse struct{
      ItemsList  []Items `json:"items"`
      MetadataList   Metadata `json:"metadata"`
}


func NewExporter(uri string) (*Exporter, error) {


        // Define the metrics that needs to be exported into Prometheus structure
        // to be defined interms of C++ kinds of structure
	return &Exporter{
               URI: uri,
               blueprints: prometheus.NewGauge(prometheus.GaugeOpts{Namespace: "cfy", Name: "total_uploaded_blueprints", Help : " The total number of blueprints uploaded in the tenant"}),
               deployments: prometheus.NewGauge(prometheus.GaugeOpts{Namespace: "cfy", Name: "total_deployments_running", Help : " The total number of current deployments created"}),

               client: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		},
        }, nil


}


func (e *Exporter) Collect(ch chan<- prometheus.Metric) {

        fmt.Printf("Starting to collect\n")

	var buffer bytes.Buffer
	for i := 0; i < len(apiMetricArray); i++ {
		
		// Pick the ith element and add it to API call
		buffer.Reset()
		buffer.WriteString(*scrapeURI)
		buffer.WriteString(apiMetricArray[i])
		fmt.Printf("Changed URL is : %s",buffer.String())
		
	
        	req, err := http.NewRequest("GET", buffer.String(), nil)

		// Add headers and parameters
        	req.Header.Add("Tenant",*tenant)
        	req.SetBasicAuth(*username,*password)
        	q :=  req.URL.Query()
        	q.Add("_include", "id")
        	req.URL.RawQuery = q.Encode()

		fmt.Printf("Request URL is: %s ", req.URL.String())

        	resp, err := e.client.Do(req)

        	if resp.StatusCode != 200 {
			fmt.Printf("Status %s (%d): %s", resp.Status, resp.StatusCode,err)
                	return
		}

        	body, readErr := ioutil.ReadAll(resp.Body)
		if readErr != nil {
			log.Fatal(readErr)
		}

		var cfyResponse CfyResponse

		jsonerr := json.Unmarshal([]byte(body), &cfyResponse)
        	if(jsonerr != nil){
            		fmt.Println("whoops:", jsonerr)
        	}
		
		switch j := apiMetricArray[i]
		{
		   case j == "blueprints" :
					e.blueprints.Set(float64(len(cfyResponse.ItemsList)))
					ch <- e.blueprints

		   case j == "deployments" :
					e.deployments.Set(float64(len(cfyResponse.ItemsList)))
					ch <- e.deployments
				
		}

		logrus.WithFields(logrus.Fields{
                	"URI": *scrapeURI,
                	"Number of items ": float64(len(cfyResponse.ItemsList)),
        	}).Info("Number fi items")

        	resp.Body.Close()
        }
	return
}


func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	e.blueprints.Describe(ch)
}

func init() {
formatter := &logrus.TextFormatter{
		FullTimestamp: true,
	}
	logrus.SetFormatter(formatter)
	logrus.SetLevel(logrus.InfoLevel)
}

func main() {
  flag.Parse()

  // Create a new exporter object and initialize it with URI that need to be scraped
  exporter, err := NewExporter(*scrapeURI)
  if err != nil {
		logrus.WithFields(logrus.Fields{
			"uri": *scrapeURI,
			"event": "starting exporter",
		}).Fatal(err)
	}
  prometheus.MustRegister(exporter)

  http.Handle(*metricsEndpoint, prometheus.Handler())
  http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
			 <head><title>Cloudify Exporter</title></head>
			 <body>
			 <h1>Cloudify Exporter</h1>
			 <p><a href='` + *metricsEndpoint + `'>Metrics</a></p>
			 </body>
			 </html>`))
	})
  logrus.WithFields(logrus.Fields{
		"port": *listeningAddress,
		"path": *metricsEndpoint,
		"event": "listening",
	}).Info("prometheus started")

  logrus.WithFields(logrus.Fields{
		"port": *listeningAddress,
		"path": *metricsEndpoint,
		"event": "web server error",
  }).Fatal(http.ListenAndServe(*listeningAddress, nil))
}
