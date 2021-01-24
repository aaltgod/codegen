package main

import "net/http"
import "strconv"


// MyApi
func (srv *MyApi) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// MyApiSwitch
	switch r.URL.Path {
	case "/user/profile":
		srv.ProfileWrapper(w, r)
	case "/user/create":
		srv.CreateWrapper(w, r)
	default:
		http.Error(w, "", 404)
		return
	}
}

func (srv *MyApi) ProfileWrapper(w http.ResponseWriter, r *http.Request) {
	checkRequestMethod("", w, r)

	output := &ProfileParams{}

	output.Login = r.FormValue("login")
}

func (srv *MyApi) CreateWrapper(w http.ResponseWriter, r *http.Request) {
	checkRequestMethod("POST", w, r)

	output := &CreateParams{}

	output.Login = r.FormValue("login")
	if len(output.Login) < 10 {
		http.Error(w, "login len must be >= 10", 400)
		return
	}
	output.Name = r.FormValue("full_name")
	output.Status = r.FormValue("status")
	if output.Status != "" {
		switch output.Status {
		case "user":
			break
		case "moderator":
			break
		case "admin":
			break
		default:
			http.Error(w, "status must be one of [user, moderator, admin]", 400)
			return
		}
	} else {
		output.Status = "user"
	}
	output.Age, _ = strconv.Atoi(r.FormValue("age"))
	if output.Age < 0 {
		http.Error(w, "age must be >= 0", 400)
		return
	}
	if output.Age > 128 {
		http.Error(w, "age must be <= 128", 400)
		return
	}
}

// OtherApi
func (srv *OtherApi) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// OtherApiSwitch
	switch r.URL.Path {
	case "/user/create":
		srv.CreateWrapper(w, r)
	default:
		http.Error(w, "", 404)
		return
	}
}

func (srv *OtherApi) CreateWrapper(w http.ResponseWriter, r *http.Request) {
	checkRequestMethod("POST", w, r)

	output := &OtherCreateParams{}

	output.Username = r.FormValue("username")
	if len(output.Username) < 3 {
		http.Error(w, "username len must be >= 3", 400)
		return
	}
	output.Name = r.FormValue("account_name")
	output.Class = r.FormValue("class")
	if output.Class != "" {
		switch output.Class {
		case "warrior":
			break
		case "sorcerer":
			break
		case "rouge":
			break
		default:
			http.Error(w, "class must be one of [warrior, sorcerer, rouge]", 400)
			return
		}
	} else {
		output.Class = "warrior"
	}
	output.Level, _ = strconv.Atoi(r.FormValue("level"))
	if output.Level < 1 {
		http.Error(w, "level must be >= 1", 400)
		return
	}
	if output.Level > 50 {
		http.Error(w, "level must be <= 50", 400)
		return
	}
}

func checkRequestMethod(availableMethod string, w http.ResponseWriter, r *http.Request) {
	if availableMethod == r.Method || availableMethod == "" {
		return
	}

	http.Error(w, "bad method", 406)
}