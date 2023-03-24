package testing

import (
	"fmt"
	"testing"
	"time"

	rnd "math/rand"

	"github.com/alexPavlikov/go-blog/app"
	"github.com/alexPavlikov/go-blog/models"
)

func TestRecordingSessions(t *testing.T) {

	var session string
	users := []models.Users{
		{
			Login:    "a1.pavlikov2002@gmail.com",
			Password: "12341234",
			Name:     "Alex1",
		},
		{
			Login:    "a2.pavlikov2002@gmail.com",
			Password: "12341234",
			Name:     "Alex2",
		},
		{
			Login:    "a3.pavlikov2002@gmail.com",
			Password: "12341234",
			Name:     "Alex3",
		},
	}
	for i, us := range users {
		t.Run(fmt.Sprintf("Step - %d", i), func(t *testing.T) {
			session = fmt.Sprintf("Пользователь, %s (логин - %s, пароль - %s) зашел в аккаунт в %s.\n", us.Name, us.Login, us.Password, time.Now().Format("2006-01-02 15:04"))

			err := app.RecordingSessions(session)
			if err != nil {
				t.Errorf("Функция не прошла тестирования - %s", err.Error())
			} else {
				t.Logf("Success !")
			}
		})
	}

}

func TestCheckArray(t *testing.T) {
	tests := []struct {
		input []int64
		want  int
	}{
		{[]int64{1}, 2},
		{[]int64{1, 2, 3, 4, 5}, 15},
		{[]int64{}, 0},
		{[]int64{-1, -1}, -2},
		{[]int64{-1, -1, 0, 0, 0}, -2},
		{[]int64{0, 0, 0}, 2},
		{[]int64{-1, 0, 1}, 3},
		{[]int64{78, 11, 48, 91, 13, 26, 74, 99}, 440},
		{[]int64{-95, -46, -65, -63, 10}, -259},
	}

	for i, tc := range tests {
		t.Run(fmt.Sprintf("Step - %d", i), func(t *testing.T) {
			ok := app.CheckArray(tc.input, uint(tc.want))
			if !ok {
				t.Errorf("Не верный результат. Ожидаемый результат - %t, результат - %t", false, true)
			} else {
				t.Logf("Success !")
			}
		})
	}
}

func TestCreateFile(t *testing.T) {
	var check models.MessageList
	var guestId string
	var userLogin string

	check.LinkId = 1
	check.Companion = "b.helloworld@gmail.com"
	check.Main = "a.pavlikov2002@gmail.com"
	check.MessageHistory = "test.json"
	guestId = "ds.skype@gmail.com"
	userLogin = "a.pavlikov2002@gmail.com"

	err := app.CreateFile(check, guestId, userLogin)
	if err != nil {
		t.Errorf("Функция не прошла тестирования - %s", err.Error())
	} else {
		t.Logf("Success !")
	}
}

func TestTrimLeftChar(t *testing.T) {

	s := "Hello"

	val := app.TrimLeftChar(s)
	if val != "ello" {
		t.Errorf("Не верный результат. Ожидаемый результат - %s, результат - %s", "ello", val)
	} else {
		t.Logf("Success !")
	}
}

func TestJSON(t *testing.T) {
	var msg models.Message
	var Path string

	user := "a.pavlikov2002@gmail.com"
	message := "Hello, test!"
	Path = "C:/Users/admin/go/src/go-blog/data/files/message/test.json"
	msg = models.Message{
		User:    user,
		Message: message,
		Data:    time.Now().Format("2006-01-02 15:04"),
	}

	_, err := app.JSON(msg, Path, user)
	if err != nil {
		t.Errorf("Функция не прошла тестирования - %s", err.Error())
	} else {
		t.Logf("Success !")
	}
}

func TestCreateMd5Hash(t *testing.T) {
	var str = []string{
		"passsword",
		"andreythisistop21",
		"pashport",
		"wehavecheaters",
		"thiissecret",
		"discord",
		"IvanGay228",
		"12341234",
		"ilikemeet",
		"Password",
		"justrelaxbro",
		"pasport",
	}

	for i, s := range str {
		t.Run(fmt.Sprintf("Step - %d", i), func(t *testing.T) {
			hash := app.CreateMd5Hash(s)
			fmt.Println(hash, s)
			if hash != app.CreateMd5Hash(s) {
				t.Errorf("Не верный результат. Ожидаемый результат - %t, результат - %t", false, true)
			} else {
				t.Logf("Success !")
			}
		})
	}
}

func TestGiveCode(t *testing.T) {
	for i := 0; i < 100; i++ {
		rnd.Seed(time.Now().UTC().UnixNano())
		//var bytes int
		bytes := rnd.Intn(9999) + 999
		if bytes <= 1000 {
			bytes = app.GiveCode()
		}
		fmt.Println(bytes)

	}
}
