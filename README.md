# Subscription Aggregator Service

![Go](https://img.shields.io/badge/Go-1.24+-blue.svg)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-15+-blue.svg)
![Docker](https://img.shields.io/badge/Docker-Compose-blue.svg)
![Swagger](https://img.shields.io/badge/Swagger-2.0-blue.svg)

A RESTful service for managing and aggregating user subscription data with PostgreSQL backend.

## Features

- CRUDL operations for subscription records
- Aggregation of subscription costs by period
- PostgreSQL database with migration support
- Swagger API documentation
- Docker-compose deployment
- Configuration via .env/yaml files
- Unit test coverage

## Prerequisites

- Go 1.24+
- Docker 20.10+
- Docker Compose 2.0+

## Quick Start

### 1. Clone the repository
```powershell
git clone https://github.com/BirgusL/SubscriptionAggregator.git
cd SubscriptionAggregator
```
### 2. Configuration
Edit .env file to switch between local and docker configurations:

```ini
# For local development
CONFIG_PATH=config/local.yaml

# For docker deployment
CONFIG_PATH=config/docker.yaml
```
### 3. Run with Docker Compose
```powershell
docker-compose up --build
```
### 4. Access the service
- API: http://localhost:8080

- Swagger UI: http://localhost:8080/swagger/index.html

- PostgreSQL: localhost:5432

## API Documentation
Swagger documentation is automatically generated and available at:

```text
http://localhost:8080/swagger/index.html
```
## Database Migrations
Migrations are automatically applied when starting the Docker container. The migration files are located in:

```text
/migrations/
```
## Testing
To run unit tests:

```powershell
go test -v ./...
```
```powershell
go test -cover ./...
```
## Environment Variables
Key configuration variables (set in .env and config/*.yaml):

Variable	Description	Example
- DB_HOST	PostgreSQL host	db (docker)
- DB_PORT	PostgreSQL port	5432
- POSTGRES_USER	Database username	postgres
- POSTGRES_PASSWORD	Database password	yourpassword
- POSTGRES_DB	Database name	subscriptions
- CONFIG_PATH	Path to config yaml files
- SERVER_ADDRESS	HTTP server port
## Project Structure
```text
.
├── cmd/                  # Main application
├── config/               # Configuration files
│   ├── local.yaml        # Local development config
│   └── docker.yaml       # Docker deployment config
├── pkg/                  # Core application logic
│   ├── handler/          # HTTP handlers
│   ├── repository/       # Database operations
│   ├── service/          # Business logic
│   └── models/           # Data models
├── migrations/           # Database migrations
├── pkg/                  # Shared packages
├── docs/                 # Swagger documentation
├── .env                  # Environment template
├── docker-compose.yml    # Docker configuration
└── Dockerfile            # Application container
```
## API Usage Examples
### 1. Create Subscription (POST)
```powershell
$url = "http://localhost:8080/subscriptions"
$body = @{
    service_name = "Yandex Plus"
    price = 599
    user_id = "60601fee-2bf1-4721-ae6f-7636e79a0cba"
    start_date = "2025-07-01"
} | ConvertTo-Json

$response = Invoke-RestMethod -Uri $url -Method Post -Body $body -ContentType "application/json"
$response | ConvertTo-Json -Depth 10
```
### 2. Get Subscription by ID (GET)
```powershell
$subscriptionId = "YOUR_SUBSCRIPTION_ID"
$url = "http://localhost:8080/subscriptions/$subscriptionId"

$response = Invoke-RestMethod -Uri $url -Method Get
$response | ConvertTo-Json -Depth 10
``` 
### 3. Update Subscription (PUT)
```powershell
$subscriptionId = "YOUR_SUBSCRIPTION_ID"
$url = "http://localhost:8080/subscriptions/$subscriptionId"
$body = @{
    service_name = "Yandex Plus"
    price = 799
} | ConvertTo-Json

$response = Invoke-RestMethod -Uri $url -Method Put -Body $body -ContentType "application/json"
$response | ConvertTo-Json -Depth 10
```

### 4. Delete Subscription (DELETE)
```powershell
$subscriptionId = "YOUR_SUBSCRIPTION_ID"
$url = "http://localhost:8080/subscriptions/$subscriptionId"

Invoke-RestMethod -Uri $url -Method Delete
Write-Host "Subscription deleted successfully"
```

### 5. List Subscriptions with Filters (GET)
```powershell
$url = "http://localhost:8080/subscriptions?user_id=60601fee-2bf1-4721-ae6f-7636e79a0cba&from_date=2025-01-01"

$response = Invoke-RestMethod -Uri $url -Method Get
$response | ConvertTo-Json -Depth 10
```

### 6. Get Total Cost (GET)
```powershell
$url = "http://localhost:8080/subscriptions/total?user_id=60601fee-2bf1-4721-ae6f-7636e79a0cba&service_name=Yandex Plus"

$response = Invoke-RestMethod -Uri $url -Method Get
$response | ConvertTo-Json -Depth 10
```

## License
MIT License - see LICENSE for details.
