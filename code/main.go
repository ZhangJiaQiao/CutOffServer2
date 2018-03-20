package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	// "bytes"
	"crypto/md5"
	"io"
	// "math/rand"
	"net/http"
	// "strconv"
	// "time"
	"github.com/gorilla/mux"
	"github.com/unrolled/render"
	"github.com/urfave/negroni"
)

// User Entity
type User struct {
	Username string
	Password string
	Score    int
	//Apikey      string `xorm:"'api_key' text" json:"api_key"`
}

func md5Hash(data string) string {
	hash := md5.New()
	io.WriteString(hash, data)
	return fmt.Sprintf("%x", hash.Sum(nil))
}

var admin User
var db *sql.DB

//main
func main() {
	admin.Username = "admin"
	admin.Password = "admin"
	db, _ = sql.Open("mysql", "root:1254860908@tcp(127.0.0.1:3306)/cut_off_user?charset=utf8")
	port := ":2222"
	server := NewServer()
	server.Run(port)
	fmt.Println("Server start successfully!")
}

//server
const Version string = "/v1/"

func NewServer() *negroni.Negroni {
	formatter := render.New(render.Options{IndentJSON: true})

	n := negroni.Classic()
	mx := mux.NewRouter()

	initRoutes(mx, formatter)

	n.UseHandler(mx)

	return n
}

func initRoutes(mx *mux.Router, formatter *render.Render) {
	var url string
	//authorize
	url = Version + "auth"
	mx.HandleFunc(url, authHandler(formatter)).Methods("GET")
	//users
	url = Version + "users"
	mx.HandleFunc(url, usersPostHandler(formatter)).Methods("POST")
	mx.HandleFunc(url, usersGetHandler(formatter)).Methods("GET")
	//score
	url = Version + "score"
	mx.HandleFunc(url, scoreGetHandler(formatter)).Methods("GET")
	//rank
	url = Version + "rank"
	mx.HandleFunc(url, rankGetHandler(formatter)).Methods("GET")
	//addScore
	url = Version + "addScore"
	mx.HandleFunc(url, addScoreGetHandler(formatter)).Methods("GET")
}

//handler
func authHandler(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		err := req.ParseForm()
		if err != nil {
			panic(err)
		}
		var user User

		fmt.Println("username =", req.FormValue("username"))
		fmt.Println("password =", req.FormValue("password"))
		pswHash := md5Hash(req.FormValue("password"))
		err2 := db.QueryRow("select username, password from data where username=?", req.FormValue("username")).Scan(&user.Username, &user.Password)
		if err2 == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if user.Password != pswHash {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		fmt.Println("Authorize successfully")
		formatter.JSON(w, http.StatusOK, user)
	}
}

// GET  /v1/addScore
func addScoreGetHandler(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		err := req.ParseForm()
		if err != nil {
			panic(err)
		}
		var user User
		user.Username = req.FormValue("username")
		user.Score = 0
		err2 := db.QueryRow("select score from data where username=?", user.Username).Scan(&user.Score)
		if err2 == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		user.Score = int(user.Score) + 1
		stmt, err3 := db.Prepare("update data set score=? where username=?")
		if err3 != nil {
			panic(err3)
		}
		_, err4 := stmt.Exec(user.Score, user.Username)
		if err4 != nil {
			panic(err4)
		}
		formatter.JSON(w, http.StatusCreated, user)
	}
}

// POST /v1/users
func usersPostHandler(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		err := req.ParseForm()
		if err != nil {
			fmt.Println("POST parse form failed!")
			return
		}
		var user User
		username := req.FormValue("username")
		password := req.FormValue("password")
		fmt.Println("POST successfully!")
		fmt.Println(username, password)
		err2 := db.QueryRow("select id from data where username = ?", username).Scan(&user.Username, &user.Password)
		if err2 != sql.ErrNoRows {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		pswHash := md5Hash(req.FormValue("password"))
		user.Username = username
		user.Password = pswHash
		stmt, err2 := db.Prepare("insert data set username=?, password=?")
		if err2 != nil {
			fmt.Println(err2)
			panic(err2)
			return
		}
		res, err := stmt.Exec(username, pswHash)
		id, err := res.LastInsertId()
		fmt.Println("id =", id)
		formatter.JSON(w, http.StatusCreated, user)
	}

}

// GET /v1/users{?username}
func usersGetHandler(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		err := req.ParseForm()
		if err != nil {
			panic(err)
		}
		fmt.Println("GET successfully")
	}
}

// GET /v1/score
func scoreGetHandler(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		fmt.Println("GET score successfully")
	}
}

// GET /v1/rank
func rankGetHandler(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		err := req.ParseForm()
		if err != nil {
			fmt.Println(err)
		}
		rows, err2 := db.Query("select username, score from data order by score desc")
		if err2 != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		var count int
		count = 10
		userList := make([]User, count)
		var i int
		i = 0
		for rows.Next() {
			if (i >= 10)
				break
			var user User
			rows.Scan(&user.Username, &user.Score)
			userList[i] = user
			i++
		}
		var data = make(map[string][]User)
		data["inforlist"]=userList
		formatter.JSON(w, http.StatusOK, data)
	}
}
