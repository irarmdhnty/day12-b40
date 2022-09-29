package main

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"math"
	"my-project/connection"
	"my-project/middleware"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	r := mux.NewRouter()

	connection.ConnectDatabase()

	r.PathPrefix("/public/").Handler(http.StripPrefix("/public/", http.FileServer(http.Dir("./public"))))

	r.PathPrefix("/uploads/").Handler(http.StripPrefix("/uploads/", http.FileServer(http.Dir("./uploads/"))))

	r.HandleFunc("/", home).Methods("GET")
	r.HandleFunc("/contact", contact).Methods("GET")
	r.HandleFunc("/project", project).Methods("GET")
	r.HandleFunc("/add-project", middleware.UploadFile(addProject)).Methods("POST")
	r.HandleFunc("/detail/{id}", detail).Methods("GET")
	r.HandleFunc("/delete/{id}", delete).Methods("GET")
	r.HandleFunc("/edit/{id}", update).Methods("GET")
	r.HandleFunc("/edit-project/{id}", middleware.EditFile(editProject)).Methods("POST")
	r.HandleFunc("/register", register).Methods("GET")
	r.HandleFunc("/form-register", formRegister).Methods("POST")
	r.HandleFunc("/login", login).Methods("GET")
	r.HandleFunc("/form-login", formLogin).Methods("POST")
	r.HandleFunc("/logout", logout).Methods("GET")

	fmt.Println("server on in port 8080")
	http.ListenAndServe("localhost:8080", r)
}

type SessionData struct {
	IsLogin   bool
	UserName  string
	FlashData string
}

var Data = SessionData{}

type Project struct {
	ID                int
	Name              string
	Start_date        time.Time
	End_date          time.Time
	StartDate         string
	EndDate           string
	Duration          string
	Author            string
	Desc              string
	Technologies      []string
	Image             string
	Format_Start_date string
	Format_End_date   string
	isLogin           bool
}

type User struct {
	ID       int
	Name     string
	Email    string
	Password string
}

func home(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "text/html; charset=utf-8")
	tmpl, err := template.ParseFiles("views/index.html")

	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}

	var store = sessions.NewCookieStore([]byte("SESSION_KEY"))
	session, _ := store.Get(r, "SESSION_KEY")

	if session.Values["IsLogin"] != true {
		Data.IsLogin = false
	} else {
		Data.IsLogin = session.Values["IsLogin"].(bool)
		Data.UserName = session.Values["Name"].(string)
	}

	flashMessage := session.Flashes("message")

	var flashes []string
	if len(flashMessage) > 0 {
		sessions.Save(r, w)
		for _, fm := range flashMessage {
			flashes = append(flashes, fm.(string))
		}
	}

	Data.FlashData = strings.Join(flashes, "")

	if session.Values["IsLogin"] != true{

		data, _ := connection.Conn.Query(context.Background(), "SELECT blog.id, blog.name, description, technologies, duration, image FROM blog ORDER BY id DESC")

		var result []Project
		for data.Next() {
			var each = Project{}
	
			err := data.Scan(&each.ID, &each.Name, &each.Desc, &each.Technologies, &each.Duration, &each.Image)
	
			if err != nil {
				fmt.Println(err.Error())
				return
			}
	
			result = append(result, each)
		}

		card := map[string]interface{}{
			"DataSession": Data,
			"Add":         result,
		}

		w.WriteHeader(http.StatusOK)
		tmpl.Execute(w, card)

	} else{
		sessionID := session.Values["ID"].(int)

		data, _ := connection.Conn.Query(context.Background(), "SELECT blog.id, blog.name, description, technologies, duration, image FROM blog WHERE blog.author_id=$1 ORDER BY id DESC", sessionID)

		var result []Project
		for data.Next() {
			var each = Project{}
	
			err := data.Scan(&each.ID, &each.Name, &each.Desc, &each.Technologies, &each.Duration, &each.Image)
	
			if err != nil {
				fmt.Println(err.Error())
				return
			}
	
			result = append(result, each)
		}
		card := map[string]interface{}{
			"DataSession": Data,
			"Add":         result,
		}

		w.WriteHeader(http.StatusOK)
		tmpl.Execute(w, card)
		
	}
}

func contact(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	tmpl, err := template.ParseFiles("views/contact.html")
	if err != nil {
		w.Write([]byte(err.Error()))
	}

	tmpl.Execute(w, "")
}

func project(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "text/html; charset=utf-8")
	tmpl, err := template.ParseFiles("views/addProject.html")
	if err != nil {
		w.Write([]byte(err.Error()))
	}

	tmpl.Execute(w, "")
}

func addProject(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Fatal(err)
	}

	var name = r.PostForm.Get("inputName")
	var start_date = r.PostForm.Get("startDate")
	var end_date = r.PostForm.Get("endDate")
	var desc = r.PostForm.Get("desc")
	var technologies []string
	technologies = r.Form["technologies"]

	dataContext := r.Context().Value("dataFile")
	image := dataContext.(string)

	var store = sessions.NewCookieStore([]byte("SESSION_KEY"))
	session, _ := store.Get(r, "SESSION_KEY")

	author := session.Values["ID"].(int)
	fmt.Println(author)

	layout := "2006-01-02"
	dateStart, _ := time.Parse(layout, start_date)
	dateEnd, _ := time.Parse(layout, end_date)

	hours := dateEnd.Sub(dateStart).Hours()
	daysInHours := hours / 24
	monthInDay := math.Round(daysInHours / 30)
	yearInMonth := math.Round(monthInDay / 12)

	var duration string

	if yearInMonth > 0 {
		duration = strconv.FormatFloat(yearInMonth, 'f', 0, 64) + " Years"
		// fmt.Println(year, " Years")
	} else if monthInDay > 0 {
		duration = strconv.FormatFloat(monthInDay, 'f', 0, 64) + " Months"
		// fmt.Println(month, " Months")
	} else if daysInHours > 0 {
		duration = strconv.FormatFloat(daysInHours, 'f', 0, 64) + " Days"
		// fmt.Println(daysInHours, " Days")
	} else if hours > 0 {
		duration = strconv.FormatFloat(hours, 'f', 0, 64) + " Hours"
		// fmt.Println(hours, " Hours")
	} else {
		duration = "0 Days"
		// fmt.Println("0 Days")
	}

	_, err = connection.Conn.Exec(context.Background(), "INSERT INTO blog(name, start_date, end_date, description, technologies, duration, image, author_id) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)", name, start_date, end_date, desc, technologies, duration,image, author)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	http.Redirect(w, r, "/", http.StatusMovedPermanently)
}

func detail(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "text/html; charset=utf-8")
	tmpl, err := template.ParseFiles("views/detail.html")
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}

	var Detail = Project{}

	id, _ := strconv.Atoi(mux.Vars(r)["id"])

	err = connection.Conn.QueryRow(context.Background(), "SELECT id, name, start_date, end_date, description, technologies, duration, image FROM blog WHERE id=$1", id).Scan(
		&Detail.ID, &Detail.Name, &Detail.Start_date, &Detail.End_date, &Detail.Desc, &Detail.Technologies, &Detail.Duration, &Detail.Image)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	Detail.Format_Start_date = Detail.Start_date.Format("2 January 2006")
	Detail.Format_End_date = Detail.End_date.Format("2 January 2006")

	data := map[string]interface{}{
		"Details": Detail,
	}
	// fmt.Println(data)
	tmpl.Execute(w, data)
}

func delete(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(mux.Vars(r)["id"])

	_, err := connection.Conn.Exec(context.Background(), "DELETE FROM blog WHERE id=$1", id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	}

	http.Redirect(w, r, "/", http.StatusMovedPermanently)
}

func editProject(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Fatal(err)
	}

	id, _ := strconv.Atoi(mux.Vars(r)["id"])

	var name = r.PostForm.Get("inputName")
	var start_date = r.PostForm.Get("startDate")
	var end_date = r.PostForm.Get("endDate")
	var desc = r.PostForm.Get("desc")
	var technologies []string
	technologies = r.Form["technologies"]

	dataContext := r.Context().Value("dataFile")
	image := dataContext.(string)

	layout := "2006-01-02"
	dateStart, _ := time.Parse(layout, start_date)
	dateEnd, _ := time.Parse(layout, end_date)

	hours := dateEnd.Sub(dateStart).Hours()
	daysInHours := hours / 24
	monthInDay := math.Round(daysInHours / 30)
	yearInMonth := math.Round(monthInDay / 12)

	var duration string

	if yearInMonth > 0 {
		duration = strconv.FormatFloat(yearInMonth, 'f', 0, 64) + " Years"
	} else if monthInDay > 0 {
		duration = strconv.FormatFloat(monthInDay, 'f', 0, 64) + " Months"
	} else if daysInHours > 0 {
		duration = strconv.FormatFloat(daysInHours, 'f', 0, 64) + " Days"
	} else {
		duration = "0 Days"
	}

	_, err = connection.Conn.Exec(context.Background(), "UPDATE blog SET name=$1, start_date=$2, end_date=$3, description=$4, technologies=$5, duration=$6, image=$7 WHERE id=$8", name, dateStart, dateEnd, desc, technologies, duration, image, id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	}

	http.Redirect(w, r, "/", http.StatusMovedPermanently)
}

func update(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "text/html; charset=utf-8")
	tmpl, err := template.ParseFiles("views/editProject.html")
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}

	var Edit = Project{}
	id, _ := strconv.Atoi(mux.Vars(r)["id"])

	err = connection.Conn.QueryRow(context.Background(), "SELECT id, name, start_date, end_date, description, technologies, duration, image FROM blog WHERE id=$1", id).Scan(
		&Edit.ID, &Edit.Name, &Edit.Start_date, &Edit.End_date, &Edit.Desc, &Edit.Technologies, &Edit.Duration, &Edit.Image)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	}

	Edit.Format_Start_date = Edit.Start_date.Format("2006-01-02")
	Edit.Format_End_date = Edit.End_date.Format("2006-01-02")

	data := map[string]interface{}{
		"Id":   id,
		"Edit": Edit,
	}
	// fmt.Println(data)
	tmpl.Execute(w, data)
	http.Redirect(w, r, "/", http.StatusMovedPermanently)
}

func register(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "text/html; charset=utf-8")
	tmpl, err := template.ParseFiles("views/register.html")
	if err != nil {
		w.Write([]byte(err.Error()))
	}

	tmpl.Execute(w, nil)
}

func formRegister(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Fatal(err)
	}

	name := r.PostForm.Get("inputName")
	email := r.PostForm.Get("inputEmail")
	password := r.PostForm.Get("inputPassword")

	hashPassword, _ := bcrypt.GenerateFromPassword([]byte(password), 10)

	_, err = connection.Conn.Exec(context.Background(), "INSERT INTO tb_user(name, email, password) VALUES ($1,$2,$3)", name, email, hashPassword)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	http.Redirect(w, r, "/login", http.StatusMovedPermanently)
}

func login(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "text/html; charset=utf-8")
	tmpl, err := template.ParseFiles("views/login.html")
	if err != nil {
		w.Write([]byte(err.Error()))
	}

	var store = sessions.NewCookieStore([]byte("SESSION_KEY"))
	session, _ := store.Get(r, "SESSION_KEY")

	flashMessage := session.Flashes("message")

	var flashes []string
	if len(flashMessage) > 0 {
		sessions.Save(r, w)
		for _, fm := range flashMessage {
			flashes = append(flashes, fm.(string))
		}
	}

	Data.FlashData = strings.Join(flashes, "")
	tmpl.Execute(w, Data)
}

func formLogin(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Fatal(err)
	}

	var email = r.PostForm.Get("inputEmail")
	var password = r.PostForm.Get("inputPassword")

	user := User{}

	err = connection.Conn.QueryRow(context.Background(), "SELECT * FROM tb_user WHERE email=$1", email).Scan(&user.ID, &user.Name, &user.Email, &user.Password)
	if err != nil {
		var store = sessions.NewCookieStore([]byte("SESSION_KEY"))
		session, _ := store.Get(r, "SESSION_KEY")

		session.AddFlash("Wrong Email", "message")
		session.Save(r, w)

		http.Redirect(w, r, "/login", http.StatusMovedPermanently)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		var store = sessions.NewCookieStore([]byte("SESSION_KEY"))
		session, _ := store.Get(r, "SESSION_KEY")

		session.AddFlash("Wrong Password", "message")
		session.Save(r, w)

		http.Redirect(w, r, "/login", http.StatusMovedPermanently)
		return
	}

	var store = sessions.NewCookieStore([]byte("SESSION_KEY"))
	session, _ := store.Get(r, "SESSION_KEY")

	session.Values["Name"] = user.Name
	session.Values["Email"] = user.Email
	session.Values["ID"] = user.ID
	session.Values["IsLogin"] = true
	session.Options.MaxAge = 18000

	session.AddFlash("Login Success", "message")
	sessions.Save(r, w)

	http.Redirect(w, r, "/", http.StatusMovedPermanently)
}

func logout(w http.ResponseWriter, r *http.Request) {
	var store = sessions.NewCookieStore([]byte("SESSION_KEY"))
	session, _ := store.Get(r, "SESSION_KEY")
	session.Options.MaxAge = -1
	session.Save(r, w)

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
