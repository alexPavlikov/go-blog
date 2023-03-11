package api

import "net/http"

func handleRequest() {
	// обработчики статических данных(папок)
	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("./assets/"))))
	http.Handle("/data/", http.StripPrefix("/data/", http.FileServer(http.Dir("./data/"))))

	//обработчики всех ссылок веб-сайта
	http.HandleFunc("/", logFormHandler)
	http.HandleFunc("/auth", authHandler)
	http.HandleFunc("/registration", regHandler)
	http.HandleFunc("/blog", blogHandler)
	http.HandleFunc("/page", pageHandler)
	http.HandleFunc("/page/post", pagePostHandler)
	http.HandleFunc("/page/del_gof", pageDelGofHandler)
	http.HandleFunc("/page/rep_gof", pageRepGofHandler)
	http.HandleFunc("/setting", settingHandler)
	http.HandleFunc("/setting/refresh", refreshSettingHandler)
	http.HandleFunc("/friends", friendsHandler)
	http.HandleFunc("/friends/add", addFriendsHandler)
	http.HandleFunc("/friends/rec", recFriendsHandler)
	http.HandleFunc("/communities", communitiesHandler)
	http.HandleFunc("/communities/add", communitiesAddHandler)
	http.HandleFunc("/comments", commentsHandler)
	http.HandleFunc("/community", communityHandler)
	http.HandleFunc("/community/del", communityDelHandler)
	http.HandleFunc("/community/post", communityPostHandler)
	http.HandleFunc("/community/edit", communityEditHandler)
	http.HandleFunc("/community/market", communityMarketHandler)
	http.HandleFunc("/community/market/add", communityMarketAddHandler)
	http.HandleFunc("/community/market/del", communityMarketDelHandler)
	http.HandleFunc("/community/store/card", communityStoreCardHandler)
	http.HandleFunc("/guest", guestHandler)
	http.HandleFunc("/guest/friends", guestFriendsHandler)
	http.HandleFunc("/guest/communities", guestCommunitiesHandler)
	http.HandleFunc("/message", messageHandler)
	http.HandleFunc("/store", storeHandler)
	http.HandleFunc("/store/card", storeCardHandler)
	http.HandleFunc("/store/buy", storeBuyHandler)
	http.HandleFunc("/store/favorites", storeFavoritesHandler)
	http.HandleFunc("/favourites", favouritesPageHandler)
	http.HandleFunc("/exit", exitHandler)

	//http.HandleFunc("/fr", frHandler)
}
