package main

import (
    "fmt"
    "net/http"
    "html/template"
    "path"
    "github.com/julienschmidt/httprouter"
    "github.com/russross/blackfriday"
    "sync"
    "database/sql"
    _ "github.com/mattn/go-sqlite3"
    "log"
    "strconv"
)

type EditContent struct{
  ID string
  Content string
}

var DB *sql.DB
var lck sync.Mutex
var id int

func newDB() (*sql.DB, error) {
    db, err := sql.Open("sqlite3", "posts.db")
    if err != nil {
        log.Println(err)
        return nil, err
    }

    q := `CREATE TABLE IF NOT EXISTS
      posts(id intenger primary key not null, content text);
      `
    if _, err := db.Exec(q); err != nil {
        log.Println(err)
        return nil, err
    }

    return db, nil
}


func main() {
  var err error
  if DB, err = newDB(); err != nil {
    return
  }

  err = DB.QueryRow("SELECT MAX(id) from posts", id).Scan(&id)
  if err != nil {
    id = 0
    log.Printf("got the following error trying to get max id: %v", err)
  }


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



  rows, err := DB.Query("SELECT id from POSTS")
    if err != nil {
      log.Println(err)
      return
    }
    defer rows.Close()
    var mdfiles []string
    for rows.Next(){
      var id int
      if err := rows.Scan(&id); err != nil{
        log.Println(err)
        continue
      }
      mdfiles = append(mdfiles, fmt.Sprintf("%v", id))
    }

    if err := rows.Err(); err != nil{
      log.Println(err)
    }

  // files, err := ioutil.ReadDir("posts")
  // if err != nil {
  //   return
  // }

  // var mdfiles []string
  // for _,file := range files {
  //   name := file.Name()
  //   if strings.HasSuffix(name,".md") {
  //     name = name[:len(name)-3]
  //   }
  //   mdfiles = append(mdfiles,name)
  // }


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
  markdown := r.FormValue("body")


  // log.Printf("post: body: %v", markdown)
  lck.Lock()
	id++
	var i = id
	lck.Unlock()
  _, err := DB.Exec("INSERT INTO posts (id, content) VALUES(?, ?)", i, markdown)
    if err != nil {
      log.Printf("Got the following error trying to save content to db: %v, %v, %v", i, markdown, err)
      return
    }

  // md_input := r.FormValue("body")
  fp := path.Join("public", "create.html")
	// lck.Lock()
	// id++
	// var i = id
	// lck.Unlock()


	// f, err := os.Create(fmt.Sprintf("posts/%v.md", i))
	// if err != nil {
	// 	// don't ignore the error like I'm doing here.
	// 	return
	// }

	// fmt.Fprintln(f, string(md_input))
	// f.Close()

  htmlout := blackfriday.MarkdownCommon([]byte(markdown))


  tmpl, err := template.ParseFiles(fp)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

  if err := tmpl.Execute(rw, template.HTML(htmlout)); err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	}
}

func postShowHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
  fp := path.Join("public", "show.html")
  id := p.ByName("id")
  iid, err := strconv.Atoi(id)
    if err != nil{
      log.Printf("got bad id %v: %v", id, err)
      return
    }

  var contents string
  if err := DB.QueryRow("select content from posts where id = ?", iid).Scan(&contents);
    err != nil{
      log.Printf("got bad id %v: %v", id, err)
      return
  }

  htmlout := blackfriday.MarkdownBasic([]byte(contents))

  /*contents, err := ioutil.ReadFile("posts/"+id+".md")
    if err != nil{
      fmt.Fprintln(rw, "Post does not exist")
      return
    }

    htmlout := blackfriday.MarkdownBasic(contents)
    // rw.Write(htmlout)

    */

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
  iid, err := strconv.Atoi(id)
    if err != nil{
      log.Printf("got bad id %v: %v", id, err)
      return
    }

  content := r.FormValue("body")
  _, err = DB.Exec("update posts set content=? where id=?", content, iid)
    if err != nil{
      log.Printf("failed to update %v: %v", id, err)
      return
    }

  htmlout := blackfriday.MarkdownBasic([]byte(content))

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
    iid, err := strconv.Atoi(id)
      if err != nil{
        log.Printf("got bad id %v: %v", id, err)
        return
      }

    var contents string
    if err := DB.QueryRow("select content from posts where id = ?", iid).Scan(&contents);
    err != nil{
        log.Printf("got bad id %v: %v", id, err)
        return
    }

    /*
    markdownFilename := path.Join("posts", id+".md")
    contents, err := ioutil.ReadFile(markdownFilename)

    if err != nil{
      //Assume that the file does not exist
      fmt.Fprintln(rw, id+".md")
      return
    }
    */

    editContent := EditContent{
      ID : id,
      Content : string(contents),
    }

    tmpl, err := template.ParseFiles(fp)
  	if err != nil {
  		http.Error(rw, err.Error(), http.StatusInternalServerError)
  		return
  	}

    if err := tmpl.Execute(rw, editContent); err != nil {
  		http.Error(rw, err.Error(), http.StatusInternalServerError)
  	}

}
