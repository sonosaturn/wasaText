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

	// 1. Conversations
	rt.router.GET("/conversations", rt.requireAuth(rt.getMyConversations))
	rt.router.POST("/conversations", rt.requireAuth(rt.createConversation))

	// 2. Messages
	rt.router.GET("/conversations/:conversationId/messages", rt.requireAuth(rt.getConversationMessages))
	rt.router.POST("/conversations/:conversationId/messages", rt.requireAuth(rt.sendMessage))
	rt.router.DELETE("/conversations/:conversationId/messages/:messageId", rt.requireAuth(rt.deleteMessage))
	rt.router.PUT("/conversations/:conversationId/seen", rt.requireAuth(rt.setConversationSeen))
	rt.router.PUT("/conversations/:conversationId/received", rt.requireAuth(rt.setConversationReceived))
	rt.router.PUT("/conversations/:conversationId/messages/:messageId/reaction", rt.requireAuth(rt.commentMessage))
	rt.router.DELETE("/conversations/:conversationId/messages/:messageId/reaction", rt.requireAuth(rt.uncommentMessage))
	rt.router.POST("/conversations/:conversationId/forward", rt.requireAuth(rt.forwardMessage))
	rt.router.ServeFiles("/images/*filepath", http.Dir("./images"))

	// 3. Group management
	rt.router.PUT("/conversations/:conversationId/members/:userId", rt.requireAuth(rt.addToGroup))
	rt.router.DELETE("/conversations/:conversationId/members/:userId", rt.requireAuth(rt.leaveGroup))
	rt.router.POST("/groups", rt.requireAuth(rt.createGroup))
	rt.router.GET("/conversations/:conversationId/members", rt.requireAuth(rt.getConversationMembers))
	rt.router.PUT("/conversations/:conversationId/name", rt.requireAuth(rt.setGroupName))
	rt.router.PUT("/conversations/:conversationId/photo", rt.requireAuth(rt.setGroupPhoto))

	// 4. Users
	rt.router.GET("/users", rt.requireAuth(rt.getUsers))
	rt.router.PUT("/users/:userId/photo", rt.requireAuth(rt.setMyPhoto))
	rt.router.PUT("/users/:userId/username", rt.wrap(rt.setUsername))

	// Photos (generic upload/download)
	rt.router.POST("/photos", rt.requireAuth(rt.uploadPhoto))
	rt.router.GET("/photos/:photoId", rt.requireAuth(rt.getPhoto))

	// When a request like "/avatars/1.jpg" comes in, call the getAvatar function
	rt.router.GET("/avatars/:filename", rt.wrap(rt.getAvatar))

	return rt.router
}
