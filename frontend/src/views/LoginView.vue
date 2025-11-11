<template>
  <div class="login-container">
    <div class="login-box">
      <h1>Login to LibreCov</h1>
      
      <div v-if="loading" class="loading">
        Loading authentication options...
      </div>
      
      <div v-else-if="authStore.isOIDCEnabled" class="oidc-login">
        <p>This instance is configured to use Single Sign-On (SSO).</p>
        <button @click="loginWithOIDC" class="btn btn-primary btn-large">
          Login with SSO
        </button>
      </div>
      
      <div v-else class="internal-login">
        <p class="info-message">
          OIDC authentication is not configured. Please contact your administrator to set up authentication.
        </p>
        <p class="info-message">
          To enable OIDC authentication, configure the following environment variables:
        </p>
        <ul class="config-list">
          <li>OIDC_ISSUER</li>
          <li>OIDC_CLIENT_ID</li>
          <li>OIDC_REDIRECT_URL</li>
        </ul>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useAuthStore } from '../stores/auth'
import { useRouter } from 'vue-router'

const authStore = useAuthStore()
const router = useRouter()
const loading = ref(true)

onMounted(async () => {
  // Check if already authenticated
  if (authStore.isAuthenticated) {
    router.push('/')
    return
  }

  // Load auth config
  await authStore.loadAuthConfig()
  loading.value = false
})

const loginWithOIDC = () => {
  authStore.loginWithOIDC()
}
</script>

<style scoped>
.login-container {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 60vh;
}

.login-box {
  background: white;
  padding: 2rem;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  max-width: 500px;
  width: 100%;
}

.login-box h1 {
  margin-top: 0;
  margin-bottom: 1.5rem;
  text-align: center;
  color: #2c3e50;
}

.loading {
  text-align: center;
  padding: 2rem;
  color: #7f8c8d;
}

.oidc-login {
  text-align: center;
}

.oidc-login p {
  margin-bottom: 1.5rem;
  color: #555;
}

.internal-login {
  padding: 1rem;
}

.info-message {
  margin-bottom: 1rem;
  color: #555;
  line-height: 1.5;
}

.config-list {
  list-style-position: inside;
  color: #555;
  margin-top: 0.5rem;
}

.config-list li {
  margin: 0.5rem 0;
  font-family: monospace;
  background: #f5f5f5;
  padding: 0.5rem;
  border-radius: 4px;
}

.btn {
  padding: 0.75rem 1.5rem;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 1rem;
  transition: all 0.2s;
}

.btn-primary {
  background: #3498db;
  color: white;
}

.btn-primary:hover {
  background: #2980b9;
}

.btn-large {
  padding: 1rem 2rem;
  font-size: 1.1rem;
}
</style>
