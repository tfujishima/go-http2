package main

import(
 "net/http"
 "fmt"
 "strings"
 "log"
 "io/ioutil"
 "github.com/tomasen/fcgi_client"
)
func main(){

  webroot := "/var/www/html"
  index := "/index.php"

  fileServer := http.StripPrefix("/",http.FileServer(http.Dir(webroot)))
  fmt.Println("server start port 443")
  err := http.ListenAndServeTLS(":443","/etc/letsencrypt/live/kurone.iceclover.net/fullchain.pem","/etc/letsencrypt/live/kurone.iceclover.net/privkey.pem", http.HandlerFunc( func(w http.ResponseWriter, r *http.Request){
    pos := strings.LastIndex(r.URL.Path, ".")
    if pos == -1 {
      r.URL.Path = index
      pos = strings.LastIndex(index, ".")
    }
    if r.URL.Path[pos:] == ".php"{
      fcgi, err := fcgiclient.Dial("unix", "/var/run/php-fpm/php-fpm.sock")
      if err != nil {
        log.Println("err@sock:", err)
      }
      log.Println(webroot + r.URL.Path)
      env := make(map[string]string)
      var resp *http.Response
      env["SCRIPT_FILENAME"] = webroot + r.URL.Path
      if r.Method == "GET" {
        env["SERVER_SOFTWARE"] = "go / fcgiclient "
        env["REMOTE_ADDR"] = "127.0.0.1"
        env["QUERY_STRING"] = r.URL.RawQuery
        resp, err = fcgi.Get(env)
      }else if r.Method == "POST"{
        log.Println(r.URL.RawQuery)
        r.ParseForm()
        resp, err = fcgi.PostForm(env, r.Form)
      }
      if err != nil {
        log.Println("err@resp:", err)
      }
      content, err := ioutil.ReadAll(resp.Body)
      if err != nil {
        log.Println("err@body:", err)
      }
      fmt.Fprint(w,string(content))
    }else{
      fileServer.ServeHTTP(w,r)
    }
  }))
  if err != nil {
    log.Fatal("ListenAndServe: ", err)
  }
}

