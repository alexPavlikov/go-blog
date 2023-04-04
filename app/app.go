package app

import (
	"crypto/md5"
	"crypto/rand"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	rnd "math/rand"
	"net/http"
	"os"
	"time"

	gomail "gopkg.in/mail.v2"

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

func RecordingSessions(session string, filename string) error {
	// fmt.Println(session)
	file, err := os.OpenFile(fmt.Sprintf("C:/Users/admin/go/src/go-blog/data/files/%s", filename), os.O_WRONLY|os.O_APPEND, 0755)
	if err != nil {
		fmt.Println(err)
		// os.Exit(1)
		return err
	}
	defer file.Close()
	data := []byte(session)
	_, err = file.Write(data)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

func CreateFile(check models.MessageList, guestId string, userLogin string) error {
	Path := "C:/Users/admin/go/src/go-blog/data/files/message/"

	if check.MessageHistory != "" {
		// fmt.Println("start", check.MessageHistory)
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
			return err
		}
		value.LinkId += 1
		err = database.InsertDoubleMessengeListbyUsers(value)
		if err != nil {
			fmt.Println(err.Error())
			return err
		}
		new, err := os.Create(Path + value.MessageHistory)
		if err != nil {
			fmt.Println("Error - createFile() Create file")
			return err
		}
		fmt.Println("Create file - ", Path+value.MessageHistory)
		check = models.MessageList{}
		// Copy standart json format
		take, err := os.Open(Path + "take.json")
		if err != nil {
			fmt.Println(`Error - os.Open()`, err)
			return err
		}
		_, err = io.Copy(new, take)
		if err != nil {
			fmt.Println("Error - createFile() io.Copy()")
			return err
		}
	}
	return nil
}

func TrimLeftChar(s string) string {
	for i := range s {
		if i > 0 {
			return s[i:]
		}
	}
	return s[:0]
}

func JSON(msg models.Message, Path string, userLogin string) (models.Messenger, error) {
	// fmt.Println("JSON", Path)
	var settings models.Messenger
	rawDataIn, err := ioutil.ReadFile(Path)
	if err != nil {
		log.Fatal("Cannot load settings:", err)
		return settings, err
	}

	err = json.Unmarshal(rawDataIn, &settings)
	if err != nil {
		log.Fatal("Invalid settings format:", err)
		return settings, err
	}

	if msg.Message != "" {
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
			return settings, err
		}

		err = ioutil.WriteFile(Path, rawDataOut, 0)
		if err != nil {
			log.Fatal("Cannot write updated settings file:", err)
			return settings, err
		}
		return settings, nil
	}
	return settings, err
}

func CreateMd5Hash(text string) string {
	hasher := md5.New()
	_, err := io.WriteString(hasher, text)
	if err != nil {
		panic(err)
	}
	return hex.EncodeToString(hasher.Sum(nil))
}

func ReadCookies(r *http.Request) *http.Cookie {
	var c *http.Cookie
	for _, c = range r.Cookies() {
		fmt.Println(c)
	}
	return c
}

func GiveCode() (bytes int) {

	rnd.Seed(time.Now().UTC().UnixNano())
	//var bytes int
	bytes = rnd.Intn(9999) + 999
	fmt.Println("Code---------", bytes)
	return bytes
	// RandomCrypto, _ := rand.Prime(rand.Reader, 32)
	// id := RandomCrypto.Int64() / 20000
	// fmt.Println(id)
	// RandomIntegerwithinRange = rnd.Intn(10000) + 1
	// if RandomIntegerwithinRange <= 1000 {
	// 	RandomIntegerwithinRange = GiveCode()
	// }
	// fmt.Println(RandomIntegerwithinRange)
	// return RandomIntegerwithinRange
}

func SendCode(email string, code int, name string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", "a.pavlikov2002@gmail.com")
	m.SetHeader("To", email)

	m.SetHeader("Subject", "Gopher.go")

	message := fmt.Sprintf(`Hello %s - %d`, name, code)
	m.SetBody("text/plain", message)
	d := gomail.NewDialer("smtp.gmail.com", 587, "a.pavlikov2002@gmail.com", "isei dkte iiwl wior")

	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	if err := d.DialAndSend(m); err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

func SendNewPass(email string, pass string, name string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", "a.pavlikov2002@gmail.com")
	m.SetHeader("To", email)

	m.SetHeader("Subject", "Gopher.go")

	message := fmt.Sprintf(`Здравствуйте, %s, Ваш новый пароль - %s`, name, pass)
	m.SetBody("text/plain", message)
	d := gomail.NewDialer("smtp.gmail.com", 587, "a.pavlikov2002@gmail.com", "isei dkte iiwl wior")

	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	if err := d.DialAndSend(m); err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}
