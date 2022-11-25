# Weather API

Originally one of my past technical test submissions, this repo has since grown 
as I continue to experiment in application development when using Go.

The original test criteria were something along the lines of as follows:
- Create an HTTP service that reports on Sydney weather, using providers Openweather & Weatherstack as a source for the data.
- Sydney should be hardcoded.
- Regardless of which provider is queried, a unified response is expected.
- The service should be able to silently failover between providers (should one go down) without impacting the end user.
- Weather results should be cached for up to 3 seconds or served stale if all providers are down.
- Scalability & reliability are important.
- Document tradeoffs & future improvements.

## Requires
* Dependencies outlined in [./build/docker/Dockerfile](./build/docker/Dockerfile).
* Docker (tested with 20.10.12).

## Setup
Run `make env` and then see [.env](.env) for further instruction.

## Usage
```
make help
```

Make targets can be invoked on your host machine as normal or in a container. See
[./build/docker/README.md](./build/docker/README.md).

## Layout
    .
    ├── cmd                     
    │   └── weatherapi          # Application entrypoint.
    └── build                   # Scripts used in the build pipeline.

## Testing Manually

1. Start the application:
```bash
make run
```

2. Hit the endpoint:
```bash
curl "http://localhost:8080/v1/weather?city=Sydney"
 ```

## Trade-offs / What was left out / What I'd do different

### ENDPOINT_URLS
For both provider endpoints the default scheme used is http. This isn't ideal given API keys are exchanged but it is easier for the sake of testing (e.g Weatherstack requires a paid subscription to use TLS).

### Distributed Result Caching
The [cache implementation](internal/memorycache/memorycache.go) is a simple in-memory key/value map. This is not a suitable option for an application that needs to scale (as each process will have it's own cache). With more time it might be worth looking at leveraging something like Redis or [groupcache](https://pkg.go.dev/github.com/golang/groupcache#pkg-overview).

### Limit Result Caching
With the current way the [results are cached](internal/providerquery/queryer.go), they will be served indefinitely when all providers are down. It might be worth limiting how long stale results are served for.

### Provider Queryer
The [Provider Queryer](internal/providerquery/queryer.go) has a very simple failover mechanism - try providers one at a time. In a scenario where the first provider goes down the queryer still tries that provider first (adding the timeout time to each user request). Ideally the queryer would remember the last successful provider and query from that first.

### Robust Provider Integration
The provider implementations ([Weatherstack](internal/weatherstack/current.go) & [Openweather](internal/openweather/weather.go)) are quite simple. It'd be worth investing time into more thorough integrations. For example, Weatherstack will return a status code 200 (OK) even for non-successful requests. The current integration will assume success on 200, deserialize to the successful response without error and return it with all values zero-valued (0 temperature, 0 wind speed).

### Type Assertion Checks
There are a few type assertions that don't check the value of the second return value ('ok'). Despite these never possibly panicking in this project, they should be guarded against panic to ensure robustness (and not to mention consistency).

### More Configuration
There is still room for more application configuration (as opposed to hardcoding values).

### Richer Error Responses
Any low level errors encountered in the [weather handler](cmd/weatherapi/handlers/weather.go) are returned as internal errors (500). Ideally, if there is an opportunity to, more expressive errors could be returned.
