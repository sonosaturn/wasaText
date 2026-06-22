<script>
import { Modal } from 'bootstrap'

export default {
  props: {
    conversations: Array
  },
  emits: ['message-forwarded'],
  data() {
    return {
      modalInstance: null,
      msgToForward: null,
      searchQuery: ''
    }
  }, // Emits an event to say "done"
  computed: {
    filteredConversations() {
      if (!this.searchQuery) return this.conversations
      return this.conversations.filter(c => 
        (c.displayName || c.title).toLowerCase().includes(this.searchQuery.toLowerCase())
      )
    }
  },
  mounted() {
    this.modalInstance = new Modal(this.$refs.modalRef)
  },
  methods: {
    show(msg) {
      this.msgToForward = msg
      this.searchQuery = ''
      this.modalInstance.show()
    },
    hide() {
      this.modalInstance.hide()
    },
    async forwardTo(chat) {
      if (!this.msgToForward) return
      try {
        await this.$axios.post(`/conversations/${chat.id}/forward`, {
          messageId: this.msgToForward.id
        })
        this.hide()
        alert("Messaggio inoltrato!")
        this.$emit('message-forwarded')
      } catch (e) {
        alert("Errore inoltro: " + e.message)
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
          <h5 class="modal-title">Inoltra a...</h5>
          <button type="button" class="btn-close" @click="hide" />
        </div>
        <div class="modal-body">
          <input v-model="searchQuery" type="text" class="form-control mb-3" placeholder="Cerca chat...">
          <div class="list-group">
            <button v-for="chat in filteredConversations" :key="chat.id" class="list-group-item list-group-item-action d-flex align-items-center" @click="forwardTo(chat)">
              <div class="fw-bold">{{ chat.displayName }}</div>
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>