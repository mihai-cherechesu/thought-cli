package pkg

import (
	"encoding/json"
	"log"
	"net/http"
)

const (
	CPX_API_URL = "http://localhost:8081/"
)

type ServersResponse []string

type ServiceResponse struct {
	Cpu     string
	Memory  string
	Service string
}

func GetServers() ServersResponse {
	r, err := http.Get(CPX_API_URL + "servers")
	if err != nil {
		log.Fatalf("could not fetch servers")
	}

	defer r.Body.Close()

	var s ServersResponse
	json.NewDecoder(r.Body).Decode(&s)
	return s
}

func GetService(ip string) ServiceResponse {
	r, err := http.Get(CPX_API_URL + ip)
	if err != nil {
		log.Fatalf("could not fetch servers")
	}

	defer r.Body.Close()

	s := ServiceResponse{}
	json.NewDecoder(r.Body).Decode(&s)
	return s
}
