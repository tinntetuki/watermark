# Image Watermark Service

This is a Go-based microservice that dynamically adds a text watermark to images retrieved from an object storage backend (like AWS S3 or Cloudflare R2). It includes a configurable caching layer (Redis or local file system) to optimize performance and reduce redundant processing.

## Features

- **Dynamic Watermarking**: Adds text watermarks to images on-the-fly.
- **Configurable Storage**: Supports AWS S3 and S3-compatible services like Cloudflare R2.
- **Configurable Caching**: Choose between Redis or a local file system for caching processed images.
- **Metrics-Driven**: Exposes Prometheus metrics for monitoring and performance analysis (`/metrics` endpoint).
- **Graceful Shutdown**: Ensures the server shuts down cleanly, finishing in-flight requests.
- **Containerized**: Comes with a `Dockerfile` and `docker-compose.yml` for easy deployment.

## How It Works

1.  A client requests an image using a URL like `GET /image/my-image.jpg?text=Hello+World`.
2.  The service generates a unique cache key based on the image name and watermark text.
3.  It first checks the configured cache (Redis or local) for a processed image.
    -   **Cache Hit**: If found, the image is served directly from the cache.
    -   **Cache Miss**: If not found, the service proceeds to the next step.
4.  The original image (`my-image.jpg`) is downloaded from the configured object storage (e.g., S3).
5.  The service uses the `freetype` library to draw the requested text ("Hello World") onto the image.
6.  The newly watermarked image is then stored in the cache for future requests.
7.  The final image is sent back to the client.

## Getting Started

### Prerequisites

- Go 1.18+
- Docker & Docker Compose (for containerized setup)
- An S3-compatible object storage bucket.
- (Optional) A Redis server.

### Configuration

The service is configured via environment variables. Refer to the table below for all available options.

| Environment Variable      | Description                                                                                             | Default                  |
| ------------------------- | ------------------------------------------------------------------------------------------------------- | ------------------------ |
| `SERVER_PORT`             | Port for the HTTP server.                                                                               | `8080`                   |
| `LOG_LEVEL`               | Log level (`debug`, `info`, `warn`, `error`).                                                           | `info`                   |
| `STORAGE_PROVIDER`        | Storage backend to use. `s3` is currently the only option.                                              | `s3`                     |
| `CACHE_PROVIDER`          | Caching backend to use. Options: `redis`, `local`.                                                      | `redis`                  |
| `S3_ENDPOINT`             | The custom endpoint for S3-compatible services (e.g., Cloudflare R2).                                   | ` ` (Empty)              |
| `AWS_REGION`              | The AWS region for your S3 bucket. Set to `auto` for Cloudflare R2.                                     | `us-east-1`              |
| `S3_BUCKET`               | **Required**. The name of your S3 bucket.                                                               | ` ` (Empty)              |
| `S3_PREFIX`               | A prefix to be prepended to all image keys when fetching from S3.                                       | `qc-images/`             |
| `S3_ACCESS_KEY_ID`        | **Required**. Your S3 access key.                                                                       | ` ` (Empty)              |
| `S3_SECRET_ACCESS_KEY`    | **Required**. Your S3 secret key.                                                                       | ` ` (Empty)              |
| `REDIS_URL`               | The full URL for connecting to Redis (e.g., `redis://user:pass@host:port/db`). Overrides other REDIS vars. | ` ` (Empty)              |
| `REDIS_ADDR`              | Redis server address.                                                                                   | `localhost:6379`         |
| `REDIS_PASSWORD`          | Redis password.                                                                                         | ` ` (Empty)              |
| `REDIS_DB`                | Redis database number.                                                                                  | `0`                      |
| `LOCAL_CACHE_PATH`        | The directory path for the local file cache if `CACHE_PROVIDER=local`.                                  | `./cache`                |
| `CACHE_TTL`               | Cache Time-To-Live for processed images.                                                                | `168h` (7 days)          |
| `FONT_PATH`               | Path to the `.ttf` font file to be used for watermarks.                                                 | `./fonts/Arial.ttf`      |
| `FONT_SIZE`               | Font size for the watermark text.                                                                       | `24.0`                   |
| `WATERMARK_COLOR`         | Watermark color in hex format (e.g., `#FFFFFF`).                                                        | `#FFFFFF`                |
| `IMAGE_QUALITY`           | The quality of the output JPEG image (1-100).                                                           | `90`                     |

### Running Locally

1.  **Clone the repository:**

    ```bash
    git clone <repository-url>
    cd watermark-service
    ```

2.  **Set up your environment:**

    Create a `.env` file in the root of the project and fill it with your configuration:

    ```env
    # Server
    SERVER_PORT=8080
    LOG_LEVEL=debug

    # Storage (Cloudflare R2 Example)
    STORAGE_PROVIDER=s3
    S3_ENDPOINT="https://<your-account-id>.r2.cloudflarestorage.com"
    AWS_REGION="auto"
    S3_BUCKET="your-r2-bucket-name"
    S3_ACCESS_KEY_ID="your-r2-access-key-id"
    S3_SECRET_ACCESS_KEY="your-r2-secret-access-key"

    # Cache (Local File System Example)
    CACHE_PROVIDER=local
    LOCAL_CACHE_PATH=./tmp/cache
    CACHE_TTL=24h

    # Watermark Style
    FONT_PATH=./internal/assets/fonts/Arial.ttf
    FONT_SIZE=48
    WATERMARK_COLOR="#FFFFFF80" # White with 50% transparency
    IMAGE_QUALITY=85
    ```

3.  **Install dependencies:**

    ```bash
    go mod tidy
    ```

4.  **Run the service:**

    ```bash
    go run cmd/server/main.go
    ```

5.  **Test the service:**

    Open your browser or use a tool like `curl` to access an image. If you have an image named `test.jpg` in your bucket, you can access it at:

    ```
    http://localhost:8080/image/test.jpg?text=My+Watermark
    ```

### Running with Docker

1.  **Set up your environment:**

    Make sure your `docker-compose.yml` is configured correctly. You can pass environment variables directly or use an `env_file`.

2.  **Build and run the containers:**

    ```bash
    docker-compose up --build
    ```

    This will start the watermark service and a Redis container.

## Metrics

The service exposes the following Prometheus metrics at the `/metrics` endpoint:

-   `http_requests_total`: Total number of HTTP requests received.
-   `http_request_duration_seconds`: Latency of HTTP requests.
-   `image_processing_duration_seconds`: Histogram of the time it takes to add a watermark to an image (cache misses).
-   `image_cache_hits_total`: The total number of cache hits.
-   `image_cache_misses_total`: The total number of cache misses.

## Deployment

This service is designed to be deployed as a container. You can build a Docker image and run it on your favorite cloud platform (e.g., AWS ECS, Google Cloud Run, DigitalOcean App Platform).

### Build the Docker Image

```bash
docker build -t watermark-service .
```
