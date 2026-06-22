<script>
import ErrorMsg from '../components/ErrorMsg.vue'

export default {
  components: { ErrorMsg },
  data() {
    return {
      username: '',
      errorMsg: ''
    }
  },
  methods: {
    async doLogin() {
      if (this.username.length < 3) return
      try {
        const res = await this.$axios.post('/session', { name: this.username })
        
        // USE SESSION STORAGE (Cleared when the browser closes)
        sessionStorage.setItem('token', res.data.identifier)
        sessionStorage.setItem('username', this.username)
        
        this.$router.push('/')
      } catch (e) {
        this.errorMsg = e.response?.data || e.message
      }
    }
  }
}
</script>

<template>
  <div class="login-wrapper">
    <div class="login-container">
      <h2 class="mb-4">WasaText</h2>
      <ErrorMsg v-if="errorMsg" :msg="errorMsg" />
      <input v-model="username" type="text" class="form-control mb-3" placeholder="Nome utente (min 3 caratteri)" @keyup.enter="doLogin">
      <button class="btn btn-success w-100" :disabled="username.length < 3" @click="doLogin">ENTRA</button>
    </div>
  </div>
</template>

<style scoped>
.login-wrapper { height: 100vh; display: flex; align-items: center; justify-content: center; background: #d1d7db; }
.login-container { width: 100%; max-width: 450px; padding: 40px; background: white; box-shadow: 0 4px 12px rgba(0,0,0,0.1); text-align: center; border-radius: 8px; }
</style>