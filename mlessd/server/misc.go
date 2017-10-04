package server

import (
	"encoding/json"
	"net/http"

	"github.com/inconshreveable/log15"
)

func jsonHandler(f func(*http.Request) (interface{}, error)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		res, err := f(r)
		if err != nil {
			// log15.Error("Returning error", "err", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if res == nil {
			return
		}

		if err := json.NewEncoder(w).Encode(res); err != nil {
			log15.Error("failed encoding answer", "err", err, "res", res)
		}
	})
}
