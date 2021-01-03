package main
////
////import (
////	"fmt"
////	"net/http"
////
////	"github.com/go-chi/chi"
////	"github.com/go-chi/jwtauth"
////)
////
////var tokenAuth *jwtauth.JWTAuth
////
////func init() {
////	tokenAuth = jwtauth.New("HS256", []byte("nTcTNah2GR8afx0MbLsymJggmDK33BQK"), nil)
////
////	// For debugging/example purposes, we generate and print
////	// a sample jwt token with claims `user_id:123` here:
////	_, tokenString, _ := tokenAuth.Encode(
////			map[string]interface{}{
////				"scraper_name": "python_scraper",
////			},
////		)
////	fmt.Printf("DEBUG: a sample jwt is %s\n\n", tokenString)
////}
////
////func main() {
////	addr := ":3333"
////	fmt.Printf("Starting server on %v\n", addr)
////	err := http.ListenAndServe(addr, router())
////	if err != nil {
////		fmt.Println("Error: ", err)
////	}
////}
////
////func router() http.Handler {
////	r := chi.NewRouter()
////
////	// Protected routes
////	r.Group(func(r chi.Router) {
////		// Seek, verify and validate JWT tokens
////		r.Use(jwtauth.Verifier(tokenAuth))
////
////		// Handle valid / invalid tokens. In this example, we use
////		// the provided authenticator middleware, but you can write your
////		// own very easily, look at the Authenticator method in jwtauth.go
////		// and tweak it, its not scary.
////		r.Use(jwtauth.Authenticator)
////
////		r.Get("/admin", func(w http.ResponseWriter, r *http.Request) {
////			_, claims, _ := jwtauth.FromContext(r.Context())
////			w.Write([]byte(fmt.Sprintf("protected area. hi %v", claims["scraper_name"])))
////		})
////	})
////
////	// Public routes
////	r.Group(func(r chi.Router) {
////		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
////			w.Write([]byte("welcome anonymous"))
////		})
////	})
////
////	return r
////}
//

