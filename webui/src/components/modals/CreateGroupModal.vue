<script>
import { Modal } from 'bootstrap'

export default {
  emits: ['group-created'],
  data() {
    return {
      modalInstance: null,
      groupName: '',
      searchQuery: '',
      searchResults: [],
      selectedMembers: []
    }
  },
  mounted() {
    this.modalInstance = new Modal(this.$refs.modalRef)
  },
  methods: {
    show() {
      this.groupName = ''
      this.searchQuery = ''
      this.searchResults = []
      this.selectedMembers = []
      this.modalInstance.show()
    },
    hide() {
      this.modalInstance.hide()
    },
    async searchUsers() {
      if (!this.searchQuery) return
      try {
        const res = await this.$axios.get(`/users?q=${this.searchQuery}`)
        const myId = sessionStorage.getItem('myId') // FIX: sessionStorage
        this.searchResults = (res.data.users || []).filter(u => 
          u.id !== myId && !this.selectedMembers.find(m => m.id === u.id)
        )
      } catch (e) { console.error(e) }
    },
    addMember(user) {
      this.selectedMembers.push(user)
      this.searchResults = this.searchResults.filter(u => u.id !== user.id)
    },
    removeMember(userId) {
      this.selectedMembers = this.selectedMembers.filter(m => m.id !== userId)
    },
    async createGroup() {
      if (!this.groupName.trim()) return
      const memberIds = this.selectedMembers.map(u => u.id)
      const myId = sessionStorage.getItem('myId') // FIX: sessionStorage
      if (myId && !memberIds.includes(myId)) {
        memberIds.push(myId)
      }

      try {
        await this.$axios.post('/groups', { name: this.groupName, members: memberIds })
        this.hide()
        this.$emit('group-created')
      } catch (e) {
        alert("Errore creazione gruppo: " + (e.response?.data || e.message))
      }
    }
  }
}
</script>

<template>
  <div ref="modalRef" class="modal fade" tabindex="-1">
    <div class="modal-dialog">
      <div class="modal-content">
        <div class="modal-header">
          <h5 class="modal-title">Nuovo Gruppo</h5>
          <button type="button" class="btn-close" @click="hide" />
        </div>
        <div class="modal-body">
          <input v-model="groupName" type="text" class="form-control mb-3" placeholder="Nome Gruppo">
          
          <div class="input-group mb-2">
            <input v-model="searchQuery" type="text" class="form-control" placeholder="Cerca membri..." @keyup.enter="searchUsers">
            <button class="btn btn-outline-secondary" @click="searchUsers">Cerca</button>
          </div>

          <div v-if="searchResults.length > 0" class="list-group mb-3">
            <button v-for="u in searchResults" :key="u.id" class="list-group-item list-group-item-action d-flex justify-content-between" @click="addMember(u)">
              {{ u.username }} <i class="bi bi-plus-circle text-primary" />
            </button>
          </div>

          <div v-if="selectedMembers.length > 0" class="p-2 border rounded bg-light">
            <span v-for="m in selectedMembers" :key="m.id" class="badge bg-secondary me-1 mb-1">
              {{ m.username }} <i class="bi bi-x-circle-fill" style="cursor: pointer" @click="removeMember(m.id)" />
            </span>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn btn-primary" :disabled="!groupName" @click="createGroup">Crea</button>
        </div>
      </div>
    </div>
  </div>
</template>