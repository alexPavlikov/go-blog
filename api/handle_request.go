package api

import "net/http"

func HandleRequest() {
	// обработчики статических данных(папок)
	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("./assets/"))))
	http.Handle("/data/", http.StripPrefix("/data/", http.FileServer(http.Dir("./data/"))))

	//обработчики всех ссылок веб-сайта
	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/entry", logFormHandler)
	http.HandleFunc("/auth", authHandler)
	http.HandleFunc("/second_auth", secondAuthHandler)
	http.HandleFunc("/change_pass", changePasswordHandler)
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
	http.HandleFunc("/community/market/sort", communityMarketSortHandler)
	http.HandleFunc("/community/market/add", communityMarketAddHandler)
	http.HandleFunc("/community/market/del", communityMarketDelHandler)
	http.HandleFunc("/community/store/card", communityStoreCardHandler)
	http.HandleFunc("/community/market/sale_list", communityMarketSaleListHandler)
	http.HandleFunc("/community/market/market/statistics", communityMarketStatisticsHandler)
	http.HandleFunc("/guest", guestHandler)
	http.HandleFunc("/guest/friends", guestFriendsHandler)
	http.HandleFunc("/guest/communities", guestCommunitiesHandler)
	http.HandleFunc("/message", messageHandler)
	http.HandleFunc("/store", storeHandler)
	http.HandleFunc("/store/sort", storeSortHandler)
	http.HandleFunc("/store/card", storeCardHandler)
	http.HandleFunc("/store/buy", storeBuyHandler)
	http.HandleFunc("/store/favorites", storeFavoritesHandler)
	http.HandleFunc("/favourites", favouritesPageHandler)
	http.HandleFunc("/help", helpHandler)
	http.HandleFunc("/help/complaint", helpComplaintHandler)
	http.HandleFunc("/exit", exitHandler)

	//admin
	http.HandleFunc("/admin", adminHandler)
	http.HandleFunc("/admin/banned", adminBanHandler)
	http.HandleFunc("/admin/del", adminDelHandler)
	http.HandleFunc("/admin/list", adminDelBanListHandler)
	http.HandleFunc("/admin/amd_list", adminListHandler)

	http.HandleFunc("/fr", frHandler)
}
