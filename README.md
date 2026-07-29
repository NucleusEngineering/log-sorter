# Log Sorter

A simple Go service that demonstrates how to consume
[Google Cloud Logging](https://cloud.google.com/logging) logs forwarded by a
[Log Sink](https://cloud.google.com/logging/docs/routing/overview) from a
[Pub/Sub push subscription](https://cloud.google.com/pubsub/docs/push), and sort
them into a [Google Cloud Storage (GCS)](https://cloud.google.com/storage)
bucket as timestamped JSON files.

## Features

- Consumes log entries from a Pub/Sub push subscription.
- Parses and validates the log entries.
- Sorts logs into GCS buckets based on their timestamp.
- Creates JSON files for each log entry.
- Handles errors and retries gracefully.

## How it Works

1.  **Cloud Logging** is configured to export logs to a **Pub/Sub topic** using
    a **Log Sink**.
2.  A **Pub/Sub push subscription** is created for the topic, with the endpoint
    pointing to this service.
3.  When new logs are generated, Cloud Logging sends them to the Pub/Sub topic.
4.  Pub/Sub pushes the log messages to this service's webhook endpoint.
5.  The service receives the push message, parses the log entry, extracts the
    timestamp, and arbitrary non-standard log fields from the `jsonPayload` of
    the log entry. In the provided example, the service looks for a `tenant_id`
    and a `job_id`.
6.  The service creates a JSON file containing the log entry and uploads it to a
    GCS bucket. The file is named using the tenant ID, job ID and the log's
    timestamp to ensure chronological order.

## Prerequisites

- A Google Cloud Platform (GCP) project.
- The `gcloud` CLI installed and configured.
- A GCS bucket.
- A Pub/Sub topic and a push subscription.

## Configuration

The service is configured using environment variables:

| Variable        | Description                               | Default |
| --------------- | ----------------------------------------- | ------- |
| `TARGET_BUCKET` | The name of the GCS bucket to store logs. | `""`    |

## Usage

1.  Clone the repository:
    ```bash
    git clone https://github.com/NucleusEngineering/log-sorter.git
    ```
2.  Set the required environment variables:
    ```bash
    export TARGET_BUCKET="your-gcs-bucket-name"
    ```
3.  Run the service:
    ```bash
    go run main.go
    ```

## Deployment

You can deploy this service to any platform that supports Go applications, such
as Google Cloud Run, Google Kubernetes Engine (GKE), or Google App Engine.

A `Dockerfile` is included for easy containerization.

### Deploying to Cloud Run

1.  Build the container image:
    ```bash
    gcloud builds submit --tag gcr.io/$PROJECT_ID/log-sorter
    ```
2.  Deploy the service to Cloud Run:
    ```bash
    gcloud run deploy log-sorter \
      --image gcr.io/$PROJECT_ID/log-sorter \
      --platform managed \
      --region us-central1 \
      --set-env-vars="TARGET_BUCKET=your-gcs-bucket-name"
    ```

## License

This project is licensed under the [Apache 2.0 License](LICENSE).
