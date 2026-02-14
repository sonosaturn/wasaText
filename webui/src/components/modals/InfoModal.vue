<script>
import { Modal } from 'bootstrap'

export default {
  props: {
    chat: Object
  },
  emits: ['chat-updated', 'chat-left'],
  data() {
    return {
      modalInstance: null,
      editingName: '',
      members: [],
      addQuery: '',
      addResults: [],
      pollInterval: null
    }
  },
  mounted() {
    this.modalInstance = new Modal(this.$refs.modalRef)
  },
  beforeUnmount() {
    if (this.pollInterval) clearInterval(this.pollInterval)
  },
  methods: {
    getFullUrl(p) {
      if (!p || p === '/default.png' || p === 'null' || p === '') return '/default.png'
      return `${this.$axios.defaults.baseURL}${p}`
    },
    async show() {
      if (this.chat.is_group) {
        this.editingName = this.chat.title
        await this.loadMembers()
        
        if (this.pollInterval) clearInterval(this.pollInterval)
        this.pollInterval = setInterval(() => {
          this.loadMembers()
        }, 3000)

      } else {
        this.editingName = this.chat.displayName
      }
      this.addQuery = ''
      this.addResults = []
      this.modalInstance.show()
    },
    hide() {
      if (this.pollInterval) clearInterval(this.pollInterval)
      this.pollInterval = null
      this.modalInstance.hide()
    },
    async loadMembers() {
      try {
        const res = await this.$axios.get(`/conversations/${this.chat.id}/members`)
        this.members = res.data.members || []
      } catch (e) {}
    },
    async updateName() {
      if (!this.chat.is_group) return
      try {
        await this.$axios.put(`/conversations/${this.chat.id}/name`, { name: this.editingName })
        this.$emit('chat-updated') 
      } catch (e) { alert(e.message) }
    },
    async uploadPhoto() {
      const file = this.$refs.photoInput.files[0]
      if (!file) return
      const fd = new FormData()
      fd.append('file', file)
      try {
        await this.$axios.put(`/conversations/${this.chat.id}/photo`, fd)
        this.$emit('chat-updated')
      } catch (e) { alert(e.message) }
    },
    async searchToAdd() {
      if (!this.addQuery) return
      try {
        const res = await this.$axios.get(`/users?q=${this.addQuery}`)
        this.addResults = (res.data.users || []).filter(u => !this.members.find(m => m.id === u.id))
      } catch (e) {}
    },
    async addMember(user) {
      try {
        await this.$axios.put(`/conversations/${this.chat.id}/members/${user.id}`)
        await this.loadMembers()
        this.addResults = []
        this.addQuery = ''
      } catch (e) { alert(e.message) }
    },
    async leaveChat() {
      if (!confirm("Sei sicuro?")) return
      try {
        const myId = sessionStorage.getItem('myId') // FIX: sessionStorage
        await this.$axios.delete(`/conversations/${this.chat.id}/members/${myId}`)
        this.hide()
        this.$emit('chat-left')
      } catch (e) { alert(e.message) }
    }
  }
}
</script>

<template>
  <div ref="modalRef" class="modal fade" tabindex="-1">
    <div class="modal-dialog">
      <div v-if="chat" class="modal-content">
        <div class="modal-header">
          <h5 class="modal-title">Info {{ chat.is_group ? 'Gruppo' : 'Contatto' }}</h5>
          <button type="button" class="btn-close" @click="hide" />
        </div>
        <div class="modal-body text-center">
          <div class="position-relative d-inline-block mb-3" :style="{cursor: chat.is_group ? 'pointer' : 'default'}" @click="chat.is_group ? $refs.photoInput.click() : null">
            <img :src="getFullUrl(chat.photo_url)" class="avatar" style="width: 100px; height: 100px; object-fit: cover; border-radius: 50%">
            <input v-if="chat.is_group" ref="photoInput" type="file" class="d-none" accept="image/*" @change="uploadPhoto">
          </div>

          <div v-if="chat.is_group">
            <input v-model="editingName" type="text" class="form-control text-center fw-bold mb-3" @blur="updateName" @keyup.enter="updateName">
          </div>
          <h4 v-else>{{ chat.displayName }}</h4>

          <div v-if="chat.is_group" class="text-start">
            <h6 class="text-muted">Partecipanti ({{ members.length }})</h6>
            <ul class="list-group mb-3" style="max-height: 200px; overflow-y: auto;">
              <li v-for="m in members" :key="m.id" class="list-group-item d-flex align-items-center">
                <img :src="getFullUrl(m.photo_url)" style="width: 30px; height: 30px; border-radius: 50%; margin-right: 10px;">
                {{ m.username }}
              </li>
            </ul>
            
            <div class="input-group mb-2">
              <input v-model="addQuery" type="text" class="form-control form-control-sm" placeholder="Aggiungi utente..." @keyup.enter="searchToAdd">
              <button class="btn btn-outline-secondary btn-sm" @click="searchToAdd">Cerca</button>
            </div>
            <div v-if="addResults.length > 0" class="list-group">
              <button v-for="u in addResults" :key="u.id" class="list-group-item list-group-item-action" @click="addMember(u)">
                {{ u.username }} <i class="bi bi-plus text-success" />
              </button>
            </div>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn btn-danger w-100" @click="leaveChat">
            {{ chat.is_group ? 'Abbandona Gruppo' : 'Elimina Chat' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>