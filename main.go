package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Used to pass between javascript and go
// as json does not like the recursion in the
// Next and Prev fields
type TaskModel struct {
	ID         uuid.UUID
	Title      string
	Desc       string
	StartTime  time.Time
	FinishTime time.Time
	Complete   bool
}

func (t *Task) toTaskModel() TaskModel {
	return TaskModel{t.ID,
		t.Title,
		t.Desc,
		t.StartTime,
		t.FinishTime,
		t.Complete}
}

func TaskArrayToTaskModels(ta []*Task) []TaskModel {
	tms := make([]TaskModel, len(ta))
	for i := range ta {
		tms[i] = ta[i].toTaskModel()
	}
	return tms
}

// r = request
// w = ?
func submitTask(w http.ResponseWriter, r *http.Request) {
	// Allow CORS here By * or specific origin
	// asserts that only allowed Origins are recieved.
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:1080")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// assert that request is post
	if r.Method != http.MethodPost {
		return
	}
	// assert that body is not too large
	r.Body = http.MaxBytesReader(w, r.Body, 1048576)

	// make decoder for r.Body
	dec := json.NewDecoder(r.Body)

	// Make newTask and try to decode info into it
	//var bucket interface{}
	bucket := make(map[string]string)
	err := dec.Decode(&bucket)

	// assert that no errors occurred
	if err != nil {
		log.Println("Error: ", err)
		log.Println("Value: ", bucket)
		return
	}

	completeness, _ := strconv.ParseBool(bucket["Complete"])
	newTask := MakeTask(bucket["Title"], bucket["Desc"], completeness)
	Mastertasklist.AppendTask(*newTask)

	//log.Println(newTask.ToString())
}

func submitChanges(w http.ResponseWriter, r *http.Request) {
	// Allow CORS here By * or specific origin
	// asserts that only allowed Origins are recieved.
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:1080")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Access-Control-Allow-Methods", "PUT")

	// assert that request is post
	if r.Method != http.MethodPut {
		return
	}
	// assert that body is not too large
	r.Body = http.MaxBytesReader(w, r.Body, 1048576)

	// make decoder for r.Body
	dec := json.NewDecoder(r.Body)

	// Make newTask and try to decode info into it
	//var bucket interface{}
	bucket := make(map[string]string)
	err := dec.Decode(&bucket)

	// assert that no errors occurred
	if err != nil {
		log.Println("Error: ", err)
		log.Println("Value: ", bucket)
		return
	}

	completeness, _ := strconv.ParseBool(bucket["Complete"])
	id, _ := uuid.Parse(bucket["ID"])
	Mastertasklist.UpdateTask(&Options{ID: id, Title: bucket["Title"], Desc: bucket["Desc"], Complete: completeness})
}

func removeTaskByTitle(w http.ResponseWriter, r *http.Request) {
	// print the method type
	fmt.Println("method: ", r.Method)
	//Allow CORS here By * or specific origin
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:1080")
	w.Header().Set("Access-Control-Allow-Methods", "DELETE")

	// assert that request is post
	if r.Method != http.MethodDelete {
		return
	}

	// Get title from URL
	words, ok := r.URL.Query()["title"]
	log.Println(words)
	title := strings.Join(words, " ")
	// make sure that the info exists
	if !ok || len(title) < 1 {
		log.Println("Url Param 'title' is missing")
		return
	}
	log.Println("Title-Value:", title)

	// Get Task
	err := Mastertasklist.RemoveTaskByTitle(title)
	// check for error
	if err != nil {
		log.Println(err)
		return
	}
}

func removeTaskByID(w http.ResponseWriter, r *http.Request) {
	// print the method type
	fmt.Println("method: ", r.Method)
	//Allow CORS here By * or specific origin
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:1080")
	w.Header().Set("Access-Control-Allow-Methods", "DELETE")

	// assert that request is post
	if r.Method != http.MethodDelete {
		return
	}

	// Get title from URL
	words, ok := r.URL.Query()["title"]
	log.Println(words)
	rawID := strings.Join(words, "")
	// make sure that the info exists
	if !ok || len(rawID) < 1 {
		log.Println("Url Param 'id' is missing")
		return
	}

	// convert to type UUID
	id, err := uuid.Parse(rawID)
	if err != nil {
		log.Println(err)
		return
	}
	log.Println(id)

	// remove task
	err = Mastertasklist.RemoveTaskByID(id)
	// check for error
	if err != nil {
		log.Println(err)
		// construct task with error description
		log.Println("Task id \"" + id.String() + "\" not found")
		return
	}
}

// Takes url search info and plugs it into getTaskById()
// returns task back to requester
func requestTaskByTitle(w http.ResponseWriter, r *http.Request) {
	// print the method type
	fmt.Println("method: ", r.Method)
	//Allow CORS here By * or specific origin
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:1080")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Access-Control-Allow-Methods", "GET")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// assert that request is post
	if r.Method != http.MethodGet {
		return
	}

	// Get title from URL
	words, ok := r.URL.Query()["title"]
	log.Println(words)
	title := strings.Join(words, " ")
	// make sure that the info exists
	if !ok || len(title) < 1 {
		log.Println("Url Param 'title' is missing")
		return
	}
	log.Println(title)

	// Get Task
	task, err := Mastertasklist.GetTaskByTitle(title)
	// check for error
	if err != nil {
		log.Println(err)
		// construct task with error description
		message := "Task title \"" + title + "\" not found"
		task = &Task{uuid.UUID{}, message, message, time.Time{}, time.Time{}, false, nil, nil}
	}
	//log.Printf("TASK-DATA: %s", task.ToString())

	// Start Sending Task Back
	// convert task into json friendly TaskModel
	taskModel := task.toTaskModel()
	// convert taskModel into json
	data, err := json.Marshal(taskModel)
	// assert nothing went wrong
	if err != nil {
		log.Println(err)
		return
	}
	//log.Printf("JSON-Data: %s", data)
	// send task back
	w.Write(data)
}

// Takes url search info and plugs it into getTaskById()
// returns task back to requester
func requestTaskByID(w http.ResponseWriter, r *http.Request) {
	// print the method type
	fmt.Println("method: ", r.Method)
	//Allow CORS here By * or specific origin
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:1080")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Access-Control-Allow-Methods", "GET")

	// assert that request is get
	if r.Method != http.MethodGet {
		return
	}

	// Get title from URL
	words, ok := r.URL.Query()["id"]
	log.Println(words)
	rawID := strings.Join(words, "")
	// make sure that the info exists
	if !ok || len(rawID) < 1 {
		log.Println("Url Param 'id' is missing")
		return
	}

	// convert to type UUID
	id, err := uuid.Parse(rawID)
	if err != nil {
		log.Println(err)
		return
	}
	log.Println(id)

	// Get Task
	task, err := Mastertasklist.GetTaskByID(id)
	// check for error
	if err != nil {
		log.Println(err)
		// construct task with error description
		message := "Task ID \"" + id.String() + "\" not found"
		task = &Task{uuid.UUID{}, message, message, time.Time{}, time.Time{}, false, nil, nil}
	}
	//log.Printf("TASK-DATA: %s", task.ToString())

	// Start Sending Task Back
	// convert task into json friendly TaskModel
	taskModel := task.toTaskModel()
	// convert taskModel into json
	data, err := json.Marshal(taskModel)
	// assert nothing went wrong
	if err != nil {
		log.Println(err)
		return
	}
	//log.Printf("JSON-Data: %s", data)
	// send task back
	w.Write(data)
}

func requestAllTasks(w http.ResponseWriter, r *http.Request) {
	// print the method type
	fmt.Println("method: ", r.Method)
	//Allow CORS here By * or specific origin
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:1080")
	w.Header().Set("Access-Control-Allow-Methods", "GET")

	// assert that request is get
	if r.Method != http.MethodGet {
		return
	}

	// Get All Tasks
	tasks := Mastertasklist.GetAllTasks()

	// Start Sending Task Back
	// convert tasks into json friendly TaskModels
	taskModels := TaskArrayToTaskModels(tasks)
	// convert taskModels into json
	data, err := json.Marshal(taskModels)
	// assert nothing went wrong
	if err != nil {
		log.Println(err)
		return
	}
	//log.Printf("JSON-Data: %s", data)
	// send task back
	w.Write(data)
}

// master tasklist
var Mastertasklist TaskList = TaskList{nil, nil, 0}

const (
	FRONTEND_DIR = "./frontend"
)

func main() {
	// create waitgroup to manage listening
	wg := &sync.WaitGroup{}

	// This sets up the API as in the server requests
	wg.Add(1)
	go func() {
		log.Println("Listening 9090")
		// handler functions
		http.HandleFunc("/make", submitTask)
		http.HandleFunc("/requestTaskByTitle", requestTaskByTitle)
		http.HandleFunc("/requestTaskByID", requestTaskByID)
		http.HandleFunc("/requestAllTasks", requestAllTasks)
		http.HandleFunc("/submitChanges", submitChanges)
		http.HandleFunc("/removeTaskByTitle", removeTaskByTitle)
		http.HandleFunc("/removeTaskByID", removeTaskByID)
		// start server
		err := http.ListenAndServe(":9090", nil) // setting listening port
		if err != nil {
			fmt.Printf("%v", err)
		}
		wg.Done()
	}()

	// frontend, AKA html pages / user interface
	wg.Add(1)
	go func() {
		http.Handle("/", http.FileServer(http.Dir(FRONTEND_DIR)))
		log.Println("Listening 1080")
		err := http.ListenAndServe(":1080", nil)
		if err != nil {
			fmt.Printf("%v", err)
		}
		wg.Done()
	}()

	// initailize listening to both ports
	wg.Wait()

	log.Println("WTF the program stopped")

}
