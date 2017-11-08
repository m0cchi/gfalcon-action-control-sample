package main

import (
	"container/list"
	"fmt"

	"html/template"
	"net/http"
	"os"
	"strconv"
	"sync"
)

const MaxContentLength = 512
const MaxSize = 10

var templates *template.Template
var messages *list.List
var messagesLock *sync.Mutex

func init() {
	messages = list.New()
	messagesLock = new(sync.Mutex)
	templates = template.Must(template.ParseGlob("resources/templates/*"))
}

func pushMessage(message string) {
	messagesLock.Lock()
	if messages.Len() == MaxSize {
		messages.Remove(messages.Front())
	}
	messages.PushBack(message)
	messagesLock.Unlock()
}

func toSlice(messages *list.List) []string {
	ret := make([]string, 0, messages.Len())
	for e := messages.Back(); e != nil; e = e.Prev() {
		ret = append(ret, fmt.Sprintf("%v", e.Value))
	}
	return ret
}

func handle(w http.ResponseWriter, r *http.Request) {
	args := map[string]interface{}{}

	if r.Method == "POST" {
		// push message
		if r.ContentLength < MaxContentLength {
			r.ParseForm()
			message := r.Form.Get("message")
			pushMessage(message)
		} else {
			args["systemMessage"] = "too long message"
		}
	}

	// show message
	args["messages"] = toSlice(messages)

	if err := templates.ExecuteTemplate(w, "chat.html.tmpl", args); err != nil {
		fmt.Fprintf(w, "error")
	}
}

func main() {
	portstr := os.Getenv("PORT")
	port, err := strconv.Atoi(portstr)
	if err != nil {
		fmt.Println("export PORT=50000")
		os.Exit(1)
	}
	http.HandleFunc("/", handle)
	http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}
