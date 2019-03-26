A prometheus exporter in Go to fetch a basic metric - Number of Blueprints from Cloudify cluster. 

Note : The exporter still need to be enhanced to cover other metrics such as deployments

# Getting Started

## Prerequisites
* Golang 
* Cloudify that uses API version 3.1


## Installing

* Clone the repository.
* Build the exporter from the directory : 
`go build;go install`

## Running the exporter
`cfy exporter <flags>`
Flags take the following options

```
 -cfy_password string
        Password to access cloudify cluster (default "admin")
  -cfy_username string
        User credentials to access  cloudify cluster (default "admin")
  -listen-port string
        Address on which to expose metrics. (default ":9117")
  -metrics-path string
        Path under which to expose metrics. (default "/metrics")
  -scrape_uri string
        Cloudify URI endpoint. Suffix /api/v3.1/ to the HTTP URL for example http://<IP address>/api/v3.1/
  -tenant string
        Tenant from which metrics are fetched (default "default_tenant")


```

## Running the Tests
- To be Added

## Deployment
-------------------

* To deploy, build the binary: go build cfy_exporter.go



