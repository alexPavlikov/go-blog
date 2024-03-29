package database

import (
	"context"
	"crypto/rand"
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/alexPavlikov/go-blog/models"
	"github.com/lib/pq"
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
	ORDER BY "Posts"."Date" DESC;`)
	if err != nil {
		fmt.Println("Error - selectPosts()", err.Error())
	}
	defer rows.Close()
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
Функция выборки всех постов из таблицы Posts по подписке
*/
func SelectPostsByUserSubs(user string) []models.Posts {
	query := fmt.Sprintf(`SELECT "Posts".*, "Communities"."Photo"
	FROM "Posts" 
	JOIN "Communities" ON "Posts"."Communities" = "Communities"."Name"
	WHERE "Posts"."Communities" IN (SELECT "Subscribers"."Communities"
	FROM "Subscribers"
	WHERE "Subscribers"."User" = '%s');`, user)
	rows, err := DB.Query(query)
	if err != nil {
		fmt.Println("Error - SelectPostsByUserSubs()", err.Error())
	}
	defer rows.Close()
	post := models.Posts{}
	posts := []models.Posts{}

	for rows.Next() {

		err = rows.Scan(&post.Id, &post.Title, &post.Content, &post.Like, &post.View, &post.Date, &post.Communities, &post.Photo, &post.Category, &post.CommunitiesPhot)
		if err != nil {
			fmt.Println("Error - SelectPostsByUserSubs() / rows.Next()", err.Error())
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
	defer rows.Close()
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
	JOIN "Communities" ON "Posts"."Communities" = "Communities"."Name" WHERE "Communities" = '%s' ORDER BY "Date" DESC`, communities)
	rows, err := DB.Query(query)
	if err != nil {
		fmt.Println("Error - SelectPostByCommunities()", err.Error())
	}
	defer rows.Close()
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
func DeleteCommunitiesPostByTime(community string, time string) error {
	res, err := DB.Exec(`DELETE FROM "Posts" WHERE "Communities" = ($1) AND "Date" = ($2)`, community, time)
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
	query := `INSERT INTO "Posts"("Id","Title", "Content", "Like", "View", "Date", "Communities", "Photo", "Category") VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()
	stmt, err := DB.PrepareContext(ctx, query)
	if err != nil {
		log.Printf("Error %s when preparing SQL statement", err)
		return post, err
	}
	defer stmt.Close()
	res, err := stmt.ExecContext(ctx, post.Id, post.Title, post.Content, post.Like, post.View, post.Date, post.Communities, post.Photo, post.Category)
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

func UpdateViewInCommunityPost(name string) error {
	query := fmt.Sprintf(`UPDATE "Posts" SET "View" = "View" + 1 WHERE "Communities" = '%s'`, name)
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
	defer rows.Close()
	var user models.Users
	var users []models.Users
	for rows.Next() {
		err = rows.Scan(&user.Login, &user.Password, &user.Name, &user.Access, &user.Photo, &user.Birthdate, &user.Wallet, pq.Array(&user.Gallery), pq.Array(&user.Music))
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
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&user.Login, &user.Password, &user.Name, &user.Access, &user.Photo, &user.Birthdate, &user.Wallet, pq.Array(&user.Gallery), pq.Array(&user.Music))
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
	rows, err := DB.Query(fmt.Sprintf(`SELECT * FROM "Users" WHERE "%s" = '%s' AND "Access" = 'User'`, column, value))
	if err != nil {
		fmt.Println("Error - SelectUsersByColumn()", err)
		return users, err
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&user.Login, &user.Password, &user.Name, &user.Access, &user.Photo, &user.Birthdate, &user.Wallet, pq.Array(&user.Gallery), pq.Array(&user.Music))
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
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&user.Login, &user.Password, &user.Name, &user.Access, &user.Photo, &user.Birthdate, &user.Wallet, pq.Array(&user.Gallery), pq.Array(&user.Music))
		if err != nil {
			fmt.Println("Error - SelectUserByColumn() rows.Next()", err.Error())
			return user, err
		}
	}
	return user, nil
}

func SelectUserWallet(login string, password string) (models.Users, error) {
	var user models.Users
	rows, err := DB.Query(fmt.Sprintf(`SELECT * FROM "Users" WHERE "Login" = '%s' AND "Password" = '%s'`, login, password))
	if err != nil {
		fmt.Println("Error - SelectUserWallet()", err)
		return user, err
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&user.Login, &user.Password, &user.Name, &user.Access, &user.Photo, &user.Birthdate, &user.Wallet, pq.Array(&user.Gallery), pq.Array(&user.Music))
		if err != nil {
			fmt.Println("Error - SelectUserWallet() rows.Next()", err.Error())
			return user, err
		}
	}
	return user, nil
}

func SelectAdmins() ([]models.Users, error) {
	var user models.Users
	var users []models.Users
	rows, err := DB.Query(`SELECT * FROM "Users" WHERE "Access" != 'User' AND "Access" != 'Banned'`)
	if err != nil {
		fmt.Println("Error - SelectAdmins()", err)
		return users, err
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&user.Login, &user.Password, &user.Name, &user.Access, &user.Photo, &user.Birthdate, &user.Wallet, pq.Array(&user.Gallery), pq.Array(&user.Music))
		if err != nil {
			fmt.Println("Error - SelectAdmins() rows.Next()", err.Error())
			return users, err
		}
		users = append(users, user)
	}
	return users, nil
}

func FindUsers(query string, login string) ([]models.Users, error) {
	var user models.Users
	var users []models.Users
	query = "%" + query + "%"
	rows, err := DB.Query(fmt.Sprintf(`SELECT "Users".* FROM "Users"
	WHERE "Users"."Login" NOT IN 
	(SELECT "Friends"."Login" FROM "Friends" WHERE "Friends"."Login" = '%s' OR "Friends"."Friend" = '%s') 
	 AND "Access" = 'User'
	 AND "Name" ILIKE '%s'
	 AND "Login" != '%s'`, login, login, query, login))
	if err != nil {
		fmt.Println("Error - FindUsers() ", err)
		return users, err
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&user.Login, &user.Password, &user.Name, &user.Access, &user.Photo, &user.Birthdate, &user.Wallet, pq.Array(&user.Gallery), pq.Array(&user.Music))
		if err != nil {
			fmt.Println("Error - FindUsers() ", err)
			return users, err
		}
		users = append(users, user)
	}
	return users, err
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
	query := `INSERT INTO "Users"("Login", "Password", "Name", "Access", "Photo", "Birthdate", "Wallet", "Gallery", "Music") VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()
	stmt, err := DB.PrepareContext(ctx, query)
	if err != nil {
		log.Printf("Error %s when preparing SQL statement", err)
		return user, err
	}
	defer stmt.Close()
	res, err := stmt.ExecContext(ctx, user.Login, user.Password, user.Name, user.Access, user.Photo, user.Birthdate, user.Wallet, pq.Array(&user.Gallery), pq.Array(&user.Music))
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
	query := fmt.Sprintf(`UPDATE "Users" SET "Password" = '%s', "Name" = '%s', "Access" = '%s', "Photo" = '%s', "Birthdate" = '%s', "Wallet" = %f WHERE "Login" = '%s' AND "Password" =  '%s'`, user.Password, user.Name, user.Access, user.Photo, user.Birthdate, user.Wallet, login, password)
	_, err := DB.Query(query)
	if err != nil {
		fmt.Println(err)
		return err
	}
	fmt.Println(query)
	return nil
}

/*
Функция обновления кошелька пользователя из таблицы Users по введенным Login и Password
*/
func UpdateUserWalletByLogPass(login string, password string, money float64) error {
	query := fmt.Sprintf(`UPDATE "Users" SET "Wallet" = "Wallet"+'%f' WHERE "Login" = '%s' AND "Password" =  '%s'`, money, login, password)
	_, err := DB.Query(query)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

/*
Функция-конструктор, вводится два значения column и value и подставляются в запрос,
что позволяет не писать повторяющие запросы SELECT
*/
func UpdateUserByColumn(column string, value string, login string, pass string) (user models.Users, err error) {
	query := fmt.Sprintf(`UPDATE "Users" SET "%s" = '%s' WHERE "Login" = '%s' AND "Password" =  '%s'`, column, value, login, pass)
	_, err = DB.Query(query)
	if err != nil {
		fmt.Println(err)
		return user, err
	}
	fmt.Println(query)
	return user, err
}

func SelectUserBannedAllByAdmin(admin string) ([]models.UserBanned, error) {
	var user models.UserBanned
	var users []models.UserBanned
	rows, err := DB.Query(fmt.Sprintf(`SELECT * FROM "UserBanned" WHERE "Admin" = '%s'`, admin))
	if err != nil {
		fmt.Println("Error - SelectUserBanned()", err)
		return users, err
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&user.Id, &user.User, &user.Reason, &user.Time, &user.Admin)
		if err != nil {
			fmt.Println("Error - SelectUserBanned() rows.Next()", err.Error())
			return users, err
		}
		users = append(users, user)
	}
	return users, nil
}

func SelectUserBannedByAdmin(admin string) ([]models.UserBanned, error) {
	var user models.UserBanned
	var users []models.UserBanned
	rows, err := DB.Query(fmt.Sprintf(`SELECT * FROM "UserBanned" WHERE "Admin" = '%s' AND "Reason" != 'Удаленный аккаунт'`, admin))
	if err != nil {
		fmt.Println("Error - SelectUserBannedByAdmin()", err)
		return users, err
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&user.Id, &user.User, &user.Reason, &user.Time, &user.Admin)
		if err != nil {
			fmt.Println("Error - SelectUserBannedByAdmin() rows.Next()", err.Error())
			return users, err
		}
		users = append(users, user)
	}
	return users, nil
}

func SelectUserDeletedByAdmin(admin string) ([]models.UserBanned, error) {
	var user models.UserBanned
	var users []models.UserBanned
	rows, err := DB.Query(fmt.Sprintf(`SELECT * FROM "UserBanned" WHERE "Admin" = '%s' AND "Reason" = 'Удаленный аккаунт'`, admin))
	if err != nil {
		fmt.Println("Error - SelectUserDeletedByAdmin()", err)
		return users, err
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&user.Id, &user.User, &user.Reason, &user.Time, &user.Admin)
		if err != nil {
			fmt.Println("Error - SelectUserDeletedByAdmin() rows.Next()", err.Error())
			return users, err
		}
		users = append(users, user)
	}
	return users, nil
}

func InsertUserToBanList(ban models.UserBanned) error {
	query := `INSERT INTO "UserBanned"("User", "Reason", "Time", "Admin") VALUES ($1, $2, $3, $4)`
	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()
	stmt, err := DB.PrepareContext(ctx, query)
	if err != nil {
		log.Printf("Error %s when preparing SQL statement", err)
		return err
	}
	defer stmt.Close()
	res, err := stmt.ExecContext(ctx, ban.User, ban.Reason, ban.Time, ban.Admin)
	if err != nil {
		log.Printf("Error %s when inserting row into User table", err)
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		log.Printf("Error %s when finding rows affected", err)
		return err
	}
	log.Printf("%d user created ", rows)
	return err
}

//Добавить еще обновление фото, имени, доступа

// --------------------Query from Communities table--------------------

func CheckCommunity(user, community string) ([]models.Communities, bool) {
	var communiti models.Communities
	var communities []models.Communities
	ok := false
	rows, err := DB.Query(fmt.Sprintf(`SELECT "Communities".* FROM "Communities","Subscribers"
	WHERE "Communities"."Name" = '%s'
	AND "Subscribers"."User" IN (SELECT "User" FROM "Subscribers" WHERE "User" = '%s')
	AND "Subscribers"."Communities" = '%s';`, community, user, community))
	if err != nil {
		fmt.Println("Error - CheckCommunity()", err.Error())
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&communiti.Name, &communiti.Author, &communiti.Photo, &communiti.Category)
		if err != nil {
			fmt.Println("Error - CheckCommunity() rows.Next()", err.Error())
		}
		communities = append(communities, communiti)
	}
	if len(communities) == 0 {
		ok = true
		return communities, ok
	}
	return communities, ok
}

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
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&communities.Name, &communities.Author, &communities.Photo, &communities.Category)
		if err != nil {
			fmt.Println("Error - SelectCommunities() rows.Next()", err.Error())
		}
		communitiesArr = append(communitiesArr, communities)
	}
	return communitiesArr
}

func FindCommunities(query string, login string) ([]models.Communities, error) {
	var communitie models.Communities
	var communities []models.Communities
	query = "%" + query + "%"
	rows, err := DB.Query(fmt.Sprintf(`SELECT * FROM "Communities"
	WHERE "Name" ILIKE '%s' AND "Name" NOT IN
	(SELECT "Communities" FROM "Subscribers" WHERE "User" = '%s');`, query, login))
	if err != nil {
		fmt.Println("Error - FindCommunities() ", err)
		return communities, err
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&communitie.Name, &communitie.Author, &communitie.Photo, &communitie.Category)
		if err != nil {
			fmt.Println("Error - FindCommunities() ", err)
			return communities, err
		}
		communities = append(communities, communitie)
	}
	return communities, err
}

func SelectCommunitiesWithOutSub(user string) []models.Communities {
	var communities models.Communities
	var communitiesArr []models.Communities
	rows, err := DB.Query(fmt.Sprintf(`
	SELECT "Communities".* FROM "Communities"
		WHERE "Communities"."Name" NOT IN 
		(SELECT "Communities"."Name"
		FROM "Communities"
		JOIN "Subscribers" ON "Subscribers"."Communities" = "Communities"."Name"
		WHERE "Subscribers"."User"='%s')
	`, user))
	if err != nil {
		fmt.Println("Error - SelectCommunities()", err.Error())
	}
	defer rows.Close()
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
	defer rows.Close()
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
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&communities.Name, &communities.Author, &communities.Photo, &communities.Category)
		if err != nil {
			fmt.Println("Error - SelectCommunitiesByColumn() rows.Next()", err.Error())
		}
	}
	return communities
}

func UpdateCommunity(communities models.Communities, name string) error {
	query := fmt.Sprintf(`UPDATE "Communities" SET "Name" = '%s', "Photo" = '%s', "Category" = '%s' WHERE "Name" = '%s'`, communities.Name, communities.Photo, communities.Category, name)
	_, err := DB.Query(query)
	if err != nil {
		fmt.Println(err)
		return err
	}
	fmt.Println(query)
	return nil
}

func UpdateCommunityAuthor(communities models.Communities, name string) error {
	query := fmt.Sprintf(`UPDATE "Communities" SET "Author" = '%s' WHERE "Name" = '%s'`, communities.Author, name)
	_, err := DB.Query(query)
	if err != nil {
		fmt.Println(err)
		return err
	}
	fmt.Println(query)
	return nil
}

/*
Функция выборки все сообщества из таблицы Communities определенного пользователя
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
	defer rows.Close()
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
	defer rows.Close()
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
	query := `INSERT INTO "Communities"("Name", "Author", "Photo", "Category") VALUES ($1, $2, $3, $4)`
	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()
	stmt, err := DB.PrepareContext(ctx, query)
	if err != nil {
		log.Printf("Error %s when preparing SQL statement", err)
		return com, err
	}
	defer stmt.Close()
	res, err := stmt.ExecContext(ctx, com.Name, com.Author, com.Photo, com.Category)
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
	defer rows.Close()
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
	defer rows.Close()
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
	defer rows.Close()
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
func SelectFriends() []models.JoinUser {
	var friend models.JoinUser
	var friends []models.JoinUser
	rows, err := DB.Query(`
	SELECT "Friends"."Login", "Friends"."Friend", "Friends"."Status",
	"Users"."Name", "Users"."Photo", "Users"."Birthdate"
	FROM "Friends"
	JOIN "Users" ON "Users"."Login" = "Friends"."Friend"
`)
	if err != nil {
		fmt.Println("Error - SelectFriendsByColumn()", err.Error())
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&friend.Login, &friend.Friend, &friend.Status, &friend.Name, &friend.Photo, &friend.Birthdate)
		if err != nil {
			fmt.Println("Error - SelectFriendsByColumn() rows.Next()", err.Error())
		}
		friends = append(friends, friend)
	}
	return friends
}

func CheckFriends(user string, frd string) (friend models.JoinUser, ok bool) {

	rows, err := DB.Query(fmt.Sprintf(`
	SELECT "Friends"."Login", "Friends"."Friend", "Friends"."Status",
	"Users"."Name", "Users"."Photo", "Users"."Birthdate"
	FROM "Friends"
	JOIN "Users" ON "Users"."Login" = "Friends"."Friend"
	WHERE "Friends"."Login" = '%s' AND "Friends"."Friend" = '%s';
`, user, frd))
	if err != nil {
		fmt.Println("Error - SelectFriendsByColumn()", err.Error())
	}
	for rows.Next() {
		err = rows.Scan(&friend.Login, &friend.Friend, &friend.Status, &friend.Name, &friend.Photo, &friend.Birthdate)
		if err != nil {
			fmt.Println("Error - SelectFriendsByColumn() rows.Next()", err.Error())
		}
	}
	defer rows.Close()
	if friend.Login == "" || friend.Friend == "" {
		ok = true
	} else {
		ok = false
	}
	fmt.Println(ok)
	return friend, ok
}

func DeleteFriendByLogin(user string, sub string) (string, error) {
	res, err := DB.Exec(`DELETE FROM "Friends" WHERE "Login" IN ($1, $2) AND "Friend" IN ($3, $4)`, user, sub, user, sub)
	if err == nil {
		count, err := res.RowsAffected()
		if err == nil {
			fmt.Println(count)
		}
		return sub, nil
	}
	return sub, err
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
	defer rows.Close()
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
	defer rows.Close()
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
	defer rows.Close()
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
	defer rows.Close()
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
	query := `INSERT INTO "RepostPost"("Post", "User") VALUES ($1, $2)`
	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()
	stmt, err := DB.PrepareContext(ctx, query)
	if err != nil {
		log.Printf("Error %s when preparing SQL statement", err)
		return reppost, err
	}
	defer stmt.Close()
	res, err := stmt.ExecContext(ctx, reppost.Post, reppost.User)
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

func DeleteRepoPostByUser(id int, user string) error {
	query := fmt.Sprintf(`DELETE FROM "RepostPost" WHERE "Post" = %d AND "User" = '%s';`, id, user)
	fmt.Println(query)
	res, err := DB.Exec(query)
	if err == nil {
		count, err := res.RowsAffected()
		if err == nil {
			fmt.Println(count)
		}
		return nil
	}
	return err
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
	defer rows.Close()
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

func DeleteSubAllOnCommunity(communities string) error {
	res, err := DB.Exec(`DELETE FROM "Subscribers" WHERE "Communities" = ($1)`, communities)
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
	defer rows.Close()
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
	defer rows.Close()
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
	defer rows.Close()
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

func InsertDoubleMessengeListbyUsers(data models.MessageList) error {
	query := `INSERT INTO "MessageList"("LinkId", "Main", "Companion", "MessageHistory") VALUES($1, $2, $3, $4)`
	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()
	stmt, err := DB.PrepareContext(ctx, query)
	if err != nil {
		log.Printf("Error %s when preparing SQL statement", err)
		return err
	}
	defer stmt.Close()
	res, err := stmt.ExecContext(ctx, data.LinkId, data.Companion, data.Main, data.MessageHistory)
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

// --------------------Query from Friends table--------------------

func SelectUserSub(user string) []models.Users {
	var sub models.Users
	var subs []models.Users
	query := fmt.Sprintf(`
	SELECT "Users".* FROM "UserSubs"
	JOIN "Users" ON "Users"."Login" = "UserSubs"."Sub"
	WHERE "UserSubs"."User" = '%s';`, user)
	rows, err := DB.Query(query)
	if err != nil {
		fmt.Println("Error - SelectUserSub()")
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&sub.Login, &sub.Password, &sub.Name, &sub.Access, &sub.Photo, &sub.Birthdate, &sub.Wallet, pq.Array(&sub.Gallery), pq.Array(&sub.Music))
		if err != nil {
			fmt.Println("Error - SelectUserSub() rows.Next()", err)
		}
		subs = append(subs, sub)
	}
	return subs
}

func InsertUserSub(user string, sub string) error {
	query := `INSERT INTO "UserSubs"("Id", "User", "Sub") VALUES($1, $2, $3)`
	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()
	stmt, err := DB.PrepareContext(ctx, query)
	if err != nil {
		log.Printf("Error %s when preparing SQL statement", err)
		return err
	}
	defer stmt.Close()
	RandomCrypto, _ := rand.Prime(rand.Reader, 32)
	id := RandomCrypto.Int64() / 20000
	fmt.Println(id, user, sub)
	res, err := stmt.ExecContext(ctx, id, user, sub)
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

func DeleteUserSub(user string, sub string) error {
	res, err := DB.Exec(`DELETE FROM "UserSubs" WHERE "User" = ($1) AND "Sub" = ($2)`, user, sub)
	if err == nil {
		count, err := res.RowsAffected()
		if err == nil {
			fmt.Println(count)
		}
		return nil
	}
	return err
}

//--

func SelectRecomendationFriends(user string) []models.Users {
	var sub models.Users
	var subs []models.Users
	query := fmt.Sprintf(`
	SELECT "Users".* FROM "Users"
	WHERE "Users"."Login" NOT IN 
	(SELECT "Friends"."Login" FROM "Friends" WHERE "Friends"."Login" = '%s' OR "Friends"."Friend" = '%s') AND "Access" = 'User'
	AND "Users"."Login" != '%s'`, user, user, user)
	rows, err := DB.Query(query)
	if err != nil {
		fmt.Println("Error - SelectRecomendationFriends()")
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&sub.Login, &sub.Password, &sub.Name, &sub.Access, &sub.Photo, &sub.Birthdate, &sub.Wallet, pq.Array(&sub.Gallery), pq.Array(&sub.Music))
		if err != nil {
			fmt.Println("Error - SelectRecomendationFriends() rows.Next()", err)
		}
		subs = append(subs, sub)
	}
	return subs
}

func SelectRecCommunities(user string) []models.Communities {
	var communities models.Communities
	var communitiesArr []models.Communities
	query := fmt.Sprintf(`
	SELECT * FROM "Communities" WHERE "Communities"."Category" IN 
	(SELECT "Communities"."Category" FROM "Subscribers" JOIN "Communities" ON "Communities"."Name" = "Subscribers"."Communities" WHERE "Subscribers"."User" = 's')
	OR "Communities"."Name" NOT IN
	(SELECT "Communities" FROM "Subscribers" WHERE "User" = '%s');
	`, user)
	rows, err := DB.Query(query)
	if err != nil {
		fmt.Println("Error - SelectRecCommunities()", err.Error())
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&communities.Name, &communities.Author, &communities.Photo, &communities.Category)
		if err != nil {
			fmt.Println("Error - SelectRecCommunities() rows.Next()", err.Error())
		}
		communitiesArr = append(communitiesArr, communities)
	}
	return communitiesArr
}

// ---
func SelectPostCategory() []string {
	var category string
	var categories []string
	rows, err := DB.Query(`SELECT * FROM "PostCatogory"`)
	if err != nil {
		fmt.Println("Error - SelectPostCategory()")
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&category)
		if err != nil {
			fmt.Println("Error - SelectPostCategory() rows.Next()")
		}
		categories = append(categories, category)
	}
	return categories
}

//---

func SelectCommunitiesCategory() []string {
	var category string
	var categories []string
	rows, err := DB.Query(`SELECT * FROM "CommunitiesCategory"`)
	if err != nil {
		fmt.Println("Error - SelectCommunitiesCategory()")
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&category)
		if err != nil {
			fmt.Println("Error - SelectCommunitiesCategory() rows.Next()")
		}
		categories = append(categories, category)
	}
	return categories
}

//---

func SelectOnlineFriends(user string) []models.JoinUser {
	var friend models.JoinUser
	var friends []models.JoinUser
	query := fmt.Sprintf(`SELECT "Friends"."Login", "Friends"."Friend", "Friends"."Status",
	"Users"."Name", "Users"."Photo", "Users"."Birthdate"
	FROM "Friends"
	JOIN "Users" ON "Users"."Login" = "Friends"."Friend"
	WHERE "Friends"."Login" =  '%s' AND "Friend" IN 
(SELECT * FROM "Online");`, user)
	rows, err := DB.Query(query)
	if err != nil {
		fmt.Println("Error - SelectOnlineFriends()")
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&friend.Login, &friend.Friend, &friend.Status, &friend.Name, &friend.Photo, &friend.Birthdate, pq.Array(&friend.Gallery), pq.Array(&friend.Music))
		if err != nil {
			fmt.Println("Error - SelectOnlineFriends() rows.Next()")
		}
		friends = append(friends, friend)
	}
	return friends
}
func InsertUserToOnline(user string) error {
	query := `INSERT INTO "Online"("User") VALUES($1)`
	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()
	stmt, err := DB.PrepareContext(ctx, query)
	if err != nil {
		log.Printf("Error %s when preparing SQL statement", err)
		return err
	}
	defer stmt.Close()
	fmt.Println(user)
	res, err := stmt.ExecContext(ctx, user)
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

func DeleteOnlineUser(user string) error {
	res, err := DB.Exec(`DELETE FROM "Online" WHERE "User" = ($1)`, user)
	if err == nil {
		count, err := res.RowsAffected()
		if err == nil {
			fmt.Println(count)
		}
		return nil
	}
	return err
}

// --------------------Query from CustomPosts table--------------------

func SelectGopherByOwner(user string) []models.JoinGopher {
	var gopher models.JoinGopher
	var gophers []models.JoinGopher
	rows, err := DB.Query(fmt.Sprintf(`SELECT "CustomPosts".*, "Users"."Photo", "Users"."Name" FROM "CustomPosts"
	JOIN "Users" ON "Users"."Login" = "CustomPosts"."Creator"
	WHERE "CustomPosts"."Owner" = '%s'`, user))
	if err != nil {
		fmt.Println("Error - SelectGopherByOwner()")
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&gopher.Id, &gopher.Creator, &gopher.Owner, &gopher.Title, &gopher.Content, &gopher.Like, &gopher.View, &gopher.Date, &gopher.CreatorPhoto, &gopher.CreatorName)
		if err != nil {
			fmt.Println("Error - SelectGopherByOwner() rows.Next()")
		}
		gophers = append(gophers, gopher)
	}
	return gophers
}

func SelectGopherById(id int) models.JoinGopher {
	var gopher models.JoinGopher
	rows, err := DB.Query(fmt.Sprintf(`SELECT "CustomPosts".*, "Users"."Photo", "Users"."Name" FROM "CustomPosts"
	JOIN "Users" ON "Users"."Login" = "CustomPosts"."Creator"
	WHERE "CustomPosts"."Id" = '%d'`, id))
	if err != nil {
		fmt.Println("Error - SelectGopherById()")
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&gopher.Id, &gopher.Creator, &gopher.Owner, &gopher.Title, &gopher.Content, &gopher.Like, &gopher.View, &gopher.Date, &gopher.CreatorPhoto, &gopher.CreatorName)
		if err != nil {
			fmt.Println("Error - SelectGopherById() rows.Next()")
		}
	}
	return gopher
}

func InsertGopher(gof models.Gopher) error {
	query := `INSERT INTO "CustomPosts"("Creator", "Owner", "Title", "Content", "Like", "View", "Date") VALUES($1, $2, $3, $4, $5, $6, $7)`
	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()
	stmt, err := DB.PrepareContext(ctx, query)
	if err != nil {
		log.Printf("Error %s when preparing SQL statement", err)
		return err
	}
	defer stmt.Close()
	res, err := stmt.ExecContext(ctx, gof.Creator, gof.Owner, gof.Title, gof.Content, gof.Like, gof.View, gof.Date)
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

func DeleteGopherByUser(user string, id int) error {
	res, err := DB.Exec(`DELETE FROM "CustomPosts" WHERE "Owner" = ($1) AND "Id" = ($2)`, user, id)
	if err == nil {
		count, err := res.RowsAffected()
		if err == nil {
			fmt.Println(count)
		}
		return nil
	}
	return err
}

func InsertLikeToGopher(id string) error {
	idInt, err := strconv.Atoi(id)
	if err != nil {
		fmt.Println(err.Error())
	}
	query := fmt.Sprintf(`UPDATE "CustomPosts" SET "Like" = "Like" + 1 WHERE "Id" = '%d'`, idInt)
	_, err = DB.Query(query)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

// store

func SelectStoreItems() ([]models.Store, error) {
	var product models.Store
	var products []models.Store
	rows, err := DB.Query(`SELECT * FROM "Store"`)
	if err != nil {
		fmt.Println("Error - SelectStoreItems()", err)
		return products, err
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&product.Id, &product.Name, &product.Photo, &product.Price, &product.NewPrice, &product.Description, &product.Category, &product.Sex, &product.Community)
		if err != nil {
			fmt.Println("Error - SelectStoreItems() rows.Next()", err.Error())
			return products, err
		}
		products = append(products, product)
	}
	return products, nil
}

func SelectStoreItemsByCommunity(community string) ([]models.Store, error) {
	var product models.Store
	var products []models.Store
	rows, err := DB.Query(fmt.Sprintf(`SELECT * FROM "Store" WHERE "Community" = '%s'`, community))
	if err != nil {
		fmt.Println("Error - SelectStoreItemsByCommunity()", err)
		return products, err
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&product.Id, &product.Name, &product.Photo, &product.Price, &product.NewPrice, &product.Description, &product.Category, &product.Sex, &product.Community)
		if err != nil {
			fmt.Println("Error - SelectStoreItemsByCommunity() rows.Next()", err.Error())
			return products, err
		}
		products = append(products, product)
	}
	return products, nil
}

func SelectStoreItemsByCategory(category string) ([]models.Store, error) {
	var product models.Store
	var products []models.Store
	rows, err := DB.Query(fmt.Sprintf(`SELECT * FROM "Store" WHERE "Category" = '%s'`, category))
	if err != nil {
		fmt.Println("Error - SelectStoreItemsByCategory()", err)
		return products, err
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&product.Id, &product.Name, &product.Photo, &product.Price, &product.NewPrice, &product.Description, &product.Category, &product.Sex, &product.Community)
		if err != nil {
			fmt.Println("Error - SelectStoreItemsByCategory() rows.Next()", err.Error())
			return products, err
		}
		products = append(products, product)
	}
	return products, nil
}

func SelectStoreItemsByCommunityAndCategory(category string, community string) ([]models.Store, error) {
	var product models.Store
	var products []models.Store
	rows, err := DB.Query(fmt.Sprintf(`SELECT * FROM "Store" WHERE "Community" = '%s' AND "Category" = '%s'`, community, category))
	if err != nil {
		fmt.Println("Error - SelectStoreItemsByCategory()", err)
		return products, err
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&product.Id, &product.Name, &product.Photo, &product.Price, &product.NewPrice, &product.Description, &product.Category, &product.Sex, &product.Community)
		if err != nil {
			fmt.Println("Error - SelectStoreItemsByCategory() rows.Next()", err.Error())
			return products, err
		}
		products = append(products, product)
	}
	return products, nil
}

func SelectStoreItemById(id int) (models.StorePlus, error) {
	var product models.StorePlus
	rows, err := DB.Query(fmt.Sprintf(`SELECT * FROM "Store" WHERE "Id" = %d`, id))
	if err != nil {
		fmt.Println("Error - SelectStoreItemById()", err)
		return product, err
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&product.Id, &product.Name, &product.Photo, &product.Price, &product.NewPrice, &product.Description, &product.Category, &product.Sex, &product.Community)
		if err != nil {
			fmt.Println("Error - SelectStoreItemById() rows.Next()", err.Error())
			return product, err
		}
	}
	return product, nil
}

//favorites

func SelectFavouritesByUser(login string) ([]models.Store, bool) {
	var product models.Store
	var products []models.Store
	ok := false
	rows, err := DB.Query(fmt.Sprintf(`SELECT "Store".* FROM "Favourites"
	JOIN "Store" ON "Store"."Id" = ANY("Favourites"."Product")
	WHERE "Favourites"."User" = '%s'`, login))
	if err != nil {
		fmt.Println("Error - SelectFavouritesByUser()", err)
		return products, ok
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&product.Id, &product.Name, &product.Photo, &product.Price, &product.NewPrice, &product.Description, &product.Category, &product.Sex, &product.Community)
		if err != nil {
			fmt.Println("Error - SelectFavouritesByUser() rows.Next()", err.Error())
			return products, ok
		}
		products = append(products, product)
	}
	if products[0].Id > 0 && products[0].Name != "" {
		ok = true
	}
	return products, ok
}

func SelectFavouritesByUserCheck(login string) ([]int64, bool) {
	var fav models.Favorites
	ok := false
	sel := `SELECT "Product" FROM "Favourites" WHERE "User" = $1`

	// wrap the output parameter in pq.Array for receiving into it
	if err := DB.QueryRow(sel, login).Scan(pq.Array(&fav.Product)); err != nil {
		log.Fatal(err, "SelectFavouritesByUser2")
	}

	if fav.Product == nil {
		ok = true
	}

	return fav.Product, ok
}

func InsertFavoritesToUser(login string, product []int) error {
	query := `INSERT INTO "Favourites"("User", "Product") VALUES($1, $2)`
	_, err := DB.Exec(query, login, pq.Array(product))
	if err != nil {
		return err
	} else {
		fmt.Println("\nRow inserted successfully!")
		return nil
	}
}

func UpdateFavouritesToUser(user string, product uint) error {
	query := fmt.Sprintf(`UPDATE "Favourites" SET "Product" = array_append("Product", %d) WHERE "User" = '%s';`, product, user)
	_, err = DB.Query(query)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

func DeleteFavouritesUser(user string, product uint) error {
	res, err := DB.Exec(fmt.Sprintf(`UPDATE "Favourites" SET "Product" = array_remove("Product", %d) WHERE "User" = '%s'`, product, user))
	if err == nil {
		count, err := res.RowsAffected()
		if err == nil {
			fmt.Println(count)
		}
		return nil
	}
	return err
}

// SALES

func SelectSalesByUser(login string) ([]models.JoinStore, bool) {
	var product models.JoinStore
	var products []models.JoinStore
	ok := false
	rows, err := DB.Query(fmt.Sprintf(`SELECT "Store".*, "Sales"."Address", "Sales"."Date" FROM "Sales"
	JOIN "Store" ON "Store"."Id" = "Sales"."Product"
	WHERE "Sales"."User" =  '%s'`, login))
	if err != nil {
		fmt.Println("Error - SelectSalesByUser()", err)
		return products, ok
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&product.Id, &product.Name, &product.Photo, &product.Price, &product.NewPrice, &product.Description, &product.Category, &product.Sex, &product.Community, &product.Address, &product.Date)
		if err != nil {
			fmt.Println("Error - SelectSalesByUser() rows.Next()", err.Error())
			return products, ok
		}
		products = append(products, product)
	}
	if products[0].Id > 0 && products[0].Name != "" {
		ok = true
	}
	return products, ok
}

func SelectSalesByCommunity(community string) ([]models.JoinStorePlus, bool) {
	var product models.JoinStorePlus
	var products []models.JoinStorePlus
	ok := false
	rows, err := DB.Query(fmt.Sprintf(`SELECT "Sales"."Id", "Store".*, "Sales"."User", "Sales"."Address", "Sales"."Date" FROM "Sales"
	JOIN "Store" ON "Store"."Id" = "Sales"."Product"
	WHERE "Store"."Community" =  '%s'`, community))
	if err != nil {
		fmt.Println("Error - SelectSalesByCommunity()", err)
		return products, ok
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&product.Id, &product.Article, &product.Name, &product.Photo, &product.Price, &product.NewPrice, &product.Description, &product.Category, &product.Sex, &product.Community, &product.User, &product.Address, &product.Date)
		if err != nil {
			fmt.Println("Error - SelectSalesByCommunity() rows.Next()", err.Error())
			return products, ok
		}
		products = append(products, product)
	}
	if products != nil {
		ok = true
	}
	return products, ok
}

func InsertToSalesProduct(sale models.Sales) error {
	query := `INSERT INTO "Sales"("Product", "User", "Address", "Date") VALUES($1, $2, $3, $4)`
	_, err := DB.Exec(query, sale.Product, sale.User, sale.Address, sale.Date)
	if err != nil {
		return err
	} else {
		fmt.Println("\nRow inserted successfully!")
		return nil
	}
}

func SelectTotalPurchaseByUser(user string) (result int) {
	rows, err := DB.Query(fmt.Sprintf(`SELECT SUM("Store"."NewPrice") FROM "Sales" 
	JOIN "Store" ON "Sales"."Product" = "Store"."Id"
	WHERE "Sales"."User" = '%s'`, user))
	if err != nil {
		fmt.Println("Error - SelectTotalPurchaseByUser()", err)
		return result
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&result)
		if err != nil {
			fmt.Println("Error - SelectTotalPurchaseByUser() rows.Next()", err.Error())
			return result
		}
	}
	return result
}

// sex

func SelectSex() ([]string, error) {
	var sx string
	var sex []string
	rows, err := DB.Query(`SELECT * FROM "Sex"`)
	if err != nil {
		fmt.Println("Error - SelectSex()", err)
		return sex, err
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&sx)
		if err != nil {
			fmt.Println("Error - SelectSex() rows.Next()", err.Error())
			return sex, err
		}
		sex = append(sex, sx)
	}
	return sex, nil
}

//category store

func SelectStoreCategory() ([]string, error) {
	var ct string
	var category []string
	rows, err := DB.Query(`SELECT * FROM "StoreCategory"`)
	if err != nil {
		fmt.Println("Error - SelectStoreCategory()", err)
		return category, err
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&ct)
		if err != nil {
			fmt.Println("Error - SelectStoreCategory() rows.Next()", err.Error())
			return category, err
		}
		category = append(category, ct)
	}
	return category, nil
}

func InsertToStoreProduct(prd models.Store) error {
	query := `INSERT INTO "Store"("Name", "Photo", "Price", "NewPrice", "Description", "Category", "Sex", "Community") VALUES($1, $2, $3, $4, $5, $6, $7, $8)`
	_, err := DB.Exec(query, prd.Name, prd.Photo, prd.Price, prd.NewPrice, prd.Description, prd.Category, prd.Sex, prd.Community)
	if err != nil {
		return err
	} else {
		fmt.Println("\nRow inserted successfully!")
		return nil
	}
}

func DeleteStore(id int) error {
	res, err := DB.Exec(`DELETE FROM "Store" WHERE "Id" = ($1)`, id)
	if err == nil {
		count, err := res.RowsAffected()
		if err == nil {
			fmt.Println(count)
		}
		return nil
	}
	return err
}

//complaints

func SelectComplaints() ([]models.Complaints, error) {
	var complaint models.Complaints
	var complaints []models.Complaints
	rows, err := DB.Query(`SELECT "Id", "Criminals", "Complaint", "Author", "Status", COALESCE("Comment", ''), COALESCE("Admin", '') FROM "Complaints"`)
	if err != nil {
		fmt.Println("Error - SelectComplaints()", err)
		return complaints, err
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&complaint.Id, &complaint.Criminal, &complaint.Complaint, &complaint.Author, &complaint.Status, &complaint.Comment, &complaint.Admin)
		if err != nil {
			fmt.Println("Error - SelectComplaints() rows.Next()", err.Error())
			return complaints, err
		}
		complaints = append(complaints, complaint)
	}
	return complaints, nil
}

func SelectComplaint(id int) (models.Complaints, error) {
	var complaint models.Complaints
	rows, err := DB.Query(fmt.Sprintf(`SELECT "Id", "Criminals", "Complaint", "Author", "Status", COALESCE("Comment", ''), COALESCE("Admin", '') FROM "Complaints" WHERE "Id" = %d`, id))
	if err != nil {
		fmt.Println("Error - SelectComplaint()", err)
		return complaint, err
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&complaint.Id, &complaint.Criminal, &complaint.Complaint, &complaint.Author, &complaint.Status, &complaint.Comment, &complaint.Admin)
		if err != nil {
			fmt.Println("Error - SelectComplaint() rows.Next()", err.Error())
			return complaint, err
		}
	}
	return complaint, nil
}

func UpdateComplaintById(com models.Complaints) error {
	fmt.Println("_________UpdateComplaintById___________", com)
	//UPDATE "Complaints" SET "Status" = 'решена', "Comment" = 'Все выполнил', "Admin" = 'a.pavlikov2002@gmail.com' WHERE "Id" = 2;
	query := fmt.Sprintf(`UPDATE "Complaints" SET "Status" = '%s', "Comment" = '%s', "Admin" = '%s' WHERE "Id" = %d;`, com.Status, com.Comment, com.Admin, com.Id)
	fmt.Println(query)
	_, err = DB.Query(query)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

func InsertComplaint(prd models.Complaints) error {
	query := `INSERT INTO "Complaints"("Criminals", "Complaint", "Author", "Status") VALUES($1, $2, $3, $4)`
	_, err := DB.Exec(query, prd.Criminal, prd.Complaint, prd.Author, prd.Status)
	if err != nil {
		return err
	} else {
		fmt.Println("\nRow inserted successfully!")
		return nil
	}
}

// music

func SelectMusicByUser(login string) []models.MusicSub {
	var music models.MusicSub
	var musics []models.MusicSub
	rows, err := DB.Query(fmt.Sprintf(`SELECT "Music".*, "Users"."Login" FROM "Users"
	JOIN "Music" ON "Music"."Id" = ANY("Users"."Music")
	WHERE "Users"."Login" = '%s'`, login))
	if err != nil {
		fmt.Println("Error - SelectMusicByUser()", err)
		return musics
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&music.Id, &music.Name, &music.Author, &music.Genre, &music.Subs, &music.Link, &music.Login)
		if err != nil {
			fmt.Println("Error - SelectMusicByUser() rows.Next()", err.Error())
			return musics
		}
		musics = append(musics, music)
	}
	return musics
}

func InsertMusic(music models.Music) error {
	query := `INSERT INTO "Music"("Name", "Author", "Genre", "Subs", "Link") VALUES($1, $2, $3, $4, $5)`
	_, err := DB.Exec(query, music.Name, music.Author, music.Genre, music.Subs, music.Link)
	if err != nil {
		return err
	} else {
		fmt.Println("\nRow inserted successfully!")
		return nil
	}
}

func UpdateMusicToUser(user string, music int) error {
	query := fmt.Sprintf(`UPDATE "Users" SET "Music" = array_append("Music", %d) WHERE "Login" = '%s';`, music, user)
	_, err = DB.Query(query)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

func UpdateMusicSub(music int) error {
	query := fmt.Sprintf(`UPDATE "Music" SET "Subs" = "Subs" + 1 WHERE "Id" = %d;`, music)
	_, err = DB.Query(query)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

func DeleteMusicUser(user string, music int) error {
	res, err := DB.Exec(fmt.Sprintf(`UPDATE "Users" SET "Music" = array_remove("Music", %d) WHERE "Login" = '%s'`, music, user))
	if err == nil {
		count, err := res.RowsAffected()
		if err == nil {
			fmt.Println(count)
		}
		return nil
	}
	return err
}

// gallery LIMIT

func SelectUserGallery(user string) (photos []string) {
	rows, err := DB.Query(fmt.Sprintf(`SELECT "Gallery" FROM "Users" WHERE "Login" = '%s'`, user))
	if err != nil {
		fmt.Println("Error - SelectUserGallery()", err)
		return photos
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(pq.Array(&photos))
		if err != nil {
			fmt.Println("Error - SelectUserGallery() rows.Next()", err.Error())
			return photos
		}
	}
	return photos
}

func SelectUserGalleryLimit(user string) (photos []string) {
	rows, err := DB.Query(fmt.Sprintf(`SELECT "Gallery" FROM "Users" WHERE "Login" = '%s' LIMIT 5`, user))
	if err != nil {
		fmt.Println("Error - SelectUserGallery()", err)
		return photos
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(pq.Array(&photos))
		if err != nil {
			fmt.Println("Error - SelectUserGallery() rows.Next()", err.Error())
			return photos
		}
	}
	return photos
}

func UpdateUserGallery(user string, photo string) error {
	query := fmt.Sprintf(`UPDATE "Users" SET "Gallery" = array_append("Gallery", '%s') WHERE "Login" = '%s';`, photo, user)
	_, err = DB.Query(query)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}
