# Load Balancer

A simple HTTP load balancer written in Go.

## Features

- Round-robin load balancing
- Health checks for upstream servers
- Rate limiting

## Tech Stack

- **Go:** For the core application logic.
- **Docker & Docker Compose:** For containerization and local development.
- **Makefile:** To simplify common commands.

## Getting Started

### Prerequisites

- [Docker](https://docs.docker.com/get-docker/)
- [Make](https://www.gnu.org/software/make/)

### Configuration

The application is configured via environment variables. You can create a `.env` file in the root of the project to set your own values. Start by copying the example:

```sh
cp .env.example .env
```

Here is an example of the `.env.example` file:
```dotenv
# Server Configuration
PORT=8080
SHUTDOWN_TIMEOUT=30s

# Logging Configuration
LOG_LEVEL=info

# Load Balancer Configuration
UPSTREAMS=http://testserver1:9001,http://testserver2:9002,http://testserver3:9003,http://testserver4:9004
HEALTH_CHECK_INTERVAL=5s

# Rate Limiter Configuration
RATE_LIMIT_DEFAULT_CAPACITY=100
RATE_LIMIT_DEFAULT_RATE=10
```

### Running Locally

1.  **Clone the repository:**
    ```sh
    git clone https://github.com/your-username/load-balancer.git
    cd load-balancer
    ```

2.  **Start the services:**
    This command will build the Docker images and start the load balancer along with four test servers.
    ```sh
    make up
    ```
    The load balancer will be accessible at `http://localhost:8080`.

3.  **Send a test request:**
    You can send a simple request using `curl`:
    ```sh
    curl http://localhost:8080
    ```

4.  **Run benchmark (optional):**
    To run a simple benchmark against the load balancer, you can use `make bench`. This uses `ab` (ApacheBench).

    **Installing `ab`:**
    - **macOS (via Homebrew):**
      ```sh
      brew install httpd
      ```
    - **Debian/Ubuntu:**
      ```sh
      sudo apt-get update && sudo apt-get install apache2-utils
      ```
    - **CentOS/RHEL:**
      ```sh
      sudo yum install httpd-tools
      ```

    Once installed, run the benchmark:
    ```sh
    make bench
    ```

5.  **Stop the services:**
    To stop and remove all running containers:
    ```sh
    make down
    ```

## Makefile Commands

- `make up`: Builds and starts all services using Docker Compose.
- `make down`: Stops and removes all services.
- `make logs`: Tails the logs from all running services.
- `make bench`: Runs a benchmark test using `ab` (ApacheBench).
