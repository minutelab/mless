package server

import (
	"fmt"
	"net/http"

	"github.com/minutelab/mless/formation"
)

var functions *formation.Functions

// Start the mlessd http server using the specified formation and port
func Start(port int, funcs *formation.Functions) error {
	functions = funcs

	srvr := http.NewServeMux()
	srvr.Handle("/invoke", jsonHandler(invokeHandler))

	return http.ListenAndServe(fmt.Sprintf(":%d", port), srvr)
}
