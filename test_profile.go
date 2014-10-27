package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

type TestProfile struct {
	Comment  string
	Host     string
	Port     float64
	Token    string
	AppToken string
	API      []TestAPI
}

type Net int8

const (
	UNKNOWNNET Net = iota
	HTTP
	WEBSOCKET
	TCP
)

func NetFromString(n string) Net {
	switch n {
	case "http":
		return HTTP
	case "websocket":
		return WEBSOCKET
	case "tcp":
		return TCP
	}
	return UNKNOWNNET
}

type Protocol int8

const (
	UNKNOWNPRO Protocol = iota
	JSONRPC
)

func ProtocolFromString(p string) Protocol {
	switch p {
	case "jsonrpc":
		return JSONRPC
	}
	return UNKNOWNPRO
}

type TestAPI struct {
	Comment     string
	Performance bool
	Net         string
	KeepAlive   bool
	Timeout     float64
	HTTPMethod  string
	UserAgent   string
	Path        string
	Protocol    string
	Case        string
	Repeat      float64
	Concurrency  float64
}

func ParseProfileFromFile(filename string) (*TestProfile, error) {
	profileByte, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Println("read profile", filename, err)
		return nil, err
	}
	var profile TestProfile
	jsonErr := json.Unmarshal(profileByte, &profile)
	if jsonErr != nil {
		fmt.Println("json unmarshal profile", jsonErr)
		return nil, jsonErr
	}
	return &profile, nil
}
