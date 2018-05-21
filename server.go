package main

import(
 "net/http"
 "fmt"
 "strings"
 "log"
)
func main(){
  fileServer := http.StripPrefix("/static/",http.FileServer(http.Dir("static")))
  fmt.Println("server start port 443")
  err := http.ListenAndServeTLS(":443","/etc/letsencrypt/live/kurone.iceclover.net/fullchain.pem","/etc/letsencrypt/live/kurone.iceclover.net/privkey.pem", http.HandlerFunc( func(w http.ResponseWriter, r *http.Request){
    if strings.HasPrefix(r.URL.Path, "/static/"){
      fileServer.ServeHTTP(w,r)
    }else{
      fmt.Fprint(w, "Hello world!!")
    }
  }))
  if err != nil {
    log.Fatal("ListenAndServe: ", err)
  }
}

