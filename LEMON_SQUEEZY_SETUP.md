# Настройка Lemon Squeezy для SpeakApper

Это руководство поможет вам настроить интеграцию с платежной системой Lemon Squeezy.

## 1. Создание аккаунта Lemon Squeezy

1. Перейдите на [lemonsqueezy.com](https://lemonsqueezy.com)
2. Зарегистрируйтесь и создайте новый аккаунт
3. Подтвердите email и завершите настройку профиля

## 2. Создание магазина (Store)

1. В панели управления создайте новый магазин
2. Заполните информацию о магазине (название, описание, логотип)
3. Настройте налоговые параметры и валюту
4. Запишите **Store ID** - он понадобится для настройки

## 3. Создание продукта и вариантов

### Создание продукта:
1. Перейдите в раздел "Products"
2. Создайте новый продукт "SpeakApper Subscription"
3. Выберите тип "Subscription"
4. Запишите **Product ID**

### Создание вариантов подписки:
1. **Basic Plan**:
   - Название: "Basic Plan"
   - Цена: $9.99
   - Интервал: Monthly
   - Запишите **Variant ID** для Basic

2. **Premium Plan**:
   - Название: "Premium Plan" 
   - Цена: $19.99
   - Интервал: Monthly
   - Запишите **Variant ID** для Premium

## 4. Получение API ключей

1. Перейдите в "Settings" → "API"
2. Создайте новый API ключ
3. Выберите необходимые разрешения:
   - `checkouts:write`
   - `subscriptions:read`
   - `customers:read`
4. Скопируйте **API Key**

## 5. Настройка Webhook

1. В разделе "Settings" → "Webhooks" создайте новый webhook
2. URL: `https://your-domain.com/api/subscription/webhook`
3. Выберите события:
   - `subscription_created`
   - `subscription_updated`
   - `subscription_cancelled`
   - `subscription_expired`
4. Создайте **Webhook Secret** и запишите его

## 6. Настройка переменных окружения

Создайте файл `.env` на основе `env.example` и заполните:

```bash
# Lemon Squeezy Configuration
LEMON_SQUEEZY_API_KEY=lssk_your_api_key_here
LEMON_SQUEEZY_STORE_ID=12345
LEMON_SQUEEZY_PRODUCT_ID=67890
LEMON_SQUEEZY_BASIC_VARIANT_ID=11111
LEMON_SQUEEZY_PREMIUM_VARIANT_ID=22222
LEMON_SQUEEZY_WEBHOOK_SECRET=your_webhook_secret_here
```

## 7. Тестирование интеграции

### Проверка API эндпоинтов:

1. **Получение планов**:
```bash
curl -H "Authorization: Bearer YOUR_JWT_TOKEN" \
     http://localhost:8080/api/subscription/plans
```

2. **Создание checkout сессии**:
```bash
curl -X POST \
     -H "Content-Type: application/json" \
     -H "Authorization: Bearer YOUR_JWT_TOKEN" \
     -d '{"plan_id": "premium"}' \
     http://localhost:8080/api/subscription/checkout
```

3. **Проверка статуса подписки**:
```bash
curl -H "Authorization: Bearer YOUR_JWT_TOKEN" \
     http://localhost:8080/api/subscription/status
```

## 8. Настройка тестового режима

Для разработки используйте тестовый режим Lemon Squeezy:

1. В настройках магазина включите "Test mode"
2. Используйте тестовые карты для проверки платежей
3. Все транзакции будут помечены как тестовые

### Тестовые карты:
- **Успешная оплата**: 4242 4242 4242 4242
- **Отклоненная карта**: 4000 0000 0000 0002
- **Недостаточно средств**: 4000 0000 0000 9995

## 9. Безопасность

### Проверка webhook подписи:
Обновите функцию `verifyWebhookSignature` в `payments.go`:

```go
func verifyWebhookSignature(body []byte, signature, secret string) bool {
    mac := hmac.New(sha256.New, []byte(secret))
    mac.Write(body)
    expectedSignature := hex.EncodeToString(mac.Sum(nil))
    return hmac.Equal([]byte(signature), []byte(expectedSignature))
}
```

### Рекомендации:
- Никогда не коммитьте API ключи в репозиторий
- Используйте HTTPS для production
- Регулярно ротируйте API ключи
- Мониторьте webhook события

## 10. Переход в production

1. Отключите тестовый режим в Lemon Squeezy
2. Обновите webhook URL на production домен
3. Проверьте все переменные окружения
4. Протестируйте полный цикл оплаты

## Поддержка

При возникновении проблем:
1. Проверьте логи backend сервера
2. Убедитесь в правильности API ключей
3. Проверьте статус webhook в панели Lemon Squeezy
4. Обратитесь к [документации Lemon Squeezy API](https://docs.lemonsqueezy.com/api)

## Полезные ссылки

- [Lemon Squeezy Dashboard](https://app.lemonsqueezy.com)
- [API Documentation](https://docs.lemonsqueezy.com/api)
- [Webhook Events](https://docs.lemonsqueezy.com/api/webhooks)
- [Testing Guide](https://docs.lemonsqueezy.com/guides/testing)
