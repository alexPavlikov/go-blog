package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/cookiejar"
	"text/template"

	"github.com/alexPavlikov/go-blog/database"
	"github.com/alexPavlikov/go-blog/models"
	"github.com/alexPavlikov/go-blog/setting"

	_ "github.com/lib/pq"
)

var posts map[string]*models.Posts
var client http.Client

func init() {
	setting.Config()
}

func main() {
	fmt.Println("Listen on:", models.Cfg.ServerPort)

	db, err := database.Connect()
	if err != nil {
		log.Fatal(err.Error())
	}
	defer db.Close()

	//posts = make(map[string]*models.Post, 0)
	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("./assets/"))))
	http.Handle("/data/", http.StripPrefix("/data/", http.FileServer(http.Dir("./data/"))))
	http.HandleFunc("/", logFormHandler)
	http.HandleFunc("/auth", authHandler)
	http.HandleFunc("/registration", regHandler)
	http.HandleFunc("/blog", blogHandler)
	http.HandleFunc("/page", pageHandler)
	//defer r.Body.Close()
	//handleRequest()
	http.ListenAndServe(":"+models.Cfg.ServerPort, nil)
}

func logFormHandler(w http.ResponseWriter, r *http.Request) { //реализовать проверку данных в бд и занос новых пользователей
	tmpl, err := template.ParseFiles("html/login.html")
	if err != nil {
		http.NotFound(w, r)
	}

	tmpl.ExecuteTemplate(w, "login", nil)
}

func authHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	if r.Method == "GET" { //авторизация пользователя
		logs := r.FormValue("email2")
		pass := r.FormValue("pswd2")
		if logs != "" || pass != "" {
			fmt.Println("This is auth", logs, pass)
			cookie := &http.Cookie{
				Name:   "id",
				Value:  "abcd",
				MaxAge: 300,
			}
			Cookies()
			http.SetCookie(w, cookie)
		}
		//if true функция выдачи action для формы в слючие совпадения пароля и логина
		http.Redirect(w, r, "/blog", http.StatusOK)
	}
}

func regHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	if r.Method == "POST" { // регистрация пользователя
		logs := r.FormValue("email1")
		pass := r.FormValue("pswd1")
		txt := r.FormValue("txt")
		if logs != "" || pass != "" || txt != "" {
			fmt.Println("This is reg", logs, pass, txt)
		}
		//if true функция выдачи action для формы в слючие совпадения пароля и логина
		http.Redirect(w, r, "/blog", http.StatusOK)
	}
}

func blogHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("html/blog.html", "html/header.html", "html/footer.html")
	if err != nil {
		http.NotFound(w, r)
	}
	title := map[string]string{"Title": models.Cfg.BlogTitle}
	tmpl.ExecuteTemplate(w, "header", title)
	tmpl.ExecuteTemplate(w, "blog", nil)
}

func pageHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("html/page.html", "html/header.html", "html/footer.html")
	if err != nil {
		http.NotFound(w, r)
	}
	title := map[string]string{"Title": "моя страница"}
	tmpl.ExecuteTemplate(w, "header", title) //сделать запрос на выборку имени пользователя и вставить в title
	tmpl.ExecuteTemplate(w, "page", nil)
}

func handleRequest() {
	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("./assets/"))))

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/write", writeHandler)
	http.HandleFunc("/edit", editHandler)
	http.HandleFunc("/SavePost", savePostHandler)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/index.html", "templates/header.html", "templates/footer.html")
	if err != nil {
		http.NotFound(w, r)
	}
	data := database.SelectPosts()
	tmpl.ExecuteTemplate(w, "index", data)
}

func writeHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/write.html", "templates/header.html", "templates/footer.html")
	if err != nil {
		http.NotFound(w, r)
	}

	tmpl.ExecuteTemplate(w, "write", nil)
}

func editHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/write.html", "templates/header.html", "templates/footer.html")
	if err != nil {
		http.NotFound(w, r)
	}

	id := r.FormValue("id")
	post, found := posts[id]
	if !found {
		http.NotFound(w, r)
	}

	tmpl.ExecuteTemplate(w, "write", post)
}

func savePostHandler(w http.ResponseWriter, r *http.Request) {
	// r.ParseForm()
	// id := GenerateId()
	// title := r.FormValue("title")
	// content := r.FormValue("content")
	// post := models.NewPost(id, title, content)
	// posts[post.Id] = post

	http.Redirect(w, r, "/", http.StatusFound)
}

func Cookies() {
	fmt.Println(client)
	jar, err := cookiejar.New(nil)
	if err != nil {
		log.Fatalf("Got error while creating cookie jar %s", err.Error())
	}
	client = http.Client{
		Jar: jar,
	}
	fmt.Println(client)
}
