Install [tesseract](https://tesseract-ocr.github.io/tessdoc/Installation.html)

Install [go](https://go.dev/doc/install)

Install [locust](https://locust.io/)

```sh
# Cài đặt các gói liên quan
$ go mod tidy

# Khởi chạy ứng dụng demo
$ go run main.go
```

```sh
# Benchmark
$ pip install locust

$ cd benchmark
$ locust

# Go to http://localhost:8089
# Pick number of max users, ramp-up user rate, host, load time
```

