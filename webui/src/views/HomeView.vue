<script>
import ChatSidebar from '../components/ChatSidebar.vue'
import ChatArea from '../components/ChatArea.vue'
import CreateGroupModal from '../components/modals/CreateGroupModal.vue'
import ForwardModal from '../components/modals/ForwardModal.vue'
import InfoModal from '../components/modals/InfoModal.vue'
import LoadingSpinner from '../components/LoadingSpinner.vue'

export default {
  components: { 
    ChatSidebar, 
    ChatArea, 
    CreateGroupModal, 
    ForwardModal, 
    InfoModal,
    LoadingSpinner
  },
  data() {
    return {
      // FIX: Use sessionStorage to reset on browser restart
      token: sessionStorage.getItem('token'),
      myId: sessionStorage.getItem('myId'),
      username: sessionStorage.getItem('username'),
      myPhoto: sessionStorage.getItem('myPhoto'),
      conversations: [],
      messages: [],
      selectedChat: null,
      pollInterval: null,
      usersCache: {},
      loading: true
    }
  },
  async mounted() {
    if (!this.token) {
      this.$router.push('/login')
      return
    }
    this.$axios.defaults.headers.common['Authorization'] = `Bearer ${this.token}`
    this.loading = true
    await this.refreshProfile() // This now checks whether you still exist in the DB
    await this.loadConversations()
    this.loading = false
    
    this.pollInterval = setInterval(() => {
      this.loadConversations()
      this.refreshProfile() 
      if (this.selectedChat) this.loadMessages()
    }, 3000)
  },
  beforeUnmount() {
    clearInterval(this.pollInterval)
  },
  methods: {
    // --- API & DATA LOADING ---
    async refreshProfile() {
      try {
        const res = await this.$axios.get(`/users?q=${this.username}`)
        const me = res.data.users?.find(u => u.username === this.username)
        
        // CRITICAL FIX: If the DB has been deleted, the user no longer exists.
        // If 'me' is undefined, log the user out.
        if (!me) {
          this.logout()
          return
        }

        if (this.myId !== me.id || this.myPhoto.split('?')[0] !== me.photo_url) {
          this.myId = me.id
          this.myPhoto = me.photo_url ? `${me.photo_url}?t=${new Date().getTime()}` : '/default.png'
          
          sessionStorage.setItem('myId', me.id)
          sessionStorage.setItem('myPhoto', me.photo_url)
        }
      } catch (e) {
        // If the API fails badly (e.g. 500 or 401), log out
        if (e.response && (e.response.status === 401 || e.response.status === 404)) {
            this.logout()
        }
      }
    },

    async loadConversations() {
      try {
        const res = await this.$axios.get('/conversations')
        
        this.conversations = (res.data.conversations || []).map(c => {
          c.displayName = c.title 
          return c
        })
        
        if (this.selectedChat) {
          const updated = this.conversations.find(c => c.id === this.selectedChat.id)
          if (updated) {
            this.selectedChat = { ...updated }
            this.selectedChat.unread_count = 0
          } else {
            this.selectedChat = null
          }
        }
      } catch (e) {}
    },
    
    // --- CHAT INTERACTION ---
    async selectChat(chat) {
      this.selectedChat = chat
      this.messages = []
      await this.loadMessages()
    },

    async loadMessages() {
      if (!this.selectedChat) return
      try {
        const res = await this.$axios.get(`/conversations/${this.selectedChat.id}/messages`)
        this.messages = res.data.messages || []
        await this.$axios.put(`/conversations/${this.selectedChat.id}/seen`)
      } catch (e) {
        this.selectedChat = null 
      }
    },

    async sendMessage({ content, file, replyTo }) {
      const fd = new FormData()
      fd.append('content', content || '')
      if (file) fd.append('file', file)
      if (replyTo) fd.append('reply_to', replyTo.id)
      
      try {
        await this.$axios.post(`/conversations/${this.selectedChat.id}/messages`, fd)
        await this.loadMessages()
        await this.loadConversations()
      } catch (e) {
        alert('Errore invio: ' + e.message)
      }
    },
    
    // --- PROFILE ACTIONS ---
    async updatePhoto(file) {
      const fd = new FormData()
      fd.append('file', file)
      try {
        await this.$axios.put(`/users/${this.myId}/photo`, fd)
        await this.refreshProfile()
        alert("Foto aggiornata!")
      } catch (e) {
        alert("Errore caricamento foto")
      }
    },

    async updateUsername(newName) {
      if (!newName || newName.length < 3 || newName.length > 16) {
        alert("Il nome deve essere tra 3 e 16 caratteri")
        return
      }
      try {
        await this.$axios.put(`/users/${this.myId}/username`, { username: newName })
        this.username = newName
        sessionStorage.setItem('username', newName)
        await this.refreshProfile()
        alert("Username aggiornato!")
      } catch (e) {
        if (e.response?.status === 409) alert("Nome già in uso")
        else alert("Errore durante il cambio nome")
      }
    },

    // --- SIDEBAR ACTIONS ---
    async searchUser(query) {
       if (!query) return
       try {
         const res = await this.$axios.get(`/users?q=${query}`)
         const target = res.data.users?.find(u => u.username.toLowerCase() === query.toLowerCase() && u.id !== this.myId)
         if (target) {
           const cr = await this.$axios.post('/conversations', { otherUserId: target.id })
           await this.loadConversations()
           const newChat = this.conversations.find(c => c.id === cr.data.conversationId)
           if (newChat) this.selectChat(newChat)
         } else {
           alert("Utente non trovato")
         }
       } catch (e) { alert("Errore ricerca: " + e.message) }
    },

    openCreateGroup() {
      this.$refs.createGroupModal.show()
    },

    logout() {
      // Clears the current session
      sessionStorage.clear()
      this.$router.push('/login')
    },

    openInfo() {
      if (this.selectedChat) this.$refs.infoModal.show()
    },

    openForward(msg) {
      this.$refs.forwardModal.show(msg)
    },

    onChatLeft() {
      this.selectedChat = null
      this.loadConversations()
    }
  }
}
</script>

<template>
  <LoadingSpinner :loading="loading">
    <div class="main-container">
      <ChatSidebar 
        :conversations="conversations" 
        :selected-chat-id="selectedChat?.id"
        :my-photo="myPhoto"
        :username="username"
        :my-id="myId"
        @select-chat="selectChat"
        @search-user="searchUser"
        @create-group="openCreateGroup"
        @logout="logout"
        @update-photo="updatePhoto"
        @update-username="updateUsername"
      />
      
      <ChatArea 
        :chat="selectedChat" 
        :messages="messages" 
        :my-id="myId"
        @send-message="sendMessage"
        @open-info="openInfo"
        @forward-message="openForward"
        @message-deleted="loadMessages" 
      />

      <CreateGroupModal ref="createGroupModal" @group-created="loadConversations" />
      <InfoModal ref="infoModal" :chat="selectedChat" @chat-updated="loadConversations" @chat-left="onChatLeft" />
      <ForwardModal ref="forwardModal" :conversations="conversations" />
    </div>
  </LoadingSpinner>
</template>

<style scoped>
.main-container { 
  display: flex; 
  height: 100vh;
  width: 100vw;
  max-width: 100%; 
  margin: 0; 
  background: white; 
  overflow: hidden;
}

@media (min-width: 1600px) {
  .main-container {
    height: calc(100vh - 40px);
    margin: 20px auto;
    max-width: 1600px;
    box-shadow: 0 4px 12px rgba(0,0,0,0.1);
  }
}
</style>