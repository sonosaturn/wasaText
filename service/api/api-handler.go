package api

import (
	"net/http"
)

// Handler returns an instance of httprouter.Router that handle APIs registered here
func (rt *_router) Handler() http.Handler {
	// --- Public routes ---
	rt.router.GET("/liveness", rt.liveness)
	rt.router.POST("/session", rt.wrap(rt.doLogin))

	// --- Protected routes (Bearer auth) ---

	// 1. Conversazioni
	rt.router.GET("/conversations", rt.requireAuth(rt.getMyConversations))
	rt.router.POST("/conversations", rt.requireAuth(rt.createConversation))

	// 2. Messaggi
	rt.router.GET("/conversations/:conversationId/messages", rt.requireAuth(rt.getConversationMessages))
	rt.router.POST("/conversations/:conversationId/messages", rt.requireAuth(rt.sendMessage))
	rt.router.DELETE("/conversations/:conversationId/messages/:messageId", rt.requireAuth(rt.deleteMessage))
	rt.router.PUT("/conversations/:conversationId/read", rt.requireAuth(rt.markAsRead))
	rt.router.PUT("/conversations/:conversationId/messages/:messageId/reaction", rt.requireAuth(rt.reactToMessage))
	rt.router.POST("/conversations/:conversationId/forward", rt.requireAuth(rt.forwardMessage))
	rt.router.ServeFiles("/images/*filepath", http.Dir("./images"))

	// 3. Gestione Gruppi
	rt.router.PUT("/conversations/:conversationId/members/:userId", rt.requireAuth(rt.addToGroup))
	rt.router.DELETE("/conversations/:conversationId/members/:userId", rt.requireAuth(rt.leaveGroup))
	rt.router.POST("/groups", rt.requireAuth(rt.createGroup))
	rt.router.GET("/conversations/:conversationId/members", rt.requireAuth(rt.getConversationMembers))
	rt.router.PUT("/conversations/:conversationId/name", rt.requireAuth(rt.setGroupName))
	rt.router.PUT("/conversations/:conversationId/photo", rt.requireAuth(rt.setGroupPhoto))

	// 4. Utenti
	rt.router.GET("/users", rt.requireAuth(rt.searchUsers))
	rt.router.PUT("/users/:userId/photo", rt.requireAuth(rt.setMyPhoto))
	rt.router.GET("/users/:userId", rt.requireAuth(rt.getUserById))
	rt.router.PUT("/users/:userId/username", rt.wrap(rt.setUsername))

	// Quando arriva una richiesta tipo "/avatars/1.jpg", chiama la funzione getAvatar
	rt.router.GET("/avatars/:filename", rt.wrap(rt.getAvatar))

	return rt.router
}
