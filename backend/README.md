Install [tesseract](https://tesseract-ocr.github.io/tessdoc/Installation.html)

Install [go](https://go.dev/doc/install)

Install [locust](https://locust.io/)

```sh
# Cài đặt các gói liên quan
$ go mod tidy
```

## V1: Single Queue + Split-Image
```sh
$ go run main.go
```

## V2: Multiple Queues
```sh
$ go run main_2.go
```

## V3: Multiple Queues(RabbitMQ)
Create .env file from .env.example
```sh
$ docker-compose up

$ go run main_3.go

# New terminal
$ go run ocr_worker.go

# New terminal
$ go run translate_worker.go
```


```sh
# Benchmark
$ pip install locust

$ cd benchmark
$ locust

# Go to http://localhost:8089
# Pick number of max users, ramp-up user rate, host, load time
```

