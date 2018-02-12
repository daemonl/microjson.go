package microjson

import (
	"fmt"
	"log"
	"net/http"
	"runtime/debug"
)

// HandlerWrap just wraps the JSON req->res methods, no transaction support
func HandlerWrap(cb func(*http.Request) (interface{}, error)) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		defer func() {
			if r := recover(); r != nil {
				log.Println("Recovered Panic:", r)
				log.Println(string(debug.Stack()))
				SendError(rw, req, fmt.Errorf("Panic"))
			}
		}()

		obj, err := cb(req)
		if err != nil {
			SendError(rw, req, err)
			return
		}
		SendObject(rw, req, 200, obj)
	})
}
