package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/lxfontes/lunister/messages"
)

func broadcastHandler(w http.ResponseWriter, r *http.Request) {
	var req messages.ARPRequest

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), 422)
		return
	}
	defer r.Body.Close()

	err = json.Unmarshal(data, &req)
	if err != nil {
		http.Error(w, err.Error(), 422)
		return
	}

	fmt.Println("Looking for", req.DestinationIP)

	var resp messages.ARPResponse

	if req.DestinationIP == "10.0.0.1" {
		resp.DestinationAddress = "a6:9b:5f:b1:06:25"
	}

	if req.DestinationIP == "10.0.0.2" {
		resp.DestinationAddress = "9a:57:5f:50:e5:38"
	}

	b, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.Write(b)
}

func main() {
	http.HandleFunc("/broadcast", broadcastHandler)
	http.ListenAndServe(":8080", nil)
}
