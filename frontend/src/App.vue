<template>
  <div id="app">
    <header class="header">
      <div class="container">
        <h1 class="logo">LibreCov</h1>
        <nav class="nav">
          <router-link to="/" class="nav-link">Projects</router-link>
          <router-link v-if="authStore.isAuthenticated" to="/admin" class="nav-link">Admin</router-link>
          <div v-if="authStore.isAuthenticated" class="user-info">
            <span>{{ authStore.user?.name || authStore.user?.email }}</span>
            <button @click="logout" class="btn btn-secondary">Logout</button>
          </div>
          <div v-else>
            <button @click="login" class="btn btn-primary">Login</button>
          </div>
        </nav>
      </div>
    </header>
    <main class="main">
      <div class="container">
        <router-view />
      </div>
    </main>
    <footer class="footer">
      <div class="container">
        <p>&copy; 2024 LibreCov - Open Source Code Coverage History</p>
      </div>
    </footer>
  </div>
</template>

<script setup lang="ts">
import { useAuthStore } from './stores/auth'

const authStore = useAuthStore()

const login = () => {
  window.location.href = '/auth/login'
}

const logout = async () => {
  await authStore.logout()
}
</script>

<style scoped>
.header {
  background: #2c3e50;
  color: white;
  padding: 1rem 0;
}

.container {
  max-width: 1200px;
  margin: 0 auto;
  padding: 0 1rem;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.logo {
  margin: 0;
  font-size: 1.5rem;
}

.nav {
  display: flex;
  gap: 1rem;
  align-items: center;
}

.nav-link {
  color: white;
  text-decoration: none;
  padding: 0.5rem 1rem;
}

.nav-link:hover {
  background: rgba(255, 255, 255, 0.1);
  border-radius: 4px;
}

.user-info {
  display: flex;
  gap: 1rem;
  align-items: center;
}

.main {
  min-height: calc(100vh - 200px);
  padding: 2rem 0;
}

.footer {
  background: #34495e;
  color: white;
  padding: 1rem 0;
  text-align: center;
}

.btn {
  padding: 0.5rem 1rem;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 0.9rem;
}

.btn-primary {
  background: #3498db;
  color: white;
}

.btn-secondary {
  background: #95a5a6;
  color: white;
}

.btn:hover {
  opacity: 0.9;
}
</style>
