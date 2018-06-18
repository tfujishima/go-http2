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
		mux.HandleFunc("/", create_server_handle(v.WebRoot, v.Index, conf.Php))
		for _, vhost := range v.Vhosts {
			mux.HandleFunc(vhost.Name+"/", create_server_handle(vhost.WebRoot, vhost.Index, conf.Php))
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

func create_server_handle(web_root string, index string, php config.Php) func(http.ResponseWriter, *http.Request) {
	fileServer := http.StripPrefix("/", http.FileServer(http.Dir(web_root)))
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.Method + ": " + web_root + r.URL.Path)
		if index != "index.html" && r.URL.Path[len(r.URL.Path)-1:] == "/" {
			r.URL.Path += index
		}
		if php.Enabled && strings.HasSuffix(r.URL.Path, ".php") {
			if r.Method == "GET" {
				phpfpm.Get(w, r, web_root, php.FpmSock)
			} else if r.Method == "POST" {
				phpfpm.Post(w, r, web_root, php.FpmSock)
			} else {
				phpfpm.OtherHttpMethod(w)
			}
		} else {
			fileServer.ServeHTTP(w, r)
		}
	}
}
