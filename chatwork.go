package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
)

const (
	API_ENDPOINT_BASE   = "https://api.chatwork.com/v2/"
	API_REQ_HEADER_BASE = "X-ChatWorkToken"
)

type Api struct {
	ApiToken string
	RoomHash map[string]int
	Rooms    []Room
}

type Room struct {
	RoomId int    `json:"room_id"`
	Name   string `json:"name"`
}

type RequestData struct {
	Endpoint string
	Method   string
	Payload  map[string]string
}

func (api *Api) GetRooms() error {
	endpoint := API_ENDPOINT_BASE + "rooms"
	body, err := api.apiRequest(RequestData{
		Endpoint: endpoint,
	})
	if err != nil {
		return errors.Wrap(err, "")
	}

	rooms := []Room{}
	if err := json.Unmarshal(body, &rooms); err != nil {
		return errors.Wrap(err, "Failed to decode Response")
	}
	api.Rooms = rooms

	api.RoomHash = make(map[string]int)
	for _, v := range rooms {
		api.RoomHash[v.Name] = v.RoomId
	}
	return nil
}

func (api *Api) GetRoomId(name string) (int, bool, error) {
	if api.RoomHash == nil {
		if err := api.GetRooms(); err != nil {
			return 0, false, errors.Wrap(err, "Failed to get Rooms related to the account")
		}
	}
	id, exists := api.RoomHash[name]
	return id, exists, nil
}

func (api *Api) apiRequest(req_data RequestData) ([]byte, error) {
	if req_data.Endpoint == "" {
		return nil, errors.New("Endpoint isn't set in RequestData")
	}

	vals := url.Values{}
	for k, v := range req_data.Payload {
		vals.Add(k, v)
	}
	req, err := http.NewRequest(
		map[bool]string{false: req_data.Method, true: "GET"}[req_data.Method == ""],
		req_data.Endpoint,
		strings.NewReader(vals.Encode()),
	)
	if err != nil {
		return nil, errors.New("Failed to create Request")
	}
	req.Header.Add(API_REQ_HEADER_BASE, api.ApiToken)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, errors.New("Failed to Request")
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to read Response")
	}
	return body, nil
}

func (api *Api) apiRequestWithFile(req_data RequestData) ([]byte, error) {
	if req_data.Endpoint == "" {
		return nil, errors.New("Endpoint isn't set in RequestData")
	}

	body := &bytes.Buffer{}
	mw := multipart.NewWriter(body)

	mw.WriteField("message", req_data.Payload["message"])

	file, _ := os.Open(req_data.Payload["filepath"])
	fw, _ := mw.CreateFormFile("file", req_data.Payload["filepath"])
	io.Copy(fw, file)

	contentType := mw.FormDataContentType()
	mw.Close()

	req, err := http.NewRequest(
		"POST",
		req_data.Endpoint,
		body,
	)
	if err != nil {
		return nil, errors.New("Failed to create Request")
	}
	req.Header.Add(API_REQ_HEADER_BASE, api.ApiToken)
	req.Header.Add("Content-Type", contentType)

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, errors.New("Failed to Request")
	}
	defer res.Body.Close()

	resbody, err := ioutil.ReadAll(res.Body)
	fmt.Println(string(resbody))
	if err != nil {
		return nil, errors.Wrap(err, "Failed to read Response")
	}
	return resbody, nil
}

func (api *Api) SendMessageByName(name string, msg string) ([]byte, error) {
	room_id, exists, err := api.GetRoomId(name)
	if err != nil {
		return nil, errors.Wrap(err, "")
	}
	if !exists {
		return nil, errors.New("Room [" + name + "] isn't exist in your account.")
	}
	endpoint := API_ENDPOINT_BASE + "rooms/" + strconv.Itoa(room_id) + "/messages"

	res, err := api.apiRequest(RequestData{
		Endpoint: endpoint,
		Method:   "POST",
		Payload: map[string]string{
			"body": msg,
		},
	})
	return res, err
}

func (api *Api) SendMessageByNameWithFile(name string, msg string, file string) ([]byte, error) {
	room_id, exists, err := api.GetRoomId(name)
	if err != nil {
		return nil, errors.Wrap(err, "")
	}
	if !exists {
		return nil, errors.New("Room [" + name + "] isn't exist in your account.")
	}
	endpoint := API_ENDPOINT_BASE + "rooms/" + strconv.Itoa(room_id) + "/files"

	res, err := api.apiRequestWithFile(RequestData{
		Endpoint: endpoint,
		Method:   "POST",
		Payload: map[string]string{
			"message":  msg,
			"filepath": file,
		},
	})
	return res, err
}
