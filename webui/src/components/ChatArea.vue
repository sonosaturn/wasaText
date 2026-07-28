<script>
export default {
  props: {
    chat: Object,
    messages: Array,
    myId: String
  },
  emits: ['send-message', 'open-info', 'forward-message', 'message-deleted'], // Added 'message-deleted'
  data() {
    return {
      newMessage: '',
      replyingTo: null,
      showReactionsId: null,
      selectedFile: null
    }
  },
  watch: {
    messages: {
      handler(newVal, oldVal) {
        if (oldVal && newVal.length === oldVal.length) return

        this.$nextTick(() => {
          const box = this.$refs.msgBox
          if (!box) return
          const distanceToBottom = box.scrollHeight - box.scrollTop - box.clientHeight
          const isUserNearBottom = distanceToBottom < 150 || !oldVal || oldVal.length === 0
          if (isUserNearBottom) {
            box.scrollTop = box.scrollHeight
          }
        })
      },
      deep: true
    }
  },
  methods: {
    getFullUrl(p) {
      if (!p || p === '/default.png' || p === 'null' || p === '') return '/default.png'
      if (p.startsWith('http')) return p
      const path = p.startsWith('/') ? p : '/' + p
      return path
    },
    handleImgError(e) {
      if (!e.target.src.includes('default.png')) e.target.src = '/default.png'
    },
    formatTime(iso) {
      return iso ? new Date(iso).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' }) : ''
    },
    
    // --- FILE HANDLING ---
    onFileSelected(e) {
      const file = e.target.files[0]
      if (file) {
        this.selectedFile = file
      }
      e.target.value = ''
    },
    removeFile() {
      this.selectedFile = null
    },

    // --- SENDING ---
    send() {
      if (!this.newMessage.trim() && !this.selectedFile) return

      this.$emit('send-message', { 
        content: this.newMessage, 
        replyTo: this.replyingTo,
        file: this.selectedFile
      })
      
      this.newMessage = ''
      this.replyingTo = null
      this.selectedFile = null
      
      this.scrollToBottom(true)
    },

    // --- DELETION (NEW) ---
    async deleteMsg(msg) {
      if (!confirm("Delete this message?")) return
      try {
        await this.$axios.delete(`/conversations/${this.chat.id}/messages/${msg.id}`)
        // Emit event to tell Home to reload the messages
        this.$emit('message-deleted')
      } catch (e) {
        alert("Delete failed: " + e.message)
      }
    },

    // --- REPLIES ---
    startReply(msg) {
      this.replyingTo = msg
      this.$refs.inputField?.focus()
    },
    cancelReply() {
      this.replyingTo = null
    },
    getReplyContent(replyId) {
      const original = this.messages.find(m => m.id === replyId)
      if (!original) return { text: "Message unavailable", hasPhoto: false }
      return { text: original.content || '', hasPhoto: !!original.photo_url }
    },

    // --- REACTIONS ---
    toggleReactionMenu(msgId) {
      this.showReactionsId = this.showReactionsId === msgId ? null : msgId
    },
    async addReaction(msgId, emoji) {
      try {
        const msg = this.messages.find(m => m.id === msgId)
        const myReaction = msg.reactions?.find(r => r.user_id === this.myId && r.emoji === emoji)
        if (myReaction) {
          // removes the existing reaction (uncommentMessage)
          await this.$axios.delete(`/conversations/${this.chat.id}/messages/${msgId}/reaction`)
        } else {
          // adds/updates the reaction (commentMessage)
          await this.$axios.put(`/conversations/${this.chat.id}/messages/${msgId}/reaction`, { emoji })
        }
        this.showReactionsId = null
      } catch (e) {
        alert("Reaction failed: " + e.message)
      }
    },

    // --- HELPERS ---
    scrollToBottom(force = false) {
      this.$nextTick(() => {
        const box = this.$refs.msgBox
        if (box) box.scrollTop = box.scrollHeight
      })
    }
  }
}
</script>

<template>
  <div class="chat-area-container">
    <div v-if="chat" class="h-100 d-flex flex-column">
      <div class="chat-header">
        <img :src="getFullUrl(chat.photo_url)" class="avatar-sm" @error="handleImgError">
        <div class="flex-grow-1 ms-3" style="cursor: pointer;" @click="$emit('open-info')">
          <span class="fw-bold d-block">{{ chat.displayName }}</span>
          <small class="text-muted">click for info</small>
        </div>
      </div>

      <div ref="msgBox" class="messages-box">
        <div v-for="msg in messages" :key="msg.id" class="msg-row" :class="msg.senderId === myId ? 'mine' : ''">
          <div class="msg" :class="msg.senderId === myId ? 'msg-out' : 'msg-in'">
            <div v-if="chat.is_group && msg.senderId !== myId" class="sender-name">
              {{ msg.sender_username }}
            </div>

            <div v-if="msg.forwarded" class="forwarded-label">
              <i class="bi bi-forward" /> Forwarded
            </div>

            <div v-if="msg.reply_to_id" class="reply-snippet">
              <div class="reply-bar" />
              <small class="text-muted d-block">Replying to:</small>
              <div class="text-truncate small d-flex align-items-center gap-1">
                <i v-if="getReplyContent(msg.reply_to_id).hasPhoto" class="bi bi-image text-secondary" />
                <span>{{ getReplyContent(msg.reply_to_id).text || (getReplyContent(msg.reply_to_id).hasPhoto ? '' : '...') }}</span>
              </div>
            </div>

            <div v-if="msg.photo_url" class="mb-1">
              <img :src="getFullUrl(msg.photo_url)" class="msg-image">
            </div>
            
            <div class="msg-content">{{ msg.content }}</div>
            
            <div class="msg-footer">
              <span class="msg-time">{{ formatTime(msg.timestamp) }}</span>
              <span v-if="msg.senderId === myId" class="msg-status">
                <i v-if="msg.status === 2" class="bi bi-check2-all text-info" />
                <i v-else-if="msg.status === 1" class="bi bi-check2-all" />
                <i v-else class="bi bi-check2" />
              </span>
            </div>

            <div v-if="msg.reactions && msg.reactions.length > 0" class="reactions-display">
              <span v-for="(r, idx) in msg.reactions" :key="idx" class="badge bg-light text-dark border me-1" :title="r.username">
                {{ r.emoji }}
              </span>
            </div>

            <div class="msg-actions">
              <button title="React" @click="toggleReactionMenu(msg.id)">😊</button>
              <button title="Reply" @click="startReply(msg)"><i class="bi bi-reply" /></button>
              <button title="Forward" @click="$emit('forward-message', msg)"><i class="bi bi-forward" /></button>
              <button v-if="msg.senderId === myId" title="Delete" class="text-danger" @click="deleteMsg(msg)">
                <i class="bi bi-trash" />
              </button>
            </div>

            <div v-if="showReactionsId === msg.id" class="reaction-picker">
              <span @click="addReaction(msg.id, '👍')">👍</span>
              <span @click="addReaction(msg.id, '😂')">😂</span>
              <span @click="addReaction(msg.id, '❤️')">❤️</span>
              <span @click="addReaction(msg.id, '😮')">😮</span>
              <span @click="addReaction(msg.id, '😢')">😢</span>
              <span @click="addReaction(msg.id, '🙏')">🙏</span>
            </div>
          </div>
        </div>
      </div>

      <div class="input-container">
        <div v-if="replyingTo" class="reply-banner">
          <div class="d-flex justify-content-between align-items-center">
            <div>
              <span class="text-primary small">Replying to...</span>
              <div class="small text-truncate text-muted" style="max-width: 300px;">
                {{ replyingTo.content || '[Photo]' }}
              </div>
            </div>
            <button class="btn btn-sm btn-link text-danger" @click="cancelReply"><i class="bi bi-x-lg" /></button>
          </div>
        </div>

        <div v-if="selectedFile" class="file-banner">
          <div class="d-flex justify-content-between align-items-center">
            <div>
              <i class="bi bi-image text-success me-2" />
              <span class="small fw-bold">{{ selectedFile.name }}</span>
            </div>
            <button class="btn btn-sm btn-link text-danger" @click="removeFile"><i class="bi bi-x-lg" /></button>
          </div>
        </div>

        <div class="message-input-area">
          <input ref="fileInput" type="file" class="d-none" accept="image/*" @change="onFileSelected">
          <button class="btn btn-link text-secondary" @click="$refs.fileInput.click()">
            <i class="bi bi-paperclip fs-5" />
          </button>
          <input ref="inputField" v-model="newMessage" type="text" class="form-control shadow-none" placeholder="Type a message..." @keyup.enter="send">
          <button class="btn btn-link text-muted" @click="send"><i class="bi bi-send-fill fs-5" /></button>
        </div>
      </div>
    </div>

    <div v-else class="empty-state">
      <div class="text-center">
        <i class="bi bi-chat-left-text fs-1 mb-3 d-block" />
        <h3>WasaText</h3>
        <p>Select a conversation to start</p>
      </div>
    </div>
  </div>
</template>

<style scoped>
.chat-area-container { flex: 1; display: flex; flex-direction: column; background: #efeae2; height: 100%; overflow: hidden; }
.chat-header { padding: 10px 16px; background: #f0f2f5; display: flex; align-items: center; height: 60px; border-bottom: 1px solid #e9edef; flex-shrink: 0; }
.messages-box { flex: 1; overflow-y: auto; padding: 20px 5%; display: flex; flex-direction: column; gap: 4px; }
.msg-row { display: flex; margin-bottom: 2px; width: 100%; position: relative; }
.msg-row.mine { justify-content: flex-end; }
.msg { 
  max-width: 75%; 
  padding: 6px 10px; border-radius: 8px; font-size: 0.9rem; 
  box-shadow: 0 1px 1px rgba(0,0,0,0.1); position: relative; 
}
.msg:hover .msg-actions { display: flex; }
.msg-in { background: white; border-top-left-radius: 0; }
.msg-out { background: #d9fdd3; border-top-right-radius: 0; }

.sender-name { font-size: 0.75rem; font-weight: bold; color: #d64f00; margin-bottom: 2px; }
.forwarded-label { font-size: 0.75rem; color: #667781; font-style: italic; margin-bottom: 2px; display: flex; align-items: center; gap: 4px; }

.msg-image { 
  max-width: 450px; 
  max-height: 500px;
  width: auto;
  height: auto;
  object-fit: cover;
  border-radius: 4px; 
  margin-bottom: 5px;
  cursor: pointer;
}

.msg-footer { display: flex; align-items: center; justify-content: flex-end; gap: 4px; font-size: 0.7rem; color: #667781; margin-top: 2px; }
.avatar-sm { width: 40px; height: 40px; border-radius: 50%; object-fit: cover; }
.empty-state { flex: 1; display: flex; align-items: center; justify-content: center; color: #8696a0; }

.input-container { background: #f0f2f5; flex-shrink: 0; }
.message-input-area { padding: 10px 16px; display: flex; align-items: center; gap: 10px; }

/* Banner */
.reply-banner { padding: 5px 15px; background: #e9edef; border-left: 5px solid #00a884; }
.file-banner { padding: 5px 15px; background: #e9edef; border-left: 5px solid #0d6efd; border-top: 1px solid #ddd; }
.reply-snippet { background: rgba(0,0,0,0.05); border-radius: 4px; padding: 5px; margin-bottom: 5px; position: relative; overflow: hidden; }
.reply-bar { position: absolute; left: 0; top: 0; bottom: 0; width: 4px; background: #00a884; }

/* Reactions */
.msg-actions {
  display: none; position: absolute; top: -15px; right: 10px; 
  background: white; border-radius: 15px; box-shadow: 0 1px 3px rgba(0,0,0,0.2);
  padding: 2px 5px; z-index: 100;
}
.msg-actions button { border: none; background: none; font-size: 1rem; cursor: pointer; padding: 0 5px; }
.msg-actions button:hover { color: #00a884; }
.reaction-picker {
  position: absolute; bottom: 100%; right: 0; background: white; 
  border-radius: 20px; padding: 5px 10px; box-shadow: 0 2px 5px rgba(0,0,0,0.2);
  display: flex; gap: 5px; z-index: 101;
}
.reaction-picker span { cursor: pointer; font-size: 1.2rem; transition: transform 0.1s; }
.reaction-picker span:hover { transform: scale(1.3); }
.reactions-display { margin-top: -5px; margin-bottom: 5px; position: relative; bottom: -8px; }
.reactions-display span { cursor: default; user-select: none; }
</style>