package setting

import (
	"encoding/json"
	"io/fs"
	"log"
	"os"

	"github.com/alexPavlikov/go-blog/models"
)

func Config() {
	file, err := os.Open("./setting/setting.cfg")
	if err != nil {
		log.Fatal("Error - Открыть файл ", err.Error())
	}
	defer file.Close()

	var stat fs.FileInfo
	stat, err = file.Stat()
	if err != nil {
		log.Fatal("Error - Статистика файла ", err.Error())
	}
	reabByte := make([]byte, stat.Size())

	_, err = file.Read(reabByte)
	if err != nil {
		log.Fatal("Error - Чтение файла ", err.Error())
	}

	err = json.Unmarshal(reabByte, &models.Cfg)
	if err != nil {
		log.Fatal("Error - Unmarshal файла ", err.Error())
	}
}
