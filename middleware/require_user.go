package middleware

import (
	"fmt"
	"net/http"

	"github.com/WawinyEdwin/lenslocked.com/context"
	"github.com/WawinyEdwin/lenslocked.com/models"
)

type RequireUser struct {
	models.UserService
}

//return an http.HandleFunc that will check to see if a user is logged in and then either call next(w, r) if they are, or redirect them to the login page if they are not
func (mw *RequireUser) ApplyFn(next http.HandlerFunc) http.HandlerFunc {
	//we want to return dynamically created
	//func (http.ResponseWriter, *http.Request)
	//but we also need convert it into an http.Handlefunc.
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//Check if a user is logged in.
		cookie, err := r.Cookie("remember_token")
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusFound)
		}
		user, err := mw.UserService.ByRemember(cookie.Value)
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		fmt.Println("User found: ", user)
		//get the context from our request
		ctx := r.Context()
		//create a new context from the existing one that has our user stored in it with the private user key
		ctx = context.WithUser(ctx, user)
		//create a new request from the existing one with our context attached to it and assign it back to 'r
		r = r.WithContext(ctx)
		//call next(w,r) with our updated context
		next(w, r)
	})
}

//Applyng our middleware to http.Handler interfaces
func (mw *RequireUser) Apply(next http.Handler) http.HandlerFunc {
	return mw.ApplyFn(next.ServeHTTP)
}

// func someHandler(w http.ResponseWriter, r *http.Request) {
// 	ctx := r.Context()
// 	newCtx := context.WithValue(ctx, "my-key", "my-value")
// 	myKey := newCtx.Value("my-key")
// 	if myKeyStr, ok := myKey.(string); !ok {

// 	}
// }
