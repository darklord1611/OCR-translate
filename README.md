# OCR-Translate

This repository contains the implementation of a distributed OCR (Optical Character Recognition) and translation system. The project is designed to process images containing text, extract the text using OCR, translate it into another language (currently configured for English to English), and generate a PDF with the translated text. The system is built with scalability, modularity, and performance in mind, leveraging modern software architecture principles and design patterns. It supports both local file storage and AWS S3 for image and PDF handling, and can run in synchronous or asynchronous modes with RabbitMQ and Redis.

---

## Repository Structure

The repository is organized into two main directories: `backend` and `frontend`.

### Frontend

The frontend is implemented using Nuxt 3 (Vue.js framework) and provides a user-friendly interface for uploading images, monitoring job statuses, and downloading the resulting PDFs.

### Backend

The backend is implemented in Go and is responsible for handling the OCR, translation, and PDF generation tasks. It is designed as a distributed system with support for asynchronous processing using RabbitMQ for message queuing and Redis for job status tracking and caching.


#### Key Features:

- **Image Upload**: Users can upload one or more images (up to 10 per batch) for OCR and translation.
- **Job Monitoring**: Polls the backend for real-time status updates of submitted jobs.
- **PDF Preview and Download**: Displays generated PDFs in iframes and allows users to download them.
- **Responsive Design**: Styled with Tailwind CSS and DaisyUI for a modern and responsive user experience.
- **Client-Side Rate Limiting**: Implements a basic rate limit for conversion requests in the UI ([`frontend/pages/preview.vue`](frontend/pages/preview.vue)).
- **Configuration**: Uses `.env` file for backend URL configuration.

---

## Design Patterns

The project employs several design patterns to enhance its architecture:

1.  **Producer-Consumer Pattern**:
    *   The backend API acts as a producer, publishing job messages (e.g., OCR tasks, translation tasks) to RabbitMQ queues (`ocr-queue`, `translation-queue`).
    *   Dedicated worker processes ([`ocr_worker.go`](backend/ocr_worker.go), [`translate_worker.go`](backend/translate_worker.go), etc.) act as consumers, processing these messages asynchronously.

2.  **Pipes-and-Filters Pattern**:
    *   The overall process follows a pipeline: Image Upload -> (Optional Caching Check) -> OCR Task -> Translation Task -> PDF Generation -> Result Storage. Each stage processes data and passes it to the next.

3.  **Claim-Check Pattern**:
    *  We implement claim-check pattern, storing the actual files at S3 buckets instead of embedding the files directly in the message to the workers. This would reduce the message size and potentially benefit worker's processing time.

4.  **Rate Limiting**:
    *   A server-side rate limiting middleware is available in [`backend/middleware/rate_limiter.go`](backend/middleware/rate_limiter.go), using a token bucket algorithm (via `golang.org/x/time/rate`) to control request frequency per client IP.
    *   The frontend also implements a simple client-side rate limit for initiating conversions in [`frontend/pages/preview.vue`](frontend/pages/preview.vue).

5.  **Cache-Aside Strategy**:
    *   The backend (async modes) uses the SHA256 hash of the uploaded image file as a `jobID`. Before processing a new upload, it checks Redis using this hash. If a completed job with the same hash exists, it can return the existing result, avoiding redundant processing ([`backend/main_async_single.go`](backend/main_async_single.go), [`backend/main_async_cluster.go`](backend/main_async_cluster.go)).

---

## Getting Started

### Prerequisites

*   [Go (version 1.23.2 or higher)](https://go.dev/doc/install)
*   [Node.js (LTS version recommended)](https://nodejs.org/)
*   [Docker and Docker Compose](https://www.docker.com/get-started)
*   [Tesseract OCR](https://tesseract-ocr.github.io/tessdoc/Installation.html) (ensure `tesseract` command is in PATH for backend)

### Backend Setup

1.  **Navigate to the backend directory:**
    ```sh
    cd backend
    ```
2.  **Install Go dependencies:**
    ```sh
    go mod tidy
    ```
3.  **Configure Environment Variables:**
    *   Copy [`.env.example`](backend/.env.example) to `.env`: `cp .env.example .env`
    *   Edit `.env` with your settings for RabbitMQ, Redis, AWS (if using S3 storage), and default port.

4.  **Running the Backend:**

    *   **Synchronous Mode:**
        ```sh
        go run main_sync.go --port 8081
        ```

    *   **Asynchronous Mode (Single Redis & Nginx):**
        1.  Start services:
            ```sh
            docker-compose -f docker-compose-single-redis.yml up -d
            ```
            This starts RabbitMQ, a single Redis instance, and Nginx (listening on port 8090 by default).
        2.  Start the API server (choose one, e.g., on port 8081):
            ```sh
            go run main_async_single.go --port 8081 --storage local # or --storage s3
            ```
            (If Nginx is configured for multiple backend instances, you can run another on port 8082)
        3.  Start OCR workers (in new terminals):
            ```sh
            bash start_multiple_ocr_worker.sh <number_of_ocr_workers>
            # or for segmentation
            # bash start_multiple_ocr_segment_worker.sh <number_of_ocr_segment_workers>
            ```
        4.  Start Translation workers (in new terminals):
            ```sh
            bash start_multiple_translate_worker.sh <number_of_translation_workers>
            ```

    *   **Asynchronous Mode (Redis Cluster):**
        1.  Start services:
            ```sh
            docker-compose -f docker-compose-redis-cluster.yml up -d
            ```
            This starts RabbitMQ and a Redis cluster.
        2.  Start the API server:
            ```sh
            go run main_async_cluster.go --port 8081 --storage local # or --storage s3
            ```
        3.  Start OCR and Translation workers as described above (translation workers should be [translate_worker_cluster.go](http://_vscodecontentref_/24) if specific logic is needed, though current [start_multiple_translate_worker.sh](http://_vscodecontentref_/25) runs [translate_worker.go](http://_vscodecontentref_/26)).

### Frontend Setup

1.  **Navigate to the frontend directory:**
    ```sh
    cd frontend
    ```
2.  **Install Node.js dependencies:**
    ```sh
    npm install
    ```
3.  **Configure Environment Variables:**
    *   Copy [.env.example](http://_vscodecontentref_/27) to `.env`: `cp [.env.example](http://_vscodecontentref_/28) .env`
    *   Edit `.env` and set `VITE_BACKEND_URL` to your backend API endpoint (e.g., [http://localhost:8090](http://_vscodecontentref_/29) if using Nginx, or [http://localhost:8081](http://_vscodecontentref_/30) if direct).

4.  **Start the development server:**
    ```sh
    npm run dev
    ```
    The frontend will be accessible at [http://localhost:3000](http://_vscodecontentref_/31) by default.

---

## Testing and Benchmarking

*   **Load Testing**: The backend includes Locust scripts for load testing in the benchmark directory.

    ```sh
    cd backend/benchmark
    locust -f async_load_test.py # or sync_load_test.py
    ```
    
    Access the Locust UI at `http://localhost:8089`.
