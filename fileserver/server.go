package main

import (
    "fmt"
    "net/http"
    "html/template"
    "path"
    "github.com/julienschmidt/httprouter"
    "github.com/russross/blackfriday"
    "sync"
    "os"
    "io/ioutil"
    "strings"
)

var lck sync.Mutex
var id int

func main() {
    r := httprouter.New()
    r.GET("/", homeHandler)
    r.ServeFiles("/css/*filepath", http.Dir("public/css"))

    // Posts collection
    r.GET("/posts", postsIndexHandler)
    r.POST("/posts", postsCreateHandler)

    // Posts singular
    r.GET("/posts/:id", postShowHandler)
    r.PUT("/posts/:id", postUpdateHandler)
    r.GET("/posts/:id/edit", postEditHandler)

    fmt.Println("Starting server on :8080")
    http.ListenAndServe(":8080", r)
}

func homeHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
    fmt.Fprintln(rw, "Home")
}

func postsIndexHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {

	fp := path.Join("public", "index.html")

  files, err := ioutil.ReadDir("posts")
  if err != nil {
    return
  }

  var mdfiles []string
  for _,file := range files {
    name := file.Name()
    if strings.HasSuffix(name,".html") {
      name = name[:len(name)-5]
    }
    mdfiles = append(mdfiles,name)
  }


	tmpl, err := template.ParseFiles(fp)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := tmpl.Execute(rw, mdfiles); err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	}

}

func postsCreateHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {

	markdown := blackfriday.MarkdownCommon([]byte(r.FormValue("body")))
  fp := path.Join("public", "show.html")
	lck.Lock()
	id++
	var i = id
	lck.Unlock()
	f, err := os.Create(fmt.Sprintf("posts/%v.html", i))
	if err != nil {
		// don't ignore the error like I'm doing here.
		return
	}
	fmt.Fprintln(f, string(markdown))
	f.Close()

  tmpl, err := template.ParseFiles(fp)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

  if err := tmpl.Execute(rw, template.HTML(markdown)); err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	}
}

func postShowHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
    id := p.ByName("id")
    fmt.Fprintln(rw, "showing post", id)
}

func postUpdateHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
    fmt.Fprintln(rw, "post update")
}

func postDeleteHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
    fmt.Fprintln(rw, "post delete")
}

func postEditHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
    fmt.Fprintln(rw, "post edit")
}
