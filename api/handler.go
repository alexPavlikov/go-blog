package api

import (
	"crypto/rand"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/alexPavlikov/go-blog/app"
	"github.com/alexPavlikov/go-blog/database"
	"github.com/alexPavlikov/go-blog/models"
)

func rootHandler(w http.ResponseWriter, r *http.Request) {
	c := app.ReadCookies(r)
	var err error
	fmt.Println(c)
	if c == nil {
		http.Redirect(w, r, "/entry", http.StatusSeeOther)
		//tmpl.ExecuteTemplate(w, "login", nil)
	} else if userAuth.Login != "" {
		userAuth, err = database.SelectUserByColumn("Login", userAuth.Login)
		if err != nil {
			fmt.Println("SelectUserByColumn()", err.Error())
		}
		http.Redirect(w, r, "/page", http.StatusSeeOther)
	} else {
		userAuth.Login = c.Value
		userAuth, err = database.SelectUserByColumn("Login", userAuth.Login)
		if err != nil {
			fmt.Println("SelectUserByColumn()", err.Error())
		}
		http.Redirect(w, r, "/page", http.StatusSeeOther)
	}
}

func logFormHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("html/login.html")
	if err != nil {
		http.NotFound(w, r)
	}
	Done := true
	data := map[string]interface{}{"Done": Done}
	tmpl.ExecuteTemplate(w, "login", data)
}

func frHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("html/fr.html")
	if err != nil {
		http.NotFound(w, r)
	}

	tmpl.ExecuteTemplate(w, "fr", nil)
}

func authHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	var authLog string
	var authPass string
	authLog = r.FormValue("email2")
	authPass = r.FormValue("pswd2")
	authPass = app.CreateMd5Hash(authPass)
	if r.Method == "GET" { //авторизация пользователя
		fmt.Println("This is auth", authLog, authPass)
		//if true функция выдачи action для формы в слючие совпадения пароля и логина
		var err error
		userAuth, _ = database.SelectUserByLogPass(authLog, authPass)
		if userAuth.Login == "" && userAuth.Password == "" {
			w.Write([]byte("Неверный логин или пароль"))
			// time.Sleep(3 * time.Second)
			// http.Redirect(w, r, "/auth", http.StatusSeeOther)
		} else if userAuth.Access == "Banned" {
			http.NotFound(w, r) //отправлять письмо - вы в бане и сколько осталось часов
		} else {
			err = database.InsertUserToOnline(userAuth.Login)
			if err != nil {
				fmt.Println("Error - authHandler() InsertUserToOnline()", err)
				// http.NotFound(w, r)
			}
			expires := time.Now().AddDate(1, 0, 0)
			cookie := &http.Cookie{
				Name:  "id",
				Value: userAuth.Login,
				//MaxAge:  300,
				Expires: expires,
			}
			http.SetCookie(w, cookie)
			if userAuth.Login != "" {
				code = app.GiveCode()
				err = app.SendCode(userAuth.Login, code, userAuth.Name)
				if err != nil {
					http.NotFound(w, r)
				}
			}
			app.RecordingSessions(fmt.Sprintf("Пользователь, %s (логин - %s, пароль - %s) зашел в аккаунт в %s.\n", userAuth.Name, userAuth.Login, userAuth.Password, time.Now().Format("2006-01-02 15:04")), "listOfVisits.txt")
			http.Redirect(w, r, "/second_auth", http.StatusSeeOther)
		}
	}
}

func secondAuthHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("html/secondAuth.html")
	if err != nil {
		http.NotFound(w, r)
	}

	if r.Method == "POST" {
		r.ParseForm()
		cd := r.FormValue("code")
		if cd != "" {
			inputCode, err := strconv.Atoi(cd)
			if err != nil {
				w.Write([]byte("Не верно указанный код"))
			}
			if inputCode == code {
				http.Redirect(w, r, "/blog", http.StatusSeeOther)
			} else {
				w.Write([]byte("Не верно указанный код"))
				http.Redirect(w, r, "/exit", http.StatusSeeOther)
			}
		}
	}

	tmpl.ExecuteTemplate(w, "secondAuth", nil)
}

func regHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	logs := r.FormValue("email1")
	pass := r.FormValue("pswd1")
	pass = app.CreateMd5Hash(pass)
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
		userAuth.Photo = "/data/image/blog/standart_avatar.png"
		userAuth.Birthdate = time.Now().Format("02.01.2006")
		fmt.Println(userAuth)
		_, err := database.InsertUser(userAuth)
		if err != nil {
			fmt.Println("Error = regHandler() InsertUser()")
			log.Fatal(err)
			http.Redirect(w, r, "/", http.StatusBadRequest)
		}
		expires := time.Now().AddDate(1, 0, 0)
		cookie := &http.Cookie{
			Name:  "id",
			Value: userAuth.Login,
			//MaxAge: 300,
			Expires: expires,
		}
		http.SetCookie(w, cookie)
		http.Redirect(w, r, "/setting", http.StatusSeeOther)
	}
}

func findUserHandler(w http.ResponseWriter, r *http.Request) { //FindFR
	tmpl, err := template.ParseFiles("html/friends.html", "html/header.html", "html/footer.html")
	if err != nil {
		http.NotFound(w, r)
	}

	var users []models.Users
	var communities []models.Communities

	if r.Method == "POST" {
		r.ParseForm()
		value := r.FormValue("find")
		users, err = database.FindUsers(value, userAuth.Login)
		if err != nil {
			fmt.Println("Error - findUserHandler() FindUsers()", err.Error())
		}
		fmt.Println(users)
		communities, err = database.FindCommunities(value, userAuth.Login)
		if err != nil {
			fmt.Println("Error - findUserHandler() FindCommunities()", err.Error())
		}
		fmt.Println(communities)
	}
	okuser := false
	okcom := false
	if users == nil {
		okuser = true
	}
	if communities == nil {
		okcom = true
	}

	data := map[string]interface{}{"User": userAuth, "Find": users, "Done": "FindFR", "Friends": "", "Subs": "", "Online": "", "OKU": okuser, "OKC": okcom, "Communities": communities}
	title := map[string]interface{}{"Title": models.Cfg.FriendsTitile, "User": userAuth}
	tmpl.ExecuteTemplate(w, "header", title)
	tmpl.ExecuteTemplate(w, "footer", title)
	tmpl.ExecuteTemplate(w, "friends", data)

}

func settingHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("html/setting.html", "html/header.html", "html/footer.html")
	if err != nil {
		http.NotFound(w, r)
	}
	title := map[string]interface{}{"Title": models.Cfg.SettingTitle, "User": userAuth}
	account := map[string]interface{}{"User": userAuth}
	tmpl.ExecuteTemplate(w, "header", title)
	tmpl.ExecuteTemplate(w, "footer", title)
	tmpl.ExecuteTemplate(w, "setting", account)
}

func changePasswordHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("html/login.html")
	if err != nil {
		http.NotFound(w, r)
	}

	if r.Method == "POST" {
		r.ParseForm()
		log := r.FormValue("login")
		if log != "" {
			user, err := database.SelectUserByColumn("Login", log)
			if err != nil {
				fmt.Println("Error - changePasswordHandler() SelectUserByColumn()", err.Error())
			}
			if user.Login != "" {
				RandomCrypto, _ := rand.Prime(rand.Reader, 32)
				p := "Gopher" + fmt.Sprint(RandomCrypto.Int64()/20000)
				pass := app.CreateMd5Hash(p)
				_, err := database.UpdateUserByColumn("Password", pass, user.Login, user.Password)
				if err != nil {
					http.NotFound(w, r)
				}
				err = app.SendNewPass(user.Login, p, user.Name)
				if err != nil {
					fmt.Println("Error - SendNewPass()", err.Error())
				}
				http.Redirect(w, r, "/", http.StatusSeeOther)
			}
		}
	}

	Done := false
	data := map[string]interface{}{"Done": Done}
	tmpl.ExecuteTemplate(w, "login", data)
}

func refreshSettingHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	if r.Method == "GET" {
		login := userAuth.Login
		newPass := r.FormValue("newPass")
		oldPass := r.FormValue("oldPass")
		newName := r.FormValue("newName")
		date := r.FormValue("newHB")
		oldPass = app.CreateMd5Hash(oldPass)

		fmt.Println(login, newPass, oldPass, newName, date)
		if oldPass != "" {
			_, err := database.SelectUserByLogPass(login, oldPass)
			if err == nil {
				if newName != "" {
					_, err = database.UpdateUserByColumn("Name", newName, login, oldPass)
					if err != nil {
						fmt.Println(err.Error())
					}
				}
				if date != "" {
					_, err = database.UpdateUserByColumn("Birthdate", date, login, oldPass)
					if err != nil {
						fmt.Println(err.Error())
					}
				}
				if newPass != "" {
					newPass = app.CreateMd5Hash(newPass)
					_, err = database.UpdateUserByColumn("Password", newPass, login, oldPass)
					if err != nil {
						fmt.Println(err.Error())
					}
				}
			} else {
				http.NotFound(w, r)
			}
		}
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
	if r.Method == "POST" {
		r.ParseForm()

		file, fileHeader, err := r.FormFile("files")
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		defer file.Close()
		imgPath = fmt.Sprintf("/data/image/blog/%d%s", time.Now().UnixNano(), filepath.Ext(fileHeader.Filename))
		dst, err := os.Create(imgPath)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		defer dst.Close()

		_, err = io.Copy(dst, file)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		us, err := database.UpdateUserByColumn("Photo", imgPath, userAuth.Login, userAuth.Password)
		if err != nil {
			fmt.Println("Error - refreshSettingHandler() POST UpdateUserByColumn()", err.Error())
		}
		fmt.Println(us)
		http.Redirect(w, r, "/exit", http.StatusSeeOther)
	}
}

func blogHandler(w http.ResponseWriter, r *http.Request) {
	c := app.ReadCookies(r)
	if c.Value == userAuth.Login {
		tmpl, err := template.ParseFiles("html/blog.html", "html/header.html", "html/footer.html")
		if err != nil {
			http.NotFound(w, r)
		}

		communitiesName = ""
		communitiesPhoto = ""

		posts := database.SelectPostsByUserSubs(userAuth.Login)

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

		title := map[string]interface{}{"Title": models.Cfg.BlogTitle, "User": userAuth}
		blog := map[string]interface{}{"Post": posts, "User": userAuth}
		tmpl.ExecuteTemplate(w, "header", title)
		tmpl.ExecuteTemplate(w, "footer", title)
		tmpl.ExecuteTemplate(w, "blog", blog)
	} else if c.Value != "" {
		var err error
		userAuth.Login = c.Value
		userAuth, err = database.SelectUserByColumn("Login", userAuth.Login)
		if err != nil {
			fmt.Println(err.Error())
		}
	} else {
		http.NotFound(w, r)
	}
}

func pageHandler(w http.ResponseWriter, r *http.Request) {
	c := app.ReadCookies(r)
	if c.Value == userAuth.Login {
		tmpl, err := template.ParseFiles("html/page.html", "html/header.html", "html/footer.html")
		if err != nil {
			http.NotFound(w, r)
		}
		fmt.Println("page", postId)
		var rep models.Repost
		var DoneGopher bool
		var GalOK bool
		rep.Post, _ = strconv.Atoi(postId)
		rep.User = userAuth.Login
		frd := database.SelectAllFriendsUser(userAuth.Login)
		comnt := database.SelectAllCommunitiesUser("User", userAuth.Login)
		gopher := database.SelectGopherByOwner(userAuth.Login)
		if gopher != nil {
			DoneGopher = true
		} else {
			DoneGopher = false
		}

		type statistics struct {
			FrinedsLen     int
			CommunitiesLen int
			HappyBithday   string
		}

		var stat statistics
		stat.CommunitiesLen = len(comnt)
		stat.FrinedsLen = len(frd)
		stat.HappyBithday = userAuth.Birthdate

		if r.Method == "POST" {
			r.ParseForm()
			GofId := r.FormValue("goLike")
			fmt.Println(GofId)
			if GofId != "" {
				err = database.InsertLikeToGopher(GofId)
				if err != nil {
					fmt.Println(err.Error())
				}
			}
		}

		if rep.Post > 0 && rep.User != "" {
			result, err := database.InsertRepoPost(rep)
			if err != nil {
				fmt.Println("Error - pageHandler() InsertRepoPost()")
			}
			fmt.Println(result)
		}
		DonePost := true
		data := database.SelectRepoPostByUser(userAuth.Login)
		if len(data) == 0 {
			DonePost = false
		}
		access := false
		if userAuth.Access != "User" {
			access = true
		}

		photo := database.SelectUserGalleryLimit(userAuth.Login)
		if photo != nil {
			GalOK = true
		} else {
			GalOK = false
		}

		title := map[string]interface{}{"Title": userAuth.Name, "User": userAuth}
		tmpl.ExecuteTemplate(w, "header", title)
		tmpl.ExecuteTemplate(w, "footer", title)

		sendUser := map[string]interface{}{"User": userAuth, "Repo": data, "Statistics": stat, "Done": DonePost, "DoneGopher": DoneGopher, "Gopher": gopher, "Access": access, "GalOK": GalOK}
		tmpl.ExecuteTemplate(w, "page", sendUser)
	} else if c.Value != "" {
		var err error
		userAuth.Login = c.Value
		userAuth, err = database.SelectUserByColumn("Login", userAuth.Login)
		if err != nil {
			fmt.Println(err.Error())
		}
	} else {
		http.NotFound(w, r)
	}
}

func pagePostHandler(w http.ResponseWriter, r *http.Request) {
	c := app.ReadCookies(r)
	if c.Value == userAuth.Login {
		var gopher models.Gopher
		if r.Method == "GET" {
			r.ParseForm()
			gopher.Title = r.FormValue("title")
			gopher.Content = r.FormValue("content")
			gopher.Creator = userAuth.Login
			gopher.Date = time.Now().Format("2006-01-02 15:04")
			gopher.Like = 0
			gopher.View = 1
			gopher.Owner = userAuth.Login
			// gopher.Owner = r.FormValue()
			fmt.Println("My", gopher)
			err := database.InsertGopher(gopher)
			if err != nil {
				fmt.Println("Error - pagePostHandler() InsertGopher() r.Method == GET")
			}
			http.Redirect(w, r, "/page", http.StatusSeeOther)
		}
		if r.Method == "POST" {
			r.ParseForm()
			gopher.Title = r.FormValue("title")
			gopher.Content = r.FormValue("content")
			gopher.Creator = userAuth.Login
			gopher.Date = time.Now().Format("2006-01-02 15:04")
			gopher.Like = 0
			gopher.View = 1
			gopher.Owner = guestId
			fmt.Println("Guest", gopher)
			err := database.InsertGopher(gopher)
			if err != nil {
				fmt.Println("Error - pagePostHandler() InsertGopher() r.Method == POST")
			}
			http.Redirect(w, r, "/guest", http.StatusSeeOther)
		}
	} else if c.Value != "" {
		var err error
		userAuth.Login = c.Value
		userAuth, err = database.SelectUserByColumn("Login", userAuth.Login)
		if err != nil {
			fmt.Println(err.Error())
		}
	} else {
		http.NotFound(w, r)
	}
}

func pageDelGofHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		r.ParseForm()
		gof := r.FormValue("gof")
		if gof != "" {
			id, err := strconv.Atoi(gof)
			if err != nil {
				http.NotFound(w, r)
			}
			err = database.DeleteGopherByUser(userAuth.Login, id)
			if err != nil {
				fmt.Println("Error - pageDelGofHandler() DeleteGopherByUser()", err.Error())
			}
		}
	}
	if r.Method == "GET" {
		r.ParseForm()
		gof := r.FormValue("rep_id")
		id, _ := strconv.Atoi(gof)
		if gof != "" {
			err := database.DeleteRepoPostByUser(id, userAuth.Login)
			if err != nil {
				fmt.Println("Error - pageDelGofHandler() DeleteRepoPostByUser()", err.Error())
			}
		}
	}
	http.Redirect(w, r, "/page#", http.StatusSeeOther)
}

func pageRepGofHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		r.ParseForm()
		gof := r.FormValue("gof")
		if gof != "" {
			id, err := strconv.Atoi(gof)
			if err != nil {
				http.NotFound(w, r)
			}
			goph := database.SelectGopherById(id)
			fmt.Println(goph)
			var newGof models.Gopher
			newGof.Creator = goph.Creator
			newGof.Content = goph.Content
			newGof.Date = goph.Date
			newGof.Like = 0
			newGof.Owner = userAuth.Login
			newGof.Title = goph.Title
			newGof.View = 1
			err = database.InsertGopher(newGof)
			if err != nil {
				fmt.Println("Error - pageRepGofHandler() InsertGopher()", err.Error())
			}
		}
	}
	if r.Method == "GET" {
		r.ParseForm()
		gof := r.FormValue("gof2")
		if gof != "" {
			id, err := strconv.Atoi(gof)
			if err != nil {
				http.NotFound(w, r)
			}
			var post models.Repost
			post.Post = id
			post.User = userAuth.Login
			_, err = database.InsertRepoPost(post)
			if err != nil {
				fmt.Println("Error - pageRepGofHandler() InsertRepoPost()", err.Error())
			}

		}
	}
	http.Redirect(w, r, "/page", http.StatusSeeOther)
}

func commentsHandler(w http.ResponseWriter, r *http.Request) {
	c := app.ReadCookies(r)
	if c.Value == userAuth.Login {
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
				comment.Posts, _ = strconv.Atoi(postId) // err
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
		tmpl.ExecuteTemplate(w, "footer", title)
		tmpl.ExecuteTemplate(w, "comments", data)
	} else if c.Value != "" {
		var err error
		userAuth.Login = c.Value
		userAuth, err = database.SelectUserByColumn("Login", userAuth.Login)
		if err != nil {
			fmt.Println(err.Error())
		}
	} else {
		http.NotFound(w, r)
	}
}

func friendsHandler(w http.ResponseWriter, r *http.Request) {
	c := app.ReadCookies(r)
	if c.Value == userAuth.Login {
		tmpl, err := template.ParseFiles("html/friends.html", "html/header.html", "html/footer.html")
		if err != nil {
			http.NotFound(w, r)
		}

		r.ParseForm()
		if r.Method == "POST" {
			id = r.FormValue("friend_id")
			fmt.Println(id)
			idN, err := database.DeleteFriendByLogin(userAuth.Login, id)
			if err != nil {
				fmt.Println("Error - friendsDelHandler()1 DeleteFriendByLogin()")
			}
			err = database.InsertUserSub(userAuth.Login, idN)
			if err != nil {
				fmt.Println("Error - InsertUserSub() DeleteFriendsById()", err)
			}
		}
		if r.Method == "GET" {
			//r.ParseForm()
			guestId = r.FormValue("Id")
			fmt.Println("GET friendsHandler()", guestId)
			_, err = database.SelectMessengeListbyUsers(userAuth.Login, guestId)
			if err != nil {
				fmt.Println("Error - friendsHandler() SelectMessengeListbyUsers()")
			}
		}

		rec := database.SelectRecomendationFriends(userAuth.Login)
		subs := database.SelectUserSub(userAuth.Login)
		friends := database.SelectAllFriendsUser(userAuth.Login)
		online := database.SelectOnlineFriends(userAuth.Login)

		data := map[string]interface{}{"User": userAuth, "Friends": friends, "Subs": subs, "Rec": rec, "Online": online, "Done": "Friends"}
		title := map[string]interface{}{"Title": models.Cfg.FriendsTitile, "User": userAuth}
		tmpl.ExecuteTemplate(w, "header", title)
		tmpl.ExecuteTemplate(w, "footer", title)
		tmpl.ExecuteTemplate(w, "friends", data)
	} else if c.Value != "" {
		var err error
		userAuth.Login = c.Value
		userAuth, err = database.SelectUserByColumn("Login", userAuth.Login)
		if err != nil {
			fmt.Println(err.Error())
		}
	} else {
		http.NotFound(w, r)
	}
}

func addFriendsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		r.ParseForm()
		sub := r.FormValue("friend_id")
		frd, _ := database.CheckFriends(userAuth.Login, sub)
		if frd.Login == "" {
			var friend = models.Friends{
				Login:  userAuth.Login,
				Status: "Friend",
				Friend: sub,
			}
			var friend1 = models.Friends{
				Login:  sub,
				Status: "Friend",
				Friend: userAuth.Login,
			}
			_, err := database.InsertFriends(friend)
			if err != nil {
				fmt.Println("Error - addFriendsHandler() InsertFriends1()", err)
			}
			_, err = database.InsertFriends(friend1)
			if err != nil {
				fmt.Println("Error - addFriendsHandler() InsertFriends2()", err)
			}
			err = database.DeleteUserSub(userAuth.Login, sub)
			if err != nil {
				fmt.Println("Error - addFriendsHandler() DeleteUserSub()", err)
			}
			var val models.MessageList
			arr := database.SelectMessengeListbyLogin(userAuth.Login, sub)
			if arr == nil {
				err := app.CreateFile(val, sub, userAuth.Login)
				if err != nil {
					fmt.Println("Error - addFriendsHandler() CreateFile()", err.Error())
				}
			} else {
				fmt.Println(arr)
			}

			http.Redirect(w, r, "/friends", http.StatusSeeOther)
		}
	}
}

func recFriendsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		r.ParseForm()
		sub := r.FormValue("friend_id")
		err := database.InsertUserSub(sub, userAuth.Login)
		if err != nil {
			fmt.Println("Error - recFriendsHandler() InsertUserSub()", err)
		}
		http.Redirect(w, r, "/friends", http.StatusSeeOther)
	}
}

func guestFriendsHandler(w http.ResponseWriter, r *http.Request) {
	c := app.ReadCookies(r)
	if c.Value == userAuth.Login {
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
			// guestId = r.FormValue("Id2")
			fmt.Println("GET - guestFriendsHandler", guestId)
		}
		gst, _ := database.SelectUserByColumn("Login", guestId)
		friends := database.SelectAllFriendsUser(guestId)
		rec := database.SelectRecomendationFriends(guestId)
		subs := database.SelectUserSub(guestId)
		online := database.SelectOnlineFriends(guestId)

		data := map[string]interface{}{"User": userAuth, "Friends": friends, "Guest": gst, "Subs": subs, "Rec": rec, "Done": "GuestFR", "Online": online}
		title := map[string]interface{}{"Title": models.Cfg.FriendsTitile, "User": userAuth}
		tmpl.ExecuteTemplate(w, "header", title)
		tmpl.ExecuteTemplate(w, "footer", title)
		tmpl.ExecuteTemplate(w, "friends", data)
	} else if c.Value != "" {
		var err error
		userAuth.Login = c.Value
		userAuth, err = database.SelectUserByColumn("Login", userAuth.Login)
		if err != nil {
			fmt.Println(err.Error())
		}
	} else {
		http.NotFound(w, r)
	}
}

func communitiesHandler(w http.ResponseWriter, r *http.Request) {
	c := app.ReadCookies(r)
	if c.Value == userAuth.Login {
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
		//	comWithOutSub := database.SelectCommunitiesWithOutSub(userAuth.Login)
		catComm := database.SelectCommunitiesCategory()
		communities := database.SelectAllCommunitiesUser("User", userAuth.Login)
		recCommunities := database.SelectRecCommunities(userAuth.Login)
		data := map[string]interface{}{"User": userAuth, "Communities": communities, "Done": done, "RecCom": recCommunities, "CommCat": catComm}
		title := map[string]interface{}{"Title": models.Cfg.CommunitiesTitile, "User": userAuth}
		tmpl.ExecuteTemplate(w, "header", title)
		tmpl.ExecuteTemplate(w, "footer", title)
		tmpl.ExecuteTemplate(w, "communities", data)
	} else if c.Value != "" {
		var err error
		userAuth.Login = c.Value
		userAuth, err = database.SelectUserByColumn("Login", userAuth.Login)
		if err != nil {
			fmt.Println(err.Error())
		}
	} else {
		http.NotFound(w, r)
	}
}

func communitiesAddHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		r.ParseForm()

		file, fileHeader, err := r.FormFile("photo")
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		defer file.Close()
		imgPath = fmt.Sprintf("./data/image/blog/%d%s", time.Now().UnixNano(), filepath.Ext(fileHeader.Filename))
		dst, err := os.Create(imgPath)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		defer dst.Close()

		_, err = io.Copy(dst, file)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		r.ParseForm()
		var communities models.Communities
		communities.Author = userAuth.Login
		communities.Category = r.FormValue("selectuser")
		communities.Name = r.FormValue("title")
		communities.Photo = imgPath
		fmt.Println(communities)

		communitiesName = communities.Name

		_, err = database.InsertCommunities(communities)
		if err != nil {
			fmt.Println("Error - communitiesAddHandler() InsertCommunities()", err)
			http.NotFound(w, r)
		}
		err = database.InsertSubscribersToUser(userAuth.Login, communitiesName)
		if err != nil {
			fmt.Println("Error - communitiesAddHandler() InsertSubscribersToUser()", err)
		}
		http.Redirect(w, r, "/community", http.StatusSeeOther)
	}
}

func guestCommunitiesHandler(w http.ResponseWriter, r *http.Request) {
	c := app.ReadCookies(r)
	if c.Value == userAuth.Login {
		tmpl, err := template.ParseFiles("html/communities.html", "html/header.html", "html/footer.html")
		if err != nil {
			http.NotFound(w, r)
		}
		fmt.Println(guestLogin)
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
		var comWithOutSub []models.Communities
		var communities []models.JoinCommunities
		if guestLogin != "" {
			comWithOutSub = database.SelectCommunitiesWithOutSub(guestLogin)
			communities = database.SelectAllCommunitiesUser("User", guestLogin)
		} else {
			comWithOutSub = database.SelectCommunitiesWithOutSub(guestId)
			communities = database.SelectAllCommunitiesUser("User", guestId)
		}

		data := map[string]interface{}{"User": userAuth, "Communities": communities, "RecCommunities": comWithOutSub, "Done": done}
		// data := map[string]interface{}{"User": userAuth, "Communities": communities, "Done": done, "RecCom": recCommunities, "CommCat": catComm}
		title := map[string]interface{}{"Title": models.Cfg.CommunitiesTitile, "User": userAuth}
		tmpl.ExecuteTemplate(w, "header", title)
		tmpl.ExecuteTemplate(w, "footer", title)
		tmpl.ExecuteTemplate(w, "communities", data)
	} else if c.Value != "" {
		var err error
		userAuth.Login = c.Value
		userAuth, err = database.SelectUserByColumn("Login", userAuth.Login)
		if err != nil {
			fmt.Println(err.Error())
		}
	} else {
		http.NotFound(w, r)
	}
}

func communityHandler(w http.ResponseWriter, r *http.Request) {
	c := app.ReadCookies(r)
	if c.Value == userAuth.Login {
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
		if communitiesName == "" && r.Method == "POST" {
			r.ParseForm()
			communitiesName = r.FormValue("community_id")
		}
		if communitiesName != "" {
			post.Communities = communitiesName
			post.CommunitiesPhot = communitiesPhoto
		}
		var Foo Comm
		Foo.Name = post.Communities
		Foo.Photo = post.CommunitiesPhot

		catComm := database.SelectCommunitiesCategory()
		posts = database.SelectPostByCommunities(post.Communities)
		subs := database.SelectSubscribersBtCommunities(post.Communities)
		author, names := database.SelectCommunitiesAuthorByName(post.Communities)
		category := database.SelectPostCategory()
		Users := database.SelectSubscribersBtCommunities(post.Communities)
		store, err := database.SelectStoreItemsByCommunity(post.Communities)
		if err != nil {
			fmt.Println(err)
		}
		comm := database.SelectCommunitiesByColumn("Name", post.Communities)
		fmt.Println("SelectCommunitiesByColumn", comm.Name)
		ok := false
		if r.Method == "GET" {
			fmt.Println("GET")
			r.ParseForm()
			postId = r.FormValue("postId")
			val, _ := strconv.Atoi(postId)
			err := database.UpdateLikeInPost(val)
			if err != nil {
				fmt.Println("Error - communityHandler() UpdateLikeInPost()")
			}
			postId = ""
			if communitiesName != "" {
				_, ok = database.CheckCommunity(userAuth.Login, communitiesName)
				fmt.Println("31231", ok)
			} else {
				_, ok = database.CheckCommunity(userAuth.Login, post.Communities)
				fmt.Println("45543", ok)
			}
		}

		if r.Method == "POST" {
			fmt.Println("POST")
			r.ParseForm()

			guestId = r.FormValue("guestId")
			fmt.Println("guestId", guestId)

			postId = r.FormValue("postId")
			fmt.Println("POST", postId)
			err = database.UpdateViewInCommunityPost(communitiesName)
			if err != nil {
				fmt.Println("Error - communityHandler() UpdateViewInPost()")
			}
		}

		title := map[string]interface{}{"Title": Foo.Name, "User": userAuth}
		blog := map[string]interface{}{"Post": posts, "User": userAuth, "Users": Users, "Subs": subs, "Author": author, "Names": names, "Communities": Foo, "PostCat": category, "CommCat": catComm, "SetCom": comm, "OK": ok, "Store": store}
		tmpl.ExecuteTemplate(w, "header", title)
		tmpl.ExecuteTemplate(w, "footer", title)
		tmpl.ExecuteTemplate(w, "community", blog)
	} else if c.Value != "" {
		var err error
		userAuth.Login = c.Value
		userAuth, err = database.SelectUserByColumn("Login", userAuth.Login)
		if err != nil {
			fmt.Println(err.Error())
		}
	} else {
		http.NotFound(w, r)
	}
}

func communityAuthorHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		r.ParseForm()
		com := r.FormValue("commun")
		user := r.FormValue("userSel")
		fmt.Println(com, user)

		upCom := database.SelectCommunitiesByColumn("Name", com)

		upCom.Author = user

		err := database.UpdateCommunityAuthor(upCom, com)
		if err != nil {
			fmt.Println(err.Error())
		}

		http.Redirect(w, r, "/community", http.StatusSeeOther)
	}
}

func communityDelHandler(w http.ResponseWriter, r *http.Request) { // сделать удаление сначала всех подписок с такой группой, а потом саму группу
	if r.Method == "POST" {
		r.ParseForm()
		value := r.FormValue("inputName")
		fmt.Println("communityDelHandler() ", value)
		err := database.DeleteSubAllOnCommunity(value)
		if err != nil {
			fmt.Println("Error - communityDelHandler() DeleteSubAllOnCommunity()", err)
		}
		err = database.DeleteCommunitiesByName(value)
		if err != nil {
			fmt.Println("Error - communityDelHandler() DeleteCommunitiesByName()", err)
		}
		http.Redirect(w, r, "/communities", http.StatusSeeOther)
	}
}

func communityPostHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method == "POST" {
		r.ParseForm()

		file, fileHeader, err := r.FormFile("photo")
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		defer file.Close()
		imgPath = fmt.Sprintf("./data/image/blog/%d%s", time.Now().UnixNano(), filepath.Ext(fileHeader.Filename))
		dst, err := os.Create(imgPath)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		defer dst.Close()

		_, err = io.Copy(dst, file)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		r.ParseForm()
		title = r.FormValue("title")
		content = r.FormValue("content")
		var post models.Posts
		RandomCrypto, _ := rand.Prime(rand.Reader, 32)
		post.Id = fmt.Sprint(RandomCrypto.Int64() / 20000)
		post.Title = title
		post.Content = content
		post.Like = 0
		post.View = 1
		post.Date = time.Now().Format("2006-01-02 15:04")
		post.Communities = communitiesName
		post.Photo = imgPath
		post.Category = r.FormValue("selectuser")
		fmt.Println(post)

		_, err = database.InsertPost(post)
		if err != nil {
			fmt.Println("Error - communityPostHandler() InsertPost()", err)
			http.NotFound(w, r)
		}
		http.Redirect(w, r, "/community", http.StatusSeeOther)
	}

	if r.Method == "GET" {
		r.ParseForm()
		time := r.FormValue("time")
		fmt.Println(time)
		err := database.DeleteCommunitiesPostByTime(communitiesName, time)
		if err != nil {
			fmt.Println("Error - communityPostHandler() DeleteCommunitiesPostByTime()", err)
			http.NotFound(w, r)
		}
		http.Redirect(w, r, "/community", http.StatusSeeOther)
	}
}

func communityEditHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method == "POST" {
		r.ParseForm()
		group := database.SelectCommunitiesByColumn("Name", communitiesName)
		file, fileHeader, err := r.FormFile("photo")
		if err != nil {
			// http.Error(w, err.Error(), http.StatusBadRequest)
			// return
			imgPath = group.Photo
		} else {
			defer file.Close()
			imgPath = fmt.Sprintf("./data/image/blog/%d%s", time.Now().UnixNano(), filepath.Ext(fileHeader.Filename))
			dst, err := os.Create(imgPath)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			defer dst.Close()

			_, err = io.Copy(dst, file)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		r.ParseForm()
		var communities models.Communities
		communities.Author = userAuth.Login
		communities.Category = r.FormValue("selectuser")
		communities.Name = r.FormValue("title")
		communities.Photo = imgPath
		fmt.Println("New Communities", communities)

		communitiesName = communities.Name

		err = database.UpdateCommunity(communities, group.Name)
		if err != nil {
			fmt.Println("Error - communityEditHandler() UpdateCommunity()", err)
			http.NotFound(w, r)
		}

		http.Redirect(w, r, "/community", http.StatusSeeOther)

	}
}

func communityMarketHandler(w http.ResponseWriter, r *http.Request) {
	c := app.ReadCookies(r)
	if c.Value == userAuth.Login {
		tmpl, err := template.ParseFiles("html/store.html", "html/header.html", "html/footer.html")
		if err != nil {
			http.NotFound(w, r)
		}
		var Ok bool
		Access := false
		Ok = true
		store, err := database.SelectStoreItemsByCommunity(post.Communities)
		if err != nil {
			fmt.Println(err)
		}
		if store == nil {
			Ok = false
		}

		grp := database.SelectCommunitiesByColumn("Name", post.Communities)
		if grp.Author == userAuth.Login {
			Access = true
		}

		sex, err := database.SelectSex()
		if err != nil {
			fmt.Println(err)
		}
		cat, err := database.SelectStoreCategory()
		if err != nil {
			fmt.Println(err)
		}

		url := "/community/market/sort"
		root := "/community/market"

		user, _ := database.SelectUserWallet(userAuth.Login, userAuth.Password)
		title := map[string]interface{}{"Title": "Товары - " + communitiesName, "User": userAuth}
		tmpl.ExecuteTemplate(w, "header", title)
		tmpl.ExecuteTemplate(w, "footer", title)
		data := map[string]interface{}{"User": userAuth, "Wallet": user.Wallet, "Store": store, "OK": Ok, "Access": Access, "Sex": sex, "Category": cat, "Community": communitiesName, "URL": url, "Root": root}
		tmpl.ExecuteTemplate(w, "store", data)
	} else if c.Value != "" {
		var err error
		userAuth.Login = c.Value
		userAuth, err = database.SelectUserByColumn("Login", userAuth.Login)
		if err != nil {
			fmt.Println(err.Error())
		}
	} else {
		http.NotFound(w, r)
	}
}

func communityMarketSortHandler(w http.ResponseWriter, r *http.Request) {
	c := app.ReadCookies(r)
	if c.Value == userAuth.Login {
		tmpl, err := template.ParseFiles("html/store.html", "html/header.html", "html/footer.html")
		if err != nil {
			http.NotFound(w, r)
		}
		var store []models.Store
		if r.Method == "GET" {
			r.ParseForm()
			val := r.FormValue("Condition")
			if val != "" {
				store, err = database.SelectStoreItemsByCommunityAndCategory(val, communitiesName)
				if err != nil {
					fmt.Println(err)
				}
			}
		}

		Access := false
		Ok := true
		if store == nil {
			Ok = false
		}

		grp := database.SelectCommunitiesByColumn("Name", communitiesName)
		if grp.Author == userAuth.Login {
			Access = true
		}

		sex, err := database.SelectSex()
		if err != nil {
			fmt.Println(err)
		}
		cat, err := database.SelectStoreCategory()
		if err != nil {
			fmt.Println(err)
		}

		user, _ := database.SelectUserWallet(userAuth.Login, userAuth.Password)
		url := "/community/market/sort"
		root := "/community/market"

		// title := map[string]string{"Title": models.Cfg.StoreTitle}
		// tmpl.ExecuteTemplate(w, "header", title)
		// data := map[string]interface{}{"User": userAuth, "Wallet": user.Wallet, "Store": store, "OK": Ok, "Access": Access, "Category": ct, "URL": url}
		// tmpl.ExecuteTemplate(w, "store", data)

		title := map[string]interface{}{"Title": models.Cfg.StoreTitle, "User": userAuth}
		tmpl.ExecuteTemplate(w, "header", title)
		tmpl.ExecuteTemplate(w, "footer", title)
		data := map[string]interface{}{"User": userAuth, "Wallet": user.Wallet, "Store": store, "OK": Ok, "Access": Access, "Sex": sex, "Category": cat, "Community": communitiesName, "URL": url, "Root": root}
		tmpl.ExecuteTemplate(w, "store", data)
	} else if c.Value != "" {
		var err error
		userAuth.Login = c.Value
		userAuth, err = database.SelectUserByColumn("Login", userAuth.Login)
		if err != nil {
			fmt.Println(err.Error())
		}
	} else {
		http.NotFound(w, r)
	}
}

func communityMarketAddHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		r.ParseForm()
		file, fileHeader, err := r.FormFile("photo")
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		defer file.Close()
		imgPath = fmt.Sprintf("./data/image/product/%d%s", time.Now().UnixNano(), filepath.Ext(fileHeader.Filename))
		dst, err := os.Create(imgPath)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		defer dst.Close()

		_, err = io.Copy(dst, file)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		r.ParseForm()
		var product models.Store
		product.Category = r.FormValue("selectcat")
		product.Community = communitiesName
		product.Description = r.FormValue("content")
		product.Name = r.FormValue("title")
		fo1 := r.FormValue("price")
		fo2, _ := strconv.Atoi(fo1)
		product.Price = float32(fo2)
		fo3 := r.FormValue("newprice")
		var fo4 int
		if fo3 == "" {
			fo4 = fo2
		} else {
			fo4, _ = strconv.Atoi(fo3)
		}
		product.NewPrice = float32(fo4)
		product.Sex = r.FormValue("selectsex")
		product.Photo = app.TrimLeftChar(imgPath)
		fmt.Println(product)
		err = database.InsertToStoreProduct(product)
		if err != nil {
			fmt.Println(err)
			http.NotFound(w, r)
		}
	}
	http.Redirect(w, r, "/community/market", http.StatusSeeOther)
}

func communityMarketDelHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		r.ParseForm()
		data := r.FormValue("selectprod")
		fmt.Println(data)
		stringarr := strings.Split(data, ";")
		// ip, port := s[0], s[1]
		fmt.Println(stringarr[0])
		id := strings.Split(stringarr[0], ":")
		ID, _ := strconv.Atoi(id[1])
		fmt.Println(stringarr[1])
		fmt.Println(stringarr[2])
		fmt.Println(ID)
		err := database.DeleteStore(ID)
		if err != nil {
			fmt.Println(err)
		}
		http.Redirect(w, r, "/community/market", http.StatusSeeOther)
	}
}

func communityStoreCardHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		r.ParseForm()
		itemId = r.FormValue("store_id")
		url := fmt.Sprintf("/store/card?store_id=%s", itemId)
		http.Redirect(w, r, url, http.StatusSeeOther)
	}
}

func communityMarketSaleListHandler(w http.ResponseWriter, r *http.Request) {
	var sales []models.JoinStorePlus
	var ok bool
	var Done string
	if r.Method == "GET" {
		r.ParseForm()
		comName := r.FormValue("community")
		if comName != "" {
			com := database.SelectCommunitiesByColumn("Name", comName)
			if com.Author == userAuth.Login {
				sales, ok = database.SelectSalesByCommunity(comName)
				if ok {
					Done = "Store"
				}
				fmt.Println(ok)
				tmpl, err := template.ParseFiles("html/list.html", "html/header.html", "html/footer.html")
				if err != nil {
					http.NotFound(w, r)
				}
				ttl := "Список продаж " + comName
				title := map[string]interface{}{"Title": ttl, "User": userAuth}
				data := map[string]interface{}{"Sales": sales, "Done": Done, "Community": comName, "Title": ttl}
				tmpl.ExecuteTemplate(w, "header", title)
				tmpl.ExecuteTemplate(w, "footer", title)
				tmpl.ExecuteTemplate(w, "list", data)
			} else {
				http.Redirect(w, r, "/page", http.StatusSeeOther)
			}
		} else {
			http.Redirect(w, r, "/page", http.StatusSeeOther)
		}
	}
	if r.Method == "POST" {
		r.ParseForm()
		comName := r.FormValue("community")
		if comName != "" {
			com := database.SelectCommunitiesByColumn("Name", comName)
			if com.Author == userAuth.Login {
				sales, _ = database.SelectSalesByCommunity(comName)
				app.Excel(sales)
				http.Redirect(w, r, "/community/market", http.StatusSeeOther)
			}
		} else {
			http.Redirect(w, r, "/page", http.StatusSeeOther)
		}
	}
}

func communityMarketStatisticsHandler(w http.ResponseWriter, r *http.Request) {
	c := app.ReadCookies(r)
	if c.Value == userAuth.Login {
		tmpl, err := template.ParseFiles("html/statistics.html", "html/header.html", "html/footer.html")
		if err != nil {
			http.NotFound(w, r)
		}

		if r.Method == "GET" {
			r.ParseForm()
			comName := r.FormValue("community")

			title := map[string]interface{}{"Title": fmt.Sprintf("Статистика - %s", comName), "User": userAuth}
			data := map[string]interface{}{}
			tmpl.ExecuteTemplate(w, "header", title)
			tmpl.ExecuteTemplate(w, "footer", title)
			tmpl.ExecuteTemplate(w, "stat", data)
		} else {
			http.Redirect(w, r, "/page", http.StatusSeeOther)
		}
	} else if c.Value != "" {
		var err error
		userAuth.Login = c.Value
		userAuth, err = database.SelectUserByColumn("Login", userAuth.Login)
		if err != nil {
			fmt.Println(err.Error())
		}
	} else {
		http.NotFound(w, r)
	}
}

func guestHandler(w http.ResponseWriter, r *http.Request) {
	c := app.ReadCookies(r)
	if c.Value == userAuth.Login {
		tmpl, err := template.ParseFiles("html/guest.html", "html/header.html", "html/footer.html")
		if err != nil {
			http.NotFound(w, r)
		}
		var ok bool
		fmt.Println("guest", guestId)
		if r.Method == "GET" {
			r.ParseForm()
			guestLogin = r.FormValue("guestLogin")

			_, ok = database.CheckFriends(userAuth.Login, guestLogin)
			if userAuth.Login == guestLogin {
				ok = false
			}
			fmt.Println("guestLogin", guestLogin, ok)
		}
		if r.Method == "POST" {
			GofId := r.FormValue("goLike")
			fmt.Println(GofId)
			if GofId != "" {
				err = database.InsertLikeToGopher(GofId)
				if err != nil {
					fmt.Println(err.Error())
				}
			}
		}
		_, ok = database.CheckFriends(userAuth.Login, guestId)
		if userAuth.Login == guestId {
			ok = false
		}

		userGuest, err := database.SelectUserByColumn("Login", guestId)
		if err != nil {
			fmt.Println("Error - guestHandler() SelectUsersByColumn()")
		}
		// fmt.Println(userGuest)

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
		var DoneGopher bool
		gopher := database.SelectGopherByOwner(userGuest.Login)
		if gopher != nil {
			DoneGopher = true
		} else {
			DoneGopher = false
		}

		var Friend bool
		_, Friend = database.CheckFriends(userAuth.Login, userGuest.Login)

		Access := false
		if userAuth.Access != "User" && userAuth.Login != userGuest.Login {
			Access = true
		}
		var GalOK bool
		photo := database.SelectUserGalleryLimit(userGuest.Login)
		if photo != nil {
			GalOK = true
		} else {
			GalOK = false
		}

		complaint := []string{"Оскорбление личности", "Оскорбление вероисповедания", "Распространение запрещенных веществ", "Обман на деньги", "Украл аккаунт"}

		title := map[string]interface{}{"Title": userGuest.Name, "User": userAuth}
		tmpl.ExecuteTemplate(w, "header", title)
		tmpl.ExecuteTemplate(w, "footer", title)

		sendUser := map[string]interface{}{"User": userAuth, "Repo": data, "Guest": userGuest, "Statistics": stat, "Done": Done, "OK": ok, "Gopher": gopher, "DoneGopher": DoneGopher, "Friend": Friend, "Access": Access, "Complaint": complaint, "GalOK": GalOK}

		tmpl.ExecuteTemplate(w, "guest", sendUser)
	} else if c.Value != "" {
		var err error
		userAuth.Login = c.Value
		userAuth, err = database.SelectUserByColumn("Login", userAuth.Login)
		if err != nil {
			fmt.Println(err.Error())
		}
	} else {
		http.NotFound(w, r)
	}
}

func messageHandler(w http.ResponseWriter, r *http.Request) {
	c := app.ReadCookies(r)
	if c.Value == userAuth.Login {

		tmpl, err := template.ParseFiles("html/message.html", "html/header.html", "html/footer.html")
		if err != nil {
			http.NotFound(w, r)
		}

		//createFile(check)
		companion = database.SelectCompanionsByLogin(userAuth.Login)
		OKS := true
		r.ParseForm()
		usrMesg = r.FormValue("user_id")
		activeChatUser, err = database.SelectUserByColumn("Login", usrMesg)
		if err != nil {
			fmt.Println("Error - messageHandler() SelectUserByColumn()", err)
		}
		if activeChatUser.Name == "" {
			OK = false
			OKS = false
		} else {
			OK = true
			OKS = true
		}

		if r.Method == "GET" {
			r.ParseForm()
			usrMesg = r.FormValue("user_id")
			UsersLink, err = database.SelectMessengeListbyUsers(userAuth.Login, usrMesg)
			if err != nil {
				fmt.Println("Error - messageHandler() SelectMessengeListbyUsers()", err)
			}
			// activeChatUser, err = database.SelectUserByColumn("Login", usrMesg)
			// if err != nil {
			// 	fmt.Println("Error - messageHandler() SelectUserByColumn()", err)
			// }

			if UsersLink.MessageHistory != "" {

				Link = fmt.Sprintf("C:/Users/admin/go/src/go-blog/data/files/message/%s", UsersLink.MessageHistory)
				msg := models.Message{
					User:    userAuth.Login,
					Message: "",
					Data:    time.Now().Format("2006-01-02 15:04"),
					Photo:   userAuth.Photo,
				}
				Messenger, err = app.JSON(msg, Link, userAuth.Login)
				if err != nil {
					log.Fatal(err.Error())
				}
				if Messenger.Messenge == nil {
					OK = false
					OKS = false
				} else {
					OK = true
					OKS = true
				}
			}

		}
		fmt.Println("UsersLink2222", UsersLink)
		//Link = fmt.Sprintf("C:/Users/admin/go/src/go-blog/data/files/message/%s", UsersLink.MessageHistory)
		if r.Method == "POST" {

			r.ParseForm()

			message := r.FormValue("commentsInput")
			//Link = fmt.Sprintf("C:/Users/admin/go/src/go-blog/data/files/message/%s", UsersLink.MessageHistory)
			if message != "" {

				//Link = "C:/Users/admin/go/src/go-blog/data/files/message/test.json"
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
				// err = app.JSON(msg, Link, userAuth.Login)
				// if err != nil {
				// 	log.Fatal(err.Error())
				// }
				Messenger, err = app.JSON(msg, Link, userAuth.Login)
				if err != nil {
					log.Fatal(err.Error())
				}
				fmt.Println(Messenger)
			}
		}

		title := map[string]interface{}{"Title": models.Cfg.MessageTitle, "User": userAuth}
		tmpl.ExecuteTemplate(w, "header", title)
		tmpl.ExecuteTemplate(w, "footer", title)
		fmt.Println(OK)
		data := map[string]interface{}{"User": userAuth, "Done": OK, "OK": OKS, "Companions": companion, "ChatUser": activeChatUser, "Chat": Messenger.Messenge}
		tmpl.ExecuteTemplate(w, "message", data)
	} else if c.Value != "" {
		var err error
		userAuth.Login = c.Value
		userAuth, err = database.SelectUserByColumn("Login", userAuth.Login)
		if err != nil {
			fmt.Println(err.Error())
		}
	} else {
		http.NotFound(w, r)
	}
}

func storeHandler(w http.ResponseWriter, r *http.Request) {
	c := app.ReadCookies(r)
	if c.Value == userAuth.Login {
		tmpl, err := template.ParseFiles("html/store.html", "html/header.html", "html/footer.html")
		if err != nil {
			http.NotFound(w, r)
		}
		Access := false
		Ok := true
		store, err := database.SelectStoreItems()
		if err != nil {
			fmt.Println(err)
		}
		if store == nil {
			Ok = false
		}

		user, _ := database.SelectUserWallet(userAuth.Login, userAuth.Password)
		ct, err := database.SelectStoreCategory()
		if err != nil {
			fmt.Println("Error - storeHandler() SelectStoreCategory()", err.Error())
		}

		url := "store/sort"
		root := "/store"

		title := map[string]interface{}{"Title": models.Cfg.StoreTitle, "User": userAuth}
		tmpl.ExecuteTemplate(w, "header", title)
		tmpl.ExecuteTemplate(w, "footer", title)
		data := map[string]interface{}{"User": userAuth, "Wallet": user.Wallet, "Store": store, "OK": Ok, "Access": Access, "Category": ct, "URL": url, "Root": root}
		tmpl.ExecuteTemplate(w, "store", data)
	} else if c.Value != "" {
		var err error
		userAuth.Login = c.Value
		userAuth, err = database.SelectUserByColumn("Login", userAuth.Login)
		if err != nil {
			fmt.Println(err.Error())
		}
	} else {
		http.NotFound(w, r)
	}
}

func storeSortHandler(w http.ResponseWriter, r *http.Request) {
	c := app.ReadCookies(r)
	if c.Value == userAuth.Login {
		tmpl, err := template.ParseFiles("html/store.html", "html/header.html", "html/footer.html")
		if err != nil {
			http.NotFound(w, r)
		}
		var store []models.Store
		if r.Method == "GET" {
			r.ParseForm()
			val := r.FormValue("Condition")
			if val != "" {
				store, err = database.SelectStoreItemsByCategory(val)
				if err != nil {
					fmt.Println(err)
				}
			}
		}

		Access := false
		Ok := true
		if store == nil {
			Ok = false
		}

		user, _ := database.SelectUserWallet(userAuth.Login, userAuth.Password)
		ct, err := database.SelectStoreCategory()
		if err != nil {
			fmt.Println("Error - storeHandler() SelectStoreCategory()", err.Error())
		}
		url := ""
		root := "/store"

		title := map[string]interface{}{"Title": models.Cfg.StoreTitle, "User": userAuth}
		tmpl.ExecuteTemplate(w, "header", title)
		tmpl.ExecuteTemplate(w, "footer", title)
		data := map[string]interface{}{"User": userAuth, "Wallet": user.Wallet, "Store": store, "OK": Ok, "Access": Access, "Category": ct, "URL": url, "Root": root}
		tmpl.ExecuteTemplate(w, "store", data)
	} else if c.Value != "" {
		var err error
		userAuth.Login = c.Value
		userAuth, err = database.SelectUserByColumn("Login", userAuth.Login)
		if err != nil {
			fmt.Println(err.Error())
		}
	} else {
		http.NotFound(w, r)
	}
}

func storeCardHandler(w http.ResponseWriter, r *http.Request) {
	c := app.ReadCookies(r)
	if c.Value == userAuth.Login {
		tmpl, err := template.ParseFiles("html/card.html", "html/header.html", "html/footer.html")
		if err != nil {
			http.NotFound(w, r)
		}
		user, _ := database.SelectUserWallet(userAuth.Login, userAuth.Password)
		if r.Method == "GET" {
			r.ParseForm()
			itemId = r.FormValue("store_id")
		}
		Id, _ := strconv.Atoi(itemId)
		product, err = database.SelectStoreItemById(Id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
		}

		title := map[string]interface{}{"Title": product.Name, "User": userAuth}
		tmpl.ExecuteTemplate(w, "header", title)
		tmpl.ExecuteTemplate(w, "footer", title)

		var products []models.StorePlus
		arr, _ := database.SelectFavouritesByUserCheck(userAuth.Login)
		fmt.Println(arr)
		for _, i := range arr {
			if product.Id == uint(i) {
				product.Status = true
			} else {
				product.Status = false
			}
		}
		fmt.Println(product)
		products = append(products, product)
		data := map[string]interface{}{"User": userAuth, "Wallet": user.Wallet, "Market": products}
		tmpl.ExecuteTemplate(w, "card", data)
	} else if c.Value != "" {
		var err error
		userAuth.Login = c.Value
		userAuth, err = database.SelectUserByColumn("Login", userAuth.Login)
		if err != nil {
			fmt.Println(err.Error())
		}
	} else {
		http.NotFound(w, r)
	}
}

func storeBuyHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println(itemId)

	if r.Method == "POST" {
		r.ParseForm()
		paypal := r.FormValue("paypal")
		card := r.FormValue("card")
		cardholder := r.FormValue("cardholder")
		cardnumber := r.FormValue("cardnumber")
		date := r.FormValue("date")
		CVC := r.FormValue("CVC")
		money := r.FormValue("money")
		floatMoney, _ := strconv.ParseFloat(money, 32)
		fmt.Println(paypal, card, cardholder, cardnumber, date, CVC, money)
		err := database.UpdateUserWalletByLogPass(userAuth.Login, userAuth.Password, floatMoney)
		if err != nil {
			http.NotFound(w, r)
		}
		http.Redirect(w, r, "/store", http.StatusSeeOther)
	}
	if r.Method == "GET" {
		r.ParseForm()
		address := r.FormValue("address")
		if address != "" {
			err := database.UpdateUserWalletByLogPass(userAuth.Login, userAuth.Password, float64(-product.NewPrice))
			if err != nil {
				fmt.Println(err)
			}
			com := database.SelectCommunitiesByColumn("Name", product.Community)
			seller, err := database.SelectUserByColumn("Login", com.Author)
			if err != nil {
				fmt.Println("Error - storeBuyHandler() SelectUserByColumn", err)
			}
			err = database.UpdateUserWalletByLogPass(seller.Login, seller.Password, float64(product.NewPrice)*0.95)
			if err != nil {
				fmt.Println(err)
			}
			income := float64(product.NewPrice) * 0.05
			err = database.UpdateUserWalletByLogPass(`a.pavlikov2002@gmail.com`, `86a65acd94b33daa87c1c6a2d1408593`, income)
			if err != nil {
				fmt.Println(err)
			}
			var sale models.Sales
			sale.Product = product.Id
			sale.User = userAuth.Login
			sale.Address = address
			sale.Date = time.Now().Format("2006-01-02 15:04")
			err = database.InsertToSalesProduct(sale)
			if err != nil {
				fmt.Println(err.Error())
			}

			err = app.SendSales(userAuth.Login, userAuth.Name, product)
			if err != nil {
				fmt.Println("Error - SendSales()", err.Error())
			}

			http.Redirect(w, r, "/store", http.StatusSeeOther)
		} else {
			http.NotFound(w, r)
		}
	}

}

func storeFavoritesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		r.ParseForm()
		res := r.FormValue("Ok")
		fmt.Println("storeFavoritesHandler - GET", res)
		if res == "true" {
			fav, ok := database.SelectFavouritesByUserCheck(userAuth.Login) //----------------
			if ok {
				var arr []int
				arr = append(arr, int(product.Id))
				err := database.InsertFavoritesToUser(userAuth.Login, arr)
				if err != nil {
					fmt.Println(err, "InsertFavoritesToUser")
				}
			}
			res := app.CheckArray(fav, product.Id)
			if res {
				err := database.UpdateFavouritesToUser(userAuth.Login, product.Id)
				if err != nil {
					fmt.Println(err)
				}
			}
		}
		http.Redirect(w, r, "/store", http.StatusSeeOther)
	}

	if r.Method == "POST" {
		r.ParseForm()
		res := r.FormValue("Ok")
		in := r.FormValue("input")
		fmt.Println("storeFavoritesHandler - POST", res, in)
		id, _ := strconv.Atoi(in)
		if id > 0 {
			err := database.DeleteFavouritesUser(userAuth.Login, uint(id))
			if err != nil {
				fmt.Println("Error - storeFavoritesHandler() POST DeleteFavouritesUser()", err)
			}
		}
		http.Redirect(w, r, "/favourites", http.StatusSeeOther)
	}

}

func favouritesPageHandler(w http.ResponseWriter, r *http.Request) {
	c := app.ReadCookies(r)
	if c.Value == userAuth.Login {
		tmpl, err := template.ParseFiles("html/favourites.html", "html/header.html", "html/footer.html")
		if err != nil {
			http.NotFound(w, r)
		}

		fav, ok := database.SelectFavouritesByUser(userAuth.Login)
		sales, done := database.SelectSalesByUser(userAuth.Login)
		purchase := database.SelectTotalPurchaseByUser(userAuth.Login)

		title := map[string]interface{}{"Title": models.Cfg.FavTitle, "User": userAuth}
		tmpl.ExecuteTemplate(w, "header", title)
		tmpl.ExecuteTemplate(w, "footer", title)
		data := map[string]interface{}{"User": userAuth, "Done": ok, "Fav": fav, "Ok": done, "Sales": sales, "Purchase": purchase}
		tmpl.ExecuteTemplate(w, "favourites", data)
	} else if c.Value != "" {
		var err error
		userAuth.Login = c.Value
		userAuth, err = database.SelectUserByColumn("Login", userAuth.Login)
		if err != nil {
			fmt.Println(err.Error())
		}
	} else {
		http.NotFound(w, r)
	}
}

func helpHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("html/help.html", "html/header.html", "html/footer.html")
	if err != nil {
		http.NotFound(w, r)
	}

	arrCategory := []string{"Баг при работе", "Предложение", "Некорректная работа приложения", "У вас что-то украли/пропало", "Другое"}

	title := map[string]interface{}{"Title": models.Cfg.HelpTitle, "User": userAuth}
	data := map[string]interface{}{"User": userAuth, "Category": arrCategory}

	tmpl.ExecuteTemplate(w, "header", title)
	tmpl.ExecuteTemplate(w, "footer", title)
	tmpl.ExecuteTemplate(w, "help", data)
}

func helpComplaintHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" { //жалобы и предложения
		r.ParseForm()
		title := r.FormValue("title")
		description := r.FormValue("description")
		selectCat := r.FormValue("selectCat")
		sessins := fmt.Sprintf("Категория - %s; Заголовок - %s; Описание - %s", selectCat, title, description)
		err := app.RecordingSessions(sessins, "complaints_and_suggestions.txt")
		if err != nil {
			fmt.Println("Error - RecordingSessions() ", err.Error())
		}
		//записывать все жалобы и предложения в отдельный файл

		http.Redirect(w, r, "/help", http.StatusSeeOther)
	}
	if r.Method == "GET" { // жалобы на пользователей
		r.ParseForm()
		guest := r.FormValue("guest")
		complaint := r.FormValue("selectComplaint")
		status := complaintStatus[0]
		fmt.Println(guest, complaint, status)
		comp := models.Complaints{
			Criminal:  guest,
			Complaint: complaint,
			Author:    userAuth.Login,
			Status:    status,
		}

		err := database.InsertComplaint(comp)
		if err != nil {
			fmt.Println("Error - InsertComplaint() ", err.Error())
		}

		http.Redirect(w, r, "/guest", http.StatusSeeOther)
	}
}

func galleryHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("html/gallery.html", "html/header.html", "html/footer.html")
	if err != nil {
		http.NotFound(w, r)
	}
	var gallery []string
	var ok bool
	if r.Method == "POST" {
		r.ParseForm()
		user := r.FormValue("user")
		fmt.Println(user)
		if user != "" {
			gallery = database.SelectUserGallery(user)
			fmt.Println("-------------GALLERY-------------", gallery)
			if gallery == nil {
				ok = false
			} else {
				ok = true
			}
		} else {
			http.Redirect(w, r, "/page", http.StatusSeeOther)
		}
	}

	title := map[string]interface{}{"Title": models.Cfg.GalleryTitile, "User": userAuth}
	data := map[string]interface{}{"User": userAuth, "Gallery": gallery, "OK": ok}

	tmpl.ExecuteTemplate(w, "header", title)
	tmpl.ExecuteTemplate(w, "footer", title)
	tmpl.ExecuteTemplate(w, "gallery", data)
}

func galleryAddHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		r.ParseForm()
		file, fileHeader, err := r.FormFile("photo")
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		defer file.Close()
		imgPath = fmt.Sprintf("./data/image/users/%d%s", time.Now().UnixNano(), filepath.Ext(fileHeader.Filename))
		dst, err := os.Create(imgPath)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		defer dst.Close()

		_, err = io.Copy(dst, file)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		err = database.UpdateUserGallery(userAuth.Login, imgPath)
		if err != nil {
			fmt.Println("Error - galleryAddHandler() UpdateUserGallery()", err.Error())
		}

		http.Redirect(w, r, "/gallery", http.StatusSeeOther)
	}
}

func musicHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("html/music.html", "html/header.html", "html/footer.html")
	if err != nil {
		http.NotFound(w, r)
	}
	var musics []models.MusicSub
	var ok bool
	if r.Method == "POST" {
		r.ParseForm()
		user := r.FormValue("user")
		fmt.Println(user, "!!!!!!!!!!")
		musics = database.SelectMusicByUser(user)
		fmt.Println("-------------MUSIC-------------", musics)
		//and select recomendation music for user
		fmt.Println(user)
	} else {
		http.Error(w, "Error", http.StatusNotFound)
	}

	title := map[string]interface{}{"Title": models.Cfg.MusicTitle, "User": userAuth}
	data := map[string]interface{}{"User": userAuth, "OK": ok, "Musics": musics}

	tmpl.ExecuteTemplate(w, "header", title)
	tmpl.ExecuteTemplate(w, "footer", title)
	tmpl.ExecuteTemplate(w, "music", data)
}

func musicSubHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		r.ParseForm()
		music := r.FormValue("music")
		user := r.FormValue("user")
		fmt.Println(music, user)
		id, _ := strconv.Atoi(music)
		err := database.UpdateMusicToUser(user, id)
		if err != nil {
			fmt.Println("Error - musicSubHandler() UpdateMusicToUser()")
		}
		err = database.UpdateMusicSub(id)
		if err != nil {
			fmt.Println("Error - musicSubHandler() UpdateMusicSub()")
		}
		http.Redirect(w, r, "/music", http.StatusSeeOther)
	}
}

func musicUnsubHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		r.ParseForm()
		music := r.FormValue("music")
		user := r.FormValue("user")
		fmt.Println(music, user)
		id, _ := strconv.Atoi(music)
		err := database.DeleteMusicUser(user, id)
		if err != nil {
			fmt.Println("Error - musicSubHandler() DeleteMusicUser()")
		}
		http.Redirect(w, r, "/music", http.StatusSeeOther)
	}
}

func exitHandler(w http.ResponseWriter, r *http.Request) {
	err := database.DeleteOnlineUser(userAuth.Login)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
	}
	c := &http.Cookie{
		Name:    "id",
		Value:   "",
		Path:    "/",
		Expires: time.Unix(0, 0),

		HttpOnly: true,
	}

	http.SetCookie(w, c)

	userAuth = models.Users{}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// Admins Handler

func adminHandler(w http.ResponseWriter, r *http.Request) {
	c := app.ReadCookies(r)
	if c.Value == userAuth.Login {
		if userAuth.Access != "User" {
			tmpl, err := template.ParseFiles("html/admin.html", "html/header.html", "html/footer.html")
			if err != nil {
				http.NotFound(w, r)
			}

			uBan, err := database.SelectUserBannedByAdmin(userAuth.Login)
			if err != nil {
				fmt.Println("Error - SelectUserBannedByAdmin()", err.Error())
			}
			uDel, err := database.SelectUserDeletedByAdmin(userAuth.Login)
			if err != nil {
				fmt.Println("Error - SelectUserBannedByAdmin()", err.Error())
			}

			all, err := database.SelectUserBannedAllByAdmin(userAuth.Login)
			if err != nil {
				fmt.Println("Error - SelectUserBannedAllByAdmin()", err.Error())
			}

			title := map[string]interface{}{"Title": "Админ панель", "User": userAuth}
			data := map[string]interface{}{"User": userAuth, "AllUs": all, "DelUser": uDel, "BanUser": uBan, "Done": true}

			tmpl.ExecuteTemplate(w, "header", title)
			tmpl.ExecuteTemplate(w, "footer", title)
			tmpl.ExecuteTemplate(w, "admin", data)
		} else {
			http.Redirect(w, r, "/page", http.StatusSeeOther)
		}
	} else {
		http.NotFound(w, r)
	}
}

func adminBanHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		r.ParseForm()

		us := r.FormValue("user")
		title := r.FormValue("title")
		time := r.FormValue("time")

		user, err := database.SelectUserByColumn("Login", us)
		if err != nil {
			fmt.Println("Error - adminBanHandler() SelectUserByColumn()", err.Error())
		}
		user.Access = "Banned"
		pass := user.Password
		user.Password = user.Photo
		user.Photo = "/data/image/blog/banned.jpg"
		err = database.UpdateUserByLogPass(user.Login, pass, user)
		if err != nil {
			fmt.Println("Error - adminBanHandler() UpdateUserByLogPass()", err.Error())
		}

		banTime, err := strconv.Atoi(time)
		if err != nil {
			fmt.Println("Error ", err.Error())
		}
		var ban = models.UserBanned{
			Id:     0,
			User:   us,
			Reason: title,
			Time:   banTime,
			Admin:  userAuth.Login,
		}

		err = database.InsertUserToBanList(ban)
		if err != nil {
			fmt.Println("Error - adminBanHandler() InsertUserToBanList()", err.Error())
		}

		http.Redirect(w, r, "/admin", http.StatusSeeOther)
	}
}

func adminDelHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		r.ParseForm()

		us := r.FormValue("Adm-guestId")
		// fmt.Println(us)
		// err := database.DeleteUserByLogin(us)
		// if err != nil {
		// 	fmt.Println("Error - adminDelHandler() DeleteUserByLogin()", err.Error())
		// }

		user, err := database.SelectUserByColumn("Login", us)
		if err != nil {
			fmt.Println("Error - adminBanHandler() SelectUserByColumn()", err.Error())
		}
		user.Access = "Banned"
		pass := user.Password
		user.Password = user.Photo
		user.Photo = "/data/image/blog/banned.jpg"
		err = database.UpdateUserByLogPass(user.Login, pass, user)
		if err != nil {
			fmt.Println("Error - adminBanHandler() UpdateUserByLogPass()", err.Error())
		}

		var ban = models.UserBanned{
			Id:     0,
			User:   us,
			Reason: "Удаленный аккаунт",
			Time:   2147483647, //integer
			Admin:  userAuth.Login,
		}

		err = database.InsertUserToBanList(ban)
		if err != nil {
			fmt.Println("Error - adminDelHandler() InsertUserToBanList()", err.Error())
		}

		http.Redirect(w, r, "/admin", http.StatusSeeOther)
	}
}

func adminDelBanListHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("html/list.html", "html/header.html", "html/footer.html")
	if err != nil {
		http.NotFound(w, r)
	}
	done := "Admin"
	var Data []models.UserBanned
	if r.Method == "GET" {

		Data, err = database.SelectUserBannedByAdmin(userAuth.Login)
		if err != nil {
			fmt.Println("Error - SelectUserBannedByAdmin()", err.Error())
		}
	}

	if r.Method == "POST" {
		Data, err = database.SelectUserDeletedByAdmin(userAuth.Login)
		if err != nil {
			fmt.Println("Error - SelectUserBannedByAdmin()", err.Error())
		}
	}

	fmt.Println(Data)
	ttl := "Админ меню"
	title := map[string]interface{}{"Title": ttl, "User": userAuth}
	data := map[string]interface{}{"Admins": Data, "Done": done, "Title": ttl}

	tmpl.ExecuteTemplate(w, "header", title)
	tmpl.ExecuteTemplate(w, "footer", title)
	tmpl.ExecuteTemplate(w, "list", data)

}

func adminListHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("html/list.html", "html/header.html", "html/footer.html")
	if err != nil {
		http.NotFound(w, r)
	}
	done := "ListAdmin"

	admin, err := database.SelectAdmins()
	if err != nil {
		fmt.Println(err.Error())
	}

	ttl := "Список администраторов"
	title := map[string]interface{}{"Title": ttl, "User": userAuth}
	data := map[string]interface{}{"ListAdm": admin, "Done": done, "Title": ttl}

	tmpl.ExecuteTemplate(w, "header", title)
	tmpl.ExecuteTemplate(w, "footer", title)
	tmpl.ExecuteTemplate(w, "list", data)

}

func adminComplaintListHandler(w http.ResponseWriter, r *http.Request) { ///admin/complaint/list
	tmpl, err := template.ParseFiles("html/list.html", "html/header.html", "html/footer.html")
	if err != nil {
		http.NotFound(w, r)
	}
	comp, err := database.SelectComplaints()
	if err != nil {
		fmt.Println("Error - SelectComplaints()", err.Error())
	}

	done := "Complaint"

	fmt.Println(comp)

	ttl := "Список жалоб"
	title := map[string]interface{}{"Title": ttl, "User": userAuth}
	data := map[string]interface{}{"Complaint": comp, "Done": done, "Title": ttl}

	tmpl.ExecuteTemplate(w, "header", title)
	tmpl.ExecuteTemplate(w, "footer", title)
	tmpl.ExecuteTemplate(w, "list", data)
}
func adminComplaintHandler(w http.ResponseWriter, r *http.Request) { ///admin/complaint
	tmpl, err := template.ParseFiles("html/admin.html", "html/header.html", "html/footer.html")
	if err != nil {
		http.NotFound(w, r)
	}

	var comp models.Complaints

	if r.Method == "POST" {
		r.ParseForm()
		id = r.FormValue("id")
		fmt.Println("ID-------", id)
		Id, _ := strconv.Atoi(id)
		comp, err = database.SelectComplaint(Id)
		if err != nil {
			fmt.Println("Error - SelectComplaint() ", err.Error())
		}
		fmt.Println(comp)
	}

	ttl := "Жалоба - №" + fmt.Sprint(Id)
	title := map[string]interface{}{"Title": ttl, "User": userAuth}
	//"User": userAuth, "AllUs": all, "DelUser": uDel, "BanUser": uBan, "Done": true}
	data := map[string]interface{}{"User": userAuth, "Complaint": comp, "Title": ttl, "Done": false, "CompStat": complaintStatus}

	tmpl.ExecuteTemplate(w, "header", title)
	tmpl.ExecuteTemplate(w, "footer", title)
	tmpl.ExecuteTemplate(w, "admin", data)
}

func adminComplaintEditHandler(w http.ResponseWriter, r *http.Request) {
	var complaint models.Complaints
	if r.Method == "POST" {
		r.ParseForm()
		complaint.Id, _ = strconv.Atoi(r.FormValue("id"))
		complaint.Criminal = r.FormValue("field1")
		complaint.Complaint = r.FormValue("field2")
		complaint.Author = r.FormValue("field3")
		complaint.Status = r.FormValue("selectuser")
		complaint.Comment = r.FormValue("field6")
		complaint.Admin = r.FormValue("field5")
		fmt.Println(complaint)

		err := database.UpdateComplaintById(complaint)
		if err != nil {
			fmt.Println("Error - UpdateComplaintById() ", err.Error())
		}

		http.Redirect(w, r, "/admin/complaint/list", http.StatusSeeOther)
	}

}
