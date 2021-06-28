// This file is part of the JUSTtheTalkAPI distribution (https://github.com/jdudmesh/justthetalk-api).
// Copyright (c) 2021 John Dudmesh.

// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, version 3.

// This program is distributed in the hope that it will be useful, but
// WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU
// General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package server

import (
	"io"
	"net/http"
	"net/http/httptest"

	log "github.com/sirupsen/logrus"

	"github.com/gorilla/mux"

	"justthetalk/businesslogic"
	"justthetalk/handlers"
	"justthetalk/middleware"

	"sync"
)

var once sync.Once

func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, `{"alive": true}`)
}

type App struct {
	router *mux.Router

	mostActiveWorker *businesslogic.MostActiveWorker

	postProcessor *businesslogic.PostProcessor

	userCache         *businesslogic.UserCache
	folderCache       *businesslogic.FolderCache
	discussionCache   *businesslogic.DiscussionCache
	bannedWordList    *businesslogic.BannedWordsList
	bookmarkProcessor *businesslogic.BookmarkProcessor
}

func NewApp() *App {

	userCache := businesslogic.NewUserCache()
	folderCache := businesslogic.NewFolderCache()
	discussionCache := businesslogic.NewDiscussionCache(folderCache)

	app := &App{
		postProcessor:     businesslogic.NewPostProcessor(userCache, folderCache, discussionCache),
		mostActiveWorker:  businesslogic.NewMostActiveWorker(),
		userCache:         userCache,
		folderCache:       folderCache,
		discussionCache:   discussionCache,
		bannedWordList:    businesslogic.NewBannedWordsList(),
		bookmarkProcessor: businesslogic.NewBookmarkProcessor(),
	}

	app.router = app.configureRouter()

	return app

}

func (a *App) configureRouter() *mux.Router {

	databaseMiddleware := middleware.NewDatabaseMiddleware()
	sessionMiddleware := middleware.NewSessionMiddleware(a.userCache)

	router := mux.NewRouter().StrictSlash(false)
	router.Use(databaseMiddleware.Middleware, sessionMiddleware.Middleware)

	a.configureFolderRouter(router)
	a.configureFrontPageRouter(router)
	a.configureUserRouter(router)
	a.configureSearchRouter(router)
	a.configureAdminRouter(router)

	router.HandleFunc("/health", HealthCheckHandler)

	websocketHandler := handlers.NewWebsockerHandler(a.userCache)
	router.HandleFunc("/ws", websocketHandler.ServeHTTP)

	return router

}

func (a *App) configureFolderRouter(router *mux.Router) {

	folderHandler := handlers.NewFolderHandler(a.userCache, a.folderCache, a.discussionCache, a.bookmarkProcessor, a.postProcessor)

	folderRouter := router.PathPrefix("/folder").Subrouter().StrictSlash(false)
	folderRouter.HandleFunc("", folderHandler.GetFolders).Methods(http.MethodGet, http.MethodOptions)
	folderRouter.HandleFunc("/{folderId:[0-9]+}", folderHandler.GetFolder).Methods(http.MethodGet, http.MethodOptions)

	folderRouter.HandleFunc("/{folderId:[0-9]+}/discussion", folderHandler.GetDiscussions).Methods(http.MethodGet, http.MethodOptions)
	folderRouter.HandleFunc("/{folderId:[0-9]+}/discussion", folderHandler.CreateDiscussion).Methods(http.MethodPost, http.MethodOptions)
	folderRouter.HandleFunc("/{folderId:[0-9]+}/discussion/{discussionId:[0-9]+}", folderHandler.GetDiscussion).Methods(http.MethodGet, http.MethodOptions)
	folderRouter.HandleFunc("/{folderId:[0-9]+}/discussion/{discussionId:[0-9]+}", folderHandler.EditDiscussion).Methods(http.MethodPut, http.MethodOptions)
	folderRouter.HandleFunc("/{folderId:[0-9]+}/discussion/{discussionId:[0-9]+}", folderHandler.DeleteDiscussion).Methods(http.MethodDelete, http.MethodOptions)

	folderRouter.HandleFunc("/{folderId:[0-9]+}/discussion/{discussionId:[0-9]+}/post", folderHandler.GetPosts).Methods(http.MethodGet, http.MethodOptions)
	folderRouter.HandleFunc("/{folderId:[0-9]+}/discussion/{discussionId:[0-9]+}/post", folderHandler.CreatePost).Methods(http.MethodPost, http.MethodOptions)
	folderRouter.HandleFunc("/{folderId:[0-9]+}/discussion/{discussionId:[0-9]+}/post/{postId:[0-9]+}", folderHandler.EditPost).Methods(http.MethodPut, http.MethodOptions)
	folderRouter.HandleFunc("/{folderId:[0-9]+}/discussion/{discussionId:[0-9]+}/post/{postId:[0-9]+}", folderHandler.DeletePost).Methods(http.MethodDelete, http.MethodOptions)

	folderRouter.HandleFunc("/{folderId:[0-9]+}/subscription", folderHandler.SubscribeToFolder).Methods(http.MethodPost, http.MethodDelete, http.MethodOptions)
	folderRouter.HandleFunc("/{folderId:[0-9]+}/discussion/{discussionId:[0-9]+}/subscription", folderHandler.SubscribeToDiscussion).Methods(http.MethodPost, http.MethodDelete, http.MethodOptions)

}

func (a *App) configureFrontPageRouter(router *mux.Router) {

	frontPageHandler := handlers.NewFrontPageHandler(a.userCache, a.discussionCache)

	frontPageRouter := router.PathPrefix("/frontpage/{viewType}").Subrouter().StrictSlash(false)
	frontPageRouter.HandleFunc("", frontPageHandler.GetFrontPage).Methods(http.MethodGet, http.MethodOptions)

}

func (a *App) configureSearchRouter(router *mux.Router) {

	searchHandler := handlers.NewSearchHandler(a.folderCache, a.discussionCache)

	searchRouter := router.PathPrefix("/search").Subrouter().StrictSlash(false)
	searchRouter.HandleFunc("", searchHandler.SearchPosts).Methods(http.MethodGet, http.MethodOptions)

}

func (a *App) configureUserRouter(router *mux.Router) {

	userHandler := handlers.NewUserHandler(a.userCache, a.folderCache, a.discussionCache)

	userRouter := router.PathPrefix("/user").Subrouter().StrictSlash(false)
	userRouter.HandleFunc("", userHandler.GetUser).Methods(http.MethodGet, http.MethodOptions)
	userRouter.HandleFunc("/{userId}", userHandler.GetOtherUser).Methods(http.MethodGet, http.MethodOptions)
	userRouter.HandleFunc("", userHandler.CreateUser).Methods(http.MethodPost, http.MethodOptions)
	userRouter.HandleFunc("/login", userHandler.Login).Methods(http.MethodPost, http.MethodOptions)
	userRouter.HandleFunc("/logout", userHandler.Logout).Methods(http.MethodPost, http.MethodOptions)
	userRouter.HandleFunc("/refresh-token", userHandler.RefreshToken).Methods(http.MethodPost, http.MethodOptions)
	userRouter.HandleFunc("/report", userHandler.CreateReport).Methods(http.MethodPost, http.MethodOptions)
	userRouter.HandleFunc("/autosubscribe", userHandler.UpdateAutoSubscribe).Methods(http.MethodPut, http.MethodOptions)
	userRouter.HandleFunc("/sortfolders", userHandler.UpdateSortFolders).Methods(http.MethodPut, http.MethodOptions)
	userRouter.HandleFunc("/bio", userHandler.UpdateBio).Methods(http.MethodPut, http.MethodOptions)
	userRouter.HandleFunc("/password", userHandler.UpdatePassword).Methods(http.MethodPut, http.MethodOptions)
	userRouter.HandleFunc("/viewtype", userHandler.UpdateViewType).Methods(http.MethodPut, http.MethodOptions)
	userRouter.HandleFunc("/forgotpassword", userHandler.ForgotPassword).Methods(http.MethodPost, http.MethodOptions)
	userRouter.HandleFunc("/password/validatekey", userHandler.ValidatePasswordResetKey).Methods(http.MethodGet, http.MethodOptions)
	userRouter.HandleFunc("/account/confirm", userHandler.ValidateSignupConfirmationKey).Methods(http.MethodGet, http.MethodOptions)

	userRouter.HandleFunc("/discussion/{discussionId}/bookmark", userHandler.DeleteDiscussionBookmark).Methods(http.MethodDelete, http.MethodOptions)

	userRouter.HandleFunc("/ignore/{userId}", userHandler.UpdateIgnore).Methods(http.MethodPut, http.MethodOptions)
	userRouter.HandleFunc("/ignore/list", userHandler.GetIgnoredUsers).Methods(http.MethodGet, http.MethodOptions)

	userRouter.HandleFunc("/subscriptions/check", userHandler.CheckSubscriptions).Methods(http.MethodGet, http.MethodOptions)
	userRouter.HandleFunc("/subscriptions/discussion", userHandler.GetDiscussionSubscriptions).Methods(http.MethodGet, http.MethodOptions)
	userRouter.HandleFunc("/subscriptions/discussion", userHandler.DeleteDiscussionSubscriptions).Methods(http.MethodDelete, http.MethodOptions)
	userRouter.HandleFunc("/subscriptions/folder", userHandler.GetFolderSubscriptions).Methods(http.MethodGet, http.MethodOptions)
	userRouter.HandleFunc("/subscriptions/folder", userHandler.UpdateFolderSubscriptions).Methods(http.MethodPost, http.MethodOptions)
	userRouter.HandleFunc("/subscriptions/folder", userHandler.DeleteFolderSubscriptions).Methods(http.MethodDelete, http.MethodOptions)
	userRouter.HandleFunc("/subscriptions/folder/exceptions", userHandler.GetFolderSubscriptionExceptions).Methods(http.MethodGet, http.MethodOptions)
	userRouter.HandleFunc("/subscriptions/fetchorder", userHandler.UpdateSubscriptionFetchOrder).Methods(http.MethodPut, http.MethodOptions)

}

func (a *App) configureAdminRouter(router *mux.Router) {

	adminHandler := handlers.NewAdminHandler(a.userCache, a.folderCache, a.discussionCache, a.postProcessor)

	adminRouter := router.PathPrefix("/admin").Subrouter().StrictSlash(false)

	adminRouter.HandleFunc("/moderation/queue", adminHandler.GetModerationQueue).Methods(http.MethodGet, http.MethodOptions)

	adminRouter.HandleFunc("/discussion/{discussionId}/report", adminHandler.GetDiscussionReports).Methods(http.MethodGet, http.MethodOptions)
	adminRouter.HandleFunc("/discussion/{discussionId}/comment", adminHandler.GetComments).Methods(http.MethodGet, http.MethodOptions)
	adminRouter.HandleFunc("/discussion/{discussionId}/post/{postId}/report", adminHandler.CreateComment).Methods(http.MethodPost, http.MethodOptions)
	adminRouter.HandleFunc("/discussion/{discussionId}/post/{postId}/delete", adminHandler.DeletePost).Methods(http.MethodPost, http.MethodOptions)
	adminRouter.HandleFunc("/discussion/{discussionId}/post/{postId}/delete", adminHandler.UndeletePost).Methods(http.MethodDelete, http.MethodOptions)

	adminRouter.HandleFunc("/discussion/{discussionId}/lock", adminHandler.LockDiscussion).Methods(http.MethodPost, http.MethodOptions)
	adminRouter.HandleFunc("/discussion/{discussionId}/premoderate", adminHandler.PremoderateDiscussion).Methods(http.MethodPost, http.MethodOptions)
	adminRouter.HandleFunc("/discussion/{discussionId}/delete", adminHandler.DeleteDiscussion).Methods(http.MethodPost, http.MethodOptions)
	adminRouter.HandleFunc("/discussion/{discussionId}/move", adminHandler.MoveDiscussion).Methods(http.MethodPost, http.MethodOptions)
	adminRouter.HandleFunc("/discussion/{discussionId}/user/block/{userId}", adminHandler.BlockUserDiscussion).Methods(http.MethodPost, http.MethodOptions)
	adminRouter.HandleFunc("/discussion/{discussionId}/user/block/{userId}", adminHandler.UnblockUserDiscussion).Methods(http.MethodDelete, http.MethodOptions)
	adminRouter.HandleFunc("/discussion/{discussionId}/user/block", adminHandler.GetBlockedUsers).Methods(http.MethodGet, http.MethodOptions)
	adminRouter.HandleFunc("/discussion/{discussionId}", adminHandler.EraseDiscussion).Methods(http.MethodDelete, http.MethodOptions)

}

func (a *App) Serve() {

	a.postProcessor.Run()

	log.Info("Serving requests...")
	if err := http.ListenAndServe(":8080", a.router); err != nil {
		log.Errorf("%v", err)
		log.Error("HTTP Server terminated")
	}

}

func (a *App) Shutdown() {
	a.postProcessor.Close()
	a.bookmarkProcessor.Close()
	a.mostActiveWorker.Close()
}

func (a *App) ExecuteTestRequest(req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	a.router.ServeHTTP(rr, req)
	return rr
}
