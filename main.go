package main

import (
	"fmt"
	"net/http"

	"github.com/WawinyEdwin/lenslocked.com/controllers"
	"github.com/WawinyEdwin/lenslocked.com/middleware"
	"github.com/WawinyEdwin/lenslocked.com/models"
	"github.com/WawinyEdwin/lenslocked.com/views"
	"github.com/gorilla/mux"
)

//view global vars
var (
	homeView    *views.View
	contactView *views.View
	faqView     *views.View
)

//database info
const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "user"
	dbname   = "lenslocked_dev"
)

//Home handler
func home(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	homeView.Render(w, nil)
}

//contact handler
func contact(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	contactView.Render(w, nil)
}

//Faq handler
func faq(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	faqView.Render(w, nil)
}

//404
func notfound(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, "<h1>404 Not Found </h1>")
}

var h http.Handler = http.HandlerFunc(notfound)

func main() {
	fmt.Println("Server started successfully...")

	//create a db conn and use it to create our model service
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+"password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)

	//create a user service using the connection string
	services, err := models.NewServices(psqlInfo)
	if err != nil {
		panic(err)
	}
	defer services.Close()
	services.AutoMigrate()
	// us.DestructiveReset()

	r := mux.NewRouter()
	//our controllers
	staticC := controllers.NewStatic()
	usersC := controllers.NewUsers(services.User)
	galleriesC := controllers.NewGalleries(services.Gallery, r)

	requireUserMw := middleware.RequireUser{
		UserService: services.User,
	}

	newGallery := requireUserMw.Apply(galleriesC.New)
	createGallery := requireUserMw.ApplyFn(galleriesC.Create)

	r.HandleFunc("/", staticC.Home.ServeHTTP).Methods("GET")
	r.HandleFunc("/contact", staticC.Contact.ServeHTTP).Methods("GET")
	r.HandleFunc("/faq", staticC.Faq.ServeHTTP).Methods("GET")

	r.HandleFunc("/signup", usersC.New).Methods("GET")
	r.HandleFunc("/signup", usersC.Create).Methods("POST")

	r.Handle("/login", usersC.LoginView).Methods("GET")
	r.HandleFunc("/login", usersC.Login).Methods("POST")
	r.HandleFunc("/cookietest", usersC.CookieTest).Methods("GET")

	r.Handle("/galleries/new", newGallery).Methods("GET")
	r.Handle("/galleries", createGallery).Methods("POST")
	r.HandleFunc("/galleries/{id:[0-9]+}", galleriesC.Show).Methods("GET").Name(controllers.ShowGallery)
	r.HandleFunc("/galleries/{id:[0-9]+}/edit", requireUserMw.ApplyFn(galleriesC.Edit)).Methods("GET")
	r.HandleFunc("/galleries/{id:[0-9]+}/update", requireUserMw.ApplyFn(galleriesC.Update)).Methods("POST")
	r.NotFoundHandler = h

	http.ListenAndServe(":9090", r)
}
