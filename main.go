package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/mux"
)

type EventData struct {
	sync.Mutex
	Data   map[int]interface{}
	NextID int
}

var data EventData

func (e *EventData) GetData(id int) (interface{}, bool) {
	e.Lock()
	defer e.Unlock()

	d, exists := e.Data[id]
	if !exists {
		return nil, false
	}
	return d, true
}

func (e *EventData) InsertData(d interface{}) int {
	e.Lock()
	defer e.Unlock()

	if e.Data == nil {
		e.Data = make(map[int]interface{})
	}

	id := e.NextID
	e.Data[id] = d
	e.NextID = id + 1
	return id
}

func main() {
	router := mux.NewRouter()

	router.HandleFunc("/events", insertEventsHandler).Methods("POST")
	router.HandleFunc("/events/{id}", getEventsHandler).Methods("GET")

	err := http.ListenAndServe(":9035", router)
	if err != nil {
		fmt.Println(err)
	}
}

func getEventsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var err error
	var d interface{}
	var exists bool

	defer func() {
		if err != nil {
			b, _ := json.Marshal(map[string]interface{}{"error": err.Error()})
			w.WriteHeader(400)
			w.Write(b)
			return
		}

		if !exists {
			b, _ := json.Marshal(map[string]interface{}{"error": "record doesn't exists"})
			w.WriteHeader(404)
			w.Write(b)
			return
		}

		b, _ := json.Marshal(d)
		w.WriteHeader(200)
		w.Write(b)
	}()

	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return
	}

	d, exists = data.GetData(id)
}

func insertEventsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var err error
	var id int

	defer func() {
		if err != nil {
			b, _ := json.Marshal(map[string]interface{}{"error": err.Error()})
			w.WriteHeader(400)
			w.Write(b)
			return
		}

		b, _ := json.Marshal(map[string]interface{}{"id": id})
		w.WriteHeader(200)
		w.Write(b)
	}()

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return
	}

	var m interface{}
	err = json.Unmarshal(body, &m)
	if err != nil {
		return
	}

	id = data.InsertData(m)

	rem := id % 10
	switch true {
	case rem == 2:
		time.Sleep(time.Second * 2)
	case rem == 5:
		time.Sleep(time.Second * 5)
	case rem == 8:
		time.Sleep(time.Second * 8)
	default:
	}
}
