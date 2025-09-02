<template>
  <div class="settings-page">
    <header class="settings-header">
      <button class="back-btn" @click="$router.push('/dashboard')">
        <svg viewBox="0 0 24 24" width="20" height="20">
          <path fill="currentColor" d="M20 11H7.83l5.59-5.59L12 4l-8 8 8 8 1.42-1.41L7.83 13H20v-2z"/>
        </svg>
        Back
      </button>
      <h1 class="settings-title">Settings</h1>
    </header>

    <main class="settings-content">
      <!-- Account Information -->
      <section class="settings-section">
        <h2 class="section-title">Account Information</h2>
        <div class="account-card">
          <div class="account-avatar">
            <div class="avatar-circle">
              {{ userInitials }}
            </div>
          </div>
          <div class="account-info">
            <div class="info-row">
              <label>Username:</label>
              <span>{{ user.username || 'Not specified' }}</span>
            </div>
            <div class="info-row">
              <label>Email:</label>
              <span>{{ user.email || 'Not specified' }}</span>
            </div>
            <div class="info-row">
              <label>Registration date:</label>
              <span>{{ formatDate(user.created_at) }}</span>
            </div>
            <div class="info-row">
              <label>User ID:</label>
              <span class="user-id">{{ user.id }}</span>
            </div>
          </div>
        </div>
      </section>

      <!-- Statistics -->
      <section class="settings-section">
        <h2 class="section-title">Statistics</h2>
        <div class="stats-grid">
          <div class="stat-card">
            <div class="stat-number">{{ notesCount }}</div>
            <div class="stat-label">Notes</div>
          </div>
          <div class="stat-card">
            <div class="stat-number">{{ materialsCount }}</div>
            <div class="stat-label">Materials</div>
          </div>
        </div>
      </section>

      <!-- Actions -->
      <section class="settings-section">
        <h2 class="section-title">Actions</h2>
        <div class="actions-list">
          <button class="action-btn logout-btn" @click="showLogoutModal = true">
            <svg viewBox="0 0 24 24" width="20" height="20">
              <path fill="currentColor" d="M17 7l-1.41 1.41L18.17 11H8v2h10.17l-2.58 2.59L17 17l5-5zM4 5h8V3H4c-1.1 0-2 .9-2 2v14c0 1.1.9 2 2 2h8v-2H4V5z"/>
            </svg>
            Log out
          </button>
        </div>
      </section>
    </main>

    <!-- Logout Confirmation Modal -->
    <div v-if="showLogoutModal" class="modal-wrap" @keydown.esc="showLogoutModal = false" @click.self="showLogoutModal = false">
      <div class="logout-modal">
        <div class="logout-icon">
          <svg viewBox="0 0 24 24" width="48" height="48">
            <path fill="#f59e0b" d="M17 7l-1.41 1.41L18.17 11H8v2h10.17l-2.58 2.59L17 17l5-5zM4 5h8V3H4c-1.1 0-2 .9-2 2v14c0 1.1.9 2 2 2h8v-2H4V5z"/>
          </svg>
        </div>
        <div class="logout-content">
          <h3 class="logout-title">Log out?</h3>
          <p class="logout-message">
            Are you sure you want to log out of your account?
          </p>
        </div>
        <div class="logout-actions">
          <button class="btn btn-danger" @click="logout">
            <svg viewBox="0 0 24 24" width="16" height="16">
              <path fill="currentColor" d="M17 7l-1.41 1.41L18.17 11H8v2h10.17l-2.58 2.59L17 17l5-5zM4 5h8V3H4c-1.1 0-2 .9-2 2v14c0 1.1.9 2 2 2h8v-2H4V5z"/>
            </svg>
            Log out
          </button>
          <button class="btn btn-ghost" @click="showLogoutModal = false">Cancel</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'Settings',
  data() {
    return {
      user: {},
      notesCount: 0,
      materialsCount: 0,
      showLogoutModal: false
    }
  },
  computed: {
    userInitials() {
      if (this.user.username) {
        return this.user.username.charAt(0).toUpperCase()
      }
      if (this.user.email) {
        return this.user.email.charAt(0).toUpperCase()
      }
      return 'U'
    }
  },
  async mounted() {
    await this.loadUserData()
    await this.loadStats()
  },
  methods: {
    async loadUserData() {
      try {
        // Load user from localStorage first
        const userData = localStorage.getItem('user')
        if (userData) {
          this.user = JSON.parse(userData)
        }

        // Optionally fetch fresh user data from API
        const token = localStorage.getItem('token')
        if (token) {
          const response = await fetch('/api/user', {
            headers: {
              'Authorization': `Bearer ${token}`
            }
          })
          if (response.ok) {
            const freshUserData = await response.json()
            this.user = { ...this.user, ...freshUserData }
          }
        }
      } catch (error) {
        console.error('Error loading user data:', error)
      }
    },
    async loadStats() {
      try {
        const token = localStorage.getItem('token')
        if (!token) return

        // Load notes count
        const notesResponse = await fetch('/api/notes', {
          headers: {
            'Authorization': `Bearer ${token}`
          }
        })
        if (notesResponse.ok) {
          const notes = await notesResponse.json()
          this.notesCount = Array.isArray(notes) ? notes.length : 0
        }

        // Load materials count (if endpoint exists)
        try {
          const materialsResponse = await fetch('/api/materials', {
            headers: {
              'Authorization': `Bearer ${token}`
            }
          })
          if (materialsResponse.ok) {
            const materials = await materialsResponse.json()
            this.materialsCount = Array.isArray(materials) ? materials.length : 0
          }
        } catch (error) {
          // Materials endpoint might not exist yet
          this.materialsCount = 0
        }
      } catch (error) {
        console.error('Error loading stats:', error)
      }
    },
    formatDate(dateString) {
      if (!dateString) return 'Not specified'
      try {
        const date = new Date(dateString)
        return date.toLocaleDateString('en-US', {
          year: 'numeric',
          month: 'long',
          day: 'numeric'
        })
      } catch (error) {
        return 'Not specified'
      }
    },
    logout() {
      // Clear all stored data
      localStorage.removeItem('token')
      localStorage.removeItem('user')
      
      // Redirect to home page
      this.$router.push('/')
      
      // Show success message if toast method exists
      if (this.$parent && this.$parent.toast) {
        this.$parent.toast('You have successfully logged out')
      }
    }
  }
}
</script>

<style scoped>
.settings-page {
  min-height: 100vh;
  background: var(--bg);
  color: var(--text);
}

.settings-header {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 20px 24px;
  border-bottom: 1px solid var(--line);
  background: rgba(255,255,255,.02);
}

.back-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  border: 1px solid var(--line);
  border-radius: 8px;
  background: rgba(255,255,255,.04);
  color: var(--text);
  cursor: pointer;
  transition: all 0.2s ease;
  font-size: 14px;
}

.back-btn:hover {
  background: rgba(255,255,255,.08);
  border-color: rgba(0,212,255,.3);
}

.settings-title {
  font-size: 24px;
  font-weight: 700;
  margin: 0;
}

.settings-content {
  padding: 24px;
  max-width: 800px;
  margin: 0 auto;
}

.settings-section {
  margin-bottom: 32px;
}

.section-title {
  font-size: 18px;
  font-weight: 600;
  margin-bottom: 16px;
  color: var(--text);
}

.account-card {
  display: flex;
  gap: 20px;
  padding: 24px;
  border: 1px solid var(--line);
  border-radius: 12px;
  background: rgba(255,255,255,.035);
}

.account-avatar {
  flex-shrink: 0;
}

.avatar-circle {
  width: 80px;
  height: 80px;
  border-radius: 50%;
  background: linear-gradient(135deg, #00d4ff, #0099cc);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 32px;
  font-weight: 700;
  color: white;
}

.account-info {
  flex: 1;
}

.info-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 0;
  border-bottom: 1px solid rgba(255,255,255,.05);
}

.info-row:last-child {
  border-bottom: none;
}

.info-row label {
  font-weight: 500;
  color: var(--muted);
}

.info-row span {
  color: var(--text);
}

.user-id {
  font-family: monospace;
  font-size: 12px;
  background: rgba(255,255,255,.05);
  padding: 4px 8px;
  border-radius: 4px;
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
  gap: 16px;
}

.stat-card {
  padding: 20px;
  border: 1px solid var(--line);
  border-radius: 12px;
  background: rgba(255,255,255,.035);
  text-align: center;
}

.stat-number {
  font-size: 32px;
  font-weight: 700;
  color: #00d4ff;
  margin-bottom: 8px;
}

.stat-label {
  color: var(--muted);
  font-size: 14px;
}

.actions-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.action-btn {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 16px 20px;
  border: 1px solid var(--line);
  border-radius: 12px;
  background: rgba(255,255,255,.035);
  color: var(--text);
  cursor: pointer;
  transition: all 0.2s ease;
  font-size: 16px;
  text-align: left;
}

.logout-btn {
  border-color: rgba(239,68,68,.3);
  color: #ef4444;
}

.logout-btn:hover {
  background: rgba(239,68,68,.1);
  border-color: rgba(239,68,68,.5);
  transform: translateY(-1px);
}

/* Logout Modal Styles */
.logout-modal {
  max-width: 420px;
  padding: 0;
  text-align: center;
  border-radius: 16px;
  overflow: hidden;
  animation: modalIn 0.3s cubic-bezier(0.34, 1.56, 0.64, 1);
  background: var(--bg);
  border: 1px solid var(--line);
}

.logout-icon {
  padding: 24px 24px 16px;
  background: rgba(245,158,11,.1);
}

.logout-content {
  padding: 0 24px 16px;
}

.logout-title {
  font-size: 20px;
  font-weight: 600;
  margin: 0 0 8px;
  color: var(--text);
}

.logout-message {
  color: var(--muted);
  margin: 0;
  line-height: 1.5;
  font-size: 14px;
}

.logout-actions {
  display: flex;
  flex-direction: column;
  gap: 12px;
  padding: 16px 24px 24px;
  background: rgba(0,0,0,0.02);
}

.btn {
  padding: 12px 20px;
  border-radius: 8px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
}

.btn-danger {
  background: #ef4444;
  color: white;
  border: 1px solid #dc2626;
}

.btn-danger:hover {
  background: #dc2626;
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(239, 68, 68, 0.3);
}

.btn-ghost {
  background: transparent;
  color: var(--muted);
  border: 1px solid var(--line);
}

.btn-ghost:hover {
  background: rgba(255,255,255,.05);
  color: var(--text);
}

@keyframes modalIn {
  from {
    opacity: 0;
    transform: scale(0.9) translateY(-10px);
  }
  to {
    opacity: 1;
    transform: scale(1) translateY(0);
  }
}

/* Modal backdrop */
.modal-wrap {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  z-index: 1000;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 20px;
  background: rgba(0,0,0,0.5);
}

.modal-backdrop {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0,0,0,0.5);
  z-index: -1;
}

/* Responsive */
@media (max-width: 768px) {
  .settings-content {
    padding: 16px;
  }
  
  .account-card {
    flex-direction: column;
    text-align: center;
  }
  
  .info-row {
    flex-direction: column;
    gap: 4px;
    text-align: center;
  }
  
  .stats-grid {
    grid-template-columns: 1fr 1fr;
  }
}
</style>
