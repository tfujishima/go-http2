package main

import(
 "net/http"
 "fmt"
 "log"
)
func handle(w http.ResponseWriter, r *http.Request){
  fmt.Fprint(w, "Hello world!!")
}
func main(){
  http.HandleFunc("/",handle)
  fmt.Println("server start port 443")
  err := http.ListenAndServeTLS(":443","/etc/letsencrypt/live/kurone.iceclover.net/fullchain.pem","/etc/letsencrypt/live/kurone.iceclover.net/privkey.pem",nil)
  if err != nil {
    log.Fatal("ListenAndServe: ", err)
  }
}

