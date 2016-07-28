# Way
HTTP router for Go 1.7

* Deliberately simple
* Extremely fast
* Route based on HTTP methods and path
* Path parameters via `Context` (e.g. `/music/:band/:song`)
* Trailing `/` matches path prefixes

## Install

There's no need to add a dependency to Way, just copy `way.go` and `way_test.go` into your project, or [drop](https://github.com/matryer/drop) them in:

```
drop github.com/matryer/way
```

If you prefer, it is go gettable:

```
go get github.com/matryer/way
```

## Usage

* Use `NewRouter` to make a new `Router`
* Call `Handle` and `HandleFunc` to add handlers
* Specify HTTP method and path pattern for each route
* Use `Param` function to get the path parameters from the context

```go
func main() {
	router := way.NewRouter()

	router.HandleFunc("GET", "/music/:band/:song", handleReadSong)
	router.HandleFunc("PUT", "/music/:band/:song", handleUpdateSong)
	router.HandleFunc("DELETE", "/music/:band/:song", handleDeleteSong)

	log.Fatalln(http.ListenAndServe(":8080", router))
}

func handleReadSong(w http.ResponseWriter, r *http.Request) {
	band := way.Param(r.Context(), "band")
	song := way.Param(r.Context(), "song")
	// use 'band' and 'song' parameters...
}
```

## Why another HTTP router?

I know, I know. But no routers offer the simplicity of path parameters via Context, and HTTP method matching. Which covers 100% of my use cases so far.
