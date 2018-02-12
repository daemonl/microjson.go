package microjson

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

var Version string
var Hostname string
var StartTime time.Time

func init() {
	Hostname, _ = os.Hostname()
	StartTime = time.Now()
	fmt.Printf("Version %s\n", Version)
}

func UpHandlerFunc(rw http.ResponseWriter, req *http.Request) {
	rw.Header().Add("Content-Type", "application-json")
	rw.WriteHeader(200)
	json.NewEncoder(rw).Encode(map[string]interface{}{
		"status":   "OK",
		"version":  Version,
		"booted":   StartTime.Format(time.RFC3339),
		"uptime":   time.Now().Sub(StartTime).Seconds(),
		"hostname": Hostname,
	})
}
