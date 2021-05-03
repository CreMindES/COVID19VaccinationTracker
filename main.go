package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"

	Log "github.com/sirupsen/logrus"

	"github.com/cremindes/COVID19VaccinationTracker/covidtracker"
)

func main() {
	Log.SetLevel(Log.DebugLevel)

	// get port from Azure Functions
	port, exists := os.LookupEnv("FUNCTIONS_CUSTOMHANDLER_PORT")
	if !exists {
		port = "8080"
	}

	// init endpoints
	mux := http.NewServeMux()
	mux.HandleFunc("/", Greet)
	mux.HandleFunc("/greet", Greet)
	mux.HandleFunc("/twitterbotcovid19timertrigger", timerHandler)

	// startup info logging
	Log.Infof("About to listen on %s. Go to https://127.0.0.1:%s/", port, port)

	// start serving
	Log.Fatal(http.ListenAndServe(":"+port, mux))
}

func timerHandler(w http.ResponseWriter, r *http.Request) {
	Log.Debug("timerHandler | Timer handler called.")

	// get last vaccinated number from last tweet
	lastVaccinated, errLV := covidtracker.FetchCVNLast()
	// get current vaccinated number
	vaccinated, errV := covidtracker.FetchCVN()

	// see if there is a new one available
	if lastVaccinated >= vaccinated {
		Log.Debug("timerHandler | No new update")
		w.WriteHeader(http.StatusOK)

		return
	}

	// get population
	population, errP := covidtracker.FetchPopulation("Hungary")

	if errLV != nil {
		Log.Error(errLV)
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if errP != nil {
		Log.Error(errP)
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if errV != nil {
		Log.Error(errV)
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if population == -1 || vaccinated == -1 {
		Log.Error("timerHandler | invalid value, not tweeting.")
		w.WriteHeader(http.StatusNoContent)
		return
	}

	err := covidtracker.Tweet(vaccinated, population)

	if err != nil {
		Log.Error(err)
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	errJSONResponse := json.NewEncoder(w).Encode("{yay}")
	if errJSONResponse != nil {
		Log.Error(errJSONResponse)
	}
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
