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
	Login       string
	Password    string
	Name        string
	Access      string
	Communities string
	Photo       string
	Birthdate   string
}

type JoinUser struct {
	Login     string
	Friend    string
	Status    string
	Name      string
	Photo     string
	Birthdate string
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
