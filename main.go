package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"math/rand"
	"net/http"
	"net/http/cookiejar"
	"os"
	"strconv"
	"text/template"
	"time"

	"github.com/alexPavlikov/go-blog/database"
	"github.com/alexPavlikov/go-blog/models"
	"github.com/alexPavlikov/go-blog/setting"

	_ "github.com/lib/pq"
)

var client http.Client
var postId string
var id string
var posts []models.Posts
var post models.Posts
var guestId string
var communitiesName string
var communitiesPhoto string
var inputComment string
var guestLogin string
var Path string
var Communication []models.Message
var check models.MessageList

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
	http.HandleFunc("/friends", friendsHandler)
	http.HandleFunc("/communities", communitiesHandler)
	http.HandleFunc("/comments", commentsHandler)
	http.HandleFunc("/community", communityHandler)
	http.HandleFunc("/guest", guestHandler)
	http.HandleFunc("/guest/friends", guestFriendsHandler)
	http.HandleFunc("/guest/communities", guestCommunitiesHandler)
	http.HandleFunc("/message", messageHandler)
}

func logFormHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("html/login.html")
	if err != nil {
		http.NotFound(w, r)
	}

	tmpl.ExecuteTemplate(w, "login", nil)
}

var userAuth models.Users

func authHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	var authLog string
	var authPass string
	authLog = r.FormValue("email2")
	authPass = r.FormValue("pswd2")
	if r.Method == "GET" { //авторизация пользователя
		if authLog != "" || authPass != "" {
			fmt.Println("This is auth", authLog, authPass)
			cookie := &http.Cookie{
				Name:   "id",
				Value:  "abcd",
				MaxAge: 300,
			}
			Cookies()
			http.SetCookie(w, cookie)
		}
		//if true функция выдачи action для формы в слючие совпадения пароля и логина
		var err error
		userAuth, err = database.SelectUserByLogPass(authLog, authPass)
		if err != nil {
			http.NotFound(w, r)
			http.Redirect(w, r, "/", http.StatusBadRequest)
		}

		recordingSessions(fmt.Sprintf("Пользователь, %s (логин - %s, пароль - %s) зашел в аккаунт в %s.\n", userAuth.Name, userAuth.Login, userAuth.Password, time.Now().Format("2006-01-02 15:04")))
		http.Redirect(w, r, "/blog", http.StatusSeeOther)
	}
}

func regHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	logs := r.FormValue("email1")
	pass := r.FormValue("pswd1")
	txt := r.FormValue("txt")
	if r.Method == "POST" { // регистрация пользователя
		if logs != "" || pass != "" || txt != "" {
			fmt.Println("This is reg", logs, pass, txt)
		}
		//if true функция выдачи action для формы в слючие совпадения пароля и логина
		userAuth.Login = logs
		userAuth.Password = pass
		userAuth.Name = txt
		userAuth.Access = "User"
		fmt.Println(userAuth)
		_, err := database.InsertUser(userAuth)
		if err != nil {
			fmt.Println("Error = regHandler() InsertUser()")
			log.Fatal(err)
			http.Redirect(w, r, "/", http.StatusBadRequest)
		}
		http.Redirect(w, r, "/blog", http.StatusSeeOther)
	}
}

func settingHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("html/setting.html", "html/header.html", "html/footer.html")
	if err != nil {
		http.NotFound(w, r)
	}
	title := map[string]string{"Title": models.Cfg.SettingTitle}
	account := map[string]interface{}{"User": userAuth}
	tmpl.ExecuteTemplate(w, "header", title)
	tmpl.ExecuteTemplate(w, "setting", account)
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
		http.Redirect(w, r, "/page", http.StatusSeeOther)
	}
}

var export string

func blogHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("html/blog.html", "html/header.html", "html/footer.html")
	if err != nil {
		http.NotFound(w, r)
	}

	communitiesName = ""
	communitiesPhoto = ""

	posts := database.SelectPosts()

	if r.Method == "GET" {
		val, _ := strconv.Atoi(postId)
		err := database.UpdateLikeInPost(val)
		if err != nil {
			fmt.Println("Error - blogHandler() UpdateLikeInPost()")
		}
		postId = ""
		export = r.FormValue("community_id")
	}

	if r.Method == "POST" {
		r.ParseForm()
		postId = r.FormValue("post_id")
		fmt.Println("blog", postId)
		err = database.UpdateViewInPost()
		if err != nil {
			fmt.Println("Error - blogHandler() UpdateViewInPost()")
		}
	}

	title := map[string]string{"Title": models.Cfg.BlogTitle}
	blog := map[string]interface{}{"Post": posts, "User": userAuth}
	tmpl.ExecuteTemplate(w, "header", title)
	tmpl.ExecuteTemplate(w, "blog", blog)
}

func pageHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("html/page.html", "html/header.html", "html/footer.html")
	if err != nil {
		http.NotFound(w, r)
	}
	fmt.Println("page", postId)
	var rep models.Repost
	rep.Id = rand.Uint32() / 10000
	rep.Post, _ = strconv.Atoi(postId)
	rep.User = userAuth.Login
	frd := database.SelectAllFriendsUser(userAuth.Login)
	comnt := database.SelectAllCommunitiesUser("User", userAuth.Login)

	type statistics struct {
		FrinedsLen     int
		CommunitiesLen int
		HappyBithday   string
	}

	var stat statistics
	stat.CommunitiesLen = len(comnt)
	stat.FrinedsLen = len(frd)
	stat.HappyBithday = userAuth.Birthdate

	result, err := database.InsertRepoPost(rep)
	if err != nil {
		fmt.Println("Error - pageHandler() InsertRepoPost()")
	}
	fmt.Println(result)
	Done := true
	data := database.SelectRepoPostByUser(userAuth.Login)
	if len(data) == 0 {
		Done = false
	}
	title := map[string]string{"Title": userAuth.Name}
	tmpl.ExecuteTemplate(w, "header", title)

	sendUser := map[string]interface{}{"User": userAuth, "Repo": data, "Statistics": stat, "Done": Done}
	tmpl.ExecuteTemplate(w, "page", sendUser)
}

func commentsHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("html/comments.html", "html/header.html", "html/footer.html")
	if err != nil {
		http.NotFound(w, r)
	}

	fmt.Println("comments", postId)

	if r.Method == "GET" {
		r.ParseForm()
		inputComment = r.FormValue("commentsInput")
		var comment models.Comments
		if inputComment != "" {
			comment.Posts, _ = strconv.Atoi(postId)
			comment.Author = userAuth.Login
			comment.Like = 1
			comment.Text = inputComment
			fmt.Println(comment)
			_, err = database.InsertComment(comment)

			if err != nil {
				fmt.Println("Error - commentsHandler() InsertComment()")
			}
		}
	}

	commentPost := database.SelectCommentsByColumn("Posts", postId)
	res, _ := strconv.Atoi(postId)
	post := database.SelectPostById(res)
	//comment := map[string]interface{}{"Comments": commentPost}
	data := map[string]interface{}{"User": userAuth, "Comments": commentPost, "CommentsTitle": post.Title}
	title := map[string]interface{}{"Title": models.Cfg.CommentsTitle}
	tmpl.ExecuteTemplate(w, "header", title)
	tmpl.ExecuteTemplate(w, "comments", data)
}

func friendsHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("html/friends.html", "html/header.html", "html/footer.html")
	if err != nil {
		http.NotFound(w, r)
	}

	r.ParseForm()
	if r.Method == "POST" {
		id = r.FormValue("friend_id")
		fmt.Println(id)
		err := database.DeleteFriendsById(id)
		if err != nil {
			fmt.Println("Error - friendsDelHandler() DeleteFriendsById()")
		}
	}
	if r.Method == "GET" {
		r.ParseForm()
		guestId = r.FormValue("Id")
		fmt.Println("GET", guestId)
		check, err = database.SelectMessengeListbyUsers(userAuth.Login, guestId)
		if err != nil {
			fmt.Println("Error - friendsHandler() SelectMessengeListbyUsers()")
		}
	}

	friends := database.SelectAllFriendsUser(userAuth.Login)
	data := map[string]interface{}{"User": userAuth, "Friends": friends}
	title := map[string]string{"Title": models.Cfg.FriendsTitile}
	tmpl.ExecuteTemplate(w, "header", title)
	tmpl.ExecuteTemplate(w, "friends", data)
}

func guestFriendsHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("html/friends.html", "html/header.html", "html/footer.html")
	if err != nil {
		http.NotFound(w, r)
	}

	r.ParseForm()
	if r.Method == "POST" {
		id = r.FormValue("friend_id")
		fmt.Println(id)
	}
	if r.Method == "GET" {
		guestId = r.FormValue("guestId")
		fmt.Println("GET", guestId)
	}
	gst, _ := database.SelectUserByColumn("Login", guestId)
	friends := database.SelectAllFriendsUser(guestLogin)
	data := map[string]interface{}{"User": userAuth, "Friends": friends, "Guest": gst}
	title := map[string]string{"Title": models.Cfg.FriendsTitile}
	tmpl.ExecuteTemplate(w, "header", title)
	tmpl.ExecuteTemplate(w, "friends", data)
}

func communitiesHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("html/communities.html", "html/header.html", "html/footer.html")
	if err != nil {
		http.NotFound(w, r)
	}

	if r.Method == "GET" {
		r.ParseForm()
		data := r.FormValue("name_com")
		fmt.Println("GET Отписка от сообщества", data, userAuth.Login)
		if data != "" {
			err := database.DeleteSubOnCommunities(data, userAuth.Login)
			if err != nil {
				fmt.Println("Error - communitiesHandler() DeleteCommunitiesByName()")
			}
		}
	}
	if r.Method == "POST" {
		r.ParseForm()
		subCom := r.FormValue("communityRec")
		fmt.Println(subCom)
		if subCom != "" {
			err := database.InsertSubscribersToUser(userAuth.Login, subCom)
			if err != nil {
				fmt.Println("Error - communitiesHandler() InsertSubscribersToUser()", err.Error())
			}
		}
		communitiesName = r.FormValue("community_id")
		fmt.Println("POST", communitiesName)
		communities := database.SelectCommunitiesByColumn("Name", communitiesName)
		communitiesPhoto = communities.Photo
	}
	done := true
	comWithOutSub := database.SelectCommunitiesWithOutSub(userAuth.Login)
	communities := database.SelectAllCommunitiesUser("User", userAuth.Login)
	data := map[string]interface{}{"User": userAuth, "Communities": communities, "RecCommunities": comWithOutSub, "Done": done}
	title := map[string]string{"Title": models.Cfg.CommunitiesTitile}
	tmpl.ExecuteTemplate(w, "header", title)
	tmpl.ExecuteTemplate(w, "communities", data)
}

func guestCommunitiesHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("html/communities.html", "html/header.html", "html/footer.html")
	if err != nil {
		http.NotFound(w, r)
	}

	if r.Method == "POST" {
		r.ParseForm()
		subCom := r.FormValue("communityRec")
		fmt.Println(subCom)
		communitiesName = r.FormValue("community_id")
		fmt.Println("POST", communitiesName)
		communities := database.SelectCommunitiesByColumn("Name", communitiesName)
		communitiesPhoto = communities.Photo
	}
	done := false
	comWithOutSub := database.SelectCommunitiesWithOutSub(guestLogin)
	communities := database.SelectAllCommunitiesUser("User", guestLogin)
	data := map[string]interface{}{"User": userAuth, "Communities": communities, "RecCommunities": comWithOutSub, "Done": done}
	title := map[string]string{"Title": models.Cfg.CommunitiesTitile}
	tmpl.ExecuteTemplate(w, "header", title)
	tmpl.ExecuteTemplate(w, "communities", data)
}

func communityHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("html/community.html", "html/header.html", "html/footer.html")
	if err != nil {
		http.NotFound(w, r)
	}

	res, _ := strconv.Atoi(export)
	post = database.SelectPostById(res)
	type Comm struct {
		Name  string
		Photo string
	}
	if communitiesName != "" {
		post.Communities = communitiesName
		post.CommunitiesPhot = communitiesPhoto
	}
	var Foo Comm
	Foo.Name = post.Communities
	Foo.Photo = post.CommunitiesPhot

	posts = database.SelectPostByCommunities(post.Communities)
	subs := database.SelectSubscribersBtCommunities(post.Communities)
	author, names := database.SelectCommunitiesAuthorByName(post.Communities)

	if r.Method == "GET" {
		val, _ := strconv.Atoi(postId)
		err := database.UpdateLikeInPost(val)
		if err != nil {
			fmt.Println("Error - communityHandler() UpdateLikeInPost()")
		}
		postId = ""
		r.ParseForm()
		guestId = r.FormValue("guestId")
		fmt.Println("Get", guestId)
	}

	if r.Method == "POST" {
		r.ParseForm()
		postId = r.FormValue("postId")
		fmt.Println("blog", postId)
		// err = database.UpdateViewInPost()
		// if err != nil {
		// 	fmt.Println("Error - communityHandler() UpdateViewInPost()")
		// }
	}

	title := map[string]string{"Title": Foo.Name}
	blog := map[string]interface{}{"Post": posts, "User": userAuth, "Subs": subs, "Author": author, "Names": names, "Communities": Foo}
	tmpl.ExecuteTemplate(w, "header", title)
	tmpl.ExecuteTemplate(w, "community", blog)
}

func guestHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("html/guest.html", "html/header.html", "html/footer.html")
	if err != nil {
		http.NotFound(w, r)
	}

	fmt.Println("guest", guestId)
	if r.Method == "GET" {
		r.ParseForm()
		guestLogin = r.FormValue("guestLogin")
	}

	userGuest, err := database.SelectUserByColumn("Login", guestId)
	if err != nil {
		fmt.Println("Error - guestHandler() SelectUsersByColumn()")
	}
	fmt.Println(userGuest)

	frd := database.SelectAllFriendsUser(userGuest.Login)
	comnt := database.SelectAllCommunitiesUser("User", userGuest.Login)

	type statistics struct {
		FrinedsLen     int
		CommunitiesLen int
		HappyBithday   string
	}

	var stat statistics

	stat.CommunitiesLen = len(comnt)
	stat.FrinedsLen = len(frd)
	stat.HappyBithday = userGuest.Birthdate
	Done := true
	data := database.SelectRepoPostByUser(userGuest.Login)
	if len(data) == 0 {
		Done = false
	}
	title := map[string]string{"Title": userGuest.Name}
	tmpl.ExecuteTemplate(w, "header", title)

	sendUser := map[string]interface{}{"User": userAuth, "Repo": data, "Guest": userGuest, "Statistics": stat, "Done": Done}

	tmpl.ExecuteTemplate(w, "guest", sendUser)
}

var usrMesg string
var f *os.File
var Link string
var companion []models.Companions
var UsersLink models.MessageList
var activeChatUser models.Users
var OK bool
var Messenger models.Messenger

func messageHandler(w http.ResponseWriter, r *http.Request) {
	OK = true
	tmpl, err := template.ParseFiles("html/message.html", "html/header.html", "html/footer.html")
	if err != nil {
		http.NotFound(w, r)
	}

	//createFile(check)
	companion = database.SelectCompanionsByLogin(userAuth.Login)
	fmt.Println("UsersLink1111", UsersLink)
	Link = fmt.Sprintf("C:/Users/admin/go/src/go-blog/data/files/message/%s", UsersLink.MessageHistory)

	if r.Method == "GET" {
		r.ParseForm()
		usrMesg = r.FormValue("user_id")
		fmt.Println(usrMesg)
		UsersLink, err = database.SelectMessengeListbyUsers(userAuth.Login, usrMesg)
		if err != nil {
			fmt.Println("Error - messageHandler() SelectMessengeListbyUsers()", err)
		}
		activeChatUser, err = database.SelectUserByColumn("Login", usrMesg)
		if err != nil {
			fmt.Println("Error - messageHandler() SelectUserByColumn()", err)
		}
		OK = true
	}
	fmt.Println("UsersLink2222", UsersLink)
	//Link = fmt.Sprintf("C:/Users/admin/go/src/go-blog/data/files/message/%s", UsersLink.MessageHistory)
	if r.Method == "POST" {

		r.ParseForm()

		message := r.FormValue("commentsInput")
		if message != "" {
			//Link = fmt.Sprintf("C:/Users/admin/go/src/go-blog/data/files/message/%s", UsersLink.MessageHistory)
			Link = "C:/Users/admin/go/src/go-blog/data/files/message/test.json"
			fmt.Println("Link------------------", Link)
			f, err = os.OpenFile(Link, os.O_WRONLY|os.O_APPEND, 0755)
			if err != nil {
				fmt.Println("Error - messageHandler() os.OpenFile()")
			}
			defer f.Close()

			msg := models.Message{
				User:    userAuth.Login,
				Message: message,
				Data:    time.Now().Format("2006-01-02 15:04"),
				Photo:   userAuth.Photo,
			}
			Messenger = JSON(msg, Link)
			fmt.Println(Messenger)
		}
	}

	title := map[string]string{"Title": models.Cfg.MessageTitle}
	tmpl.ExecuteTemplate(w, "header", title)
	fmt.Println(OK)
	data := map[string]interface{}{"User": userAuth, "Done": OK, "OK": OK, "Companions": companion, "ChatUser": activeChatUser, "Chat": Messenger.Messenge}
	tmpl.ExecuteTemplate(w, "message", data)
}

func JSON(msg models.Message, Path string) models.Messenger {
	fmt.Println("JSON", Path)
	rawDataIn, err := ioutil.ReadFile(Path)
	fmt.Println(Path)
	if err != nil {
		log.Fatal("Cannot load settings:", err)
	}

	var settings models.Messenger
	err = json.Unmarshal(rawDataIn, &settings)
	if err != nil {
		log.Fatal("Invalid settings format:", err)
	}

	newClient := models.Message{
		User:    msg.User,
		Message: msg.Message,
		Data:    msg.Data,
		Photo:   msg.Photo,
	}

	settings.Messenge = append(settings.Messenge, newClient)
	for i := range settings.Messenge {
		if settings.Messenge[i].User == userAuth.Login {
			settings.Messenge[i].Access = 2
		} else {
			settings.Messenge[i].Access = 1
		}
	}

	rawDataOut, err := json.MarshalIndent(&settings, "", "  ")
	if err != nil {
		log.Fatal("JSON marshaling failed:", err)
	}

	err = ioutil.WriteFile(Path, rawDataOut, 0)
	if err != nil {
		log.Fatal("Cannot write updated settings file:", err)
	}
	return settings
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
	file, err := os.OpenFile("C:/Users/admin/go/src/go-blog/data/files/listOfVisits.txt", os.O_WRONLY|os.O_APPEND, 0755)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer file.Close()
	data := []byte(session)
	_, err = file.Write(data)
	if err != nil {
		fmt.Println(err)
	}
}

func createFile(check models.MessageList) {
	Path := "C:/Users/admin/go/src/go-blog/data/files/message/"
	fmt.Println("start", check.MessageHistory)
	if check.LinkId > 0 {
		fmt.Println(check.MessageHistory)
	} else {
		value := models.MessageList{
			LinkId:         rand.Uint32() / 10000,
			Main:           userAuth.Login,
			Companion:      guestId,
			MessageHistory: strconv.Itoa(int(math.Abs(float64(rand.Int31())))) + ".json",
		}
		database.InsertMessengeListbyUsers(value)
		value.LinkId += 1
		err := database.InsertDoubleMessengeListbyUsers(value)
		if err != nil {
			fmt.Println(err.Error())
		}
		_, err = os.Create(Path + value.MessageHistory)
		if err != nil {
			fmt.Println("Error - friendsHandler() Create file")
		}
		fmt.Println("Create file - ", Path+value.MessageHistory)
		check = models.MessageList{}
	}
}
