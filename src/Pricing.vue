<template>
  <div class="pricing-container">
    <div class="pricing-header">
      <h1>Выберите ваш план</h1>
      <p>Получите полный доступ к возможностям SpeakApper</p>
    </div>

    <div class="subscription-status" v-if="subscriptionStatus">
      <div class="status-card" :class="{ active: subscriptionStatus.is_active }">
        <h3>{{ subscriptionStatus.is_active ? 'Активная подписка' : 'Подписка неактивна' }}</h3>
        <p v-if="subscriptionStatus.is_active">
          План: {{ subscriptionStatus.plan || 'Premium' }}
        </p>
        <p v-if="subscriptionStatus.is_active && subscriptionStatus.days_left >= 0">
          Осталось дней: {{ subscriptionStatus.days_left }}
        </p>
      </div>
    </div>

    <div class="plans-grid" v-if="plans.length > 0">
      <div 
        class="plan-card" 
        v-for="plan in plans" 
        :key="plan.id"
        :class="{ recommended: plan.id === 'premium' }"
      >
        <div class="plan-header">
          <h3>{{ plan.name }}</h3>
          <div class="price">
            <span class="amount">${{ plan.price }}</span>
            <span class="period">/{{ plan.interval === 'month' ? 'мес' : 'год' }}</span>
          </div>
        </div>
        
        <div class="plan-description">
          <p>{{ plan.description }}</p>
        </div>

        <div class="plan-features">
          <ul>
            <li v-if="plan.id === 'basic'">✓ До 10 заметок в месяц</li>
            <li v-if="plan.id === 'basic'">✓ Базовая генерация флешкарточек</li>
            <li v-if="plan.id === 'basic'">✓ Простые квизы</li>
            
            <li v-if="plan.id === 'premium'">✓ Неограниченные заметки</li>
            <li v-if="plan.id === 'premium'">✓ Продвинутая генерация контента</li>
            <li v-if="plan.id === 'premium'">✓ Все типы квизов</li>
            <li v-if="plan.id === 'premium'">✓ YouTube транскрипция</li>
            <li v-if="plan.id === 'premium'">✓ Приоритетная поддержка</li>
          </ul>
        </div>

        <button 
          class="subscribe-btn" 
          :class="{ loading: loadingPlan === plan.id }"
          :disabled="loadingPlan === plan.id || (subscriptionStatus?.is_active && subscriptionStatus.plan === plan.id)"
          @click="subscribeToPlan(plan.id)"
        >
          <span v-if="subscriptionStatus?.is_active && subscriptionStatus.plan === plan.id">
            Текущий план
          </span>
          <span v-else-if="loadingPlan === plan.id">
            Создание платежа...
          </span>
          <span v-else>
            Выбрать план
          </span>
        </button>
      </div>
    </div>

    <div class="loading" v-else-if="loading">
      <p>Загрузка планов...</p>
    </div>

    <div class="error" v-if="error">
      <p>{{ error }}</p>
      <button @click="loadPlans" class="retry-btn">Попробовать снова</button>
    </div>
  </div>
</template>

<script>
export default {
  name: 'Pricing',
  data() {
    return {
      plans: [],
      subscriptionStatus: null,
      loading: true,
      loadingPlan: null,
      error: null
    }
  },
  async mounted() {
    await this.loadPlans()
    await this.loadSubscriptionStatus()
  },
  methods: {
    async loadPlans() {
      try {
        this.loading = true
        this.error = null
        
        const response = await fetch('/api/subscription/plans', {
          headers: {
            'Authorization': `Bearer ${localStorage.getItem('token')}`
          }
        })
        
        if (!response.ok) {
          throw new Error('Не удалось загрузить планы')
        }
        
        const data = await response.json()
        this.plans = (data && data.data && data.data.plans) ? data.data.plans : []
      } catch (err) {
        this.error = err.message
        console.error('Error loading plans:', err)
      } finally {
        this.loading = false
      }
    },

    async loadSubscriptionStatus() {
      try {
        const response = await fetch('/api/subscription/status', {
          headers: {
            'Authorization': `Bearer ${localStorage.getItem('token')}`
          }
        })
        
        if (response.ok) {
          const json = await response.json()
          this.subscriptionStatus = json && json.data ? json.data : null
        }
      } catch (err) {
        console.error('Error loading subscription status:', err)
      }
    },

    async subscribeToPlan(planId) {
      try {
        this.loadingPlan = planId
        
        const response = await fetch('/api/subscription/checkout', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${localStorage.getItem('token')}`
          },
          body: JSON.stringify({
            plan_id: planId
          })
        })
        
        if (!response.ok) {
          throw new Error('Не удалось создать платежную сессию')
        }
        
        const data = await response.json()
        
        if (data && data.data && data.data.checkout_url) {
          // Перенаправляем на страницу оплаты Lemon Squeezy
          window.location.href = data.data.checkout_url
        } else {
          throw new Error('Не получена ссылка на оплату')
        }
      } catch (err) {
        this.error = err.message
        console.error('Error creating checkout:', err)
      } finally {
        this.loadingPlan = null
      }
    }
  }
}
</script>

<style scoped>
.pricing-container {
  max-width: 1200px;
  margin: 0 auto;
  padding: 2rem;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
}

.pricing-header {
  text-align: center;
  margin-bottom: 3rem;
}

.pricing-header h1 {
  font-size: 2.5rem;
  font-weight: 700;
  color: #1a1a1a;
  margin-bottom: 1rem;
}

.pricing-header p {
  font-size: 1.2rem;
  color: #666;
  max-width: 600px;
  margin: 0 auto;
}

.subscription-status {
  margin-bottom: 2rem;
}

.status-card {
  background: #f8f9fa;
  border: 2px solid #e9ecef;
  border-radius: 12px;
  padding: 1.5rem;
  text-align: center;
}

.status-card.active {
  background: #d4edda;
  border-color: #28a745;
}

.status-card h3 {
  margin: 0 0 0.5rem 0;
  color: #333;
}

.status-card.active h3 {
  color: #155724;
}

.plans-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: 2rem;
  margin-bottom: 3rem;
}

.plan-card {
  background: white;
  border: 2px solid #e9ecef;
  border-radius: 16px;
  padding: 2rem;
  transition: all 0.3s ease;
  position: relative;
}

.plan-card:hover {
  border-color: #007bff;
  transform: translateY(-4px);
  box-shadow: 0 8px 25px rgba(0, 123, 255, 0.15);
}

.plan-card.recommended {
  border-color: #28a745;
  position: relative;
}

.plan-card.recommended::before {
  content: 'Рекомендуется';
  position: absolute;
  top: -12px;
  left: 50%;
  transform: translateX(-50%);
  background: #28a745;
  color: white;
  padding: 0.5rem 1rem;
  border-radius: 20px;
  font-size: 0.875rem;
  font-weight: 600;
}

.plan-header {
  text-align: center;
  margin-bottom: 1.5rem;
}

.plan-header h3 {
  font-size: 1.5rem;
  font-weight: 700;
  color: #1a1a1a;
  margin-bottom: 1rem;
}

.price {
  display: flex;
  align-items: baseline;
  justify-content: center;
  gap: 0.25rem;
}

.price .amount {
  font-size: 3rem;
  font-weight: 700;
  color: #007bff;
}

.price .period {
  font-size: 1rem;
  color: #666;
}

.plan-description {
  text-align: center;
  margin-bottom: 2rem;
}

.plan-description p {
  color: #666;
  font-size: 1rem;
}

.plan-features ul {
  list-style: none;
  padding: 0;
  margin: 0 0 2rem 0;
}

.plan-features li {
  padding: 0.75rem 0;
  color: #333;
  font-size: 1rem;
  border-bottom: 1px solid #f1f3f4;
}

.plan-features li:last-child {
  border-bottom: none;
}

.subscribe-btn {
  width: 100%;
  background: #007bff;
  color: white;
  border: none;
  border-radius: 8px;
  padding: 1rem 2rem;
  font-size: 1.1rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s ease;
}

.subscribe-btn:hover:not(:disabled) {
  background: #0056b3;
  transform: translateY(-1px);
}

.subscribe-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
  transform: none;
}

.subscribe-btn.loading {
  position: relative;
}

.loading, .error {
  text-align: center;
  padding: 3rem;
}

.error {
  color: #dc3545;
}

.retry-btn {
  background: #007bff;
  color: white;
  border: none;
  border-radius: 6px;
  padding: 0.75rem 1.5rem;
  margin-top: 1rem;
  cursor: pointer;
  font-size: 1rem;
}

.retry-btn:hover {
  background: #0056b3;
}

@media (max-width: 768px) {
  .pricing-container {
    padding: 1rem;
  }
  
  .pricing-header h1 {
    font-size: 2rem;
  }
  
  .plans-grid {
    grid-template-columns: 1fr;
    gap: 1.5rem;
  }
  
  .plan-card {
    padding: 1.5rem;
  }
  
  .price .amount {
    font-size: 2.5rem;
  }
}
</style>
