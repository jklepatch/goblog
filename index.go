package main

import (
  "os"
  "fmt"
  "bytes"
  "strings"
  "path"
  "path/filepath"
  "net/http"
  "io/ioutil"
  "html/template"
  "github.com/yuin/goldmark"
  "github.com/yuin/goldmark/parser"
  "github.com/yuin/goldmark-meta"
)

var URL string = ":8080"
var MARKDOWN_PATH string = "markdown"
var TEMPLATE_PATH string = "templates"
var INDEX string = "index"
var NOT_FOUND string = "404"
var markdown goldmark.Markdown
var templates map[string]*template.Template = map[string]*template.Template{}

type Page struct {
  Title string
  Description string
  Body template.HTML
}

func toPath(file string) string {
  parts := strings.Split(file, ".")
  return path.Join(MARKDOWN_PATH, parts[0] + ".md") 
}

func handler(w http.ResponseWriter, r *http.Request) {
  // Get the markdown source
  file := path.Base(r.URL.Path)
  if file == "/" {
    file = INDEX 
  }
  source, err := ioutil.ReadFile(toPath(file))
  if err != nil {
    source, err = ioutil.ReadFile(toPath(NOT_FOUND))
    if err != nil {
      panic(err)
    }
  }

  //Parse markdown
  var buf bytes.Buffer
  context := parser.NewContext()
  if err := markdown.Convert(
    []byte(source), 
    &buf, 
    parser.WithContext(context)); err != nil {
      panic(err)
  }

  //Render html and send it
  metadata := meta.Get(context)
  title := metadata["title"]
  description := metadata["title"]
  templateName := metadata["template"]
  page := &Page{
    Title: fmt.Sprintf("%v", title),
    Description: fmt.Sprintf("%v", description),
    Body: template.HTML(buf.String()),
  }
  if tpl, ok := templates[fmt.Sprintf("%v", templateName)]; ok {
    tpl.Execute(w, page)
  } else {
    fmt.Fprint(w, "Oooch... internal server error - template does not exist")
  }
}

func main() {
  err := filepath.Walk(
    TEMPLATE_PATH, 
    func(fullPath string, info os.FileInfo, err error) error {
       file := path.Base(fullPath)
       parts := strings.Split(file, ".")
       //skip folder name
       if len(parts) == 2 {
         templates[parts[0]] = template.Must(
           template.ParseFiles(fullPath))
       }
       return nil
  })
  if(err != nil) {
    panic(err)
  }

  markdown = goldmark.New(
      goldmark.WithExtensions(
          meta.Meta,
      ),
  )
  http.HandleFunc("/", handler)
  fmt.Println("server is running on url " + URL)
  http.ListenAndServe(URL, nil)
}
