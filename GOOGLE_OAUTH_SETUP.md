# Настройка Google OAuth для SpeakApper

## 1. Создание проекта в Google Cloud Console

1. Перейдите на [Google Cloud Console](https://console.cloud.google.com/)
2. Создайте новый проект или выберите существующий
3. Включите Google+ API:
   - Перейдите в "APIs & Services" > "Library"
   - Найдите "Google+ API" и включите его

## 2. Настройка OAuth 2.0

1. Перейдите в "APIs & Services" > "Credentials"
2. Нажмите "Create Credentials" > "OAuth 2.0 Client IDs"
3. Выберите тип приложения "Web application"
4. Заполните форму:
   - **Name**: SpeakApper
   - **Authorized JavaScript origins**:
     - `http://localhost:3001` (для разработки)
     - `http://localhost:3000` (альтернативный порт)
   - **Authorized redirect URIs**:
     - `http://localhost:3001/`
     - `http://localhost:3000/`

5. Нажмите "Create"
6. Скопируйте **Client ID** (понадобится для фронтенда)

## 3. Обновление фронтенда

Создайте файл `.env` в корне проекта на основе `env.example` и укажите Client ID:

```
VITE_GOOGLE_CLIENT_ID=123456789-abcdefghijklmnop.apps.googleusercontent.com
```

Vite автоматически подхватывает переменную `VITE_GOOGLE_CLIENT_ID` после перезапуска `npm run dev`. В коде она используется как `import.meta.env.VITE_GOOGLE_CLIENT_ID`.

## 4. Тестирование

1. Запустите фронтенд: `npm run dev`
2. Запустите бэкенд: `cd backend && go run .`
3. Перейдите на страницу регистрации
4. Нажмите "Continue with Google"
5. Выберите аккаунт Google
6. Разрешите доступ приложению

## 5. Безопасность

### Для продакшена:
- Добавьте домен вашего сайта в "Authorized JavaScript origins"
- Используйте HTTPS
- Настройте правильные redirect URIs
- Ограничьте доступ по IP если необходимо

### Переменные окружения:
Рекомендуется вынести Client ID в переменные окружения:

```javascript
// В .env файле
VUE_APP_GOOGLE_CLIENT_ID=your-client-id-here

// В коде
client_id: process.env.VUE_APP_GOOGLE_CLIENT_ID,
```

## 6. Возможные проблемы

### CORS ошибки:
Убедитесь, что домен добавлен в "Authorized JavaScript origins"

### "Invalid client" ошибка:
Проверьте правильность Client ID

### "Redirect URI mismatch":
Проверьте настройки redirect URIs в Google Console

## 7. Дополнительные настройки

### Scopes (области доступа):
Текущие scopes: `openid email profile`
- `openid` - для OpenID Connect
- `email` - доступ к email
- `profile` - доступ к основной информации профиля

### Дополнительные scopes (при необходимости):
- `https://www.googleapis.com/auth/user.birthday.read` - дата рождения
- `https://www.googleapis.com/auth/user.gender.read` - пол
- `https://www.googleapis.com/auth/user.phonenumbers.read` - телефон

## 8. Мониторинг

В Google Cloud Console можно отслеживать:
- Количество запросов к API
- Ошибки авторизации
- Использование квот
- Активность пользователей 