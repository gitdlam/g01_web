package main

import (
	"net/http"

	"github.com/BurntSushi/toml"
	"github.com/kardianos/osext"

	"log"
	"net/url"
	"os"

	"os/signal"

	"strings"
	"sync"
	//	"flag"
	//	"strconv"
	"github.com/vulcand/oxy/forward"
)

var global globalType

type mFn func(http.ResponseWriter, *http.Request)

type globalType struct {
	appName string
	folder  string
	config  tomlConfig

	funcMap map[string]func(http.ResponseWriter, *http.Request)
}

type global_entry struct {
	Exe         string `json:"exe"`
	Config_file string `json:"port"`
}

type appEntry struct {
	Exe          string `json:"exe"`
	Port         string `json:"port"`
	Session_port string
	Pg_code      string `json:"pg_code"`
	Args         string `json:"args"`
	Pathmap      []struct {
		Path string
		Port string
	}
}

type tomlConfig struct {
	sync.RWMutex
	HttpPort    string `toml:"port"  json:"port"`
	SessionPort string
	PgCode      string `toml:"pg_code"     json:"pg_code"`
	Apps        []appEntry
	Pathmap     []struct {
		Path string
		Port string
	}
}

func configure() {

	// os.Args[0] is name of the executable
	// remove .exe to get the name
	folder, _ := osext.ExecutableFolder()
	might_add_one := 0

	if folder != "" {
		might_add_one = 1
	}
	path, _ := osext.Executable()
	global.folder = folder

	s := strings.Split(path[len(folder)+might_add_one:], ".exe")
	global.appName = s[0]
	global.funcMap = map[string]func(http.ResponseWriter, *http.Request){}

	global.config.Lock()

	var configFile struct{ Config_file string }
	if _, err := toml.DecodeFile(folder+"/config.toml", &configFile); err != nil {
		log.Printf("%s", err)
		return
	}

	if _, err := toml.DecodeFile(configFile.Config_file, &global.config); err != nil {
		log.Printf("%s, %s", configFile.Config_file, err)
		return
	}

	for _, v := range global.config.Apps {
		if v.Exe == global.appName {
			global.config.SessionPort = v.Session_port
			global.config.HttpPort = v.Port
			global.config.PgCode = v.Pg_code

			for _, v2 := range v.Pathmap {
				global.funcMap[v2.Path] = createFunc(v2.Port)
				global.config.Pathmap = append(global.config.Pathmap, v2)
			}
		}

	}

	global.config.Unlock()
}

func waitForQuit() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for range c {
			log.Printf("Quit")
			os.Exit(1)
		}
	}()

}

func createFunc(port string) func(http.ResponseWriter, *http.Request) {

	fn := func(w http.ResponseWriter, req *http.Request) {
		if len(req.RequestURI) >= 3 && req.RequestURI[0:3] == "/g/" {
			//user := req.Context().Value("sso").(string)
			user := req.Header.Get("sso")
			log.Println("redirected check:", user)
		}

		fwd, _ := forward.New()
		req.URL, _ = url.ParseRequestURI("http://127.0.0.1:" + port + req.RequestURI)
		log.Println("http://127.0.0.1:" + port + req.RequestURI)
		fwd.ServeHTTP(w, req)

	}
	return fn
}
