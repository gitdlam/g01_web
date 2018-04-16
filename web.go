package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	//	"net/url"
	"os"
	"strings"
	"time"

	_ "github.com/lib/pq"

	"github.com/buger/jsonparser"
	"github.com/gitdlam/common"
	"github.com/go-chi/chi"
	//	"github.com/go-chi/chi/middleware"
	"github.com/gitdlam/authenticator"
	"github.com/justinas/alice"
	//"github.com/vulcand/oxy/forward"
)

func pingResponse(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, global.appName)

}

func folderResponse(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, global.folder)

}

func terminate(w http.ResponseWriter, req *http.Request) {
	os.Exit(0)
}

func sso(w http.ResponseWriter, req *http.Request) {
	if len(req.RequestURI) >= 3 && req.RequestURI[0:3] == "/g/" {
		//		user := req.Context().Value("sso").(string)
		user := req.Header.Get("remote_user")

		fmt.Fprintln(w, strings.ToUpper(user))

	}
}

func randomString() string {
	rand.Seed(time.Now().UnixNano())
	var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, 3)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
func do(w http.ResponseWriter, req *http.Request) {
	m := extractSessionInfo(req)
	log.Println(m["session_id"])
	updateStatusBegan(m["session_id"])
	//doWork(m)
	updateStatusFinished(m["session_id"])
	fmt.Fprintf(w, global.appName)

}

func extractSessionInfo(req *http.Request) map[string]string {

	b, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Fatal(err.Error())
	}
	m := make(map[string]string)
	if s, err := jsonparser.GetString(b, "session_id"); err != nil {
		m["session_id"] = ""
	} else {
		m["session_id"] = s
	}

	if s, err := jsonparser.GetString(b, "finished"); err != nil {
		m["finished"] = ""
	} else {
		m["finished"] = s
	}

	if s, err := jsonparser.GetString(b, "began"); err != nil {
		m["began"] = ""
	} else {
		m["began"] = s
	}

	if s, err := jsonparser.GetString(b, "task_name"); err != nil {
		m["task_name"] = ""
	} else {
		m["task_name"] = s
	}

	if s, err := jsonparser.GetString(b, "brcd"); err != nil {
		m["brcd"] = ""
	} else {
		m["brcd"] = s
	}

	if s, err := jsonparser.GetString(b, "printer"); err != nil {
		m["printer"] = ""
	} else {
		m["printer"] = s
	}

	if s, err := jsonparser.GetString(b, "ref"); err != nil {
		m["ref"] = ""
	} else {
		m["ref"] = s
	}

	return m
}

func updateStatusBegan(session_id string) {
	var jsonStr = []byte(`{"session_id":"` + session_id + `","began":"` + common.TimeNowString() + `"}`)
	req, err := http.NewRequest("POST", "http://127.0.0.1:"+global.config.SessionPort+"/updateStatusBegan", bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

}

func updateStatusFinished(session_id string) {
	var jsonStr = []byte(`{"session_id":"` + session_id + `","finished":"` + common.TimeNowString() + `"}`)
	req, err := http.NewRequest("POST", "http://127.0.0.1:"+global.config.SessionPort+"/updateStatusFinished", bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

}

func HTTPServe() {
	r := chi.NewRouter()

	r.Get("/ping", pingResponse)
	r.Get("/folder", folderResponse)
	FileServer(r, "/f/static/", http.Dir(global.folder+"/static"))

	for _, v := range global.config.Pathmap {
		r.HandleFunc(v.Path+"*", global.funcMap[v.Path])
	}
	r.HandleFunc("/g/sso", sso)
	r.HandleFunc("/g/sso/name", sso)

	r.Get("/terminate/"+global.config.SessionPort, terminate)
	r.Post("/do", do)

	mw := alice.New(authenticator.Authenticator2).Then(r)

	server := &http.Server{Addr: ":" + global.config.HttpPort, Handler: mw}

	//	server.SetKeepAlivesEnabled(true)

	server.ListenAndServe()
	//http.ListenAndServe(":"+global.config.HttpPort, mw)

}

func getHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Success"))
	return
}

func dbConnection() *sql.DB {
	db, err := sql.Open("postgres", fmt.Sprintf("user=data dbname=common password=%s sslmode=disable", global.config.PgCode))
	if err != nil {
		log.Panicln("Cannot connect to DB")
	}
	return db

}

func addTitle(title string, t string) string {
	if t[:6] == "<html>" {
		tt := strings.Split(t, "\n")
		t = strings.Join(tt[1:len(tt)-1], "\n")
	}
	return `<html> <head>
<title></title>
<link rel="stylesheet" type="text/css" href="/f/media/style.css"/>

<style> br { display: block; line-height: 20px;} body {margin:5px}</style>  

</head>


<body>
<div class="container">
<H4>` + title + `</H4>
<div id="responsecontainer">` + t + `
</div>


</div>
</div>
</body>
</html>`
}

func FileServer(r chi.Router, path string, root http.FileSystem) {
	if strings.ContainsAny(path, "{}*") {
		// panic("FileServer does not permit URL parameters.")
		return
	}

	fs := http.StripPrefix(path, http.FileServer(root))

	if path != "/" && path[len(path)-1] != '/' {
		r.Get(path, http.RedirectHandler(path+"/", 301).ServeHTTP)
		path += "/"
	}
	path += "*"

	r.Get(path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fs.ServeHTTP(w, r)
	}))
}
