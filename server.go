package main

import (
	"./config"
	"bytes"
	"github.com/tomasen/fcgi_client"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
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
			fcgi, err := fcgiclient.Dial("unix", php.FpmSock)
			if err != nil {
				log.Fatal("err@sock:", err)
			}
			env := make(map[string]string)
			var resp *http.Response
			env["SCRIPT_FILENAME"] = web_root + "/" + r.URL.Path
			if r.Method == "GET" {
				env["SERVER_SOFTWARE"] = "go / fcgiclient "
				env["REMOTE_ADDR"] = "127.0.0.1"
				env["QUERY_STRING"] = r.URL.RawQuery
				resp, err = fcgi.Get(env)
			} else if r.Method == "POST" {
				r.ParseForm()
				resp, err = fcgi.PostForm(env, r.Form)
			} else {
				resp = &http.Response{
					Body:       ioutil.NopCloser(bytes.NewBufferString("This http method is not supported.")),
				}
				w.WriteHeader(501)
			}
			if err != nil {
				log.Fatal("err@resp:", err)
			}
			content, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Fatal("err@body:", err)
			}
			if resp.Header.Get("Status") != "" {
				status_code, _ := strconv.Atoi(resp.Header.Get("Status")[:3])
				w.WriteHeader(status_code)
			}
			w.Write([]byte(content))
		} else {
			fileServer.ServeHTTP(w, r)
		}
	}
}
