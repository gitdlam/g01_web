package main

import (
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/go-chi/chi"

	"github.com/BurntSushi/toml"

	"log"
	"net/url"
	"os"

	"strings"

	//	"flag"
	//	"strconv"

	"github.com/vulcand/oxy/forward"
)

var global globalType

type globalType struct {
	appName string
	folder  string
	config  tomlConfig
}

type appEntry struct {
	Exe         string `json:"exe"`
	Port        string `json:"port"`
	SessionPort string
	pgCode      string `json:"pg_code"`

	Pathmap []struct {
		Path string
		Port string
	}
}

type tomlConfig struct {
	HttpPort    string `toml:"port"  json:"port"`
	SessionPort string
	PgCode      string `toml:"pg_code"     json:"pg_code"`
}

func init() {
	ex, _ := os.Executable()

	global.folder, global.appName = filepath.Split(ex)
	global.appName = strings.Split(global.appName, ".exe")[0]
	global.folder = global.folder[0 : len(global.folder)-1]

	f, err := os.OpenFile(global.folder+"/log.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening log file: %v", err)
	}

	log.SetOutput(f)
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	var configFile struct{ Config_file string }
	if _, err := toml.DecodeFile(global.folder+"/config.toml", &configFile); err != nil {
		log.Printf("%s", err)
		return
	}
	var configFile2 struct{ Apps []appEntry }

	if _, err := toml.DecodeFile(configFile.Config_file, &configFile2); err != nil {
		log.Printf("%s, %s", configFile.Config_file, err)
		return
	}

	for _, v := range configFile2.Apps {
		if v.Exe == global.appName {
			global.config.SessionPort = v.SessionPort
			global.config.HttpPort = v.Port
			global.config.PgCode = v.pgCode
			//			pathMapNormal = map[string]string{}
			//			pathMapWithLocking.m = map[string]string{}
			//			pathMapWithLocking.Lock()
			for _, v2 := range v.Pathmap {

				pathMap.Store(v2.Path, v2.Port)
				//				pathMapNormal[v2.Path] = v2.Port
				//				pathMapWithLocking.m[v2.Path] = v2.Port
			}
			//			pathMapWithLocking.Unlock()

		}

	}

}

func refreshRoutes() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var configFile struct{ Config_file string }
		if _, err := toml.DecodeFile(global.folder+"/config.toml", &configFile); err != nil {
			log.Println("%s", err)
			fmt.Fprint(w, "problem refreshing routes")
			return
		}
		var configFile2 struct{ Apps []appEntry }
		if _, err := toml.DecodeFile(configFile.Config_file, &configFile2); err != nil {
			log.Println("%s", err)
			fmt.Fprint(w, "problem refreshing routes")
			return

		}

		for _, v := range configFile2.Apps {
			if v.Exe == global.appName {

				for _, v2 := range v.Pathmap {
					pathMap.Store(v2.Path, v2.Port)
					//log.Println(v2.Path, v2.Port)

				}
			}

		}

		fmt.Fprint(w, "Refreshed")

	}

}

func createFunc(port string) func(http.ResponseWriter, *http.Request) {

	fn := func(w http.ResponseWriter, req *http.Request) {
		/*		if len(req.RequestURI) >= 3 && req.RequestURI[0:3] == "/g/" {
								user := req.Header.Get("remote_user")
								log.Println("redirected check:", user)
				}
		*/
		fwd, _ := forward.New()

		req.URL, _ = url.ParseRequestURI("http://127.0.0.1:" + port + req.RequestURI)
		//log.Println("http://127.0.0.1:" + port + req.RequestURI)
		fwd.ServeHTTP(w, req)

	}
	return fn
}

func createFunc2() func(http.ResponseWriter, *http.Request) {

	fn := func(w http.ResponseWriter, req *http.Request) {
		fwd, _ := forward.New()

		port := pathMap.MatchPort(req.RequestURI)

		if port != "" {
			req.URL, _ = url.ParseRequestURI("http://127.0.0.1:" + port + req.RequestURI)
			//			log.Println("http://127.0.0.1:" + port + req.RequestURI)

		}
		fwd.ServeHTTP(w, req)

	}
	return fn
}

func createFunc3(r *chi.Mux) func(http.ResponseWriter, *http.Request) {

	return func(w http.ResponseWriter, req *http.Request) {

		port := pathMapWithLocking.MatchPort(req.RequestURI)

		if port != "" {
			req.URL, _ = url.ParseRequestURI("http://127.0.0.1:" + port + req.RequestURI)
			//			log.Println("http://127.0.0.1:" + port + req.RequestURI)
			fwd, _ := forward.New()
			fwd.ServeHTTP(w, req)

		} else {

			r.ServeHTTP(w, req)
		}

	}

}
