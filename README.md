
# Data Processing Application

This guide outlines the development of a data processing application in Go, aimed at processing scanning results from Google Pub/Sub, and managing distinct (IP, port, service) information sets with support for horizontal scaling and at-least-once processing semantics.

## Step 1: Setting Up the Project

### Initialize the Go Module

- Navigate to your project root.
- Run `go mod init <your_module_name>` to start a new Go module.

### Install Necessary Dependencies

- **Google Cloud Pub/Sub Client:** `go get cloud.google.com/go/pubsub`
- **PostgreSQL Driver:** `go get -u github.com/lib/pq`
- **Configuration Management (Optional):** `go get github.com/spf13/viper` for advanced config handling.

## Step 2: Defining the Application Structure

Adhere to the directory structure provided, with each directory containing Go files specific to parts of the application.

## Step 3: Implementing the Application Components
```code
data-processor/
│
├── cmd/
│   └── main.go           # Entry point, sets up the application
│
├── config/
│   └── config.go         # Configuration management
│
├── logger/
│   └── logger.go         # Logging setup and configuration
│
├── models/
│   └── models.go         # Data models for the application
│
├── repositories/
│   └── scan_result_processor.go   # repositories
│
├── services/
│   └── scan_result_processor.go            # service  Layer


```

## Step 4: Processing Logic in Detail

### Fetching Messages from Pub/Sub

- Setup a subscriber to "scan-sub" for message fetching, ensuring at-least-once delivery.

### Message Processing

- Decode messages based on `data_version`. Decode base64 `response_bytes_utf8` for `data_version = 1`, or directly use `response_str` for `data_version = 2`.
- Normalize data for consistent formatting.

### Database Operations

- Insert or update records in PostgreSQL, keeping each (IP, port, service) record current with the latest timestamp and service response.

## Step 5: Horizontal Scaling and Reliability

- Design the application to be stateless for horizontal scaling.
- Utilize database transactions for data integrity.
- Ensure multiple consumer support in Pub/Sub subscriber setup without message duplication.

## Step 6: Deployment Considerations

- Use containerization (e.g., Docker) for deployment simplicity and scaling.
- Consider a managed Kubernetes service for easier scaling and management.

This roadmap provides a structured approach to developing an application that meets the requirements for efficient data processing, scalability, and reliability.
