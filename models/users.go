package models

import (
	"regexp"
	"strings"

	"github.com/WawinyEdwin/lenslocked.com/hash"
	"github.com/WawinyEdwin/lenslocked.com/rand"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"golang.org/x/crypto/bcrypt"
)

var _ UserDB = &userGorm{}
var _ UserService = &userService{}

//password peppering
var userPwPepper = "secret-random-string"

const hmacSecretKey = "secret-hmac-key"

const (
	//to be returned when a resource a resource cannot be found in the database.
	ErrNotFound modelError = "models: resource not found"
	//to be returned  when an invalid Id is provided
	ErrIDInvalid modelError = "models: ID provided was invalid"
	//to be returned when an invalid password is used.
	ErrPasswordIncorrect modelError = "models: incorrect password provided"
	//to be returned when an email address is not provided when creating a user
	ErrEmailRequired modelError = "models: email address is required"
	//to be returned when an email address provided does not match any of our reqs.
	ErrEmailInvalid modelError = "models: email address is invalid"
	//to be returned when an update or create is attempted with an email already in use
	ErrEmailTaken modelError = "models: email address is already taken"
	//to be returned when a usser tries to set a password that is less than 8 characcters long
	ErrPasswordTooShort modelError = "models: password must be at least 8 characters long"
	//to be returned when a create is attempted without a user password provided
	ErrPasswordRequired modelError = "models: password is required"
	//to be returned when a create or update is attempted without a user remember token hash
	ErrRememberRequired modelError = "models: remember token is required"
	//to be returned when a remember token is not atleast 32 bytes
	ErrRememberTooShort modelError = "models: remember token must be ar least 32 bytes"
)

type modelError string

func (e modelError) Error() string {
	return string(e)
}

func (e modelError) Public() string {
	s := strings.Replace(string(e), "models: ", "", 1)
	split := strings.Split(s, " ")
	split[0] = strings.Title(split[0])
	return strings.Join(split, " ")
}

//userDB is used to interact with he users database
//for single user queries, if any error but ErrNotFound should probably result in a 500 error until we make public.
type UserDB interface {
	//methods for querying a single user
	ByID(id uint) (*User, error)
	ByEmail(email string) (*User, error)
	ByRemember(token string) (*User, error)

	// Methods for altering users
	Create(user *User) error
	Update(user *User) error
	Delete(id uint) error
}

//userGorm represents our databse interaction layer
//and implements the UserDB interface fully
type userGorm struct {
	db *gorm.DB
}

//init our data model.
type User struct {
	gorm.Model
	Name         string
	Email        string `gorm:"not null;unique_index"`
	Password     string `gorm: "-"`
	PasswordHash string `gorm: "not null"`
	Remember     string `gorm: "-"`
	RememberHash string `gorm: "not null; unique_index"`
}

//UserService is a set of methods used to manipulate and work with
//the  user Model
type UserService interface {
	//Authenticate will verify the provided email address and password are correct if they are correct the user will be returned else errors
	Authenticate(email, password string) (*User, error)
	UserDB
}

//creating an abstraction layer for our database
type userService struct {
	UserDB
}

//connection for our new user service.
func NewUserService(db *gorm.DB) UserService {
	ug := &userGorm{db}
	hmac := hash.NewHMAC(hmacSecretKey)
	uv := newUserValidator(ug, hmac)
	return &userService{
		UserDB: uv,
	}
}

//allows us to retrieve a user from  the database using the id of the user.
//ErrNotFound results in 500 error.
func (ug *userGorm) ByID(id uint) (*User, error) {
	var user User
	db := ug.db.Where("id = ? ", id)
	err := first(db, &user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

//Create will create the provided user and backfill data
func (ug *userGorm) Create(user *User) error {
	return ug.db.Create(user).Error
}

//we query to get the first item and place into dst else ErrNotFound
func first(db *gorm.DB, dst interface{}) error {
	err := db.First(dst).Error
	if err == gorm.ErrRecordNotFound {
		return ErrNotFound
	}
	return err
}

//we query users by their email addresses and returns a user.
func (ug *userGorm) ByEmail(email string) (*User, error) {
	var user User
	db := ug.db.Where("email = ? ", email)
	err := first(db, &user)
	return &user, err
}

//Looks up a user with te given remeber token and returns  that user.Handles hashing the token for us
func (ug *userGorm) ByRemember(rememberHash string) (*User, error) {
	var user User
	err := first(ug.db.Where("remember_hash = ?", rememberHash), &user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

//updates the provided user with all of the data in the provide user object.
func (ug *userGorm) Update(user *User) error {
	return ug.db.Save(user).Error
}

//deletes a user with provided  id else returns an error.
func (ug *userGorm) Delete(id uint) error {
	user := User{Model: gorm.Model{ID: id}}
	return ug.db.Save(&user).Error
}

//userValidator is our validation layer that validates
//and normalizes data before passing it to the next
//UserDB in our interface
type userValidator struct {
	UserDB
	hmac       hash.HMAC
	emailRegex *regexp.Regexp
}

func newUserValidator(udb UserDB, hmac hash.HMAC) *userValidator {
	return &userValidator{
		UserDB: udb,
		hmac:   hmac,
		emailRegex: regexp.MustCompile(
			`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,16}$`),
	}
}

func (uv *userValidator) rememberMinBytes(user *User) error {
	if user.Remember == "" {
		return nil
	}
	n, err := rand.NBytes(user.Remember)
	if err != nil {
		return err
	}
	if n < 32 {
		return ErrRememberTooShort
	}
	return nil
}

func (uv *userValidator) remeemberHashRequired(user *User) error {
	if user.RememberHash == "" {
		return ErrRememberRequired
	}
	return nil
}
func (uv *userValidator) passwordIsRequired(user *User) error {
	if user.Password == "" {
		return ErrPasswordRequired
	}
	return nil
}

func (uv *userValidator) passwordHashRequired(user *User) error {
	if user.PasswordHash == "" {
		return ErrPasswordRequired
	}
	return nil
}

func (uv *userValidator) passwordMinLength(user *User) error {
	if user.Password == "" {
		return nil
	}
	if len(user.Password) < 8 {
		return ErrPasswordTooShort
	}
	return nil
}

func (uv *userValidator) emailIsAvail(user *User) error {
	existing, err := uv.ByEmail(user.Email)
	if err == ErrNotFound {
		return nil
	}
	if err != nil {
		return err
	}
	if user.ID != existing.ID {
		return ErrEmailTaken
	}
	return nil
}

func (uv *userValidator) emailFormat(user *User) error {
	if user.Email == "" {
		return nil
	}
	if !uv.emailRegex.MatchString(user.Email) {
		return ErrEmailInvalid
	}
	return nil
}

func (uv *userValidator) normalizeEmail(user *User) error {
	user.Email = strings.ToLower(user.Email)
	user.Email = strings.TrimSpace(user.Email)
	return nil
}

func (uv *userValidator) requireEmail(user *User) error {
	if user.Email == "" {
		return ErrEmailRequired
	}
	return nil
}

//By Email will normalize an email address before passing it on to the databse layer to perform a query
func (uv *userValidator) ByEmail(email string) (*User, error) {
	user := User{
		Email: email,
	}
	err := runUserValFns(&user, uv.normalizeEmail)
	if err != nil {
		return nil, err
	}
	return uv.UserDB.ByEmail(user.Email)
}

func (uv *userValidator) idGreaterThan(n uint) userValFn {
	return userValFn(func(user *User) error {
		if user.ID <= n {
			return ErrIDInvalid
		}
		return nil
	})
}

func (uv *userValidator) hmacRemember(user *User) error {
	if user.Remember == "" {
		return nil
	}
	user.RememberHash = uv.hmac.Hash(user.Remember)
	return nil
}

func (uv *userValidator) setRememberIfUnset(user *User) error {
	if user.Remember != "" {
		return nil
	}
	token, err := rand.RememberToken()
	if err != nil {
		return err
	}
	user.Remember = token
	return nil
}

//bcryptPassword will hash a user's password with an app-wide pepper and bcrypt, which salts for us.
func (uv *userValidator) bcryptPassword(user *User) error {
	if user.Password == "" {
		return nil
	}
	pwBytes := []byte(user.Password + userPwPepper)
	hashedBytes, err := bcrypt.GenerateFromPassword(pwBytes, bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.PasswordHash = string(hashedBytes)
	user.Password = ""
	return nil
}

//what we expect user validation function to look like.
type userValFn func(*User) error

func runUserValFns(user *User, fns ...userValFn) error {
	for _, fn := range fns {
		if err := fn(user); err != nil {
			return err
		}
	}
	return nil
}

//ByRemember will hash the remember token and then call ByRemember on the subsequent UserDB layer
func (uv *userValidator) ByRemember(token string) (*User, error) {
	user := User{
		Remember: token,
	}
	if err := runUserValFns(&user, uv.hmacRemember); err != nil {
		return nil, err
	}
	return uv.UserDB.ByRemember(user.RememberHash)
}

//Create will  create the provided user and backfill data like the ID, CreateAt, and UpdateAt fields
func (uv *userValidator) Create(user *User) error {
	if user.Remember == "" {
		token, err := rand.RememberToken()
		if err != nil {
			return err
		}
		user.Remember = token
	}
	err := runUserValFns(user,
		uv.passwordIsRequired,
		uv.passwordMinLength,
		uv.bcryptPassword,
		uv.passwordHashRequired,
		uv.setRememberIfUnset,
		uv.rememberMinBytes,
		uv.hmacRemember,
		uv.remeemberHashRequired,
		uv.normalizeEmail,
		uv.requireEmail,
		uv.emailFormat,
		uv.emailIsAvail)
	if err != nil {
		return err
	}

	return uv.UserDB.Create(user)
}

//Update will hash a remember token if it provided
func (uv *userValidator) Update(user *User) error {
	err := runUserValFns(user,
		uv.passwordMinLength,
		uv.bcryptPassword,
		uv.passwordHashRequired,
		uv.rememberMinBytes,
		uv.hmacRemember,
		uv.remeemberHashRequired,
		uv.normalizeEmail,
		uv.requireEmail,
		uv.emailFormat,
		uv.emailIsAvail)
	if err != nil {
		return err
	}
	return uv.UserDB.Update(user)
}

//Delete will delete the user with A provided ID
func (uv *userValidator) Delete(id uint) error {
	var user User
	user.ID = id
	err := runUserValFns(&user, uv.idGreaterThan(0))
	if err != nil {
		return err
	}
	return uv.UserDB.Delete(id)
}

//used to authenticate a user with the provided email and password.
//if email invalid = ErrNotFound
//if password invalid = ErrPasswordIncorrect
func (us *userService) Authenticate(email, password string) (*User, error) {
	foundUser, err := us.ByEmail(email)
	if err != nil {
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword(
		[]byte(foundUser.PasswordHash),
		[]byte(password+userPwPepper))
	switch err {
	case nil:
		return foundUser, nil
	case bcrypt.ErrMismatchedHashAndPassword:
		return nil, ErrPasswordIncorrect
	default:
		return nil, err
	}
}
