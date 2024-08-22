# splunk-otel-go-with-k8s

This example demonstrates how a simple Go application can be instrumented 
with the Splunk Distribution of OpenTelemetry Go. The application is based
on the dice roll application provided 
as part of https://opentelemetry.io/docs/languages/go/getting-started/. 

Prerequisites: 

* Go 1.22+

## Build the Sample Application

Create the go.mod file (if it doesn’t already exist): 
````
go mod init rolldice
````

Build and run the application: 
````
go run .
````

Test it with the following commmand: 
````
curl http://localhost:8080/rolldice
````

## Install and activate the Go instrumentation
````
go get github.com/signalfx/splunk-otel-go/distro
````

Set the service name and environment: 
````
export OTEL_SERVICE_NAME=rolldice
export OTEL_RESOURCE_ATTRIBUTES="deployment.environment=test"
````

Add the instrumentation to main.go using the distro package: 
````
import (
	"github.com/signalfx/splunk-otel-go/distro"
	"log"
	"context"  
	"net/http"
)

func main() {

	sdk, err := distro.Run()
	if err != nil {
		panic(err)
	}
	// Flush all spans before the application exits
	defer func() {
		if err := sdk.Shutdown(context.Background()); err != nil {
			panic(err)
		}
	}()
    ...
````

Add the net/http instrumentation library: 
````
go get go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp
````

Modify main.go to ensure net/http calls are instrumented: 

````
package main

import (
	"context"
	"github.com/signalfx/splunk-otel-go/distro"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"log"
	"net/http"
)

func main() {

	sdk, err := distro.Run()
	if err != nil {
		panic(err)
	}
	// Flush all spans before the application exits
	defer func() {
		if err := sdk.Shutdown(context.Background()); err != nil {
			panic(err)
		}
	}()

	// Wrap your httpHandler function
	handler := http.HandlerFunc(rolldice)
	wrappedHandler := otelhttp.NewHandler(handler, "rolldice")
	http.Handle("/rolldice", wrappedHandler)

	log.Fatal(http.ListenAndServe("0.0.0.0:8080", nil))
}
````

Build and run the application:
````
go run .
````

Test it with the following commmand:
````
curl http://localhost:8080/rolldice
````

Dockerize the application by adding the following Dockerfile:
````
FROM golang:alpine3.18

# Creates an app directory to hold your app’s source code
WORKDIR /app

# Copies everything from your root directory into /app
COPY . .

# Installs Go dependencies
RUN go mod download

# Builds your app with optional configuration
RUN go build -o /rolldice

EXPOSE 8080

CMD ["/rolldice"]
````

Build a new Docker image for the app:
````
docker build -t rolldice:1.0 .
````

Test the image using docker run:
````
docker run --name rolldice \
--detach \
-p 8080:8080 \
rolldice:1.0
````

Test it with the following command:
````
curl http://localhost:8080/rolldice
````

Let's push the image to Docker Hub:
````
docker tag rolldice:1.0 derekmitchell399/rolldice
docker push derekmitchell399/rolldice
````

Then, deploy the image to our Kubernetes cluster: 
````
kubectl apply -f kubernetes.yaml
````

Ensure an OpenTelemetry collector is deployed to our K8s cluster as well: 
````
helm repo add splunk-otel-collector-chart https://signalfx.github.io/splunk-otel-collector-chart
helm repo update
helm install splunk-otel-collector --set="splunkObservability.accessToken=***,clusterName=test,splunkObservability.realm=us1,gateway.enabled=false,environment=test" splunk-otel-collector-chart/splunk-otel-collector
````

Get the IP address for our rolldice service in K8s: 
````
kubectl describe svc rolldice | grep IP
````

Then use the IP to test the service: 
````
curl http://10.101.167.162:8080/rolldice
````



