// Copyright 2025 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"cloud.google.com/go/storage"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/event/datacodec/json"
)

type PubSubMessage struct {
	Message struct {
		Data []byte `json:"data,omitempty"`
		ID   string `json:"id"`
	} `json:"message"`
	Subscription string `json:"subscription"`
}

type LogEntry struct {
	JSONPayload struct {
		TenantID string `json:"tenant_id"` // NOTE:custom field inserted via structured logging: https://cloud.google.com/logging/docs/structured-logging
		JobID    string `json:"job_id"`    // NOTE:custom field inserted via structured logging
	} `json:"jsonPayload"`
	Timestamp time.Time `json:"receiveTimestamp"`
}

var (
	targetBucket string
	gcsClient    *storage.Client
)

func eventHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Unpack CloudEvent
	ce, err := cloudevents.NewEventFromHTTPRequest(r)
	if err != nil {
		log.Printf("failed to parse CloudEvent: %v", err)
		http.Error(w, "Bad Request: expected CloudEvent", http.StatusBadRequest)
		return
	}

	// Unmarshal PubSub message wrapper
	var msg PubSubMessage
	err = ce.DataAs(&msg)
	if err != nil {
		log.Printf("failed to unmarshal PubSubMessage: %v", err)
		http.Error(w, "Bad Request: expected PubSubMessage", http.StatusBadRequest)
		return
	}

	// Unmarshal Cloud Logging LogEntry
	var entry LogEntry
	err = json.Decode(ctx, msg.Message.Data, &entry)
	if err != nil {
		log.Printf("failed to unmarshal LogEntry: %v", err)
		http.Error(w, "Bad Request: expected LogEntry", http.StatusBadRequest)
		return
	}

	// Construct GCS object name
	targetObject := fmt.Sprintf(
		"%s/%s/%d.json",
		entry.JSONPayload.TenantID,
		entry.JSONPayload.JobID,
		entry.Timestamp.UnixNano(),
	)

	// Write the log entry to GCS
	writer := gcsClient.Bucket(targetBucket).Object(targetObject).NewWriter(ctx)
	if _, err := writer.Write(msg.Message.Data); err != nil {
		log.Printf("failed to write to GCS: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	writer.Close()
	w.Write([]byte("ok"))
}

func main() {
	ctx := context.Background()

	targetBucket = os.Getenv("TARGET_BUCKET")
	if targetBucket == "" {
		log.Fatal("TARGET_BUCKET not set")
		return
	}

	var err error
	gcsClient, err = storage.NewClient(ctx)
	if err != nil {
		log.Fatal("unable to init storage client")
		return
	}
	defer gcsClient.Close()

	http.HandleFunc("/", eventHandler)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Default to 8080 if PORT is not set.
	}
	log.Printf("Listening on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
