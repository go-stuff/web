# web

[![Build Status](https://cloud.drone.io/api/badges/go-stuff/web/status.svg)](https://cloud.drone.io/go-stuff/web)
[![Go Report Card](https://goreportcard.com/badge/github.com/go-stuff/web)](https://goreportcard.com/report/github.com/go-stuff/web)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

![Gopher Share](https://github.com/go-stuff/images/blob/master/GOPHER_SHARE_640x320.png)

Web page using Gorilla web toolkit and MongoDB Driver.

# Deploy

## Kubernetes

To deploy in Kubernetes run the following in the root dir:

```
kubectl apply -R -f deploy/
```

This will deploy an instance of mongodb along with the demo web app.

# Certs

I have included some test certs in the package to connect with `go-stuff\grpc`. You can generate them the following way:

``` bash
go run GOROOT/src/crypto/tls/generate_cert.go --host 127.0.0.1 --duration 17520h
```

## License

[MIT License](LICENSE)