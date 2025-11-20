<template>
  <div id="app">
    <!-- Mobile menu backdrop -->
    <div 
      v-if="mobileMenuOpen" 
      class="mobile-backdrop" 
      @click="mobileMenuOpen = false"
      aria-hidden="true"
    ></div>
    
    <header class="header">
      <div class="container">
        <h1 class="logo">LibreCov</h1>
        <button class="mobile-menu-toggle" @click="mobileMenuOpen = !mobileMenuOpen" aria-label="Toggle menu">
          <span class="hamburger-icon" :class="{ open: mobileMenuOpen }"></span>
        </button>
        <nav class="nav" :class="{ 'mobile-open': mobileMenuOpen }">
          <router-link to="/" class="nav-link" @click="mobileMenuOpen = false">Projects</router-link>
          <router-link v-if="authStore.isAuthenticated" to="/tokens" class="nav-link" @click="mobileMenuOpen = false">API Tokens</router-link>
          <router-link v-if="authStore.isAuthenticated && authStore.user?.admin" to="/admin" class="nav-link" @click="mobileMenuOpen = false">Admin</router-link>
          <div v-if="authStore.isAuthenticated" class="user-info">
            <span class="user-name">{{ authStore.user?.name || authStore.user?.email }}</span>
            <button @click="logout" class="btn btn-secondary">Logout</button>
          </div>
          <div v-else class="auth-actions">
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
      <div class="container footer-content">
        <p class="footer-text">&copy; 2024 LibreCov - Open Source Code Coverage History</p>
        <div class="footer-links">
          <a href="/swagger/index.html" target="_blank" rel="noopener noreferrer" class="footer-link">
            📚 API Documentation
          </a>
        </div>
      </div>
    </footer>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { useAuthStore } from './stores/auth'
import { useRouter } from 'vue-router'

const authStore = useAuthStore()
const router = useRouter()
const mobileMenuOpen = ref(false)

const login = () => {
  mobileMenuOpen.value = false
  // Check if OIDC is enabled and redirect accordingly
  if (authStore.isOIDCEnabled) {
    // Redirect to backend OIDC login endpoint
    window.location.href = '/auth/login'
  } else {
    // Navigate to login page
    router.push('/login')
  }
}

const logout = async () => {
  mobileMenuOpen.value = false
  await authStore.logout()
}

// Handle Escape key to close mobile menu for accessibility
const handleEscapeKey = (event: KeyboardEvent) => {
  if (event.key === 'Escape' && mobileMenuOpen.value) {
    mobileMenuOpen.value = false
  }
}

onMounted(() => {
  document.addEventListener('keydown', handleEscapeKey)
})

onUnmounted(() => {
  document.removeEventListener('keydown', handleEscapeKey)
})
</script>

<style scoped>
.header {
  background: #2c3e50;
  color: white;
  padding: 1rem 0;
  position: sticky;
  top: 0;
  z-index: 300;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
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
  font-weight: 600;
  flex-shrink: 0;
}

.mobile-menu-toggle {
  display: none;
  background: none;
  border: none;
  cursor: pointer;
  padding: 0.5rem;
  z-index: 250;
  position: relative;
}

.hamburger-icon {
  display: block;
  width: 24px;
  height: 2px;
  background: white;
  position: relative;
  transition: background 0.3s;
}

.hamburger-icon::before,
.hamburger-icon::after {
  content: '';
  position: absolute;
  width: 24px;
  height: 2px;
  background: white;
  transition: transform 0.3s;
}

.hamburger-icon::before {
  top: -8px;
}

.hamburger-icon::after {
  top: 8px;
}

.hamburger-icon.open {
  background: transparent;
}

.hamburger-icon.open::before {
  transform: rotate(45deg) translate(5px, 6px);
}

.hamburger-icon.open::after {
  transform: rotate(-45deg) translate(5px, -6px);
}

.nav {
  display: flex;
  gap: 1rem;
  align-items: center;
  flex-wrap: wrap;
}

.nav-link {
  color: white;
  text-decoration: none;
  padding: 0.5rem 1rem;
  border-radius: 4px;
  transition: background 0.2s;
  white-space: nowrap;
}

.nav-link:hover,
.nav-link.router-link-active {
  background: rgba(255, 255, 255, 0.1);
}

.user-info {
  display: flex;
  gap: 1rem;
  align-items: center;
  flex-wrap: wrap;
}

.user-name {
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  max-width: 200px;
}

.auth-actions {
  display: flex;
  gap: 0.5rem;
}

.main {
  min-height: calc(100vh - 200px);
  padding: 2rem 0;
}

.footer {
  background: #34495e;
  color: white;
  padding: 1.5rem 0;
  margin-top: auto;
}

.footer-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 1rem;
}

.footer-text {
  margin: 0;
}

.footer-links {
  display: flex;
  gap: 1.5rem;
  align-items: center;
}

.footer-link {
  color: white;
  text-decoration: none;
  padding: 0.5rem 1rem;
  border-radius: 4px;
  background: rgba(255, 255, 255, 0.1);
  transition: background 0.2s;
  white-space: nowrap;
  font-size: 0.9rem;
}

.footer-link:hover {
  background: rgba(255, 255, 255, 0.2);
}

.btn {
  padding: 0.5rem 1rem;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 0.9rem;
  transition: opacity 0.2s;
  white-space: nowrap;
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

/* Tablet breakpoint */
@media (max-width: 768px) {
  .logo {
    font-size: 1.25rem;
  }
  
  .nav {
    gap: 0.75rem;
  }
  
  .nav-link {
    padding: 0.4rem 0.75rem;
    font-size: 0.9rem;
  }
  
  .user-name {
    max-width: 150px;
  }
  
  .footer-content {
    flex-direction: column;
    text-align: center;
  }
  
  .footer-text {
    font-size: 0.9rem;
  }
}

/* Mobile breakpoint */
@media (max-width: 640px) {
  .mobile-menu-toggle {
    display: block;
  }
  
  .mobile-backdrop {
    position: fixed;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    background: rgba(0, 0, 0, 0.5);
    z-index: 150;
    transition: opacity 0.3s;
  }
  
  .nav {
    position: fixed;
    top: 0;
    right: -100%;
    height: 100vh;
    width: 250px;
    background: #2c3e50;
    flex-direction: column;
    align-items: stretch;
    padding: 5rem 1rem 1rem;
    transition: right 0.3s;
    box-shadow: -2px 0 5px rgba(0, 0, 0, 0.2);
    overflow-y: auto;
    z-index: 200;
  }
  
  .nav.mobile-open {
    right: 0;
  }
  
  .nav-link {
    padding: 0.75rem 1rem;
    border-radius: 0;
    border-bottom: 1px solid rgba(255, 255, 255, 0.1);
  }
  
  .user-info {
    flex-direction: column;
    align-items: stretch;
    padding: 1rem 0;
    border-top: 1px solid rgba(255, 255, 255, 0.1);
    gap: 0.75rem;
  }
  
  .user-name {
    max-width: none;
    padding: 0 1rem;
  }
  
  .auth-actions {
    flex-direction: column;
    padding: 1rem 0;
  }
  
  .btn {
    width: 100%;
  }
  
  .main {
    padding: 1.5rem 0;
  }
  
  .container {
    padding: 0 0.75rem;
  }
  
  .footer-link {
    font-size: 0.85rem;
    padding: 0.4rem 0.75rem;
  }
}

/* Small mobile breakpoint */
@media (max-width: 375px) {
  .logo {
    font-size: 1.1rem;
  }
  
  .footer-text {
    font-size: 0.8rem;
  }
  
  .footer-link {
    font-size: 0.8rem;
  }
}
</style>
