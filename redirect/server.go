package redirect

import (
	"html/template"
	"net/http"
	"os"
	"path"
)

func handler(w http.ResponseWriter, r *http.Request) {
	wd, _ := os.Getwd()
	templateFile := path.Join(wd, "gc-link", "redirect", "template.html")
	if _, err := os.Stat(templateFile); err != nil {
		templateFile = path.Join(wd, "redirect", "template.html")
	}

	q := r.URL.Query()
	t, _ := template.ParseFiles(templateFile)

	t.Execute(w, q)
}

func Serve(addr string) {
	http.HandleFunc("/", handler)
	http.ListenAndServe(addr, nil)
}
