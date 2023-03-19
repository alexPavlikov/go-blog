package testing

import (
	"context"
	"crypto/rand"
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"testing"
	"time"

	"github.com/alexPavlikov/go-blog/models"
)

var DB *sql.DB
var err error
var name string

func init() {
	ConnectTest()
}

func ConnectTest() (*sql.DB, error) {
	name = "Test2023031216"
	DB, err = sql.Open("postgres", fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", "127.0.0.1", "5432", "postgres", "AlexPAV2307", name))
	if err != nil {
		fmt.Printf("Error - TestConnect() - %s", err.Error())
		return DB, err
	}
	return DB, nil
}

func TestSelectPost(t *testing.T) {
	rows, err := DB.Query(`SELECT "Posts".*, "Communities"."Photo"
	FROM "Posts" 
	JOIN "Communities" ON "Posts"."Communities" = "Communities"."Name"
	ORDER BY "Posts"."Date" DESC;`)
	if err != nil {
		t.Errorf("Error - selectPosts() - %s", err.Error())
	}

	post := models.Posts{}
	posts := []models.Posts{}

	for rows.Next() {

		err = rows.Scan(&post.Id, &post.Title, &post.Content, &post.Like, &post.View, &post.Date, &post.Communities, &post.Photo, &post.Category, &post.CommunitiesPhot)
		if err != nil {
			t.Errorf("Error - selectPosts() / rows.Next() - %s", err.Error())
		}
		posts = append(posts, post)
	}
	if posts == nil {
		t.Errorf("Error - TestSelectPost() - false")
	} else {
		t.Logf("Success !")
	}
}

func TestSelectPostsByUserSubs(t *testing.T) {
	user := `a.pavlikov2002@gmail.com`
	query := fmt.Sprintf(`SELECT "Posts".*, "Communities"."Photo"
	FROM "Posts" 
	JOIN "Communities" ON "Posts"."Communities" = "Communities"."Name"
	WHERE "Posts"."Communities" IN (SELECT "Subscribers"."Communities"
	FROM "Subscribers"
	WHERE "Subscribers"."User" = '%s');`, user)
	rows, err := DB.Query(query)
	if err != nil {
		t.Errorf("Error - selectPosts() - %s", err.Error())
	}

	post := models.Posts{}
	posts := []models.Posts{}

	for rows.Next() {

		err = rows.Scan(&post.Id, &post.Title, &post.Content, &post.Like, &post.View, &post.Date, &post.Communities, &post.Photo, &post.Category, &post.CommunitiesPhot)
		if err != nil {
			t.Errorf("Error - selectPosts() / rows.Next() - %s", err.Error())
		}
		posts = append(posts, post)
	}
	if posts[0].Id == "" {
		t.Errorf("Error - TestSelectPostsByUserSubs() - false")
	} else {
		t.Logf("Success !")
	}
}

func TestSelectPostById(t *testing.T) {
	id := 1
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
	if post == (models.Posts{}) {
		t.Errorf("Error - TestSelectPostById() - false")
	} else {
		t.Logf("Success !")
	}
}

func TestSelectPostsByCommunities(t *testing.T) {
	communities := "Golang"
	var posts []models.Posts
	var post models.Posts
	query := fmt.Sprintf(`SELECT "Posts".*, "Communities"."Photo"
	FROM "Posts" 
	JOIN "Communities" ON "Posts"."Communities" = "Communities"."Name" WHERE "Communities" = '%s' ORDER BY "Date" DESC`, communities)
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
	if posts[0].Id == "" {
		t.Errorf("Error - TestSelectPostsByCommunities() - false")
	} else {
		t.Logf("Success !")
	}
}

func TestInsertPosts(t *testing.T) {

	var posts []models.Posts

	for i := 1; i < 100; i++ {
		RandomCrypto, _ := rand.Prime(rand.Reader, 32)
		Id := fmt.Sprint(RandomCrypto.Int64() / 20000)
		post := models.Posts{
			Id:              Id + strconv.Itoa(i),
			Title:           "Title" + strconv.Itoa(i),
			Content:         "Content" + strconv.Itoa(i),
			Like:            0,
			View:            1,
			Date:            time.Now().Format("2006-01-02 15:04"),
			Communities:     "Golang",
			Photo:           "photo",
			Category:        "Новостной",
			CommunitiesPhot: "photo",
		}
		posts = append(posts, post)
	}

	for i, post := range posts {
		t.Run(fmt.Sprintf("Step - %d", i), func(t *testing.T) {
			query := `INSERT INTO "Posts"("Id","Title", "Content", "Like", "View", "Date", "Communities", "Photo", "Category") VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
			ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancelfunc()
			stmt, err := DB.PrepareContext(ctx, query)
			if err != nil {
				log.Printf("Error %s when preparing SQL statement", err)
			}
			defer stmt.Close()
			res, err := stmt.ExecContext(ctx, post.Id, post.Title, post.Content, post.Like, post.View, post.Date, post.Communities, post.Photo, post.Category)
			if err != nil {
				log.Printf("Error %s when inserting row into Posts table", err)
			}
			rows, err := res.RowsAffected()
			if err != nil {
				log.Printf("Error %s when finding rows affected", err)
			}
			if rows <= 0 {
				t.Errorf("Не верный результат. Ожидаемый результат - %t, результат - %t", false, true)
			} else {
				t.Logf("Success !")
			}
		})
	}
}

func TestInsertUsers(t *testing.T) {

	var users []models.Users

	for i := 1; i < 100; i++ {
		user := models.Users{
			Login:     "test.log" + strconv.Itoa(i) + "gmail.com",
			Password:  "111111",
			Name:      "Alex" + strconv.Itoa(i),
			Access:    "User",
			Photo:     "photo",
			Birthdate: time.Now().Format("2006-01-02 15:04"),
			Wallet:    0,
		}
		users = append(users, user)
	}

	for i, user := range users {
		t.Run(fmt.Sprintf("Step - %d", i), func(t *testing.T) {
			query := `INSERT INTO "Users"("Login", "Password", "Name", "Access", "Photo", "Birthdate", "Wallet") VALUES ($1, $2, $3, $4, $5, $6, $7)`
			ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancelfunc()
			stmt, err := DB.PrepareContext(ctx, query)
			if err != nil {
				log.Printf("Error %s when preparing SQL statement", err)
			}
			defer stmt.Close()
			res, err := stmt.ExecContext(ctx, user.Login, user.Password, user.Name, user.Access, user.Photo, user.Birthdate, user.Wallet)
			if err != nil {
				log.Printf("Error %s when inserting row into User table", err)
			}
			rows, err := res.RowsAffected()
			if err != nil {
				log.Printf("Error %s when finding rows affected", err)
			}
			if rows <= 0 {
				t.Errorf("Не верный результат. Ожидаемый результат - %t, результат - %t", false, true)
			} else {
				t.Logf("Success !")
			}
		})
	}
}

func TestInsertCommunities(t *testing.T) {

	var coms []models.Communities

	for i := 1; i < 100; i++ {
		com := models.Communities{
			Name:     "Test" + strconv.Itoa(i),
			Author:   "a.pavlikov2002@gmail.com",
			Photo:    "photo",
			Category: "Блог",
		}
		coms = append(coms, com)
	}

	for i, com := range coms {
		t.Run(fmt.Sprintf("Step - %d", i), func(t *testing.T) {
			query := `INSERT INTO "Communities"("Name", "Author", "Photo", "Category") VALUES ($1, $2, $3, $4)`
			ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancelfunc()
			stmt, err := DB.PrepareContext(ctx, query)
			if err != nil {
				log.Printf("Error %s when preparing SQL statement", err)
			}
			defer stmt.Close()
			res, err := stmt.ExecContext(ctx, com.Name, com.Author, com.Photo, com.Category)
			if err != nil {
				log.Printf("Error %s when inserting row into Communities table", err)
			}
			rows, err := res.RowsAffected()
			if err != nil {
				log.Printf("Error %s when finding rows affected", err)
			}
			if rows <= 0 {
				t.Errorf("Не верный результат. Ожидаемый результат - %t, результат - %t", false, true)
			} else {
				t.Logf("Success !")
			}
		})
	}
}

func TestInsertAccess(t *testing.T) {

	var coms []models.Access

	for i := 1; i < 100; i++ {
		com := models.Access{
			Name: "Access" + strconv.Itoa(i),
		}
		coms = append(coms, com)
	}

	for i, access := range coms {
		t.Run(fmt.Sprintf("Step - %d", i), func(t *testing.T) {
			query := `INSERT INTO "Access"("Name") VALUES ($1)`
			ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancelfunc()
			stmt, err := DB.PrepareContext(ctx, query)
			if err != nil {
				log.Printf("Error %s when preparing SQL statement", err)
			}
			defer stmt.Close()
			res, err := stmt.ExecContext(ctx, access.Name)
			if err != nil {
				log.Printf("Error %s when inserting row into Access table", err)
			}
			rows, err := res.RowsAffected()
			if err != nil {
				log.Printf("Error %s when finding rows affected", err)
			}
			if rows <= 0 {
				t.Errorf("Не верный результат. Ожидаемый результат - %t, результат - %t", false, true)
			} else {
				t.Logf("Success !")
			}
		})
	}
}

func TestInsertComments(t *testing.T) {

	var coms []models.Comments

	for i := 1; i < 100; i++ {
		com := models.Comments{
			Posts:  1,
			Text:   "Test_text" + strconv.Itoa(i),
			Like:   0,
			Author: "a.pavlikov2002@gmail.com",
		}
		coms = append(coms, com)
	}

	for i, comment := range coms {
		t.Run(fmt.Sprintf("Step - %d", i), func(t *testing.T) {
			query := `INSERT INTO "Comments"("Posts", "Text", "Like", "Author") VALUES ($1, $2, $3, $4)`
			ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancelfunc()
			stmt, err := DB.PrepareContext(ctx, query)
			if err != nil {
				log.Printf("Error %s when preparing SQL statement", err)
			}
			defer stmt.Close()
			res, err := stmt.ExecContext(ctx, comment.Posts, comment.Text, comment.Like, comment.Author)
			if err != nil {
				log.Printf("Error %s when inserting row into Comment table", err)
			}
			rows, err := res.RowsAffected()
			if err != nil {
				log.Printf("Error %s when finding rows affected", err)
			}
			if rows <= 0 {
				t.Errorf("Не верный результат. Ожидаемый результат - %t, результат - %t", false, true)
			} else {
				t.Logf("Success !")
			}
		})
	}
}

func TestInsertFriends(t *testing.T) {

	var coms []models.Friends

	for i := 1; i < 100; i++ {
		com := models.Friends{
			Login:  "a.",
			Status: "",
			Friend: "",
		}
		coms = append(coms, com)
	}

	for i, friend := range coms {
		t.Run(fmt.Sprintf("Step - %d", i), func(t *testing.T) {
			query := `INSERT INTO "Friends"("Login", "Status", "Friend") VALUES ($1, $2, $3)`
			ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancelfunc()
			stmt, err := DB.PrepareContext(ctx, query)
			if err != nil {
				log.Printf("Error %s when preparing SQL statement", err)
			}
			defer stmt.Close()
			res, err := stmt.ExecContext(ctx, friend.Login, friend.Status, friend.Friend)
			if err != nil {
				log.Printf("Error %s when inserting row into Friend table", err)
			}
			rows, err := res.RowsAffected()
			if err != nil {
				log.Printf("Error %s when finding rows affected", err)
			}
			log.Printf("%d post created ", rows)
			if rows <= 0 {
				t.Errorf("Не верный результат. Ожидаемый результат - %t, результат - %t", false, true)
			} else {
				t.Logf("Success !")
			}
		})
	}
}
