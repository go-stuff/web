# web

[![Build Status](https://cloud.drone.io/api/badges/go-stuff/web/status.svg)](https://cloud.drone.io/go-stuff/web)
[![Go Report Card](https://goreportcard.com/badge/github.com/go-stuff/web)](https://goreportcard.com/report/github.com/go-stuff/web)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

![Gopher Share](https://github.com/go-stuff/images/blob/master/GOPHER_SHARE_640x320.png)

Web page using Gorilla web toolkit and MongoDB Driver.

# Deploy

## Environment Variables

Example of the environment variables needed to get started:

```conf
MONGOURL                 = "mongodb://localhost:27017"
MONGO_DB_NAME            = "test"
MONGOSTORE_SESSION_TTL   = "1200"
MONGOSTORE_HTTPS_ONLY    = "false"
GORILLA_SESSION_AUTH_KEY = "SuperSecret32ByteKey"
GORILLA_SESSION_ENC_KEY  = "SuperSecret16ByteKey"
LDAP_SERVER              = "LDAPSSL"
LDAP_PORT                = "636"
LDAP_BIND_DN             = "SuperSecretBindUsername"
LDAP_BIND_PASS           = "SuperSecretBindPassword"
LDAP_USER_BASE_DN        = "OU=Users,DC=go-stuff,DC=ca"
LDAP_USER_SEARCH_ATTR    = "CN"
LDAP_GROUP_BASE_DN       = "OU=Groups,DC=go-stuff,DC=ca"
LDAP_GROUP_OBJECT_CLASS  = "group"
LDAP_GROUP_SEARCH_ATTR   = "member"
LDAP_GROUP_SEARCH_FULL   = "true"
ADMIN_AD_GROUP           = "ADAdminGroup"
```

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