//go:build !(cgo && (darwin || linux))

// Yerel pencere yalnızca macOS / Linux'ta CGO ile derlenir; diğer derlemelerde boş program.
package main

func main() {}
