package main

import (
	"encoding/json"
	"fmt"
	"github.com/Kosench/restdb"
	"github.com/gorilla/mux"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"
)

func SliceToJSON(slice interface{}, w io.Writer) error {
	e := json.NewEncoder(w)
	return e.Encode(slice)
}

type notAllowedHandler struct{}

func (h notAllowedHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	MethodNotAllowedHandler(rw, r)
}

func DefaultHandler(rw http.ResponseWriter, r *http.Request) {
	log.Println("DefaultHandler Serving:", r.URL.Path, "from", r.Host, "with method", r.Method)
	rw.WriteHeader(http.StatusNotFound)
	Body := r.URL.Path + " is not supported. Thanks for visiting!\n"
	fmt.Fprintf(rw, "%s", Body)
}

func MethodNotAllowedHandler(rw http.ResponseWriter, r *http.Request) {
	log.Println("Serving:", r.URL.Path, "from", r.Host, "with method", r.Method)
	rw.WriteHeader(http.StatusNotFound)
	Body := "Method not allowed!\n"
	fmt.Fprintf(rw, "%s", Body)
}

func TimeHandler(rw http.ResponseWriter, r *http.Request) {
	log.Println("TimeHandler Serving:", r.URL.Path, "from", r.Host)
	rw.WriteHeader(http.StatusOK)
	t := time.Now().Format(time.RFC1123)
	Body := "The current time is: " + t + "\n"
	fmt.Fprintf(rw, "%s", Body)
}

func AddHandler(rw http.ResponseWriter, r *http.Request) {
	log.Println("AddHandler Serving:", r.URL.Path, "from", r.Host)
	d, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	if len(d) == 0 {
		rw.WriteHeader(http.StatusBadRequest)
		log.Println("No input!")
		return
	}

	var users = []restdb.User{}
	err = json.Unmarshal(d, &users)
	if err != nil {
		log.Println(err)
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	log.Println(users)
	if !restdb.IsUserAdmin(users[0]) {
		log.Println("Issued by non-admin user:", users[0].Username)
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	result := restdb.InsertUser(users[1])
	if !result {
		rw.WriteHeader(http.StatusBadRequest)
	}

}

func DeleteHandler(rw http.ResponseWriter, r *http.Request) {
	log.Println("DeleteHandler Serving:", r.URL.Path, "from", r.Host)

	// Get the ID of the user to be deleted
	id, ok := mux.Vars(r)["id"]
	if !ok {
		log.Println("ID value not set!")
		rw.WriteHeader(http.StatusNotFound)
		return
	}

	var user = restdb.User{}
	err := user.FromJSON(r.Body)
	if err != nil {
		log.Println(err)
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	if !restdb.IsUserAdmin(user) {
		log.Println("User", user.Username, "is not admin!")
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	intID, err := strconv.Atoi(id)
	if err != nil {
		log.Println("id", err)
		return
	}

	t := restdb.FindUserID(intID)
	if t.Username != "" {
		log.Println("About to delete:", t)
		deleted := restdb.DeleteUser(intID)
		if deleted {
			log.Println("User deleted:", id)
			rw.WriteHeader(http.StatusOK)
			return
		} else {
			log.Println("User ID not found:", id)
			rw.WriteHeader(http.StatusNotFound)
		}
	}
	rw.WriteHeader(http.StatusNotFound)
}

func GetAllHandler(rw http.ResponseWriter, r *http.Request) {
	log.Println("GetAllHandler Serving:", r.URL.Path, "from", r.Host)
	d, err := io.ReadAll(r.Body)
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		log.Println(err)
		return
	}

	if len(d) == 0 {
		rw.WriteHeader(http.StatusBadRequest)
		log.Println("No input!")
		return
	}

	var user = restdb.User{}
	err = json.Unmarshal(d, &user)
	if err != nil {
		log.Println(err)
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	if !restdb.IsUserAdmin(user) {
		log.Println("User", user, "is not an admin!")
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	err = SliceToJSON(restdb.ListAllUsers(), rw)
	if err != nil {
		log.Println(err)
		rw.WriteHeader(http.StatusBadRequest)
		return
	}
}

func GetIDHandler(rw http.ResponseWriter, r *http.Request) {
	log.Println("GetIDHandler Serving:", r.URL.Path, "from", r.Host)

	username, ok := mux.Vars(r)["username"]
	if !ok {
		log.Println("ID value not set")
		rw.WriteHeader(http.StatusNotFound)
		return
	}

	d, err := io.ReadAll(r.Body)
	if err != nil {
		rw.WriteHeader(http.StatusNotFound)
		log.Println("No input")
		return
	}

	if len(d) == 0 {
		rw.WriteHeader(http.StatusBadRequest)
		log.Println("No input!")
		return
	}

	var user = restdb.User{}
	err = json.Unmarshal(d, &user)
	if err != nil {
		log.Println(err)
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	log.Println("Input user:", user)
	if !restdb.IsUserAdmin(user) {
		log.Println("User", user.Username, "not an admin!")
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	t := restdb.FindUserUsername(username)
	if t.ID != 0 {
		err := t.ToJSON(rw)
		if err != nil {
			rw.WriteHeader(http.StatusBadRequest)
			log.Println(err)
		}
	} else {
		rw.WriteHeader(http.StatusNotFound)
		log.Println("User " + user.Username + "not found")
	}
}

func GetUserDataHandler(rw http.ResponseWriter, r *http.Request) {
	log.Println("GetUserDataHandler Serving:", r.URL.Path, "from", r.Host)
	id, ok := mux.Vars(r)["id"]
	if !ok {
		log.Println("ID value not set!")
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	intID, err := strconv.Atoi(id)
	if err != nil {
		log.Println(err)
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	t := restdb.FindUserID(intID)
	if t.ID != 0 {
		err = t.ToJSON(rw)
		if err != nil {
			log.Println(err)
			rw.WriteHeader(http.StatusBadRequest)
			return
		}
		return
	} else {
		log.Println("User not found:", id)
		rw.WriteHeader(http.StatusBadRequest)
	}
}

func UpdateHandler(rw http.ResponseWriter, r *http.Request) {
	log.Println("UpdateHandler Serving:", r.URL.Path, "from", r.Host)
	d, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	if len(d) == 0 {
		rw.WriteHeader(http.StatusBadRequest)
		log.Println("No input!")
		return
	}

	var users = []restdb.User{}
	err = json.Unmarshal(d, &users)
	if err != nil {
		log.Println(err)
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	if !restdb.IsUserAdmin(users[0]) {
		log.Println("Command issued by non-admin user:", users[0].Username)
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	log.Println(users)
	t := restdb.FindUserUsername(users[1].Username)
	t.Username = users[1].Username
	t.Password = users[1].Password
	t.Admin = users[1].Admin

	if !restdb.UpdateUser(t) {
		log.Println("Update failed:", t)
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	log.Println("Update successful:", t)
	rw.WriteHeader(http.StatusOK)
}

func LoginHandler(rw http.ResponseWriter, r *http.Request) {
	log.Println("LoginHandler Serving:", r.URL.Path, "from", r.Host)
	d, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	if len(d) == 0 {
		rw.WriteHeader(http.StatusBadRequest)
		log.Println("No input!")
		return
	}

	var user = restdb.User{}
	err = json.Unmarshal(d, &user)
	if err != nil {
		log.Println(err)
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	log.Println("Input user:", user)

	if !restdb.IsUserValid(user) {
		log.Println("User", user.Username, "not valid")
		return
	}

	t := restdb.FindUserUsername(user.Username)
	log.Println("Logging in:", t)

	t.LastLogin = time.Now().Unix()
	t.Active = 1
	if restdb.UpdateUser(t) {
		log.Println("User updated: ", t)
		rw.WriteHeader(http.StatusOK)
	} else {
		log.Println("Update failed:", t)
		rw.WriteHeader(http.StatusBadRequest)
	}

}

func LogoutHandler(rw http.ResponseWriter, r *http.Request) {
	log.Println("LogoutHandler Serving:", r.URL.Path, "from", r.Host)
	d, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	if len(d) == 0 {
		rw.WriteHeader(http.StatusBadRequest)
		log.Println("No input!")
		return
	}

	var user = restdb.User{}
	err = json.Unmarshal(d, &user)
	if err != nil {
		log.Println(err)
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	if !restdb.IsUserValid(user) {
		log.Println("User", user.Username, "exists!")
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	t := restdb.FindUserUsername(user.Username)
	log.Println("Logging out: ", t.Username)
	t.Active = 0
	if restdb.UpdateUser(user) {
		log.Println("User updated:", t)
		rw.WriteHeader(http.StatusOK)
	} else {
		log.Println("User failed:", t)
		rw.WriteHeader(http.StatusBadRequest)
	}

}

func LoggedUsersHandler(rw http.ResponseWriter, r *http.Request) {
	log.Println("LoggedUsersHandler Serving:", r.URL.Path, "from", r.Host)
	var user = restdb.User{}

	err := user.FromJSON(r.Body)
	if err != nil {
		log.Println(err)
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	if !restdb.IsUserValid(user) {
		log.Println("User", user.Username, "does not exist or is invalid!")
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	err = SliceToJSON(restdb.ReturnLoggedUsers(), rw)
	if err != nil {
		log.Println(err)
		rw.WriteHeader(http.StatusBadRequest)
		return
	}
}
