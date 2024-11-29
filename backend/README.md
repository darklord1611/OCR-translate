Install [tesseract](https://tesseract-ocr.github.io/tessdoc/Installation.html)

Install [go](https://go.dev/doc/install)

Install [locust](https://locust.io/)


Create .env file from .env.example, setup AWS credentials

```sh
# Cài đặt các gói liên quan
$ go mod tidy
```

## V1: Sync Request
```sh
$ go run main_sync.go
```

## V2: Async + Message Queue
```sh
$ docker compose up -d redis rabbitmq

$ go run main_async.go

# New terminal
$ source start_multiple_ocr_worker.sh ${number_of_workers}

# New terminal
$ source start_multiple_translate_worker.sh ${number_of_workers}

```

```sh
# Benchmark
$ pip install locust

$ cd benchmark
$ locust

# Go to http://localhost:8089
# Pick number of max users, ramp-up user rate, host, load time
```

