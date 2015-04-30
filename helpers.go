package mux

import (
	"net/http"
	"path/filepath"
)

// FileServer helper method for serving files from a folder.
//
// Example use: serve files from the static folder in your project. relative to the executable. under the public address.
//
// r.FileServer("/public/","static")
//
// http://localhost/public/jquery.js
func (r *Router) FileServer(path_prefix string, file_path string) {
	f, _ := filepath.Abs(file_path)
	r.PathPrefix(path_prefix).HandlerHttp(http.StripPrefix(path_prefix, http.FileServer(http.Dir(f))))
}
