package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

type Resp struct {
	Status        string         `json:"status"`
	DomainCheckup *DomainCheckup `json:"domian"`
}

type DomainCheckup struct {
	HasMX       bool   `json:"hasMX"`
	HasSPF      bool   `json:"hasSPF"`
	HasDMARC    bool   `json:"hasDMARC"`
	SpfRecord   string `json:"spfRecord"`
	DmarcRecord string `json:"dmarcRecord"`
}

/**
 * validate email
**/
func checkEmail(w http.ResponseWriter, r *http.Request) {
	//setting header content type to json
	w.Header().Set("content-Type", "application/json")

	var resp Resp

	//getting params from header
	params := mux.Vars(r)

	//get email
	email := params["email"]

	emailsplit := strings.Split(email, "@")

	if len(emailsplit) > 1 {
		domain := emailsplit[1]

		DomainCheckup, err := checkDomain(domain)
		if err != nil {
			//encodig reponse and movie to json and return
			json.NewEncoder(w).Encode(err)
			return
		}

		resp.Status = "success"
		resp.DomainCheckup = &DomainCheckup

		json.NewEncoder(w).Encode(resp)
		return
	}
	json.NewEncoder(w).Encode("error invalid email")
	return
}

/**
 * validate domain
**/
func checkDomain(domain string) (DomainCheckup, error) {
	var DomainCheckup DomainCheckup

	var hasMx, hasSPF, hasDMARC bool
	var spfRecord, dmarcRecord string

	//MX lookup
	mxRecords, err := net.LookupMX(domain)
	if err != nil {
		log.Printf("Error:%v\n", err)
		return DomainCheckup, err
	}

	if len(mxRecords) > 0 {
		hasMx = true
	} else {
		hasMx = false
	}

	//TXT lookup
	txtRecords, err := net.LookupTXT(domain)
	if err != nil {
		log.Printf("Error:%v\n", err)
		return DomainCheckup, err
	}

	//search for spf1 record
	for _, record := range txtRecords {
		if strings.HasPrefix(record, "v=spf1") {
			hasSPF = true
			spfRecord = record
			break
		}
	}

	//dmarc lookup
	dmarcRecords, err := net.LookupTXT("_dmarc." + domain)
	if err != nil {
		log.Printf("Error:%v\n", err)
		return DomainCheckup, err
	}

	//search for Dmarc1 record
	for _, record := range dmarcRecords {
		if strings.HasPrefix(record, "v=DMARC1") {
			hasDMARC = true
			dmarcRecord = record
			break
		}
	}

	DomainCheckup.HasMX = hasMx
	DomainCheckup.HasSPF = hasSPF
	DomainCheckup.HasDMARC = hasDMARC
	DomainCheckup.SpfRecord = spfRecord
	DomainCheckup.DmarcRecord = dmarcRecord
	return DomainCheckup, err

}

/**
 * main function
**/
func main() {
	//using gorilla mux for routing
	r := mux.NewRouter()

	//creating routes
	r.HandleFunc("/email/{email}", checkEmail).Methods("GET")

	//serving at port 8000
	fmt.Printf("Staring server at port 8000\n")
	log.Fatal(http.ListenAndServe(":8000", r))
}
