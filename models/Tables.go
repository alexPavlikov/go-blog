package models

type Posts struct { // structure of the Posts table
	Id              string
	Title           string
	Content         string
	Like            uint
	View            uint
	Date            string
	Communities     string
	Photo           string
	Category        string
	CommunitiesPhot string
}

type Communities struct { // structure of the Communities table
	Name     string
	Author   string
	Photo    string
	Category string
}

type JoinCommunities struct {
	Communities string
	User        string
	Photo       string
	Author      string
	Category    string
}

type Comments struct { // structure of the Comments table
	Id     uint32
	Posts  int
	Text   string
	Like   uint
	Author string
}

type JoinComments struct { // structure of the Comments table
	Posts  uint
	Author string
	Name   string
	Photo  string
	Text   string
	Like   uint
}

type Users struct { // structure of the Users table
	Login     string
	Password  string
	Name      string
	Access    string
	Photo     string
	Birthdate string
	Wallet    float64
	Gallery   []string
	Music     []int
}

type JoinUser struct {
	Login     string
	Friend    string
	Status    string
	Name      string
	Photo     string
	Birthdate string
	Wallet    float64
	Gallery   []string
	Music     []int
}

type Friends struct { // structure of the Friends table
	Id     uint
	Login  string
	Status string
	Friend string
}

type Status struct { // structure of the Status table
	Name string
}

type Access struct { // structure of the Access table
	Name string
}

type RepostPost struct {
	User             string
	Post             uint
	Title            string
	Content          string
	Like             uint
	View             uint
	Date             string
	Communities      string
	PostPhoto        string
	Categoty         string
	CommunitiesPhoto string
}

type Repost struct {
	Id   uint32
	Post int
	User string
}

type Subscribers struct {
	Id          int
	User        string
	Communities string
}

type Message struct {
	User    string `json:"User"`
	Message string `json:"Message"`
	Data    string `json:"Data"`
	Photo   string `json:"Photo"`
	Access  int    `json:"Access"`
}

type Messenger struct {
	Messenge []Message
}

type MessageList struct {
	LinkId         uint32
	Main           string
	Companion      string
	MessageHistory string
}

type Companions struct {
	LinkId         int
	Main           string
	Companion      string
	MessageHistory string
	Name           string
	Photo          string
}

type Gopher struct {
	Id      uint
	Creator string
	Owner   string
	Title   string
	Content string
	Like    uint
	View    uint
	Date    string
}

type JoinGopher struct {
	Id           uint
	Creator      string
	CreatorPhoto string
	CreatorName  string
	Owner        string
	Title        string
	Content      string
	Like         uint
	View         uint
	Date         string
}

type Store struct {
	Id          uint
	Name        string
	Photo       string
	Price       float32
	NewPrice    float32
	Description string
	Category    string
	Sex         string
	Community   string
}

type JoinStore struct {
	Id          uint
	Name        string
	Photo       string
	Price       float32
	NewPrice    float32
	Description string
	Category    string
	Sex         string
	Community   string
	Address     string
	Date        string
}

type JoinStorePlus struct {
	Id          uint
	Article     uint
	Name        string
	Photo       string
	Price       float32
	NewPrice    float32
	User        string
	Description string
	Category    string
	Sex         string
	Community   string
	Address     string
	Date        string
}

type StorePlus struct {
	Id          uint
	Name        string
	Photo       string
	Price       float32
	NewPrice    float32
	Description string
	Category    string
	Sex         string
	Community   string
	Status      bool
}

type Favorites struct {
	Id      uint64
	User    string
	Product []int64
}

type Fav struct {
	Id      uint64
	User    string
	Product []string
}

type Sales struct {
	Id      uint64
	Product uint
	User    string
	Address string
	Date    string
}

type UserBanned struct {
	Id     int
	User   string
	Reason string
	Time   int
	Admin  string
}

type Complaints struct {
	Id        int
	Criminal  string
	Complaint string
	Author    string
	Status    string
	Comment   string
	Admin     string
}

type MusicSub struct {
	Id     int
	Name   string
	Author string
	Genre  string
	Subs   int
	Link   string
	Login  string
}

type Music struct {
	Id     int
	Name   string
	Author string
	Genre  string
	Subs   int
	Link   string
}
