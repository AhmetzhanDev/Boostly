<template>
  <div class="signin-page">
    <div class="signin-container">
      <div class="signin-card">
        <!-- Header -->
        <div class="signin-header">
          <h1 class="signin-title">Sign in</h1>
          <p class="signin-subtitle">Welcome back to SpeakApper</p>
        </div>

        <!-- Login Form -->
        <form @submit.prevent="handleLogin" class="signin-form">
          <div class="form-group">
            <label for="email" class="form-label">Email</label>
            <input
              id="email"
              v-model="form.email"
              type="email"
              class="form-input"
              placeholder="Enter your email"
              required
              :disabled="loading"
            />
          </div>

          <div class="form-group">
            <label for="password" class="form-label">Password</label>
            <div class="password-input">
              <input
                id="password"
                v-model="form.password"
                :type="showPassword ? 'text' : 'password'"
                class="form-input"
                placeholder="Enter your password"
                required
                :disabled="loading"
              />
              <button
                type="button"
                class="password-toggle"
                @click="showPassword = !showPassword"
                :disabled="loading"
              >
                <svg v-if="showPassword" viewBox="0 0 24 24" width="20" height="20">
                  <path fill="currentColor" d="M12 7c2.76 0 5 2.24 5 5 0 .65-.13 1.26-.36 1.83l2.92 2.92c1.51-1.26 2.7-2.89 3.43-4.75-1.73-4.39-6-7.5-11-7.5-1.4 0-2.74.25-3.98.7l2.16 2.16C10.74 7.13 11.35 7 12 7zM2 4.27l2.28 2.28.46.46C3.08 8.3 1.78 10.02 1 12c1.73 4.39 6 7.5 11 7.5 1.55 0 3.03-.3 4.38-.84l.42.42L19.73 22 21 20.73 3.27 3 2 4.27zM7.53 9.8l1.55 1.55c-.05.21-.08.43-.08.65 0 1.66 1.34 3 3 3 .22 0 .44-.03.65-.08l1.55 1.55c-.67.33-1.41.53-2.2.53-2.76 0-5-2.24-5-5 0-.79.2-1.53.53-2.2zm4.31-.78l3.15 3.15.02-.16c0-1.66-1.34-3-3-3l-.17.01z"/>
                </svg>
                <svg v-else viewBox="0 0 24 24" width="20" height="20">
                  <path fill="currentColor" d="M12 4.5C7 4.5 2.73 7.61 1 12c1.73 4.39 6 7.5 11 7.5s9.27-3.11 11-7.5c-1.73-4.39-6-7.5-11-7.5zM12 17c-2.76 0-5-2.24-5-5s2.24-5 5-5 5 2.24 5 5-2.24 5-5 5zm0-8c-1.66 0-3 1.34-3 3s1.34 3 3 3 3-1.34 3-3-1.34-3-3-3z"/>
                </svg>
              </button>
            </div>
          </div>

          <!-- Error Message -->
          <div v-if="error" class="error-message">
            <svg viewBox="0 0 24 24" width="16" height="16">
              <path fill="currentColor" d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm-2 15l-5-5 1.41-1.41L10 14.17l7.59-7.59L19 8l-9 9z"/>
            </svg>
            {{ error }}
          </div>

          <!-- Submit Button -->
          <button type="submit" class="signin-btn" :disabled="loading">
            <svg v-if="loading" class="loading-spinner" viewBox="0 0 24 24" width="20" height="20">
              <circle cx="12" cy="12" r="10" fill="none" stroke="currentColor" stroke-width="2" stroke-dasharray="31.416" stroke-dashoffset="31.416">
                <animate attributeName="stroke-dasharray" dur="2s" values="0 31.416;15.708 15.708;0 31.416" repeatCount="indefinite"/>
                <animate attributeName="stroke-dashoffset" dur="2s" values="0;-15.708;-31.416" repeatCount="indefinite"/>
              </circle>
            </svg>
            <span v-if="!loading">Sign in</span>
            <span v-else>Signing in...</span>
          </button>
        </form>

        <!-- Divider -->
        <div class="divider">
          <span>or</span>
        </div>

        <!-- Google Sign In -->
        <button class="google-btn" @click="handleGoogleLogin" :disabled="loading">
          <svg viewBox="0 0 24 24" width="20" height="20">
            <path fill="#4285F4" d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z"/>
            <path fill="#34A853" d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"/>
            <path fill="#FBBC05" d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z"/>
            <path fill="#EA4335" d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z"/>
          </svg>
          Sign in with Google
        </button>

        <!-- Footer Links -->
        <div class="signin-footer">
          <p>Don't have an account? <router-link to="/signup" class="link">Sign up</router-link></p>
          <router-link to="/" class="link">← Back to Home</router-link>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'SignIn',
  data() {
    return {
      form: {
        email: '',
        password: ''
      },
      showPassword: false,
      loading: false,
      error: ''
    }
  },
  methods: {
    async handleLogin() {
      if (this.loading) return
      
      this.loading = true
      this.error = ''

      try {
        const response = await fetch('/api/login', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json'
          },
          body: JSON.stringify({
            email: this.form.email,
            password: this.form.password
          })
        })

        const data = await response.json()

        if (response.ok && data.success) {
          // Save token and user data
          localStorage.setItem('token', data.token)
          if (data.user) {
            localStorage.setItem('user', JSON.stringify(data.user))
          }

          // Show success message
          this.showToast('Successfully signed in!')

          // Redirect to dashboard
          this.$router.push('/dashboard')
        } else {
          this.error = data.message || 'Sign-in failed'
        }
      } catch (error) {
        console.error('Login error:', error)
        this.error = 'Failed to connect to server'
      } finally {
        this.loading = false
      }
    },

    async handleGoogleLogin() {
      if (this.loading) return
      
      this.loading = true
      this.error = ''

      try {
        // Initialize Google OAuth
        if (!window.google) {
          throw new Error('Google OAuth not loaded')
        }

        const tokenClient = window.google.accounts.oauth2.initTokenClient({
          client_id: import.meta.env.VITE_GOOGLE_CLIENT_ID,
          scope: 'openid email profile',
          callback: async (response) => {
            if (response.access_token) {
              await this.handleGoogleCallback(response.access_token)
            } else {
              this.error = 'Failed to obtain token from Google'
              this.loading = false
            }
          },
          error_callback: (error) => {
            console.error('Google OAuth error:', error)
            this.error = 'Google authorization error'
            this.loading = false
          }
        })
        
        tokenClient.requestAccessToken()
      } catch (error) {
        console.error('Google login error:', error)
        this.error = 'Google OAuth initialization error'
        this.loading = false
      }
    },

    async handleGoogleCallback(accessToken) {
      try {
        // Get user info from Google
        const userInfoResponse = await fetch(`https://www.googleapis.com/oauth2/v2/userinfo?access_token=${accessToken}`)
        const userInfo = await userInfoResponse.json()

        if (!userInfoResponse.ok) {
          throw new Error('Failed to fetch user information')
        }

        // Send to backend
        const response = await fetch('/api/google-signup', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json'
          },
          body: JSON.stringify({
            token: accessToken,
            idToken: userInfo.id,
            email: userInfo.email,
            firstName: userInfo.given_name || '',
            lastName: userInfo.family_name || ''
          })
        })

        const data = await response.json()

        if (response.ok && data.success) {
          // Save token and user data
          localStorage.setItem('token', data.token)
          if (data.user) {
            localStorage.setItem('user', JSON.stringify(data.user))
          }

          this.showToast('Successfully signed in with Google!')
          this.$router.push('/dashboard')
        } else {
          this.error = data.message || 'Google sign-in failed'
        }
      } catch (error) {
        console.error('Google callback error:', error)
        this.error = 'Error processing Google authorization'
      } finally {
        this.loading = false
      }
    },

    showToast(message) {
      // Simple toast implementation
      if (this.$parent && this.$parent.toast) {
        this.$parent.toast(message)
      } else {
        alert(message)
      }
    }
  },

  mounted() {
    // Check if already logged in
    const token = localStorage.getItem('token')
    if (token) {
      this.$router.push('/dashboard')
    }

    // Load Google OAuth script
    this.loadGoogleOAuth()
  },

  methods: {
    async handleLogin() {
      if (this.loading) return
      
      this.loading = true
      this.error = ''

      try {
        const response = await fetch('/api/login', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json'
          },
          body: JSON.stringify({
            email: this.form.email,
            password: this.form.password
          })
        })

        const data = await response.json()

        if (response.ok && data.success) {
          // Save token and user data
          localStorage.setItem('token', data.token)
          if (data.user) {
            localStorage.setItem('user', JSON.stringify(data.user))
          }

          // Show success message
          this.showToast('Successfully signed in!')

          // Redirect to dashboard
          this.$router.push('/dashboard')
        } else {
          this.error = data.message || 'Sign-in failed'
        }
      } catch (error) {
        console.error('Login error:', error)
        this.error = 'Failed to connect to server'
      } finally {
        this.loading = false
      }
    },

    async handleGoogleLogin() {
      if (this.loading) return
      this.loading = true
      this.error = ''

      try {
        await this.waitForGoogleOAuth()
        const cid = import.meta.env.VITE_GOOGLE_CLIENT_ID
        if (!cid) {
          console.error('Google Client ID not configured. Create .env with VITE_GOOGLE_CLIENT_ID="<your-client-id>" and restart the dev server (npm run dev).')
          this.loading = false
          return
        }
        const tokenClient = google.accounts.oauth2.initTokenClient({
          client_id: cid,
          scope: 'openid email profile',
          callback: async (response) => {
            if (response.error) { console.error('Google authorization error: ' + response.error); this.loading = false; return }
            try {
              const serverResponse = await fetch('/api/google-signup', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ token: response.access_token, idToken: response.credential })
              })
              const result = await serverResponse.json()
              if (result.success) {
                localStorage.setItem('token', result.token)
                if (result.user) localStorage.setItem('user', JSON.stringify(result.user))
                this.showToast('Successfully signed in with Google!')
                this.$router.push('/dashboard')
              } else {
                this.error = result.message || 'Google sign-in failed'
              }
            } catch (err) {
              console.error('Error sending token to server:', err)
              this.error = 'Error signing in with Google. Please try again.'
            } finally {
              this.loading = false
            }
          }
        })
        tokenClient.requestAccessToken()
      } catch (error) {
        console.error('Error initializing Google OAuth:', error)
        this.error = 'Error initializing Google OAuth. Please try again.'
        this.loading = false
      }
    },

    // No longer need popup/One Tap — we use the same flow as in Signup
    
    waitForGoogleOAuth() {
      return new Promise((resolve, reject) => {
        const maxAttempts = 50
        let attempts = 0
        const checkGoogle = () => {
          if (typeof google !== 'undefined' && google.accounts && google.accounts.oauth2) {
            resolve()
          } else if (attempts < maxAttempts) {
            attempts++
            setTimeout(checkGoogle, 100)
          } else {
            reject(new Error('Google OAuth SDK did not load'))
          }
        }
        checkGoogle()
      })
    },

    // Previously used One Tap and userinfo flow removed to fully match Signup

    showToast(message) {
      // Simple toast implementation
      if (this.$parent && this.$parent.toast) {
        this.$parent.toast(message)
      } else {
        alert(message)
      }
    },

    loadGoogleOAuth() {
      // Check if script already loaded
      if (window.google) {
        return
      }

      const script = document.createElement('script')
      script.src = 'https://accounts.google.com/gsi/client'
      script.async = true
      script.defer = true
      script.onload = () => {
        console.log('Google OAuth script loaded')
      }
      script.onerror = () => {
        console.error('Failed to load Google OAuth script')
      }
      document.head.appendChild(script)
    }
  }
}
</script>

<style scoped>
.signin-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 20px;
}

.signin-container {
  width: 100%;
  max-width: 400px;
}

.signin-card {
  background: rgba(255,255,255,.035);
  backdrop-filter: blur(10px);
  border-radius: 20px;
  padding: 40px;
  box-shadow: 0 12px 30px rgba(0,0,0,.35);
  border: 1px solid var(--line);
  color: var(--text);
}

.signin-header {
  text-align: center;
  margin-bottom: 32px;
}

.signin-title {
  font-size: 28px;
  font-weight: 700;
  color: var(--text);
  margin: 0 0 8px;
}

.signin-subtitle {
  color: var(--muted);
  margin: 0;
  font-size: 16px;
}

.signin-form {
  margin-bottom: 24px;
}

.form-group {
  margin-bottom: 20px;
}

.form-label {
  display: block;
  font-weight: 500;
  color: var(--text);
  margin-bottom: 8px;
  font-size: 14px;
}

.form-input {
  width: 100%;
  padding: 12px 16px;
  border: 1px solid var(--line);
  border-radius: 12px;
  font-size: 16px;
  transition: all 0.2s ease;
  background: rgba(0,0,0,.25);
  color: var(--text);
  box-sizing: border-box;
}

.form-input::placeholder {
  color: var(--muted);
  opacity: .8;
}

.form-input:focus {
  outline: none;
  border-color: rgba(124,58,237,.5);
  box-shadow: 0 0 0 2px rgba(124,58,237,.25);
}

.form-input:disabled {
  background: rgba(255,255,255,.06);
  cursor: not-allowed;
}

.password-input {
  position: relative;
}

.password-toggle {
  position: absolute;
  right: 12px;
  top: 50%;
  transform: translateY(-50%);
  background: none;
  border: none;
  color: var(--muted);
  cursor: pointer;
  padding: 4px;
  border-radius: 4px;
  transition: color 0.2s ease;
}

.password-toggle:hover { color: var(--text); }

.password-toggle:disabled {
  cursor: not-allowed;
  opacity: 0.5;
}

.error-message {
  display: flex;
  align-items: center;
  gap: 8px;
  color: #fecaca;
  font-size: 14px;
  margin-bottom: 16px;
  padding: 12px;
  background: rgba(239,68,68,.12);
  border-radius: 8px;
  border: 1px solid rgba(239,68,68,.35);
}

.signin-btn {
  width: 100%;
  padding: 14px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  border: none;
  border-radius: 12px;
  font-size: 16px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s ease;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
}

.signin-btn:hover:not(:disabled) {
  transform: translateY(-1px);
  box-shadow: 0 8px 25px rgba(102, 126, 234, 0.3);
}

.signin-btn:disabled {
  opacity: 0.7;
  cursor: not-allowed;
  transform: none;
}

.loading-spinner {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

.divider {
  text-align: center;
  margin: 24px 0;
  position: relative;
  color: var(--muted);
  font-size: 14px;
}

.divider::before {
  content: '';
  position: absolute;
  top: 50%;
  left: 0;
  right: 0;
  height: 1px;
  background: rgba(255,255,255,.12);
}

.divider span { background: transparent; padding: 0 16px; }

.google-btn {
  width: 100%;
  padding: 12px;
  background: rgba(255,255,255,.06);
  border: 1px solid var(--line);
  border-radius: 12px;
  font-size: 16px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 12px;
  color: var(--text);
}

.google-btn:hover:not(:disabled) {
  border-color: rgba(124,58,237,.35);
  background: rgba(255,255,255,.08);
}

.google-btn:disabled { opacity: 0.7; cursor: not-allowed; }

.signin-footer {
  text-align: center;
  margin-top: 24px;
  font-size: 14px;
  color: var(--muted);
}

.signin-footer p {
  margin: 0 0 12px;
}

.link {
  color: #c4b5fd;
  text-decoration: none;
  font-weight: 500;
  transition: color 0.2s ease;
}

.link:hover { color: #e9d5ff; text-decoration: underline; }

/* Responsive */
@media (max-width: 480px) {
  .signin-card {
    padding: 24px;
    margin: 10px;
  }
  
  .signin-title {
    font-size: 24px;
  }
  
  .form-input {
    font-size: 16px; /* Prevent zoom on iOS */
  }
}
</style>
