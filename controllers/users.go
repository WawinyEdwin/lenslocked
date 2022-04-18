package controllers

import (
	"fmt"
	"net/http"

	"github.com/WawinyEdwin/lenslocked.com/models"
	"github.com/WawinyEdwin/lenslocked.com/rand"
	"github.com/WawinyEdwin/lenslocked.com/views"
)

func NewUsers(us models.UserService) *Users {
	return &Users{
		NewView:   views.NewView("bootstrap", "users/new"),
		LoginView: views.NewView("bootstrap", "users/login"),
		us:        us,
	}
}

//users
type Users struct {
	NewView   *views.View
	LoginView *views.View
	us        models.UserService
}

//form for signing up
type signupForm struct {
	Name     string `schema: "name"`
	Email    string `schema: "email"`
	Password string `schema: "password"`
}

//form for login
type LoginForm struct {
	Email    string `schema: "email"`
	Password string `schema: "password`
}

//New is used to render the form where a user can
//GET /signup
func (u *Users) New(w http.ResponseWriter, r *http.Request) {
	u.NewView.Render(w, nil)
}

func (u *Users) Create(w http.ResponseWriter, r *http.Request) {
	var vd views.Data
	var form signupForm
	if err := ParseForm(r, &form); err != nil {
		vd.SetAlert(err)
		u.NewView.Render(w, vd)
		return
	}
	user := models.User{
		Name:     form.Name,
		Email:    form.Email,
		Password: form.Password,
	}
	if err := u.us.Create(&user); err != nil {
		vd.SetAlert(err)
		u.NewView.Render(w, vd)
		return
	}
	err := u.signIn(w, &user)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	//redirect to the cookie test page to test the cookie
	http.Redirect(w, r, "/cookietest", http.StatusNotFound)
}

//used to process login form when a user tries to login an existing user. //POST /login
func (u *Users) Login(w http.ResponseWriter, r *http.Request) {
	var vd views.Data
	form := LoginForm{}
	if err := ParseForm(r, &form); err != nil {
		vd.SetAlert(err)
		u.LoginView.Render(w, vd)
	}
	user, err := u.us.Authenticate(form.Email, form.Password)
	if err != nil {
		switch err {
		case models.ErrNotFound:
			vd.AlertError("No user exists with that email address")
		default:
			vd.SetAlert(err)
		}
		u.LoginView.Render(w, vd)
		return
	}
	err = u.signIn(w, user)
	if err != nil {
		vd.SetAlert(err)
		u.LoginView.Render(w, vd)
		return
	}
	http.Redirect(w, r, "/cookietest", http.StatusFound)

}

//signin is used to sign the given user in via cookie
func (u *Users) signIn(w http.ResponseWriter, user *models.User) error {
	//Implement this
	if user.Remember == "" {
		token, err := rand.RememberToken()
		if err != nil {
			return err
		}
		user.Remember = token
		err = u.us.Update(user)
		if err != nil {
			return err
		}
	}

	cookie := http.Cookie{
		Name:     "remember_token",
		Value:    user.Remember,
		HttpOnly: true,
	}
	http.SetCookie(w, &cookie)
	return nil
}

//viewing cookies with Go Code
func (u *Users) CookieTest(w http.ResponseWriter, r *http.Request) {
	//implementation
	cookie, err := r.Cookie("remember_token")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	user, err := u.us.ByRemember(cookie.Value)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintln(w, user)
}
