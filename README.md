# Way
HTTP router for Go 1.7

* Deliberately simple
* Extremely fast
* Route based on HTTP methods and path only
* Path parameters via Context
* Trailing / matches path prefixes

## Install

There's no need to add a dependency to Way, just copy and paste the appropriate files into your project, or [drop](https://github.com/matryer/drop) them in:

```
drop github.com/matryer/way
```

## Usage

```go
func main() {
	router := NewWayRouter()
	router.HandleFunc("GET", "/music/:band/:song", handleSong)
	http.Handle("/", router)
	log.Fatalln(http.ListenAndServe(":8080", nil))
}

func handleSong(w http.ResponseWriter, r *http.Request) {
	
}
```