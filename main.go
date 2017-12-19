package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

func main() {
	flag.Parse()
	//http.Handle("/echo", websocket.Handler(tailFile))
	fmt.Println(http.ListenAndServe(*addr, GetRouter()))
}

type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

var addr = flag.String("addr", ":8080", "http service address")

var upgrader = websocket.Upgrader{} // use default options

//Routes array is a
type Routes []Route

func initRoutes() Routes {
	routes := Routes{
		//Route{"index", "GET", "/", http.FileServer(http.Dir("./public"))},
		Route{"call", "GET", "/eqx/{command}", call},
		Route{"echo", "GET", "/ws/{type}/{file}", echo},
	}
	return routes
}

func GetRouter() *mux.Router {
	router := mux.NewRouter().StrictSlash(true)
	for _, route := range initRoutes() {
		var handler http.Handler
		handler = route.HandlerFunc
		handler = Logger(handler, route.Name)
		handler = AccessControlAllowOrigin(handler)

		router.
			Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			Handler(handler)
	}
	return router
}

func Logger(next http.Handler, name string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		//counter = counter + 1

		next.ServeHTTP(w, r)

		log.Printf(
			"%s\t%s\t%s\t%s",
			r.Method,
			r.RequestURI,
			name,
			time.Since(start),
			//counter,
		)
	})
}

func AccessControlAllowOrigin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, PATCH, DELETE")
		w.Header().Set("Access-Control-Allow-Headers",
			"Content-Type, Authentication, Accept, Content-Length, Accept-Encoding")
		w.Header().Set("Content-Type", "application/json")
		if r.Method == "OPTIONS" {
			return
		}
		//for websocket
		r.Header["Origin"] = nil
		next.ServeHTTP(w, r)
	})
}

func call(w http.ResponseWriter, r *http.Request) {
	var (
		parameter = mux.Vars(r)
	)
	cmdArgs := []string{}

	cmdArgs = append(cmdArgs, parameter["command"])

	if app := r.URL.Query().Get("app"); app != "" {
		cmdArgs = append(cmdArgs, app)
	}
	if process := r.URL.Query().Get("process"); process != "" {
		cmdArgs = append(cmdArgs, process)
	}
	if service := r.URL.Query().Get("service"); service != "" {
		cmdArgs = append(cmdArgs, service)
	}
	if instance := r.URL.Query().Get("instance"); instance != "" {
		cmdArgs = append(cmdArgs, instance)
	}
	out, err := exec.Command("./eqxAgent.sh", cmdArgs...).Output()
	if err != nil {
		json.NewEncoder(w).Encode(err.Error())
		return
	}
	res := string(out)
	fmt.Fprintf(w, res)
	return
}

func echo(w http.ResponseWriter, r *http.Request) {
	var (
		parameter = mux.Vars(r)
	)
	dirPaths := map[string]string{
		"log":  "/opt/equinox/log/",
		"stat": "/opt/equinox/stat/",
	}
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
	}
	defer c.CloseHandler()
	fileType, typeOk := parameter["type"]
	fileName, nameOk := parameter["file"]
	if !(typeOk && nameOk) {
		c.WriteMessage(websocket.TextMessage, []byte("BAD Request"))
	}
	typePath := dirPaths[fileType]
	fullPath := typePath + fileName
	c.WriteMessage(websocket.TextMessage, []byte(fullPath))
	//cmdArgs := []string{}
	//cmdArgs = append(cmdArgs, "-f")
	//cmdArgs = append(cmdArgs, fullPath)
	cmd := exec.Command("tail", "-f", fullPath)
	cmdReader, err := cmd.StdoutPipe()
	if err != nil {
		c.WriteMessage(websocket.TextMessage, []byte(err.Error()))
		//fmt.Fprintln(os.Stderr, "Error creating StdoutPipe for Cmd", err)
		cmd.Process.Kill()
	}

	scanner := bufio.NewScanner(cmdReader)
	go func() {
		for scanner.Scan() {
			err := c.WriteMessage(websocket.TextMessage, scanner.Bytes())
			if err != nil {
				log.Printf("Client close ws connection %s", err.Error())
				cmd.Process.Kill()
				break
			}
			//fmt.Printf("docker build out | %s\n", scanner.Text())
		}
	}()

	err = cmd.Start()
	if err != nil {
		c.WriteMessage(websocket.TextMessage, []byte(err.Error()))
		//fmt.Fprintln(os.Stderr, "Error starting Cmd", err)
		cmd.Process.Kill()
	}

	err = cmd.Wait()
	if err != nil {
		c.WriteMessage(websocket.TextMessage, []byte(err.Error()))
		//fmt.Fprintln(os.Stderr, "Error waiting for Cmd", err)
		cmd.Process.Kill()
	}
}
