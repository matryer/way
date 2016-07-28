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
	router := NewRouter()
	router.HandleFunc("GET", "/music/:band/:song", handleSong)
	http.Handle("/", router)
	log.Fatalln(http.ListenAndServe(":8080", nil))
}

func handleSong(w http.ResponseWriter, r *http.Request) {
	band, ok := Param(r.Context(), "band")
	if !ok {
		http.Error(w, "must provide band", http.StatusBadRequest)
		return
	}
	song, ok := Param(r.Context(), "song")
	if !ok {
		http.Error(w, "must provide song", http.StatusBadRequest)
		return
	}
	// use 'band' and 'song' parameters...
}
```
