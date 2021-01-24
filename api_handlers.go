package main

import "net/http"


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
	}
}

func (srv *MyApi) ProfileWrapper(w http.ResponseWriter, r *http.Request) {
	checkRequestMethod("", w, r)
}

func (srv *MyApi) CreateWrapper(w http.ResponseWriter, r *http.Request) {
	checkRequestMethod("POST", w, r)
}

// OtherApi
func (srv *OtherApi) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// OtherApiSwitch
	switch r.URL.Path {
	case "/user/create":
		srv.CreateWrapper(w, r)
	default:
		http.Error(w, "", 404)
	}
}

func (srv *OtherApi) CreateWrapper(w http.ResponseWriter, r *http.Request) {
	checkRequestMethod("POST", w, r)
}

func checkRequestMethod(availableMethod string, w http.ResponseWriter, r *http.Request) {
	if availableMethod == r.Method || availableMethod == "" {
		return
	}

	http.Error(w, "bad method", 406)
}