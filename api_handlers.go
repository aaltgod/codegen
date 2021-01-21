package main

import "net/http"
import "encoding/binary"
import "bytes"


	// URL
	var URLRaw uint32
	binary.Read(reader, binary.LittleEndian, &URLRaw)
	srv.URL = int(URLRaw)

	// Query
	var QueryLenRaw uint32
	binary.Read(reader, binary.LittleEndian, &QueryLenRaw)
	QueryRaw := make([]byte, QueryLenRaw)
	binary.Read(reader, binary.LittleEndian, QueryRaw)
	srv.Query = string(QueryRaw)
func (srv *MyApi) ServeHTTP(w http.ResponseWriter, r *http.Request) {
 reader := bytes.NewReader(r)
}
func (srv *MyApi) ServeHTTP(w http.ResponseWriter, r *http.Request) {
 reader := bytes.NewReader(r)
}
func (srv *OtherApi) ServeHTTP(w http.ResponseWriter, r *http.Request) {
 reader := bytes.NewReader(r)
}
}
