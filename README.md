
# Todo Plus

[![Go Report Card](https://goreportcard.com/badge/github.com/berikulyBeket/todo-plus)](https://goreportcard.com/report/github.com/berikulyBeket/todo-plus)
[![License](https://img.shields.io/github/license/berikulyBeket/todo-plus.svg)](https://github.com/berikulyBeket/todo-plus/blob/master/LICENSE)
[![Release](https://img.shields.io/github/v/release/berikulyBeket/todo-plus.svg)](https://github.com/berikulyBeket/todo-plus/releases/)

**Todo Plus** is a robust and scalable Golang application designed using modern architectural principles. It handles high performance and reliability at scale while maintaining a clean codebase. This application incorporates powerful features such as caching, logging, metrics, and advanced search services.

Built to tackle real-world challenges, **Todo Plus** is configured with clusters for Redis (using Sentinel), Kafka, and Elasticsearch in master/replica setups, ensuring high availability and fault tolerance. Its modular design promotes simplicity and growth, allowing for easy client switching and extension.

## Table of Contents
- [Features](#features)
- [Technologies Used](#technologies-used)
- [Installation](#installation)
- [Project Structure](#project-structure)
- [Usage](#usage)
- [Testing](#testing)
- [Clusters & High Availability](#clusters--high-availability)
- [Additional Tools](#additional-tools)
- [Contributing](#contributing)

## Features
- **Clean Architecture**: Modular, maintainable, and scalable structure based on Clean Architecture principles.
- **Redis Sentinel**: Provides high availability and automated failover for Redis clusters.
- **Master/Replica Clustering**: Seamless integration with Kafka, Redis (via Sentinel), and Elasticsearch clusters for resilience.
- **Caching**: Integrated Redis for efficient caching to enhance performance.
- **Logging & Monitoring**: Centralized logging with Prometheus metrics for effective monitoring and troubleshooting.
- **Advanced Search**: High-performance search capabilities powered by Elasticsearch.
- **Design Principles**: Utilizes SOLID principles and the observer pattern for efficient service communication.
- **Grafana Dashboards**: Real-time visualization of metrics through pre-configured Grafana dashboards.
- **Swagger API Documentation**: Automatically generated API documentation for easy access.

## Technologies Used
- **Programming Languages**: Golang (Go), Bash scripting
- **Software Architecture**: Clean Architecture, SOLID principles, Modular design
- **Databases & Caching**: Postgres, Redis
- **Messaging & Event Streaming**: Kafka
- **Search Technologies**: Elasticsearch
- **Logging & Monitoring**: Prometheus, Grafana, Centralized logging
- **High Availability & Fault Tolerance**: Redis Sentinel, Kafka Clusters, Elasticsearch Clustering
- **Containers & Orchestration**: Docker, Docker Compose
- **API Documentation**: Swagger
- **Testing & Quality Assurance**: Unit tests, Integration tests, golangci-lint
- **Version Control**: Git, GitHub

## Installation

Follow these steps to set up **Todo Plus** on your local machine:

1. **Clone the Repository**:
   ```bash
   git clone https://github.com/berikulyBeket/todo-plus.git
   ```

2. **Navigate to Project Directory**:
   ```bash
   cd todo-plus
   ```

3. **Install Dependencies**:
   ```bash
   go mod tidy
   ```

4. **Configure Environment Variables**:
   
   Make sure to configure your environment variables appropriately:
   - Copy the example environment file to create your local configuration:
     ```bash
     cp .env.example .env
     ```
   - Copy the development environment file to create your Docker configuration:
     ```bash
     cp .env.dev.example .env.dev
     ```
5. **Generate Configuration Files and SSL Keys**:
   
   - Use the provided commands to generate configuration files and SSL keys:
      ```bash
      make generate-sentinel-conf 
      make generate-redis-slave-conf
      make generate-ssl-certs
      ```
6. **Install Required Binaries**:

   - Install all necessary binaries for the project:
      ```bash
      make bin-deps
      ```
7. **Run Database Migrations**:

   - Execute the migrations:
      ```bash
      make migrate-up
      ```

## Project Structure

The project is organized for scalability and maintainability:

```
/cmd                - Application entry points
/config             - Configuration files
/docs               - Swagger documentation
/bin                - Compiled binaries
/migrations         - Database migration scripts
/integration-test   - Integration tests
/internal           - Core business logic and services
/pkg                - Reusable libraries and components
/utils              - Utility functions and helper methods
/ssl-certs          - SSL certificate files
/elastic-mappings   - Mappings for elasticsearch indexes
/grafana-dashboards - Grafana dashboard configurations
```

## Usage

### Running the Application

Get started quickly:

1. **To run app fully on Docker**:
   ```bash
   make compose-up
   ```

2. **To run services on Docker and the app locally**:
   
   - Start necessary services using Docker Compose:
      ```bash
      make compose-up-services
      ```
   - Build and run the application:
      ```bash
      make run
      ```

### Accessing API Documentation

Once the application is running, access the Swagger-generated API documentation:
```
http://localhost:8081/
```

### Monitoring Metrics with Grafana

View key metrics in the Grafana dashboard:
```
http://localhost:3000
```

## Testing

**Todo Plus** includes comprehensive testing:

- **Unit Tests**:
   ```bash
   make unit-test
   ```

- **Integration Tests**:

   - To run integration tests, first start the application and necessary services using Docker Compose:
      ```bash
      make compose-up
      ```
      Then, run the integration tests:
      ```bash
      make integration-test
      ```
   - Alternatively, you can start the application specifically for integration tests:
      ```bash
      make compose-up-integration
      ```

## Clusters & High Availability

Designed for a distributed environment, **Todo Plus** ensures high availability and fault tolerance with the following configurations:

- **Redis Clustering & Sentinel**: Set up in a master/replica configuration with automated failover.
- **Kafka Clusters**: Provides durable message streaming and reliable event handling.
- **Elasticsearch**: Configured with multiple nodes for distributed search and fault-tolerant query processing.

## Additional Tools

- **Linting**: Code quality is maintained with tools like `golangci-lint`.
- **Caching**: Redis optimizes performance by caching frequent queries.
- **Logging**: Centralized logging formatted for easy analysis.
- **Monitoring**: Prometheus metrics are ready for scraping.
- **Scripts**: Custom Bash scripts facilitate environment setup.

## Contributing

Contributions are welcome! Feel free to fork the repository, create a feature branch, and open a pull request. Ensure all tests pass before submitting.
