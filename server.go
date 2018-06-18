package main

import (
	"./config"
	"./phpfpm"
	"log"
	"net/http"
	"strings"
	"sync"
)

func main() {
	conf := config.LoadConfig()
	var wg sync.WaitGroup
	for _, v := range conf.Server {
		wg.Add(1)
		mux := http.NewServeMux()
		mux.HandleFunc("/", createServerHandle(v.WebRoot, v.Index, conf.Php))
		for _, vhost := range v.Vhosts {
			mux.HandleFunc(vhost.Name+"/", createServerHandle(vhost.WebRoot, vhost.Index, conf.Php))
		}
		log.Println("server start port ", v.Port)
		go func(server config.Server) {
			var err error
			if server.SslEnabled == true {
				err = http.ListenAndServeTLS(server.Port, server.Cert, server.Key, mux)
			} else {
				err = http.ListenAndServe(server.Port, mux)
			}
			if err != nil {
				log.Fatal("ListenAndServe: ", err)
			}
			wg.Done()
		}(v)
	}
	wg.Wait()
}

func createServerHandle(webRoot string, index string, php config.Php) func(http.ResponseWriter, *http.Request) {
	fileServer := http.StripPrefix("/", http.FileServer(http.Dir(webRoot)))
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.Method + ": " + webRoot + r.URL.Path)
		if index != "index.html" && r.URL.Path[len(r.URL.Path)-1:] == "/" {
			r.URL.Path += index
		}
		if php.Enabled && strings.HasSuffix(r.URL.Path, ".php") {
			if r.Method == "GET" {
				phpfpm.Get(w, r, webRoot, php.FpmSock)
			} else if r.Method == "POST" {
				phpfpm.Post(w, r, webRoot, php.FpmSock)
			} else {
				phpfpm.OtherHttpMethod(w)
			}
		} else {
			fileServer.ServeHTTP(w, r)
		}
	}
}
