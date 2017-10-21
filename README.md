HTTP Caching Client
===================

Go HTTP client compatible with [http.Client](https://golang.org/pkg/net/http/#Client),
but with added response caching function.

How to use it?
==============

Just change:

```
client = &http.Client{}
```

to:

```
client = &httpcaching.Client{CacheDir: "cache"}
```

and your client will start caching HTTP responses in `cache` directory.
