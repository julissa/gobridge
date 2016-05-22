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

type EditContent struct{
  ID string
  Content string
}

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
	r.POST("/posts/:id", postUpdateHandler)
	r.GET("/posts/:id/edit", postEditHandler)

	fmt.Println("Starting server on :9090")
	http.ListenAndServe(":9090", r)
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
    if strings.HasSuffix(name,".md") {
      name = name[:len(name)-3]
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
  md_input := r.FormValue("body")
  fp := path.Join("public", "create.html")
	lck.Lock()
	id++
	var i = id
	lck.Unlock()
	f, err := os.Create(fmt.Sprintf("posts/%v.md", i))
	if err != nil {
		// don't ignore the error like I'm doing here.
		return
	}
	fmt.Fprintln(f, string(md_input))
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
  fp := path.Join("public", "show.html")
    id := p.ByName("id")
    contents, err := ioutil.ReadFile("posts/"+id+".md")
    if err != nil{
      fmt.Fprintln(rw, "Post does not exist")
      return
    }

    htmlout := blackfriday.MarkdownBasic(contents)
    // rw.Write(htmlout)

    tmpl, err := template.ParseFiles(fp)
  	if err != nil {
  		http.Error(rw, err.Error(), http.StatusInternalServerError)
  		return
  	}

    if err := tmpl.Execute(rw, template.HTML(htmlout)); err != nil {
  		http.Error(rw, err.Error(), http.StatusInternalServerError)
  	}

}

func postUpdateHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
  fp := path.Join("public", "update.html")
  id := p.ByName("id")
  markdownFilename := path.Join("posts", id+".md")
  body := r.FormValue("body")

  f, err := os.Create(markdownFilename)
	if err != nil {
		// don't ignore the error like I'm doing here.
		return
	}
	fmt.Fprintln(f, body)
	f.Close()
  htmlout := blackfriday.MarkdownBasic([]byte(body))

  tmpl, err := template.ParseFiles(fp)
  if err != nil {
    http.Error(rw, err.Error(), http.StatusInternalServerError)
    return
  }

  if err := tmpl.Execute(rw, template.HTML(htmlout)); err != nil {
    http.Error(rw, err.Error(), http.StatusInternalServerError)
  }


}

func postDeleteHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
    fmt.Fprintln(rw, "post delete")
}

func postEditHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
    fp := path.Join("public", "edit.html")
    id := p.ByName("id")
    markdownFilename := path.Join("posts", id+".md")
    contents, err := ioutil.ReadFile(markdownFilename)

    if err != nil{
      //Assume that the file does not exist
      fmt.Fprintln(rw, id+".md")
      return
    }

    tmpl, err := template.ParseFiles(fp)
  	if err != nil {
  		http.Error(rw, err.Error(), http.StatusInternalServerError)
  		return
  	}

    editContent := EditContent{
      ID : id,
      Content : string(contents),
    }

    if err := tmpl.Execute(rw, editContent); err != nil {
  		http.Error(rw, err.Error(), http.StatusInternalServerError)
  	}

}
