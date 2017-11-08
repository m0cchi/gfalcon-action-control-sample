package main

import (
	"container/list"
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/m0cchi/gfalcon/model"
	"html/template"
	"net/http"
	"os"
	"strconv"
	"sync"
)

const ServiceName = "sahohime"
const MaxContentLength = 512
const MaxSize = 10

const PostActionID = "postable"

var PostAction *model.Action

var templates *template.Template
var messages *list.List
var messagesLock *sync.Mutex
var idp string
var db *sqlx.DB

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

func makeResponse(args map[string]interface{}, w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		// push message
		if r.ContentLength < MaxContentLength {
			r.ParseForm()
			message := r.Form.Get("message")
			pushMessage(message)
		} else {
			args["systemMessage"] = args["systemMessage"].(string) + "\ntoo long message"
		}
	}

	// show message
	args["messages"] = toSlice(messages)

	if err := templates.ExecuteTemplate(w, "chat.html.tmpl", args); err != nil {
		fmt.Fprintf(w, "error")
	}
}

func check(r *http.Request) error {
	cookie, err := r.Cookie("gfalcon.session")
	if err != nil {
		return errors.New("you must SignIn")
	}
	value := cookie.Value
	session, err := model.GetSession(db, value)
	if err = session.Validate(); err != nil {
		return errors.New("you must SignIn")
	}

	if r.Method == "POST" {
		user := &model.User{IID: session.UserIID}
		actionLink, err := model.GetActionLink(db, PostAction, user)
		if err != nil || actionLink.Validate() != nil {
			return errors.New("you don't have the authority to post message")
		}
	}
	return nil
}

func handle(w http.ResponseWriter, r *http.Request) {
	args := map[string]interface{}{
		"systemMessage": "",
		"IdP":           idp,
	}

	if err := check(r); err != nil {
		args["systemMessage"] = fmt.Sprintf("%v", err)
		r.Method = "!POST" // failed
	}

	makeResponse(args, w, r)
}

func initGfalcon() error {
	service, err := model.GetService(db, ServiceName)
	if err != nil {
		return err
	}

	PostAction, err = model.GetAction(db, service.IID, PostActionID)
	return err
}

func main() {
	portstr := os.Getenv("PORT")
	datasource := os.Getenv("DATASOURCE")
	idp = os.Getenv("IDP")
	port, err := strconv.Atoi(portstr)
	if err != nil {
		fmt.Println("export PORT=50000")
		os.Exit(1)
	}
	if datasource == "" {
		fmt.Println("not specify datasource")
		os.Exit(1)
	}
	db, err = sqlx.Connect("mysql", datasource)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := initGfalcon(); err != nil {
		fmt.Printf("err: %v\n", err)
		os.Exit(1)
	}

	http.HandleFunc("/", handle)
	http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}
