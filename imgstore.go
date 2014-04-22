package main

import (
	"net/http"
	"html/template"
	"regexp"
	"errors"
	"fmt"
	"image"
	_ "image/jpeg"
	"log"
	"strings"
	"strconv"
	"github.com/mmigacz/imgstore/store"
	"os"
)



var templates = template.Must(template.ParseFiles("./web/upload.html", "./web/error.html"))


func check(err error) {
	if err != nil {
		log.Printf("error: %s", err)
		panic(err)
	}
}



//
// upload
//
func fileName(fullName string) string {
	li := strings.LastIndex(fullName, ".")
	name := fullName

	if li >=0 {
		name = fullName[0:li]
	}

	return store.NormOrignalImgName(name)
}


func upload(st *store.Store) func (http.ResponseWriter, *http.Request) {
	return func (w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			err := templates.ExecuteTemplate(w, "upload.html", map[string]interface{} {
				"Files": st.ListImgNames(),
			})

			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}

		f, fh, err := r.FormFile("image")
		check(err)
		defer f.Close()

		img, _, err := image.Decode(f)
		check(err)

		log.Printf("uploaded file %s", fh.Filename)

		err = st.StoreImg(fileName(fh.Filename), img)
		check(err)

//		http.Redirect(w, r, "/upload", 302)
	}
}



//
// view
//
var imgNameValidPath = regexp.MustCompile("^/(edit|save|view)/(.+)\\??.*$")

func parseDim(s string) uint {
	dim, err := strconv.ParseUint(s, 10, 0)
	if err != nil {
		dim = 0
	}
	return uint(dim)
}

func getImgName(w http.ResponseWriter, r *http.Request) (string, uint, uint, error) {
	m := imgNameValidPath.FindStringSubmatch(r.URL.Path)

	width := parseDim(r.URL.Query().Get("w"))
	height := parseDim(r.URL.Query().Get("h"))

	if m == nil {
		http.NotFound(w, r)
		return "", 0, 0, errors.New("invalid name")
	}

	return m[2], width, height, nil // The title is the second subexpression.
}


func view(st *store.Store) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		title, width, height, err := getImgName(w, r)
		check(err)

		url, err := st.GetImageUrl(title, width, height)
		check(err)

		fmt.Fprint(w, url)
	}
}




func errorHandler(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if e, ok := recover().(error); ok {
				w.WriteHeader(500)
				templates.ExecuteTemplate(w, "error.html", e)
			}
		}()
		fn(w, r)
	}
}




func main() {
	s := &store.Store{
		AccessKey: "YourAccessKey",
		SecretKey: "YourSecretKey",
		BucketName: "YourBucketName"}

	http.HandleFunc("/upload", errorHandler(upload(s)))
	http.HandleFunc("/view/", errorHandler(view(s)))
	http.ListenAndServe(":8080", nil)
}



