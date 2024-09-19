package main

import "encoding/json"

type User struct {
	Id       string `json:"ID"`
	UserName string `json:"userName"`
	// Other fields...
}

type TestRequest struct {
	ServiceType string          `json:"service_type"`
	User        User            `json:"user"`
	DataRequest json.RawMessage `json:"data_request"`
	MethodName  string          `json:"method_name"`
}
