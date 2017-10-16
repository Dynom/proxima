# Introduction
Proxima is a small reverse-proxy intended to guard the popular [Imaginary](https://github.com/h2non/imaginary) service. It's intended as a public facing frontend for Imaginary.


```
$ ./proxima -h
Usage of ./proxima:
  -allow-hosts value
        Repeatable flag (or a comma-separated list) for hosts to allow for the URL parameter (e.g. "d2dktr6aauwgqs.cloudfront.net")
  -allowed-actions value
        A comma seperated list of actions allows to be sent upstream. If empty, everything is allowed.
  -allowed-params value
        A comma seperated list of parameters allows to be sent upstream. If empty, everything is allowed.
  -bucket-rate float
        Rate limiter bucket fill rate (req/s) (default 20)
  -bucket-size int
        Rate limiter bucket size (burst capacity) (default 500)
  -imaginary-url string
        URL to imaginary (default: http://localhost:9000) (default "http://localhost:9000")
  -listen-port int
        Port to listen on (default 8080)
  -root-path-strip string
        A section of the (left most) path to strip (e.g.: "/static"). Start with a /.
```


# Example
Start imaginary on a local interface
```bash
./imaginary -a 127.0.0.1 -p 9001 -enable-url-source
```

Start Proxima on the public interface
```bash
./proxima \
    --allowed-actions="info,crop" \
    --imaginary-url="http://127.0.0.1:9001" \
    --listen-port=9000
```

And visit http://localhost:9000/ as if it were Imaginary, e.g.:
```bash
curl \
    --data-binary @pancake.jpg \
    "http://localhost:9000/info" | jq .

{
  "width": 1400,
  "height": 919,
  "type": "jpeg",
  "space": "srgb",
  "hasAlpha": false,
  "hasProfile": false,
  "channels": 3,
  "orientation": 0
}
```

And since we only defined the `info` and `crop` actions, *enlarge* is not allowed:
```bash
curl --data-binary @pancake.jpg \
    "http://localhost:9000/enlarge?width=10000&height=10000"

Unregisterd action
```
