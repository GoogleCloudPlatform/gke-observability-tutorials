/*
Copyright 2019 Google LLC
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
https://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"

	"contrib.go.opencensus.io/exporter/stackdriver"
	"go.opencensus.io/trace"

	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/plugin/ochttp/propagation/tracecontext"

	"github.com/gorilla/mux"
)

var (
	projectID = os.Getenv("PROJECT_ID")
	location  = os.Getenv("LOCATION")
)

func mainHandler(w http.ResponseWriter, r *http.Request) {
	// create root span
	ctx, rootspan := trace.StartSpan(context.Background(), "incoming call")
	defer rootspan.End()

	// get span context from incoming request
	HTTPFormat := &tracecontext.HTTPFormat{}
	if spanContext, ok := HTTPFormat.SpanContextFromRequest(r); ok {
		// log to console
		fmt.Println("got incoming context")
		// create new span
		_, span := trace.StartSpanWithRemoteParent(ctx, "main logic", spanContext)
		defer span.End()

		// generate a random 0-2 int and sleep for that many seconds
		r := rand.Int63n(2)
		s := strconv.FormatInt(int64(r), 10) // for output and logging
		time.Sleep(time.Duration(r) * time.Second)
		fmt.Println("slept for " + s + " seconds") // to console
		fmt.Fprintf(w, "slept for "+s+" seconds")  // to client/browser

	}

	// do something with the context
	ctx.Done()
	// log basic output
	fmt.Printf("did not get context\n")
	// return basic output
	fmt.Fprintf(w, "hello world \n", r.URL.Path)

}

func main() {
	// set up Stackdriver exporter
	exporter, err := stackdriver.NewExporter(stackdriver.Options{ProjectID: projectID, Location: location})
	if err != nil {
		log.Fatal(err)
	}
	trace.RegisterExporter(exporter)
	trace.ApplyConfig(trace.Config{
		DefaultSampler: trace.AlwaysSample(),
	})

	// handle root request
	r := mux.NewRouter()
	r.HandleFunc("/", mainHandler)
	var handler http.Handler = r
	handler = &ochttp.Handler{
		Handler:     handler,
		Propagation: &tracecontext.HTTPFormat{}}

	log.Fatal(http.ListenAndServe(":8080", handler))
}
