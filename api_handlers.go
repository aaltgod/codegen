package main

import "net/http"


// MyApi
func (srv *MyApi) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	// MyApiSwitch
	switch r.URL.Path {
	case "/user/profile":
		srv.profileWrapper(w, r)
	case "/user/create":
		srv.createWrapper(w, r)
	default:
		http.Error(w, "", http.StatusBadRequest)
	}
}

// createWrapper
func (srv *MyApi) createWrapper(w http.ResponseWriter, r *http.Request) {
}

// profileWrapper
func (srv *MyApi) profileWrapper(w http.ResponseWriter, r *http.Request) {
}

// OtherApi
func (srv *OtherApi) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	// OtherApiSwitch
	switch r.URL.Path {
	case "/user/profile":
		srv.profileWrapper(w, r)
	case "/user/create":
		srv.createWrapper(w, r)
	default:
		http.Error(w, "", http.StatusBadRequest)
	}
}

// createWrapper
func (srv *OtherApi) createWrapper(w http.ResponseWriter, r *http.Request) {
}

// profileWrapper
func (srv *OtherApi) profileWrapper(w http.ResponseWriter, r *http.Request) {
}
