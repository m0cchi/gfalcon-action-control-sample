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
var AuthError = errors.New("you must SignIn")
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

func handle(args map[string]interface{}, w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" && args["postable"].(bool) {
		// push message
		if r.ContentLength < MaxContentLength {
			r.ParseForm()
			message := r.Form.Get("message")
			pushMessage(message)
		} else {
			args["systemMessage"] = "\ntoo long message"
		}
	}

	// show message
	args["messages"] = toSlice(messages)

	if err := templates.ExecuteTemplate(w, "chat.html.tmpl", args); err != nil {
		fmt.Fprintf(w, "error")
	}
}

func check(r *http.Request) error {
	sessionID, err := r.Cookie("gfalcon.session")
	if err != nil {
		return AuthError
	}
	userIID, err := r.Cookie("gfalcon.iid")
	if err != nil {
		return AuthError
	}
	IID, err := strconv.ParseUint(userIID.Value, 10, 32)
	if err != nil {
		return AuthError
	}

	session, err := model.GetSession(db, uint32(IID), sessionID.Value)
	if err = session.Validate(); err != nil {
		return AuthError
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

func withCheck(handle func(map[string]interface{}, http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		args := map[string]interface{}{
			"systemMessage": "",
			"IdP":           idp,
			"postable":      true,
		}

		if err := check(r); err != nil {
			args["systemMessage"] = fmt.Sprintf("%v", err)
			args["postable"] = false
		}
		handle(args, w, r)
	}
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

	http.HandleFunc("/", withCheck(handle))
	http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}
