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

func logFormHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("html/login.html")
	if err != nil {
		http.NotFound(w, r)
	}

	tmpl.ExecuteTemplate(w, "login", nil)
}

func authHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	var authLog string
	var authPass string
	authLog = r.FormValue("email2")
	authPass = r.FormValue("pswd2")
	if r.Method == "GET" { //авторизация пользователя
		fmt.Println("This is auth", authLog, authPass)
		cookie := &http.Cookie{
			Name:   "id",
			Value:  "abcd",
			MaxAge: 300,
		}
		// app.Cookies()
		http.SetCookie(w, cookie)

		//if true функция выдачи action для формы в слючие совпадения пароля и логина
		var err error
		userAuth, _ = database.SelectUserByLogPass(authLog, authPass)
		if userAuth.Login == "" && userAuth.Password == "" {
			http.NotFound(w, r)
			http.Redirect(w, r, "/", http.StatusBadRequest)
		} else {
			err = database.InsertUserToOnline(userAuth.Login)
			if err != nil {
				fmt.Println("Error - authHandler() InsertUserToOnline()", err)
				// http.NotFound(w, r)
			}
			app.RecordingSessions(fmt.Sprintf("Пользователь, %s (логин - %s, пароль - %s) зашел в аккаунт в %s.\n", userAuth.Name, userAuth.Login, userAuth.Password, time.Now().Format("2006-01-02 15:04")))
			http.Redirect(w, r, "/blog", http.StatusSeeOther)
		}
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
		login := userAuth.Login
		newPass := r.FormValue("newPass")
		oldPass := r.FormValue("oldPass")
		newName := r.FormValue("newName")
		date := r.FormValue("newHB")
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
}

func blogHandler(w http.ResponseWriter, r *http.Request) {
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
	var DoneGopher bool
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
	title := map[string]string{"Title": userAuth.Name}
	tmpl.ExecuteTemplate(w, "header", title)

	sendUser := map[string]interface{}{"User": userAuth, "Repo": data, "Statistics": stat, "Done": DonePost, "DoneGopher": DoneGopher, "Gopher": gopher}
	tmpl.ExecuteTemplate(w, "page", sendUser)
}

func pagePostHandler(w http.ResponseWriter, r *http.Request) {
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
		gof := r.FormValue("gof2")
		if gof != "" {
			err := database.DeleteRepoPostByUser(gof, userAuth.Login)
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
		check, err = database.SelectMessengeListbyUsers(userAuth.Login, guestId)
		if err != nil {
			fmt.Println("Error - friendsHandler() SelectMessengeListbyUsers()")
		}
	}

	rec := database.SelectRecomendationFriends(userAuth.Login)
	subs := database.SelectUserSub(userAuth.Login)
	friends := database.SelectAllFriendsUser(userAuth.Login)
	online := database.SelectOnlineFriends(userAuth.Login)

	data := map[string]interface{}{"User": userAuth, "Friends": friends, "Subs": subs, "Rec": rec, "Online": online, "Done": true}
	title := map[string]string{"Title": models.Cfg.FriendsTitile}
	tmpl.ExecuteTemplate(w, "header", title)
	tmpl.ExecuteTemplate(w, "friends", data)
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
				app.CreateFile(val, sub, userAuth.Login)
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

	data := map[string]interface{}{"User": userAuth, "Friends": friends, "Guest": gst, "Subs": subs, "Rec": rec, "Done": false, "Online": online}
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
	//	comWithOutSub := database.SelectCommunitiesWithOutSub(userAuth.Login)
	catComm := database.SelectCommunitiesCategory()
	communities := database.SelectAllCommunitiesUser("User", userAuth.Login)
	recCommunities := database.SelectRecCommunities(userAuth.Login)
	data := map[string]interface{}{"User": userAuth, "Communities": communities, "Done": done, "RecCom": recCommunities, "CommCat": catComm}
	title := map[string]string{"Title": models.Cfg.CommunitiesTitile}
	tmpl.ExecuteTemplate(w, "header", title)
	tmpl.ExecuteTemplate(w, "communities", data)
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
	store, err := database.SelectStoreItemsByCommunity(communitiesName)
	if err != nil {
		fmt.Println(err)
	}
	comm := database.SelectCommunitiesByColumn("Name", post.Communities)
	fmt.Println("SelectCommunitiesByColumn", comm.Name)

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

	_, ok := database.CheckCommunity(userAuth.Login, communitiesName)

	title := map[string]string{"Title": Foo.Name}
	blog := map[string]interface{}{"Post": posts, "User": userAuth, "Subs": subs, "Author": author, "Names": names, "Communities": Foo, "PostCat": category, "CommCat": catComm, "SetCom": comm, "OK": ok, "Store": store}
	tmpl.ExecuteTemplate(w, "header", title)
	tmpl.ExecuteTemplate(w, "community", blog)
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
	tmpl, err := template.ParseFiles("html/store.html", "html/header.html", "html/footer.html")
	if err != nil {
		http.NotFound(w, r)
	}
	var Ok bool
	Access := false
	Ok = true
	store, err := database.SelectStoreItemsByCommunity(communitiesName)
	if err != nil {
		fmt.Println(err)
	}
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
	title := map[string]string{"Title": "Товары - " + communitiesName}
	tmpl.ExecuteTemplate(w, "header", title)
	data := map[string]interface{}{"User": userAuth, "Wallet": user.Wallet, "Store": store, "OK": Ok, "Access": Access, "Sex": sex, "Category": cat}
	tmpl.ExecuteTemplate(w, "store", data)
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

	_, ok := database.CheckFriends(userAuth.Login, guestId)

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

	title := map[string]string{"Title": userGuest.Name}
	tmpl.ExecuteTemplate(w, "header", title)

	sendUser := map[string]interface{}{"User": userAuth, "Repo": data, "Guest": userGuest, "Statistics": stat, "Done": Done, "OK": ok, "Gopher": gopher, "DoneGopher": DoneGopher, "Friend": Friend}

	tmpl.ExecuteTemplate(w, "guest", sendUser)
}

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
			Link = fmt.Sprintf("C:/Users/admin/go/src/go-blog/data/files/message/%s", UsersLink.MessageHistory)
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
			Messenger = app.JSON(msg, Link, userAuth.Login)
			fmt.Println(Messenger)
		}
	}

	title := map[string]string{"Title": models.Cfg.MessageTitle}
	tmpl.ExecuteTemplate(w, "header", title)
	fmt.Println(OK)
	data := map[string]interface{}{"User": userAuth, "Done": OK, "OK": OK, "Companions": companion, "ChatUser": activeChatUser, "Chat": Messenger.Messenge}
	tmpl.ExecuteTemplate(w, "message", data)
}

func storeHandler(w http.ResponseWriter, r *http.Request) {
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

	title := map[string]string{"Title": models.Cfg.StoreTitle}
	tmpl.ExecuteTemplate(w, "header", title)
	data := map[string]interface{}{"User": userAuth, "Wallet": user.Wallet, "Store": store, "OK": Ok, "Access": Access}
	tmpl.ExecuteTemplate(w, "store", data)
}

func storeCardHandler(w http.ResponseWriter, r *http.Request) {
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

	title := map[string]string{"Title": product.Name}
	tmpl.ExecuteTemplate(w, "header", title)

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
		// fmt.Printf("Вы купили - %s\n%f", itemId, float64(-product.NewPrice))
		err := database.UpdateUserWalletByLogPass(userAuth.Login, userAuth.Password, float64(-product.NewPrice))
		if err != nil {
			fmt.Println(err)
		}
		com := database.SelectCommunitiesByColumn("Name", product.Community)
		seller, err := database.SelectUserByColumn("Login", com.Author)
		if err != nil {
			fmt.Println("Error - storeBuyHandler() SelectUserByColumn", err)
		}
		err = database.UpdateUserWalletByLogPass(seller.Login, seller.Password, float64(product.NewPrice))
		if err != nil {
			fmt.Println(err)
		}
		var sale models.Sales
		sale.Product = product.Id
		sale.User = userAuth.Login
		sale.Address = "address"
		sale.Date = time.Now().Format("2006-01-02 15:04")
		fmt.Println(sale)
		err = database.InsertToSalesProduct(sale)
		if err != nil {
			fmt.Println(err)
		}
		http.Redirect(w, r, "/store", http.StatusSeeOther)
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
	tmpl, err := template.ParseFiles("html/favourites.html", "html/header.html", "html/footer.html")
	if err != nil {
		http.NotFound(w, r)
	}

	fav, ok := database.SelectFavouritesByUser(userAuth.Login)
	sales, done := database.SelectSalesByUser(userAuth.Login)

	title := map[string]string{"Title": models.Cfg.FavTitle}
	tmpl.ExecuteTemplate(w, "header", title)
	data := map[string]interface{}{"User": userAuth, "Done": ok, "Fav": fav, "Ok": done, "Sales": sales}
	tmpl.ExecuteTemplate(w, "favourites", data)
}

func exitHandler(w http.ResponseWriter, r *http.Request) {
	err := database.DeleteOnlineUser(userAuth.Login)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
