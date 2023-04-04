package api

import (
	"net/http"
	"os"

	"github.com/alexPavlikov/go-blog/models"
)

var client http.Client
var postId string
var id string
var posts []models.Posts
var post models.Posts
var guestId string
var communitiesName string
var communitiesPhoto string
var inputComment string
var guestLogin string
var Path string
var Communication []models.Message
var check models.MessageList
var export string
var userAuth models.Users
var imgPath, title, content string
var usrMesg string
var f *os.File
var Link string
var companion []models.Companions
var UsersLink models.MessageList
var activeChatUser models.Users
var OK bool
var Messenger models.Messenger
var itemId string
var product models.StorePlus
var code int
var Id int
var complaintStatus = []string{"не решена", "выполняется", "решена"}
