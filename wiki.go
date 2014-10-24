package main

import (
        "flag"
        "html/template"
        "io/ioutil"
        "log"
        "net"
        "net/http"
        // "regexp"
        "code.google.com/p/sre2/sre2"
        "errors"
)

var templates = template.Must(template.ParseFiles("edit.html", "view.html"))
// var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")                   // change this to sre2.MustParse(...)
var validPath = sre2.MustParse("^/(edit|save|view)/([a-zA-Z0-9]+)$")
var addr = flag.Bool("addr", false, "find open address and print to final-port.txt")

type Page struct {
        Title string
        Body []byte
}

func main() {
        http.HandleFunc("/view/", makeHandler(viewHandler))
        http.HandleFunc("/edit/", makeHandler(editHandler))
        http.HandleFunc("/save/", makeHandler(saveHandler))

        if *addr {
                l, err := net.Listen("tcp", "127.0.0.1:0")
                if err != nil {
                        log.Fatal(err)
                }
                err = ioutil.WriteFile("final-port.txt", []byte(l.Addr().String()), 0644)
                if err != nil {
                        log.Fatal(err)
                }
                s := &http.Server{}
                s.Serve(l)
                return
        }
        http.ListenAndServe(":8080", nil)
}

func loadPage(title string) (*Page, error) {
        filename := title + ".txt"
        body, err := ioutil.ReadFile(filename)
        if err != nil {
                return nil, err
        }
        return &Page{Title: title, Body: body}, nil
}

func (p *Page) save() error {
        filename := p.Title + ".txt"
        return ioutil.WriteFile(filename, p.Body, 0600) // 0600 is an octal integer literal signifies that the file create has read-write permissions current user only (unix).
}

func getTitle(w http.ResponseWriter, r *http.Request) (string, error) {
        u := r.URL.Path
        // m := validPath.FindStringSubmatch(u)           // regexp
        m := validPath.MatchIndex(u)
        if m == nil {
                http.NotFound(w, r)
                return "", errors.New("Invalid Page Title")
        }
        return string(u(m[2])), nil // The title is the second subexpression of the regular expression output. Needs to change to be used with sre2
}

func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
        p, err := loadPage(title)
        if err != nil {
                http.Redirect(w, r, "/edit/"+title, http.StatusFound)
                return
        }
        renderTemplate(w, "view", p)
}

func editHandler(w http.ResponseWriter, r *http.Request, title string) {
        p, err := loadPage(title)
        if err != nil {
                p = &Page{Title: title}
        }
        renderTemplate(w, "edit", p)
}

func saveHandler(w http.ResponseWriter, r *http.Request, title string) {
        body := r.FormValue("body")
        p := &Page{Title: title, Body: []byte(body)}
        err := p.save()
        if err != nil {
                http.Error(w, err.Error(), http.StatusInternalServerError)
                return
        }
        http.Redirect(w, r, "/view/"+title, http.StatusFound)
}

func makeHandler(fn func (http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
        return func(w http.ResponseWriter, r *http.Request) {
                u := r.URL.Path
                // m := validPath.FindStringSubmatch(u)   // regexp
                m := validPath.MatchIndex(u)
                if m == nil {
                        http.NotFound(w, r)
                        return
                }
                fn(w, r, string(u(m[2]))) // Will return an int with sre2.
        }
}

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
        err := templates.ExecuteTemplate(w, tmpl+".html", p)

        if err != nil {
                http.Error(w, err.Error(), http.StatusInternalServerError)
                return
        }
}
