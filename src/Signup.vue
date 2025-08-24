<template>
  <div id="signup">
    <!-- Animated Background Stars -->
    <div class="parallax-star" style="top: 10%; left: 5%; animation-delay: 0s;"></div>
    <div class="parallax-star" style="top: 20%; left: 15%; animation-delay: 1s;"></div>
    <div class="parallax-star" style="top: 30%; left: 25%; animation-delay: 2s;"></div>
    <div class="parallax-star" style="top: 40%; left: 35%; animation-delay: 3s;"></div>
    <div class="parallax-star" style="top: 50%; left: 45%; animation-delay: 4s;"></div>
    <div class="parallax-star" style="top: 60%; left: 55%; animation-delay: 5s;"></div>
    <div class="parallax-star" style="top: 70%; left: 65%; animation-delay: 6s;"></div>
    <div class="parallax-star" style="top: 80%; left: 75%; animation-delay: 7s;"></div>
    <div class="parallax-star" style="top: 90%; left: 85%; animation-delay: 8s;"></div>
    <div class="parallax-star" style="top: 15%; left: 95%; animation-delay: 9s;"></div>
    
    <!-- Star Rays (4-pointed stars) -->
    <div class="star-ray" style="top: 8%; left: 8%; animation-delay: 0s;"></div>
    <div class="star-ray" style="top: 18%; left: 18%; animation-delay: 1s;"></div>
    <div class="star-ray" style="top: 28%; left: 28%; animation-delay: 2s;"></div>
    <div class="star-ray" style="top: 38%; left: 38%; animation-delay: 3s;"></div>
    <div class="star-ray" style="top: 48%; left: 48%; animation-delay: 4s;"></div>
    <div class="star-ray" style="top: 58%; left: 58%; animation-delay: 5s;"></div>
    <div class="star-ray" style="top: 68%; left: 68%; animation-delay: 6s;"></div>
    <div class="star-ray" style="top: 78%; left: 78%; animation-delay: 7s;"></div>
    <div class="star-ray" style="top: 88%; left: 88%; animation-delay: 8s;"></div>
    <div class="star-ray" style="top: 13%; left: 13%; animation-delay: 9s;"></div>
    
    <!-- Header -->
    <header class="header">
      <div class="container">
        <div class="logo">
          <h1>SpeakApper AI</h1>
        </div>
        <nav class="nav">
          <a href="#" @click="goBack" class="nav-link">Back to Home</a>
        </nav>
      </div>
    </header>

    <!-- Signup Section -->
    <section class="signup-section">
      <div class="container">
        <div class="signup-content">
          <div class="signup-form-container">
            <div class="signup-header">
              <h2 class="signup-title">Sign Up</h2>
              <p class="signup-subtitle">Create notes in minutes. No credit card required.</p>
            </div>
            
            <form class="signup-form" @submit.prevent="handleSignup">
              <!-- Google Sign Up Button -->
              <button type="button" class="btn btn-google" @click="signupWithGoogle">
                <div class="google-icon">
                  <svg width="18" height="18" viewBox="0 0 18 18" xmlns="http://www.w3.org/2000/svg">
                    <g fill="none" fill-rule="evenodd">
                      <path d="M17.64 9.205c0-.639-.057-1.252-.164-1.841H9v3.481h4.844a4.14 4.14 0 0 1-1.796 2.716v2.259h2.908c1.702-1.567 2.684-3.875 2.684-6.615z" fill="#4285F4"/>
                      <path d="M9 18c2.43 0 4.467-.806 5.956-2.18l-2.908-2.259c-.806.54-1.837.86-3.048.86-2.344 0-4.328-1.584-5.036-3.711H.957v2.332A8.997 8.997 0 0 0 9 18z" fill="#34A853"/>
                      <path d="M3.964 10.71A5.41 5.41 0 0 1 3.682 9c0-.593.102-1.17.282-1.71V4.958H.957A8.996 8.996 0 0 0 0 9c0 1.452.348 2.827.957 4.042l3.007-2.332z" fill="#FBBC05"/>
                      <path d="M9 3.58c1.321 0 2.508.454 3.44 1.345l2.582-2.58C13.463.891 11.426 0 9 0A8.997 8.997 0 0 0 .957 4.958L3.964 7.29C4.672 5.163 6.656 3.58 9 3.58z" fill="#EA4335"/>
                    </g>
                  </svg>
                </div>
                Continue with Google
              </button>
              
              <div class="divider">
                <span>OR</span>
              </div>
              
              <!-- Form Fields -->
              <div class="form-row">
                <div class="form-group">
                  <label for="firstName">First name</label>
                  <input 
                    type="text" 
                    id="firstName" 
                    v-model="formData.firstName" 
                    required
                    class="form-input"
                  >
                </div>
                <div class="form-group">
                  <label for="lastName">Last name</label>
                  <input 
                    type="text" 
                    id="lastName" 
                    v-model="formData.lastName" 
                    required
                    class="form-input"
                  >
                </div>
              </div>
              
              <div class="form-group">
                <label for="email">Email</label>
                <input 
                  type="email" 
                  id="email" 
                  v-model="formData.email" 
                  required
                  class="form-input"
                >
              </div>
              
              <div class="form-group">
                <label for="password">Password</label>
                <input 
                  type="password" 
                  id="password" 
                  v-model="formData.password" 
                  required
                  class="form-input"
                >
              </div>
              
              <button type="submit" class="btn btn-primary btn-signup">
                Create an account
              </button>
            </form>
            
            <div class="signup-footer">
              <p class="login-link">
                Already have an account? <a href="#" @click="goToLogin">Sign in</a>
              </p>
              <p class="terms">
                By creating or entering an account, you agree to the 
                <a href="#">Terms of Service</a> and <a href="#">Privacy Policy</a>.
              </p>
            </div>
          </div>
          
          <!-- Visual Side -->
          <div class="signup-visual">
            <div class="ai-assistant-demo">
              <div class="demo-header">
                <div class="demo-title">AI Assistant Demo</div>
              </div>
              <div class="demo-content">
                <div class="ai-message">
                  <div class="ai-avatar">🤖</div>
                  <div class="message-text">Hi! I'm your AI assistant. I'll help you create amazing notes and documents.</div>
                </div>
                <div class="user-message">
                  <div class="message-text">Can you help me organize my thoughts?</div>
                  <div class="user-avatar">👤</div>
                </div>
                <div class="ai-message">
                  <div class="ai-avatar">🤖</div>
                  <div class="message-text">Absolutely! Just start typing and I'll help structure your ideas.</div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </section>
  </div>
</template>

<script>
export default {
  name: 'Signup',
  data() {
    return {
      formData: {
        firstName: '',
        lastName: '',
        email: '',
        password: ''
      }
    }
  },
  methods: {
    async handleSignup() {
      try {
        const response = await fetch('http://localhost:8080/api/signup', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(this.formData)
        })
        const result = await response.json()
        if (result.success) {
          localStorage.setItem('token', result.token)
          localStorage.setItem('user', JSON.stringify(result.user))
          this.$router.push('/dashboard')
        } else {
          console.warn(`Ошибка: ${result.message}`)
        }
      } catch (error) {
        console.error('Ошибка при регистрации:', error)
        console.error('Ошибка при регистрации. Попробуйте еще раз.')
      }
    },

    waitForGoogleOAuth() {
      return new Promise((resolve, reject) => {
        const maxAttempts = 50 // Максимум 5 секунд
        let attempts = 0
        const checkGoogle = () => {
          if (typeof google !== 'undefined' && google.accounts && google.accounts.oauth2) {
            resolve()
          } else if (attempts < maxAttempts) {
            attempts++
            setTimeout(checkGoogle, 100)
          } else {
            reject(new Error('Google OAuth SDK не загрузился'))
          }
        }
        checkGoogle()
      })
    },

    async signupWithGoogle() {
      try {
        await this.waitForGoogleOAuth()
        const cid = import.meta.env.VITE_GOOGLE_CLIENT_ID
        if (!cid) {
          console.error('Google Client ID не настроен. Создайте .env с VITE_GOOGLE_CLIENT_ID="<your-client-id>" и перезапустите dev-сервер (npm run dev).')
          return
        }
        const tokenClient = google.accounts.oauth2.initTokenClient({
          client_id: cid,
          scope: 'openid email profile',
          callback: async (response) => {
            if (response.error) { console.error('Ошибка авторизации Google: ' + response.error); return }
            try {
              const serverResponse = await fetch('http://localhost:8080/api/google-signup', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ token: response.access_token, idToken: response.credential })
              })
              const result = await serverResponse.json()
              if (result.success) {
                localStorage.setItem('token', result.token)
                localStorage.setItem('user', JSON.stringify(result.user))
                this.$router.push('/dashboard')
              } else {
                console.warn(`Ошибка: ${result.message}`)
              }
            } catch (error) {
              console.error('Ошибка при отправке токена на сервер:', error)
              console.error('Ошибка при регистрации через Google. Попробуйте еще раз.')
            }
          }
        })
        tokenClient.requestAccessToken()
      } catch (error) {
        console.error('Ошибка при инициализации Google OAuth:', error)
        console.error('Ошибка при инициализации Google OAuth. Попробуйте еще раз.')
      }
    },

    goToLogin() {
      console.log('Navigate to login')
    },
    goBack() { this.$router.push('/') }
  }
}
</script>

<style scoped>
/* Signup specific styles */
.signup-section {
  min-height: 100vh;
  padding: 120px 0 80px;
  background: rgba(10, 10, 10, 0.8);
  position: relative;
  overflow: hidden;
}

.signup-section::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: radial-gradient(circle at 30% 20%, rgba(0, 212, 255, 0.1) 0%, transparent 50%),
              radial-gradient(circle at 70% 80%, rgba(124, 58, 237, 0.1) 0%, transparent 50%);
  z-index: 1;
}

.signup-section .container {
  position: relative;
  z-index: 2;
}

.signup-content {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 80px;
  align-items: center;
  max-width: 1000px;
  margin: 0 auto;
}

.signup-form-container {
  background: rgba(20, 20, 40, 0.6);
  backdrop-filter: blur(25px);
  border: 1px solid rgba(0, 212, 255, 0.25);
  border-radius: 20px;
  padding: 40px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.2), 0 0 20px rgba(0, 212, 255, 0.1);
}

.signup-header {
  text-align: center;
  margin-bottom: 40px;
}

.signup-title {
  font-size: 36px;
  font-weight: 700;
  color: #ffffff;
  margin-bottom: 16px;
  text-shadow: 0 0 20px rgba(255, 255, 255, 0.3);
}

.signup-subtitle {
  font-size: 18px;
  color: #b0b0b0;
  text-shadow: 0 0 10px rgba(176, 176, 176, 0.3);
}

.signup-form {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.btn-google {
  background: rgba(255, 255, 255, 0.1);
  backdrop-filter: blur(20px);
  border: 1px solid rgba(255, 255, 255, 0.2);
  color: #ffffff;
  padding: 16px 24px;
  border-radius: 12px;
  font-size: 16px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.3s ease;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 12px;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.1);
}

.btn-google:hover {
  background: rgba(255, 255, 255, 0.15);
  transform: translateY(-2px);
  box-shadow: 0 8px 25px rgba(0, 0, 0, 0.2);
}

.google-icon {
  font-size: 20px;
}

.divider {
  text-align: center;
  position: relative;
  margin: 20px 0;
}

.divider::before {
  content: '';
  position: absolute;
  top: 50%;
  left: 0;
  right: 0;
  height: 1px;
  background: rgba(255, 255, 255, 0.2);
}

.divider span {
  background: rgba(20, 20, 40, 0.6);
  padding: 0 20px;
  color: #b0b0b0;
  font-size: 14px;
}

.form-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.form-group label {
  font-size: 14px;
  font-weight: 600;
  color: #ffffff;
  text-shadow: 0 0 10px rgba(255, 255, 255, 0.3);
}

.form-input {
  background: rgba(255, 255, 255, 0.1);
  backdrop-filter: blur(20px);
  border: 1px solid rgba(255, 255, 255, 0.2);
  border-radius: 8px;
  padding: 12px 16px;
  color: #ffffff;
  font-size: 16px;
  transition: all 0.3s ease;
}

.form-input:focus {
  outline: none;
  border-color: rgba(0, 212, 255, 0.6);
  box-shadow: 0 0 20px rgba(0, 212, 255, 0.2);
}

.form-input::placeholder {
  color: rgba(255, 255, 255, 0.5);
}

.btn-signup {
  margin-top: 16px;
  padding: 16px 24px;
  font-size: 16px;
  font-weight: 600;
}

.signup-footer {
  margin-top: 32px;
  text-align: center;
}

.login-link {
  color: #b0b0b0;
  margin-bottom: 16px;
}

.login-link a {
  color: #00d4ff;
  text-decoration: none;
  text-shadow: 0 0 10px rgba(0, 212, 255, 0.3);
  transition: all 0.3s ease;
}

.login-link a:hover {
  text-shadow: 0 0 15px rgba(0, 212, 255, 0.5);
}

.terms {
  font-size: 12px;
  color: #888;
  line-height: 1.5;
}

.terms a {
  color: #00d4ff;
  text-decoration: none;
  text-shadow: 0 0 10px rgba(0, 212, 255, 0.3);
}

.terms a:hover {
  text-shadow: 0 0 15px rgba(0, 212, 255, 0.5);
}

/* Visual Side */
.signup-visual {
  display: flex;
  justify-content: center;
  align-items: center;
}

.ai-assistant-demo {
  background: rgba(20, 20, 40, 0.6);
  backdrop-filter: blur(25px);
  border: 1px solid rgba(0, 212, 255, 0.25);
  border-radius: 20px;
  padding: 24px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.2), 0 0 20px rgba(0, 212, 255, 0.1);
  width: 100%;
  max-width: 400px;
}

.demo-header {
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
  padding-bottom: 16px;
  margin-bottom: 20px;
}

.demo-title {
  font-size: 16px;
  color: #00d4ff;
  text-align: center;
  text-shadow: 0 0 10px rgba(0, 212, 255, 0.3);
}

.demo-content {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.ai-message, .user-message {
  display: flex;
  align-items: flex-start;
  gap: 12px;
}

.user-message {
  flex-direction: row-reverse;
}

.ai-avatar, .user-avatar {
  font-size: 24px;
  flex-shrink: 0;
}

.message-text {
  background: rgba(255, 255, 255, 0.1);
  backdrop-filter: blur(20px);
  border: 1px solid rgba(255, 255, 255, 0.2);
  border-radius: 12px;
  padding: 12px 16px;
  color: #ffffff;
  font-size: 14px;
  line-height: 1.4;
  max-width: 280px;
}

.user-message .message-text {
  background: rgba(0, 212, 255, 0.2);
  border-color: rgba(0, 212, 255, 0.3);
}

/* Responsive Design */
@media (max-width: 768px) {
  .signup-content {
    grid-template-columns: 1fr;
    gap: 40px;
  }
  
  .signup-form-container {
    padding: 30px 20px;
  }
  
  .form-row {
    grid-template-columns: 1fr;
  }
  
  .signup-title {
    font-size: 28px;
  }
  
  .signup-visual {
    order: -1;
  }
}

@media (max-width: 480px) {
  .signup-section {
    padding: 100px 0 60px;
  }
  
  .signup-form-container {
    padding: 24px 16px;
  }
  
  .signup-title {
    font-size: 24px;
  }
}
</style> 