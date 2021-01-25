package main

import "net/http"
import "fmt"
import "strconv"
import "encoding/json"

type Response map[string]interface{}


// MyApi
func (srv *MyApi) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// MyApiSwitch
	switch r.URL.Path {
	case "/user/profile":
		srv.ProfileWrapper(w, r)
	case "/user/create":
		srv.CreateWrapper(w, r)
	default:
		response, _ := json.Marshal(&Response{
			"error": "unknown method",
		})
		w.WriteHeader(http.StatusNotFound)
		w.Write(response)
		return
	}
}

func (srv *MyApi) ProfileWrapper(w http.ResponseWriter, r *http.Request) {
	if err := checkRequestMethod("", r); err != nil {
		response, _ := json.Marshal(&Response{
				"error": err.Error(),
		})
		w.WriteHeader(http.StatusNotAcceptable)
		w.Write(response)
		return
	}

	output := ProfileParams{}

	output.Login = r.FormValue("login")
	if output.Login == "" {
		response, _ := json.Marshal(&Response{
			"error": "login must me not empty",
		})

		w.WriteHeader(http.StatusBadRequest)
		w.Write(response)

		return
	}

	res, err := srv.Profile(r.Context(), output)
	if err != nil {
		switch err.(type) {
		case ApiError:
			err := err.(ApiError)
			response, _ := json.Marshal(&Response{
			"error": err.Error(),
			})

			w.WriteHeader(err.HTTPStatus)
			w.Write(response)
		case error:
			response, _ := json.Marshal(&Response{
			"error": err.Error(),
			})

			w.WriteHeader(http.StatusInternalServerError)
			w.Write(response)
		}
		
		return
	}

	response, _ := json.Marshal(&Response{
		"error": "",
		"response": res, 
	})

	w.Write(response)

}

func (srv *MyApi) CreateWrapper(w http.ResponseWriter, r *http.Request) {
	if err := checkRequestMethod("POST", r); err != nil {
		response, _ := json.Marshal(&Response{
				"error": err.Error(),
		})
		w.WriteHeader(http.StatusNotAcceptable)
		w.Write(response)
		return
	}
	if err := checkAuth(r); err != nil {
		response, _ := json.Marshal(&Response{
			"error": err.Error(),
		})
		w.WriteHeader(http.StatusForbidden)
		w.Write(response)
		return
	}

	output := CreateParams{}

	output.Login = r.FormValue("login")
	if output.Login == "" {
		response, _ := json.Marshal(&Response{
			"error": "login must me not empty",
		})

		w.WriteHeader(http.StatusBadRequest)
		w.Write(response)

		return
	}

	if len(output.Login) < 10 {
		response, _ := json.Marshal(&Response{
			"error": "login len must be >= 10",
		})

		w.WriteHeader(http.StatusBadRequest)
		w.Write(response)

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
			response, _ := json.Marshal(&Response{
			"error": "status must be one of [user, moderator, admin]",
			})

			w.WriteHeader(http.StatusBadRequest)
			w.Write(response)

			return
		}
	} else {
		output.Status = "user"
	}

	var err error	
	output.Age, err = strconv.Atoi(r.FormValue("age"))
	if err != nil {
		response, _ := json.Marshal(&Response{
			"error": "age must be int",
		})

		w.WriteHeader(http.StatusBadRequest)
		w.Write(response)
		
		return
	}
	if output.Age < 0 {
		response, _ := json.Marshal(&Response{
			"error": "age must be >= 0",
		})

		w.WriteHeader(http.StatusBadRequest)
		w.Write(response)
		
		return
	}

	if output.Age > 128 {
		response, _ := json.Marshal(&Response{
			"error": "age must be <= 128",
		})

		w.WriteHeader(http.StatusBadRequest)
		w.Write(response)

		return
	}

	res, err := srv.Create(r.Context(), output)
	if err != nil {
		switch err.(type) {
		case ApiError:
			err := err.(ApiError)
			response, _ := json.Marshal(&Response{
			"error": err.Error(),
			})

			w.WriteHeader(err.HTTPStatus)
			w.Write(response)
		case error:
			response, _ := json.Marshal(&Response{
			"error": err.Error(),
			})

			w.WriteHeader(http.StatusInternalServerError)
			w.Write(response)
		}
		
		return
	}

	response, _ := json.Marshal(&Response{
		"error": "",
		"response": res, 
	})

	w.Write(response)

}

// OtherApi
func (srv *OtherApi) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// OtherApiSwitch
	switch r.URL.Path {
	case "/user/create":
		srv.CreateWrapper(w, r)
	default:
		response, _ := json.Marshal(&Response{
			"error": "unknown method",
		})
		w.WriteHeader(http.StatusNotFound)
		w.Write(response)
		return
	}
}

func (srv *OtherApi) CreateWrapper(w http.ResponseWriter, r *http.Request) {
	if err := checkRequestMethod("POST", r); err != nil {
		response, _ := json.Marshal(&Response{
				"error": err.Error(),
		})
		w.WriteHeader(http.StatusNotAcceptable)
		w.Write(response)
		return
	}
	if err := checkAuth(r); err != nil {
		response, _ := json.Marshal(&Response{
			"error": err.Error(),
		})
		w.WriteHeader(http.StatusForbidden)
		w.Write(response)
		return
	}

	output := OtherCreateParams{}

	output.Username = r.FormValue("username")
	if output.Username == "" {
		response, _ := json.Marshal(&Response{
			"error": "username must me not empty",
		})

		w.WriteHeader(http.StatusBadRequest)
		w.Write(response)

		return
	}

	if len(output.Username) < 3 {
		response, _ := json.Marshal(&Response{
			"error": "username len must be >= 3",
		})

		w.WriteHeader(http.StatusBadRequest)
		w.Write(response)

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
			response, _ := json.Marshal(&Response{
			"error": "class must be one of [warrior, sorcerer, rouge]",
			})

			w.WriteHeader(http.StatusBadRequest)
			w.Write(response)

			return
		}
	} else {
		output.Class = "warrior"
	}

	var err error	
	output.Level, err = strconv.Atoi(r.FormValue("level"))
	if err != nil {
		response, _ := json.Marshal(&Response{
			"error": "level must be int",
		})

		w.WriteHeader(http.StatusBadRequest)
		w.Write(response)
		
		return
	}
	if output.Level < 1 {
		response, _ := json.Marshal(&Response{
			"error": "level must be >= 1",
		})

		w.WriteHeader(http.StatusBadRequest)
		w.Write(response)
		
		return
	}

	if output.Level > 50 {
		response, _ := json.Marshal(&Response{
			"error": "level must be <= 50",
		})

		w.WriteHeader(http.StatusBadRequest)
		w.Write(response)

		return
	}

	res, err := srv.Create(r.Context(), output)
	if err != nil {
		switch err.(type) {
		case ApiError:
			err := err.(ApiError)
			response, _ := json.Marshal(&Response{
			"error": err.Error(),
			})

			w.WriteHeader(err.HTTPStatus)
			w.Write(response)
		case error:
			response, _ := json.Marshal(&Response{
			"error": err.Error(),
			})

			w.WriteHeader(http.StatusInternalServerError)
			w.Write(response)
		}
		
		return
	}

	response, _ := json.Marshal(&Response{
		"error": "",
		"response": res, 
	})

	w.Write(response)

}

func checkRequestMethod(availableMethod string, r *http.Request) error {
	if availableMethod == r.Method || availableMethod == "" {
		return nil
	}
	
	return fmt.Errorf("%s", "bad method")
}

func checkAuth(r *http.Request) error {
	auth := r.Header.Get("X-Auth")
	if auth == "100500" {
		return nil
	}
	
	return fmt.Errorf("%s", "unauthorized")
}
