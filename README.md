# API

## Get

```Bash
    go get -u github.com/lordtor/go-base-api/api
```

## Configuration

|Parameter|Required|Type|Default| Description|
|---|---|---|---| --- |
| ListenPort      | int         | | 8080 | Service port |
| WriteTimeout    | int         | | 30 | WriteTimeout is a time limit imposed on client connecting to the server via http from the time the server has completed reading the request header up to the time it has finished writing the response. |
| ReadTimeout     | int         | | 30 | ReadTimeout is a timing constraint on the client http request imposed by the server from the moment of initial connection up to the time the entire request body has been read. |
| GracefulTimeout | int         | | 15 | Shutdown gracefully shuts down the server without interrupting any active connections. Shutdown works by first closing all open listeners, then closing all idle connections, and then waiting indefinitely for connections to return to idle and then shut down. If the provided context expires before the shutdown is complete, Shutdown returns the context’s error, otherwise it returns any error returned from closing the Server’s underlying Listener(s). |
| IdleTimeout     | int         | | 60 | This timeout is also applicable to a connection pool. Idle Connection Timeout specifies how much time an unused connection should be kept around. |
| Swagger         | bool        | | false | Enable swagger |
| Prometheus      | bool        | | false | Enable metrics Prometheus |
| LocalSwagger    | bool        | | false | Use for test swagger on localhost or local IP (dev mode) |
| Schema          | string      | | http | base schema |
| App             | string      | * | nil | Service name |
| Host            | string      | * | nil | Service host name |
| ApiHost         | string      | | nil | If not set auto value |
| AllowedOrigins  | []string    | | * | Set allowed origins (CORS) |
| AllowedHeaders  | []string    | | "X-Requested-With", "Content-Type", "Authorization", "SERVICE-AGENT", "Access-Control-Allow-Methods", "Date", "X-FORWARDED-FOR", "Accept", "Content-Length", "Accept-Encoding", "Service-Agent" | Set allowed headers (CORS) |
| AllowedMethods  | []string    | | "GET", "POST", "PUT", "PATCH", "HEAD", "OPTIONS" | Set allowed methods (CORS) |
| AppConfig       | interface{} | * | nil | Main config for show by method `/env` use json `json:"-"` anotation for secret data. |

## Example

```Bash
    cd cmd
    go run main.go
```

### Update swagger doc

* Get swag

```Bash
    go install github.com/swaggo/swag/cmd/swag@latest
```

* Documentation update:

```Bash
    swag init  --parseDependency --generatedTime --parseInternal  --parseDepth 5
```
