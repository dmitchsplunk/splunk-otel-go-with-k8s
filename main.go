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
