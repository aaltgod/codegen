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
		http.Error(w, "", http.StatusBadRequest)
	}
}

func (srv *MyApi) ProfileWrapper(w http.ResponseWriter, r *http.Request) {
}

func (srv *MyApi) CreateWrapper(w http.ResponseWriter, r *http.Request) {
}

// OtherApi
func (srv *OtherApi) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// OtherApiSwitch
	switch r.URL.Path {
	case "/user/create":
		srv.CreateWrapper(w, r)
	default:
		http.Error(w, "", http.StatusBadRequest)
	}
}

func (srv *OtherApi) CreateWrapper(w http.ResponseWriter, r *http.Request) {
}
