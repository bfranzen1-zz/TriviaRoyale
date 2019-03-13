package handlers

import "net/http"

/* TODO: implement a CORS middleware handler, as described
in https://drstearns.github.io/tutorials/cors/ that responds
with the following headers to all requests:

  Access-Control-Allow-Origin: *
  Access-Control-Allow-Methods: GET, PUT, POST, PATCH, DELETE
  Access-Control-Allow-Headers: Content-Type, Authorization
  Access-Control-Expose-Headers: Authorization
  Access-Control-Max-Age: 600
*/

//CorsHeader is a middleware handler that adds headers to satisfy
//CORS constraints
type CorsHeader struct {
	handler http.Handler
}

//ServeHTTP handles the request by passing it to the real
//handler and adding the necessary CORS headers
func (l *CorsHeader) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	//add CORS headers to response
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, PUT, POST, PATCH, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Access-Control-Expose-Headers", "Authorization")
	w.Header().Set("Access-Control-Max-Age", "600")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
	}
	//call wrapped handler
	l.handler.ServeHTTP(w, r)
}

//NewCORS constructs a new CORS Header middleware handler
func NewCORS(handlerToWrap http.Handler) *CorsHeader {
	return &CorsHeader{handlerToWrap}
}
