package app

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/alexPavlikov/go-blog/database"
	"github.com/alexPavlikov/go-blog/models"
)

func CheckArray(fav []int64, id uint) bool {
	for _, i := range fav {
		if i == int64(id) {
			return false
		}
	}
	return true
}

// func Cookies() {
// 	fmt.Println(client)
// 	jar, err := cookiejar.New(nil)
// 	if err != nil {
// 		log.Fatalf("Got error while creating cookie jar %s", err.Error())
// 	}
// 	client = http.Client{
// 		Jar: jar,
// 	}
// 	fmt.Println(client)
// }

func RecordingSessions(session string) {
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

func CreateFile(check models.MessageList, guestId string, userLogin string) {
	Path := "C:/Users/admin/go/src/go-blog/data/files/message/"
	fmt.Println("start", check.MessageHistory)
	if check.MessageHistory != "" {
		fmt.Println("createFile!!!!", check.MessageHistory)
	} else {
		RandomCrypto, _ := rand.Prime(rand.Reader, 32)
		value := models.MessageList{
			LinkId:         uint32(RandomCrypto.Int64() / 20000),
			Main:           userLogin,
			Companion:      guestId,
			MessageHistory: time.Now().Format("20060102150405") + ".json",
		}
		fmt.Println(value)
		err := database.InsertMessengeListbyUsers(value)
		if err != nil {
			fmt.Println(err.Error())
		}
		value.LinkId += 1
		err = database.InsertDoubleMessengeListbyUsers(value)
		if err != nil {
			fmt.Println(err.Error())
		}
		new, err := os.Create(Path + value.MessageHistory)
		if err != nil {
			fmt.Println("Error - createFile() Create file")
		}
		fmt.Println("Create file - ", Path+value.MessageHistory)
		check = models.MessageList{}
		// Copy standart json format
		take, err := os.Open(Path + "take.json")
		if err != nil {
			fmt.Println(`Error - os.Open()`, err)
		}
		_, err = io.Copy(new, take)
		if err != nil {
			fmt.Println("Error - createFile() io.Copy()")
		}
	}
}

func TrimLeftChar(s string) string {
	for i := range s {
		if i > 0 {
			return s[i:]
		}
	}
	return s[:0]
}

func JSON(msg models.Message, Path string, userLogin string) models.Messenger {
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
		if settings.Messenge[i].User == userLogin {
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
