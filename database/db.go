package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/alexPavlikov/go-blog/models"
)

var DB *sql.DB
var err error

func Connect() (*sql.DB, error) {
	DB, err = sql.Open("postgres", fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", models.Cfg.PgHost, models.Cfg.PgPort, models.Cfg.PgUser, models.Cfg.PgPass, models.Cfg.PgName))
	if err != nil {
		fmt.Println("Error - database/Connect()", err)
		return DB, err
	}
	return DB, nil
}

// --------------------Query from Posts table--------------------

/*
Функция выборки всех постов из таблицы Posts
*/
func SelectPosts() []models.Posts {

	rows, err := DB.Query(`SELECT "Posts".*, "Communities"."Photo"
	FROM "Posts" 
	JOIN "Communities" ON "Posts"."Communities" = "Communities"."Name"
	ORDER BY "View" DESC;`)
	if err != nil {
		fmt.Println("Error - selectPosts()", err.Error())
	}

	post := models.Posts{}
	posts := []models.Posts{}

	for rows.Next() {

		err = rows.Scan(&post.Id, &post.Title, &post.Content, &post.Like, &post.View, &post.Date, &post.Communities, &post.Photo, &post.Category, &post.CommunitiesPhot)
		if err != nil {
			fmt.Println("Error - selectPosts() / rows.Next()", err.Error())
		}
		posts = append(posts, post)
	}
	return posts
}

/*
Функция выборки определенного поста из таблицы Posts по Id
*/
func SelectPostById(id int) models.Posts {
	var post models.Posts
	query := fmt.Sprintf(`SELECT "Posts".*, "Communities"."Photo"
	FROM "Posts" 
	JOIN "Communities" ON "Posts"."Communities" = "Communities"."Name"
	WHERE "Posts"."Id" = '%d' ORDER BY "View" DESC;`, id)
	rows, err := DB.Query(query)
	if err != nil {
		fmt.Println("Error - SelectPostById()", err.Error())
	}
	for rows.Next() {
		err = rows.Scan(&post.Id, &post.Title, &post.Content, &post.Like, &post.View, &post.Date, &post.Communities, &post.Photo, &post.Category, &post.CommunitiesPhot)
		if err != nil {
			fmt.Println("Error - SelectPostById() / rows.Next()", err.Error())
		}
	}
	return post
}

/*
Функция выборки определенного поста из таблицы Posts по Communitites
*/
func SelectPostByCommunities(communities string) []models.Posts {
	var posts []models.Posts
	var post models.Posts
	query := fmt.Sprintf(`SELECT "Posts".*, "Communities"."Photo"
	FROM "Posts" 
	JOIN "Communities" ON "Posts"."Communities" = "Communities"."Name" WHERE "Communities" = '%s'`, communities)
	rows, err := DB.Query(query)
	if err != nil {
		fmt.Println("Error - SelectPostByCommunities()", err.Error())
	}
	for rows.Next() {
		err = rows.Scan(&post.Id, &post.Title, &post.Content, &post.Like, &post.View, &post.Date, &post.Communities, &post.Photo, &post.Category, &post.CommunitiesPhot)
		if err != nil {
			fmt.Println("Error - SelectPostByCommunities() / rows.Next()", err.Error())
		}
		posts = append(posts, post)
	}
	return posts
}

/*
Функция удаления определенного поста из таблицы Posts по Id
*/
func DeletePostById(id string) error {
	res, err := DB.Exec(`DELETE FROM "Posts" WHERE "Id" = ($1)`, id)
	if err == nil {
		count, err := res.RowsAffected()
		if err == nil {
			fmt.Println(count)
		}
		return nil
	}
	return err
}

/*
Функция удаления определенного поста из таблицы Posts по определенному Communities
*/
func DeletePostByCommunities(communities string) error {
	res, err := DB.Exec(`DELETE FROM "Posts" WHERE "Communities" = ($1)`, communities)
	if err == nil {
		count, err := res.RowsAffected()
		if err == nil {
			fmt.Println(count)
		}
		return nil
	}
	return err
}

/*
Функция добавления поста в таблицу Posts
*/
func InsertPost(post models.Posts) (models.Posts, error) {
	query := `INSERT INTO "Posts"("Title", "Content", "Like", "View", "Date", "Communities") VALUES ($1, $2, $3, $4, $5, $6, $7)`
	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()
	stmt, err := DB.PrepareContext(ctx, query)
	if err != nil {
		log.Printf("Error %s when preparing SQL statement", err)
		return post, err
	}
	defer stmt.Close()
	res, err := stmt.ExecContext(ctx, post.Title, post.Content, post.Like, post.View, post.Date, post.Communities, post.Photo)
	if err != nil {
		log.Printf("Error %s when inserting row into Posts table", err)
		return post, err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		log.Printf("Error %s when finding rows affected", err)
		return post, err
	}
	log.Printf("%d post created ", rows)
	return post, nil
}

/*
Функция обновления поста в таблице Posts по определенному Id
*/
func UpdatePostById(id string, post models.Posts) error {
	query := fmt.Sprintf(`UPDATE "Posts" SET "Title" = %s, "Content" = %s, "Like" = '%d', "View" = '%d', "Date" = %s, "Communities" = %s, "Photo" = %s WHERE "Id" =  %s`, post.Title, post.Content, post.Like, post.View, post.Date, post.Communities, post.Photo, post.Id)
	_, err := DB.Query(query)
	if err != nil {
		fmt.Println(err)
		return err
	}
	fmt.Println(query)
	return nil
}

/*
Функция обновления лайков поста в таблице Posts по определенному Id
*/
func UpdateLikeInPost(id int) error {
	query := fmt.Sprintf(`UPDATE "Posts" SET "Like" = "Like" + 1 WHERE "Id" = '%d'`, id)
	_, err := DB.Query(query)
	if err != nil {
		fmt.Println(err)
		return err
	}
	fmt.Println(query)
	return nil
}

/*
Функция обновления просмотров поста в таблице Posts
*/
func UpdateViewInPost() error {
	query := `UPDATE "Posts" SET "View" = "View" + 1`
	_, err := DB.Query(query)
	if err != nil {
		fmt.Println(err)
		return err
	}
	fmt.Println(query)
	return nil
}

/*
Функция обновления фотографии поста в таблице Posts по определенному Id
*/
func UpdatePhotoInPost(id string, photo string) error {
	query := fmt.Sprintf(`UPDATE "Posts" SET "Photo" = %s WHERE "Id" = %s`, photo, id)
	_, err := DB.Query(query)
	if err != nil {
		fmt.Println(err)
		return err
	}
	fmt.Println(query)
	return nil
}

// --------------------Query from Users table--------------------

/*
Функция выборки всех пользователей из таблицы Users
*/
func SelectUsers() []models.Users {
	rows, err := DB.Query(`SELECT * FROM "Users"`)
	if err != nil {
		fmt.Println("Error - SelecetUser()", err)
	}
	var user models.Users
	var users []models.Users
	for rows.Next() {
		err = rows.Scan(&user.Login, &user.Password, &user.Name, &user.Access, &user.Communities, &user.Photo, &user.Birthdate)
		if err != nil {
			fmt.Println("Error - SelectUser() rows.Next()", err.Error())
		}
		users = append(users, user)
	}
	return users
}

/*
Функция выборки пользователя из таблицы Users по определенному Login и Password
*/
func SelectUserByLogPass(log string, pass string) (user models.Users, err error) {
	query := fmt.Sprintf(`SELECT * FROM "Users" WHERE "Login" = '%s' AND "Password" = '%s'`, log, pass)
	fmt.Println(query)
	rows, err := DB.Query(query)
	if err != nil {
		fmt.Println("Error - SelectUserByLogPass()", err)
		return user, err
	}
	for rows.Next() {
		err = rows.Scan(&user.Login, &user.Password, &user.Name, &user.Access, &user.Communities, &user.Photo, &user.Birthdate)
		if err != nil {
			fmt.Println("Error - SelectUserByLogPass() rows.Next()", err)
		}
	}
	return user, nil
}

/*
Функция-конструктор, вводится два значения column и value и подставляются в запрос,
что позволяет не писать повторяющие запросы SELECT
*/
func SelectUsersByColumn(column string, value string) ([]models.Users, error) {
	var user models.Users
	var users []models.Users
	rows, err := DB.Query(fmt.Sprintf(`SELECT * FROM "Users" WHERE "%s" = '%s'`, column, value))
	if err != nil {
		fmt.Println("Error - SelectUsersByColumn()", err)
		return users, err
	}

	for rows.Next() {
		err = rows.Scan(&user.Login, &user.Password, &user.Name, &user.Access, &user.Communities, &user.Photo, &user.Birthdate)
		if err != nil {
			fmt.Println("Error - SelectUsersByColumn() rows.Next()", err.Error())
			return users, err
		}
		users = append(users, user)
	}
	return users, nil
}

func SelectUserByColumn(column string, value string) (models.Users, error) {
	var user models.Users
	rows, err := DB.Query(fmt.Sprintf(`SELECT * FROM "Users" WHERE "%s" = '%s'`, column, value))
	if err != nil {
		fmt.Println("Error - SelectUserByColumn()", err)
		return user, err
	}

	for rows.Next() {
		err = rows.Scan(&user.Login, &user.Password, &user.Name, &user.Access, &user.Communities, &user.Photo, &user.Birthdate)
		if err != nil {
			fmt.Println("Error - SelectUserByColumn() rows.Next()", err.Error())
			return user, err
		}
	}
	return user, nil
}

/*
Функция удаления пользователя из таблицы Users по определенному Login
*/
func DeleteUserByLogin(login string) error {
	res, err := DB.Exec(`DELETE FROM "Users" WHERE "Login" = ($1)`, login)
	if err == nil {
		count, err := res.RowsAffected()
		if err == nil {
			fmt.Println(count)
		}
		return nil
	}
	return err
}

/*
Функция добавления пользователя в таблицу Users по введенным значениям
*/
func InsertUser(user models.Users) (models.Users, error) {
	query := `INSERT INTO "Users"("Login", "Password", "Name", "Access") VALUES ($1, $2, $3, $4)`
	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()
	stmt, err := DB.PrepareContext(ctx, query)
	if err != nil {
		log.Printf("Error %s when preparing SQL statement", err)
		return user, err
	}
	defer stmt.Close()
	res, err := stmt.ExecContext(ctx, user.Login, user.Password, user.Name, user.Access)
	if err != nil {
		log.Printf("Error %s when inserting row into User table", err)
		return user, err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		log.Printf("Error %s when finding rows affected", err)
		return user, err
	}
	log.Printf("%d user created ", rows)
	return user, nil
}

/*
Функция обновления пользователя из таблицы Users по введенным Login и Password
*/
func UpdateUserByLogPass(login string, password string, user models.Users) error {
	query := fmt.Sprintf(`UPDATE "Users" SET "Login" = %s, "Password" = %s, "Name" = %s, "Access" = %s, "Communities" = %s, "Photo" = %s "Birthdate" = %s WHERE "Login" = %s AND "Password" =  %s`, user.Login, user.Password, user.Name, user.Access, user.Communities, user.Photo, user.Birthdate, login, password)
	_, err := DB.Query(query)
	if err != nil {
		fmt.Println(err)
		return err
	}
	fmt.Println(query)
	return nil
}

/*
Функция-конструктор, вводится два значения column и value и подставляются в запрос,
что позволяет не писать повторяющие запросы SELECT
*/
func UpdateUserByColumn(column string, value string, login string, pass string) (user models.Users, err error) {
	query := fmt.Sprintf(`UPDATE "Users" SET %s = %s WHERE "Login" = %s AND "Password" =  %s`, column, value, login, pass)
	_, err = DB.Query(query)
	if err != nil {
		fmt.Println(err)
		return user, err
	}
	fmt.Println(query)
	return user, err
}

//Добавить еще обновление фото, имени, доступа

// --------------------Query from Communities table--------------------

/*
Функция выборки всех сообществ из таблицы Communities
*/
func SelectCommunities() []models.Communities {
	var communities models.Communities
	var communitiesArr []models.Communities
	rows, err := DB.Query(`SELECT * FROM "Communities"`)
	if err != nil {
		fmt.Println("Error - SelectCommunities()", err.Error())
	}
	for rows.Next() {
		err = rows.Scan(&communities.Name, &communities.Author, &communities.Photo, &communities.Category)
		if err != nil {
			fmt.Println("Error - SelectCommunities() rows.Next()", err.Error())
		}
		communitiesArr = append(communitiesArr, communities)
	}
	return communitiesArr
}

func SelectCommunitiesWithOutSub(user string) []models.Communities {
	var communities models.Communities
	var communitiesArr []models.Communities
	rows, err := DB.Query(fmt.Sprintf(`SELECT "Communities".*
	FROM "Communities"
	JOIN "Subscribers" ON "Subscribers"."Communities" <> "Communities"."Name"
	WHERE "Subscribers"."User"='%s';`, user))
	if err != nil {
		fmt.Println("Error - SelectCommunities()", err.Error())
	}
	for rows.Next() {
		err = rows.Scan(&communities.Name, &communities.Author, &communities.Photo, &communities.Category)
		if err != nil {
			fmt.Println("Error - SelectCommunities() rows.Next()", err.Error())
		}
		communitiesArr = append(communitiesArr, communities)
	}
	return communitiesArr
}

func SelectCommunitiesAuthorByName(name string) (author string, names string) {

	rows, err := DB.Query(fmt.Sprintf(`SELECT "Communities"."Author", "Users"."Name"
	FROM "Communities" 
	JOIN "Users" ON "Users"."Login" = "Communities"."Author"
	WHERE "Communities"."Name" = '%s'`, name))
	if err != nil {
		fmt.Println("Error - SelectCommunitiesAuthorByName()", err.Error())
	}
	for rows.Next() {
		err = rows.Scan(&author, &names)
		if err != nil {
			fmt.Println("Error - SelectCommunitiesAuthorByName() rows.Next()", err.Error())
		}
	}
	return author, names
}

/*
Функция-конструктор, вводится два значения column и value и подставляются в запрос,
что позволяет не писать повторяющие запросы SELECT
*/
func SelectCommunitiesByColumn(column string, value string) models.Communities {
	var communities models.Communities
	rows, err := DB.Query(fmt.Sprintf(`SELECT * FROM "Communities" WHERE "%s" = '%s'`, column, value))
	if err != nil {
		fmt.Println("Error - SelectCommunitiesByColumn()", err.Error())
	}
	for rows.Next() {
		err = rows.Scan(&communities.Name, &communities.Author, &communities.Photo, &communities.Category)
		if err != nil {
			fmt.Println("Error - SelectCommunitiesByColumn() rows.Next()", err.Error())
		}
	}
	return communities
}

/*
Функция выборки все сообщества из таблицы Communities определенного пользователя
SELECT  "Subscribers"."Communities", "Subscribers"."User",

	"Communities"."Photo", "Communities"."Author", "Communities"."Category"

FROM "Subscribers"
JOIN "Communities" ON "Communities"."Name" = "Subscribers"."Communities"
WHERE "Subscribers"."User" = 'a.pavlikov2002@gmail.com';
*/
func SelectAllCommunitiesUser(column string, key string) []models.JoinCommunities {
	var communitie models.JoinCommunities
	var communities []models.JoinCommunities
	query := fmt.Sprintf(`
		SELECT "Subscribers"."Communities", "Subscribers"."User",
		"Communities"."Photo", "Communities"."Author", "Communities"."Category"
		FROM "Subscribers"
		JOIN "Communities" ON "Communities"."Name" = "Subscribers"."Communities"
		WHERE "Subscribers"."%s" = '%s';
	`, column, key)
	rows, err := DB.Query(query)
	if err != nil {
		fmt.Println("Error - SelectFriendsByColumn()", err.Error())
	}
	for rows.Next() {
		err = rows.Scan(&communitie.Communities, &communitie.User, &communitie.Photo, &communitie.Author, &communitie.Category)
		if err != nil {
			fmt.Println("Error - SelectFriendsByColumn() rows.Next()", err.Error())
		}
		communities = append(communities, communitie)
	}
	return communities
}

func SelectCountSubscribersByCommunities(name string) (count uint) {
	query := fmt.Sprintf(`SELECT count(*) FROM "Subscribers" WHERE "Communities" = '%s'`, name)
	rows, err := DB.Query(query)
	if err != nil {
		fmt.Println("Error - SelectFriendsByColumn()", err.Error())
	}
	for rows.Next() {
		err = rows.Scan(count)
		if err != nil {
			fmt.Println("Error - SelectFriendsByColumn() rows.Next()", err.Error())
		}
	}
	return count
}

/*
Функция удаления сообщества из таблицы Communities по введенному имени
*/
func DeleteCommunitiesByName(name string) error {
	res, err := DB.Exec(`DELETE FROM "Communities" WHERE "Name" = ($1)`, name)
	if err == nil {
		count, err := res.RowsAffected()
		if err == nil {
			fmt.Println(count)
		}
		return nil
	}
	return err
}

/*
Функция добавления сообщества в таблицу Communities
*/
func InsertCommunities(com models.Communities) (models.Communities, error) {
	query := `INSERT INTO "Communities"("Name", "Author") VALUES ($1, $2)`
	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()
	stmt, err := DB.PrepareContext(ctx, query)
	if err != nil {
		log.Printf("Error %s when preparing SQL statement", err)
		return com, err
	}
	defer stmt.Close()
	res, err := stmt.ExecContext(ctx, com.Name, com.Author, com.Photo)
	if err != nil {
		log.Printf("Error %s when inserting row into Communities table", err)
		return com, err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		log.Printf("Error %s when finding rows affected", err)
		return com, err
	}
	log.Printf("%d post created ", rows)
	return com, nil
}

// --------------------Query from Access table--------------------

/*
Функция выборки доступа из таблицы Access
*/
func SelectAccess() []models.Access {
	var access models.Access
	var accessArr []models.Access
	rows, err := DB.Query(`SELECT * FROM "Access"`)
	if err != nil {
		fmt.Println("Error - SelectAccess()", err.Error())
	}
	for rows.Next() {
		err = rows.Scan(&access.Name)
		if err != nil {
			fmt.Println("Error - SelectAccess() rows.Next()")
		}
		accessArr = append(accessArr, access)
	}
	return accessArr
}

/*
Функция удаления доступа из таблицы Access
*/
func DeleteAccess(name string) error {
	res, err := DB.Exec(`DELETE FROM "Access" WHERE "Name" = ($1)`, name)
	if err == nil {
		count, err := res.RowsAffected()
		if err == nil {
			fmt.Println(count)
		}
		return nil
	}
	return err
}

/*
Функция обновления значения доступа из таблицы Access по введенному значению
*/
func UpdateAccess(name string, newName string) error {
	query := fmt.Sprintf(`UPDATE "Access" SET "Name" = %s WHERE "Name" = %s`, newName, name)
	_, err := DB.Query(query)
	if err != nil {
		fmt.Println(err)
		return err
	}
	fmt.Println(query)
	return nil
}

/*
Функция добавления доступа в таблицу Access
*/
func InsertAccess(access models.Access) (models.Access, error) {
	query := `INSERT INTO "Access"("Name") VALUES ($1)`
	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()
	stmt, err := DB.PrepareContext(ctx, query)
	if err != nil {
		log.Printf("Error %s when preparing SQL statement", err)
		return access, err
	}
	defer stmt.Close()
	res, err := stmt.ExecContext(ctx, access.Name)
	if err != nil {
		log.Printf("Error %s when inserting row into Access table", err)
		return access, err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		log.Printf("Error %s when finding rows affected", err)
		return access, err
	}
	log.Printf("%d post created ", rows)
	return access, err
}

// --------------------Query from Comments table--------------------

/*
Функция выборки комментария из таблицы Comments
*/
func SelectComments() []models.Comments {
	var comment models.Comments
	var comments []models.Comments
	rows, err := DB.Query(`SELECT * FROM "Comments"`)
	if err != nil {
		fmt.Println("Error - SelectComments()")
	}
	for rows.Next() {
		err = rows.Scan(&comment.Id, &comment.Posts, &comment.Text, &comment.Like, &comment.Author)
		if err != nil {
			fmt.Println("Error - SelectComments() rows.Next()")
		}
		comments = append(comments, comment)
	}
	return comments
}

/*
Функция-конструктор, вводится два значения column и value и подставляются в запрос,
что позволяет не писать повторяющие запросы SELECT
*/
func SelectCommentsByColumn(column string, value string) []models.JoinComments {
	var comment models.JoinComments
	var comments []models.JoinComments
	rows, _ := DB.Query(fmt.Sprintf(
		`SELECT "Comments"."Posts", "Comments"."Author", "Users"."Name" , "Users"."Photo",
		"Comments"."Text", "Comments"."Like"
		FROM "Users"
		JOIN "Comments" ON "Comments"."Author" = "Users"."Login"
		WHERE "Comments"."%s" = '%s'`, column, value))
	// if err != nil {
	// 	fmt.Println("Error - SelectCommentsByColumn()")
	// }
	for rows.Next() {
		err = rows.Scan(&comment.Posts, &comment.Author, &comment.Name, &comment.Photo, &comment.Text, &comment.Like)
		if err != nil {
			fmt.Println("Error - SelectCommentsByColumn() rows.Next()")
		}
		comments = append(comments, comment)
	}
	return comments
}

/*
Функция удаления комментария из таблицы Comments по введенному Id
*/
func DeleteCommentsById(id uint) error {
	res, err := DB.Exec(`DELETE FROM "Comments" WHERE "Id" = ($1)`, id)
	if err == nil {
		count, err := res.RowsAffected()
		if err == nil {
			fmt.Println(count)
		}
		return nil
	}
	return err
}

/*
Функция добавления комментария в таблицу Comments
*/
func InsertComment(comment models.Comments) (models.Comments, error) {
	query := `INSERT INTO "Comments"("Posts", "Text", "Like", "Author") VALUES ($1, $2, $3, $4)`
	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()
	stmt, err := DB.PrepareContext(ctx, query)
	if err != nil {
		log.Printf("Error %s when preparing SQL statement", err)
		return comment, err
	}
	defer stmt.Close()
	res, err := stmt.ExecContext(ctx, comment.Posts, comment.Text, comment.Like, comment.Author)
	if err != nil {
		log.Printf("Error %s when inserting row into Comment table", err)
		return comment, err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		log.Printf("Error %s when finding rows affected", err)
		return comment, err
	}
	log.Printf("%d post created ", rows)
	return comment, err
}

// --------------------Query from Friends table--------------------

/*
Функция выборки друзей из таблицы Friends
*/
func SelectFriends() []models.Friends {
	var friend models.Friends
	var friends []models.Friends
	rows, err := DB.Query(`SELECT * FROM "Friends"`)
	if err != nil {
		fmt.Println("Error - SelectFriends()", err.Error())
	}
	for rows.Next() {
		err = rows.Scan(&friend.Id, &friend.Login, &friend.Status, &friend.Friend)
		if err != nil {
			fmt.Println("Error - SelectFriends() rows.Next()", err.Error())
		}
		friends = append(friends, friend)
	}
	return friends
}

/*
Функция-конструктор, вводится два значения column и value и подставляются в запрос,
что позволяет не писать повторяющие запросы SELECT
*/
func SelectFriendsByColumn(column string, value string) []models.Friends {
	var friend models.Friends
	var friends []models.Friends
	query := fmt.Sprintf(`SELECT * FROM "Friends" WHERE "%s" = '%s'`, column, value)
	fmt.Println(query)
	rows, err := DB.Query(query)
	if err != nil {
		fmt.Println("Error - SelectFriendsByColumn()", err.Error())
	}
	for rows.Next() {
		err = rows.Scan(&friend.Id, &friend.Login, &friend.Status, &friend.Friend)
		if err != nil {
			fmt.Println("Error - SelectFriendsByColumn() rows.Next()", err.Error())
		}
		friends = append(friends, friend)
	}
	return friends
}

/*
Функция выборки всех друзей из таблицы Friends определенного пользователя
*/
func SelectAllFriendsUser(key string) []models.JoinUser {
	var friend models.JoinUser
	var friends []models.JoinUser
	query := fmt.Sprintf(`
		SELECT "Friends"."Login", "Friends"."Friend", "Friends"."Status",
		"Users"."Name", "Users"."Photo", "Users"."Birthdate"
		FROM "Friends"
		JOIN "Users" ON "Users"."Login" = "Friends"."Friend"
		WHERE "Friends"."Login" = '%s'
	`, key)
	rows, err := DB.Query(query)
	if err != nil {
		fmt.Println("Error - SelectFriendsByColumn()", err.Error())
	}
	for rows.Next() {
		err = rows.Scan(&friend.Login, &friend.Friend, &friend.Status, &friend.Name, &friend.Photo, &friend.Birthdate)
		if err != nil {
			fmt.Println("Error - SelectFriendsByColumn() rows.Next()", err.Error())
		}
		friends = append(friends, friend)
	}
	return friends
}

/*
Функция удаления друзей из таблицы Friends по введенному Id
*/
func DeleteFriendsById(id string) error {
	res, err := DB.Exec(`DELETE FROM "Friends" WHERE "Id" = ($1)`, id)
	if err == nil {
		count, err := res.RowsAffected()
		if err == nil {
			fmt.Println(count)
		}
		return nil
	}
	return err
}

/*
Функция добавления друзей в таблицу Friends
*/
func InsertFriends(friend models.Friends) (models.Friends, error) {
	query := `INSERT INTO "Friends"("Login", "Status", "Friend") VALUES ($1, $2, $3)`
	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()
	stmt, err := DB.PrepareContext(ctx, query)
	if err != nil {
		log.Printf("Error %s when preparing SQL statement", err)
		return friend, err
	}
	defer stmt.Close()
	res, err := stmt.ExecContext(ctx, friend.Login, friend.Status, friend.Friend)
	if err != nil {
		log.Printf("Error %s when inserting row into Friend table", err)
		return friend, err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		log.Printf("Error %s when finding rows affected", err)
		return friend, err
	}
	log.Printf("%d post created ", rows)
	return friend, err
}

// --------------------Query from Status table--------------------

/*
Функция выборки статуса из таблицы Status
*/
func SelectStatus() []models.Status {
	var status models.Status
	var statusArr []models.Status
	rows, err := DB.Query(`SELECT * FROM "Status"`)
	if err != nil {
		fmt.Println("Error - SelectStatus()", err.Error())
	}
	for rows.Next() {
		err = rows.Scan(&status.Name)
		if err != nil {
			fmt.Println("Error - SelectStatus() rows.Next()")
		}
		statusArr = append(statusArr, status)
	}
	return statusArr
}

/*
Функция удаления статуса из таблицы Status
*/
func DeleteStatus(name string) error {
	res, err := DB.Exec(`DELETE FROM "Status" WHERE "Name" = ($1)`, name)
	if err == nil {
		count, err := res.RowsAffected()
		if err == nil {
			fmt.Println(count)
		}
		return nil
	}
	return err
}

/*
Функция обновления статуса в таблице Status
*/
func UpdateStatus(name string, newName string) error {
	query := fmt.Sprintf(`UPDATE "Status" SET "Name" = %s WHERE "Name" = %s`, newName, name)
	_, err := DB.Query(query)
	if err != nil {
		fmt.Println(err)
		return err
	}
	fmt.Println(query)
	return nil
}

/*
Функция добавления статуса в таблицу Status
*/
func InsertStatus(status models.Status) (models.Status, error) {
	query := `INSERT INTO "Status"("Name") VALUES ($1)`
	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()
	stmt, err := DB.PrepareContext(ctx, query)
	if err != nil {
		log.Printf("Error %s when preparing SQL statement", err)
		return status, err
	}
	defer stmt.Close()
	res, err := stmt.ExecContext(ctx, status.Name)
	if err != nil {
		log.Printf("Error %s when inserting row into Status table", err)
		return status, err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		log.Printf("Error %s when finding rows affected", err)
		return status, err
	}
	log.Printf("%d post created ", rows)
	return status, err
}

// --------------------Query from RepoPost table--------------------

func SelectRepoPostByUser(user string) []models.RepostPost {
	var repoPost models.RepostPost
	var repoPosts []models.RepostPost
	query := fmt.Sprintf(`SELECT "RepostPost"."User", "RepostPost"."Post", 
	"Posts"."Title", "Posts"."Content", "Posts"."Like", 
	"Posts"."View", "Posts"."Date", "Posts"."Communities", "Posts"."Photo", "Posts"."Category", "Communities"."Photo"
	FROM "RepostPost"
	JOIN "Posts" ON "Posts"."Id" = "RepostPost"."Post"
	JOIN "Communities" ON "Communities"."Name" = "Posts"."Communities"
	WHERE "User"='%s';`, user)
	rows, err := DB.Query(query)
	if err != nil {
		fmt.Println("Error - SelectRepoPostByUser()")
	}
	for rows.Next() {
		err = rows.Scan(&repoPost.User, &repoPost.Post, &repoPost.Title, &repoPost.Content, &repoPost.Like, &repoPost.View, &repoPost.Date, &repoPost.Communities, &repoPost.PostPhoto, &repoPost.Categoty, &repoPost.CommunitiesPhoto)
		if err != nil {
			fmt.Println("Error - SelectRepoPostByUser() rows.Scan()")
		}
		repoPosts = append(repoPosts, repoPost)
	}
	return repoPosts
}

func InsertRepoPost(reppost models.Repost) (models.Repost, error) {
	query := `INSERT INTO "RepostPost"("Id", "Post", "User") VALUES ($1, $2, $3)`
	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()
	stmt, err := DB.PrepareContext(ctx, query)
	if err != nil {
		log.Printf("Error %s when preparing SQL statement", err)
		return reppost, err
	}
	defer stmt.Close()
	res, err := stmt.ExecContext(ctx, reppost.Id, reppost.Post, reppost.User)
	if err != nil {
		log.Printf("Error %s when inserting row into RepostPost table", err)
		return reppost, err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		log.Printf("Error %s when finding rows affected", err)
		return reppost, err
	}
	log.Printf("%d post created ", rows)
	return reppost, err
}

// --------------------Query from Subscribers table--------------------

func SelectSubscribersBtCommunities(communities string) []models.Subscribers {
	var sub models.Subscribers
	var subs []models.Subscribers
	query := fmt.Sprintf(`SELECT * FROM "Subscribers" WHERE "Communities" = '%s'`, communities)
	rows, err := DB.Query(query)
	if err != nil {
		fmt.Println("Error - SelectSubscribersBtCommunities()")
	}
	for rows.Next() {
		err = rows.Scan(&sub.Id, &sub.User, &sub.Communities)
		if err != nil {
			fmt.Println("Error - SelectSubscribersBtCommunities() rows.Next()")
		}
		subs = append(subs, sub)
	}
	return subs
}

/*
Функция удаления статуса из таблицы Subscribers
*/
func DeleteSubOnCommunities(communities string, user string) error {
	res, err := DB.Exec(`DELETE FROM "Subscribers" WHERE "Communities" = ($1) AND "User" = ($2)`, communities, user)
	if err == nil {
		count, err := res.RowsAffected()
		if err == nil {
			fmt.Println(count)
		}
		return nil
	}
	return err
}

func InsertSubscribersToUser(user string, communities string) error {
	query := `INSERT INTO "Subscribers"("User", "Communities") VALUES($1, $2)`
	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()
	stmt, err := DB.PrepareContext(ctx, query)
	if err != nil {
		log.Printf("Error %s when preparing SQL statement", err)
		return err
	}
	defer stmt.Close()
	res, err := stmt.ExecContext(ctx, user, communities)
	if err != nil {
		log.Printf("Error %s when inserting row into RepostPost table", err)
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		log.Printf("Error %s when finding rows affected", err)
		return err
	}
	log.Printf("%d post created ", rows)
	return err
}

// --------------------Query from MessengeList table--------------------

func SelectMessengeListbyUsers(author string, guest string) (models.MessageList, error) {
	var messageList models.MessageList
	query := fmt.Sprintf(`SELECT * FROM "MessageList" 
	WHERE "Main" = '%s' 
	AND "Companion" = '%s';`, author, guest)
	rows, err := DB.Query(query)
	if err != nil {
		fmt.Println("Error - SelectMessengeListbyUsers()")
		return messageList, err
	}
	for rows.Next() {
		err = rows.Scan(&messageList.LinkId, &messageList.Main, &messageList.Companion, &messageList.MessageHistory)
		if err != nil {
			fmt.Println("Error - SelectMessengeListbyUsers() rows.Next()")
			return messageList, err
		}
	}
	return messageList, nil
}

func SelectMessengeListbyLogin(author string, guest string) []models.MessageList {
	var messageList models.MessageList
	var msgArr []models.MessageList
	query := fmt.Sprintf(`SELECT "MessageList".*, "Users"."Name", "Users"."Photo" FROM "MessageList"
	JOIN "Users" ON "Users"."Login" = "MessageList"."Main"
	OR "Users"."Login" = "MessageList"."Companion"
	WHERE "Main" = '%s' 
	AND "Companion" = '%s';`, author, guest)
	rows, err := DB.Query(query)
	if err != nil {
		fmt.Println("Error - SelectMessengeListbyUsers()")
	}
	for rows.Next() {
		err = rows.Scan(&messageList.LinkId, &messageList.Main, &messageList.Companion, &messageList.MessageHistory)
		if err != nil {
			fmt.Println("Error - SelectMessengeListbyUsers() rows.Next()")
		}
		msgArr = append(msgArr, messageList)
	}
	return msgArr
}

func SelectCompanionsByLogin(user string) []models.Companions {
	var companion models.Companions
	var companions []models.Companions
	query := fmt.Sprintf(`
	SELECT "MessageList".*, "Users"."Name", "Users"."Photo"
	FROM "MessageList"
	JOIN "Users" ON "Users"."Login" = "MessageList"."Companion"
	WHERE "Main" = '%s';`, user)
	rows, err := DB.Query(query)
	if err != nil {
		fmt.Println("Error - SelectCompanionsByLogin()")
	}
	for rows.Next() {
		err = rows.Scan(&companion.LinkId, &companion.Main, &companion.Companion, &companion.MessageHistory, &companion.Name, &companion.Photo)
		if err != nil {
			fmt.Println("Error - SelectCompanionsByLogin() rows.Next()")
		}
		companions = append(companions, companion)
	}
	return companions
}

func InsertMessengeListbyUsers(data models.MessageList) error {
	query := `INSERT INTO "MessageList"("LinkId", "Main", "Companion", "MessageHistory") VALUES($1, $2, $3, $4)`
	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()
	stmt, err := DB.PrepareContext(ctx, query)
	if err != nil {
		log.Printf("Error %s when preparing SQL statement", err)
		return err
	}
	defer stmt.Close()
	res, err := stmt.ExecContext(ctx, data.LinkId, data.Main, data.Companion, data.MessageHistory)
	if err != nil {
		log.Printf("Error %s when inserting row into RepostPost table", err)
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		log.Printf("Error %s when finding rows affected", err)
		return err
	}
	log.Printf("%d post created ", rows)
	return err
}
