mux
===
[![Build Status](https://travis-ci.org/gorilla/mux.png?branch=master)](https://travis-ci.org/gorilla/mux)

gorilla/mux is a powerful URL router and dispatcher.

Read the full documentation here: http://www.gorillatoolkit.org/pkg/mux

Experimental changes to gorilla mux to use a non standard http.handler. The idea is to not use the gorilla.context at all but pass data via a context parameter.

This fork is updated to use a nonstandard handler. 

Example
```go
package main

import (
  "fmt"
  "net/http"
  "github.com/zdebeer99/mux"
)

func main() {
  r := mux.NewRouter()
  r.HandleFunc("/{id}", func(c *mux.Context) {
    fmt.Fprintf(c.Response, "Hello %v", c.Vars)
  })
  http.ListenAndServe(":3001", r)
}
```

File Server Example:

```go
  r := mux.NewRouter()
  //Serve files from the static folder in your project folder and access Ex: http://localhost/public/jquery.js
  r.FileServer("/public/","static")
```

Custom Context Example:
```go
  r := mux.NewRouter()
```
