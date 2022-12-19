package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/cookiejar"
	"os"
	"text/template"

	"github.com/alexPavlikov/go-blog/database"
	"github.com/alexPavlikov/go-blog/models"
	"github.com/alexPavlikov/go-blog/setting"

	_ "github.com/lib/pq"
)

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

	//defer r.Body.Close()
	handleRequest()
	http.ListenAndServe(":"+models.Cfg.ServerPort, nil)
}

func handleRequest() {
	// обработчики статических данных(папок)
	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("./assets/"))))
	http.Handle("/data/", http.StripPrefix("/data/", http.FileServer(http.Dir("./data/"))))

	//обработчики всех ссылок веб-сайта
	http.HandleFunc("/", logFormHandler)
	http.HandleFunc("/auth", authHandler)
	http.HandleFunc("/registration", regHandler)
	http.HandleFunc("/blog", blogHandler)
	http.HandleFunc("/page", pageHandler)
	http.HandleFunc("/setting", settingHandler)
	http.HandleFunc("/setting/refresh", refreshSettingHandler)
}

func logFormHandler(w http.ResponseWriter, r *http.Request) {
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
		user, err := database.SelectUserByLogPass(logs, pass)
		if err != nil {
			http.NotFound(w, r)
			http.Redirect(w, r, "/", http.StatusBadRequest)
		}

		recordingSessions(fmt.Sprintf("Пользователь, %s (логин - %s, пароль - %s) зашел в аккаунт.\n", user.Name, user.Login, user.Password))
		http.Redirect(w, r, "/blog", http.StatusOK)
	}
}

func regHandler(w http.ResponseWriter, r *http.Request) {
	var user models.Users
	r.ParseForm()
	if r.Method == "POST" { // регистрация пользователя
		logs := r.FormValue("email1")
		pass := r.FormValue("pswd1")
		txt := r.FormValue("txt")
		if logs != "" || pass != "" || txt != "" {
			fmt.Println("This is reg", logs, pass, txt)
		}
		//if true функция выдачи action для формы в слючие совпадения пароля и логина
		user.Login = logs
		user.Password = pass
		user.Name = txt
		fmt.Println(user)
		_, err := database.InsertUser(user)
		if err != nil {
			fmt.Println("Error = regHandler() InsertUser()")
			log.Fatal(err)
			http.Redirect(w, r, "/", http.StatusBadRequest)
		}
		http.Redirect(w, r, "/blog", http.StatusOK)
	}
}

func settingHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("html/setting.html", "html/header.html", "html/footer.html")
	if err != nil {
		http.NotFound(w, r)
	}
	title := map[string]string{"Title": models.Cfg.SettingTitle}
	tmpl.ExecuteTemplate(w, "header", title)
	tmpl.ExecuteTemplate(w, "setting", nil)
}

func refreshSettingHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	if r.Method == "GET" {
		login := r.FormValue("login")
		newPass := r.FormValue("newPass")
		oldPass := r.FormValue("oldPass")
		newName := r.FormValue("newName")
		date := r.FormValue("newHB")
		fmt.Println(login, newPass, oldPass, newName, date)
		if oldPass != "" {
			_, err := database.SelectUsersByColumn("Password", oldPass)
			if err == nil {
				if newPass != "" {
					_, err = database.UpdateUserByColumn("Password", newPass, login, oldPass)
					if err != nil {
						fmt.Println(err.Error())
					}
				} else if newName != "" {
					_, err = database.UpdateUserByColumn("Name", newName, login, oldPass)
					if err != nil {
						fmt.Println(err.Error())
					}
				} else if date != "" {
					_, err = database.UpdateUserByColumn("Birthdate", date, login, oldPass)
					// сделать проверку чтобы нельзя вписать дату больше сегодняшней
					if err != nil {
						fmt.Println(err.Error())
					}
				}
			} else {
				http.NotFound(w, r)
			}
		}
		http.Redirect(w, r, "/page", http.StatusOK)
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

	var postId string
	if r.Method == "POST" {
		r.ParseForm()
		postId = r.FormValue("post_id")
		fmt.Println(postId)
	}
	//id, _ := strconv.Atoi(postId)
	//fmt.Println(id) // Сделать добавления поста на страницу пользователя

	title := map[string]string{"Title": "моя страница"}
	tmpl.ExecuteTemplate(w, "header", title) //сделать запрос на выборку имени пользователя и вставить в title
	tmpl.ExecuteTemplate(w, "page", nil)
}

func Cookies() { // доделать/сделать
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

func recordingSessions(session string) {
	fmt.Println(session)
	file, err := os.Open("/data/files/list of visits.txt")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer file.Close()
	data := []byte(session)
	file.Write(data)
}
