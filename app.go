package main

import (
	"html/template"
	"net/http"
)

type WebData struct {
	Title string
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	//command
	/*var (
		cmdOut []byte
		err    error
	)
	cmdName := "git"
	cmdArgs := []string{}
	if cmdOut, err = exec.Command(cmdName).Output(); err != nil {
		fmt.Fprintln(os.Stderr, "There was an error running git rev-parse command: ", err)
		os.Exit(1)
	}
	res := string(cmdOut)*/
	//firstSix := sha[:6]
	//fmt.Println("The first six chars of the SHA at HEAD in this repo are", firstSix)
	//render
	tmpl, _ := template.ParseFiles("layout.html", "index.html")

	wd := WebData{
		Title: "res",
	}
	tmpl.Execute(w, &wd)
}

func pageHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, _ := template.ParseFiles("layout.html", "page-1.html")
	wd := WebData{
		Title: "Page",
	}
	tmpl.Execute(w, &wd)
}

/*func main() {
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/page-1", pageHandler)
	http.ListenAndServe(":8080", nil)

}*/
