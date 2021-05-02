package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"

	Log "github.com/sirupsen/logrus"
)

func main() {
	Log.SetLevel(Log.DebugLevel)

	port, exists := os.LookupEnv("FUNCTIONS_CUSTOMHANDLER_PORT")
	if !exists {
		port = "8080"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", Greet)
	mux.HandleFunc("/greet", Greet)

	Log.Infof("About to listen on %s. Go to https://127.0.0.1:%s/", port, port)

	Log.Fatal(http.ListenAndServe(":"+port, mux))
}

func Greet(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method == "GET" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		err := json.NewEncoder(w).Encode("{Blip, blop, I'm a Twitter bot.}")
		if err != nil {
			Log.Errorf("could write greet response to GET | %v", err)
		}
	} else {
		w.WriteHeader(http.StatusOK)

		body, _ := ioutil.ReadAll(r.Body)
		n, err := w.Write(body)

		if n == 0 || err != nil {
			Log.Errorf("greet error: %d %v", n, err)
		}
	}
}
