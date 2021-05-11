package backend

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

type App struct {
	Router *mux.Router
	Model  *Model
}

func NewApp(model *Model) *App {
	a := &App{
		Router: mux.NewRouter(),
		Model:  model,
	}
	a.initializeRoutes()
	return a
}

func (a *App) initializeRoutes() {
	a.Router.HandleFunc("/vitals", a.vitals).Methods("GET")
	a.Router.HandleFunc("/stations", a.stations).Methods("GET")
	a.Router.HandleFunc("/journey_query", a.journeyQuery).Methods("GET")
}

func (a *App) Run(addr string) {
	// TODO: pick a better restriction. For now, allow anyone.
	originsOk := handlers.AllowedOrigins([]string{"*"})

	log.Fatal(http.ListenAndServe(addr, handlers.CORS(originsOk)(a.Router)))
}

func (a *App) vitals(w http.ResponseWriter, _ *http.Request) {
	v, err := a.Model.Vitals()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondWithJSON(w, http.StatusOK, v)
}

func (a *App) stations(w http.ResponseWriter, _ *http.Request) {
	v, err := a.Model.GetStations()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondWithJSON(w, http.StatusOK, v)
}

func (a *App) journeyQuery(w http.ResponseWriter, r *http.Request) {
	var src, dst Circle
	var err error

	if src.Center.Lat, err = strconv.ParseFloat(r.FormValue("src_lat"), 64); err != nil {
		respondWithError(w, http.StatusBadRequest, fmt.Sprintf("Invalid src_lat: %v", err))
		return
	}
	if src.Center.Long, err = strconv.ParseFloat(r.FormValue("src_long"), 64); err != nil {
		respondWithError(w, http.StatusBadRequest, fmt.Sprintf("Invalid src_long: %v", err))
		return
	}
	if src.RadiusKm, err = strconv.ParseFloat(r.FormValue("src_radius"), 64); err != nil {
		respondWithError(w, http.StatusBadRequest, fmt.Sprintf("Invalid src_radius: %v", err))
		return
	}
	if dst.Center.Lat, err = strconv.ParseFloat(r.FormValue("dst_lat"), 64); err != nil {
		respondWithError(w, http.StatusBadRequest, fmt.Sprintf("Invalid dst_lat: %v", err))
		return
	}
	if dst.Center.Long, err = strconv.ParseFloat(r.FormValue("dst_long"), 64); err != nil {
		respondWithError(w, http.StatusBadRequest, fmt.Sprintf("Invalid dst_long: %v", err))
		return
	}
	if dst.RadiusKm, err = strconv.ParseFloat(r.FormValue("dst_radius"), 64); err != nil {
		respondWithError(w, http.StatusBadRequest, fmt.Sprintf("Invalid dst_radius: %v", err))
		return
	}

	v, err := a.Model.JourneyQuery(src, dst)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, v)
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}
