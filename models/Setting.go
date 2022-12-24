package models

type Setting struct {
	ServerHost        string
	ServerPort        string
	PgHost            string
	PgPort            string
	PgUser            string
	PgPass            string
	PgName            string
	Data              string
	Assets            string
	Html              string
	Email             string
	BlogTitle         string
	SettingTitle      string
	FriendsTitile     string
	CommunitiesTitile string
}

var Cfg Setting
