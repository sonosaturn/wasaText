<script>
export default {
  props: {
    conversations: Array,
    selectedChatId: String,
    myPhoto: String,
    username: String,
    myId: String
  },
  emits: ['select-chat', 'search-user', 'create-group', 'logout', 'update-photo', 'update-username'],
  data() {
    return {
      searchQuery: '',
      isEditingName: false,
      tempUsername: ''
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
    startEdit() {
      this.tempUsername = this.username
      this.isEditingName = true
      // Focus automatico sull'input quando appare
      this.$nextTick(() => {
        const input = this.$refs.nameInput
        if (input) input.focus()
      })
    },
    saveName() {
      // FIX 1: Se non stiamo più editando (es. invio ha già chiuso), esci subito.
      // Questo blocca il secondo evento (blur) generato dall'apertura dell'alert.
      if (!this.isEditingName) return

      if (this.tempUsername === this.username || !this.tempUsername.trim()) {
        this.isEditingName = false
        return
      }

      // FIX 2: Chiudiamo l'editing PRIMA di emettere l'evento.
      // Così l'input sparisce dal DOM e non può generare altri eventi.
      this.isEditingName = false
      
      this.$emit('update-username', this.tempUsername)
    },
    onFileChange(e) {
      const file = e.target.files[0]
      if (file) this.$emit('update-photo', file)
    }
  }
}
</script>

<template>
  <div class="sidebar-container">
    <div class="sidebar-header">
      <div class="position-relative" style="cursor: pointer;" @click="$refs.fileInput.click()">
        <img :src="getFullUrl(myPhoto)" class="avatar" alt="Me" @error="handleImgError">
        <input ref="fileInput" type="file" class="d-none" accept="image/*" @change="onFileChange">
      </div>
      <div class="flex-grow-1 ms-3 d-flex align-items-center">
        <div v-if="!isEditingName" class="d-flex align-items-center">
          <span class="fw-bold me-2">{{ username }}</span>
          <i class="bi bi-pencil-fill text-muted small" style="cursor: pointer" @click="startEdit" />
        </div>
        <input v-else ref="nameInput" v-model="tempUsername" type="text" class="form-control form-control-sm" @keyup.enter="saveName" @blur="saveName">
      </div>
      <button class="btn btn-sm btn-link text-muted" @click="$emit('logout')"><i class="bi bi-box-arrow-right fs-5" /></button>
    </div>

    <div class="search-box">
      <div class="input-group mb-2">
        <input v-model="searchQuery" type="text" class="form-control shadow-none" placeholder="Cerca utente..." @keyup.enter="$emit('search-user', searchQuery)">
        <button class="btn btn-light border" @click="$emit('search-user', searchQuery)"><i class="bi bi-chat-dots" /></button>
      </div>
      <button class="btn btn-outline-primary w-100 btn-sm" @click="$emit('create-group')">Nuovo Gruppo</button>
    </div>

    <div class="chat-list">
      <div v-for="chat in conversations" :key="chat.id" class="chat-item" :class="{ active: selectedChatId === chat.id }" @click="$emit('select-chat', chat)">
        <img :src="getFullUrl(chat.photo_url)" class="avatar" alt="Chat" @error="handleImgError">
        <div class="flex-grow-1 overflow-hidden">
          <div class="d-flex justify-content-between">
            <span class="fw-bold text-truncate">{{ chat.displayName }}</span>
            <span v-if="chat.last_message_at" class="small text-muted">{{ formatTime(chat.last_message_at) }}</span>
          </div>
          <div class="d-flex justify-content-between align-items-center">
            <small class="text-muted text-truncate d-flex align-items-center">
              
              <span v-if="chat.last_message_sender_id == myId" class="d-flex align-items-center me-1">
                <span class="me-1">Tu:</span>
                <i v-if="chat.last_message_status === 2" class="bi bi-check2-all text-info" />
                <i v-else-if="chat.last_message_status === 1" class="bi bi-check2-all" />
                <i v-else class="bi bi-check2" />
              </span>

              <span>{{ chat.last_message_preview || '...' }}</span>
            </small>
            
            <div v-if="chat.unread_count > 0 && selectedChatId !== chat.id && chat.last_message_sender_id != myId" class="unread-badge">
              {{ chat.unread_count }}
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.sidebar-container { width: 350px; border-right: 1px solid #e9edef; display: flex; flex-direction: column; background: #fff; height: 100%; position: relative; z-index: 10; }
.sidebar-header { padding: 10px 16px; background: #f0f2f5; display: flex; align-items: center; height: 60px; flex-shrink: 0; }
.avatar { width: 40px; height: 40px; border-radius: 50%; object-fit: cover; }
.search-box { padding: 10px; background: #fff; border-bottom: 1px solid #f0f2f5; flex-shrink: 0; }
.chat-list { flex: 1; overflow-y: auto; }
.chat-item { padding: 12px 15px; cursor: pointer; display: flex; align-items: center; gap: 12px; border-bottom: 1px solid #f5f6f6; transition: background 0.2s; }
.chat-item:hover { background-color: #f5f6f6; }
.chat-item.active { background-color: #f0f2f5; }
.unread-badge { background-color: #00a884; color: white; border-radius: 50%; min-width: 20px; height: 20px; font-size: 0.7rem; display: flex; align-items: center; justify-content: center; padding: 0 5px; }
</style>