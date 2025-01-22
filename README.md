# Spotify to Telegram Bio

Автоматически обновляет биографию в Telegram, показывая текущий трек из Spotify.

## Первый запуск

1. Создайте приложение в [Spotify Developer Dashboard](https://developer.spotify.com/dashboard):
   - Получите Client ID и Client Secret
   - В настройках приложения добавьте Redirect URI: `http://localhost:8080/callback`

2. Создайте приложение в [Telegram API](https://my.telegram.org/apps):
   - Получите API ID (числовой) и API Hash

3. Запустите программу. При первом запуске она попросит ввести:
   - Spotify Client ID
   - Spotify Client Secret
   - Telegram API ID
   - Telegram API Hash
   - Номер телефона в международном формате (например: +79001234567)

4. После ввода данных:
   - Откроется браузер для авторизации в Spotify
   - В консоли появится запрос кода подтверждения от Telegram (придет в приложение)

## Изменение конфигурации

Конфигурация хранится в файле `config.json` в той же папке, где находится программа.

Параметры конфигурации:
```json
{
  "spotify_client_id": "ваш_client_id",
  "spotify_client_secret": "ваш_client_secret",
  "telegram_api_id": "ваш_api_id",
  "telegram_api_hash": "ваш_api_hash",
  "telegram_phone": "ваш_номер_телефона",
  "update_interval": 45,
  "http_proxy": "ip:port"
}
```

- `update_interval`: интервал обновления статуса в секундах
- `http_proxy`: (опционально) прокси для доступа к Spotify в формате "ip:port"

## Использование

- Для остановки программы нажмите Ctrl+C - оригинальная биография будет восстановлена
- При перезапуске программы повторный ввод данных не требуется

## Использование в странах, где Spotify недоступен

Для работы программы необходимо:

1. Создать аккаунт Spotify через VPN:
   - Включите VPN страны, где доступен Spotify (например, США)
   - Зарегистрируйтесь на [spotify.com](https://www.spotify.com)
   - При регистрации укажите страну, где Spotify доступен
   - Хотя бы раз войдите в веб-плеер Spotify через VPN

2. После создания аккаунта для работы программы можно использовать:
   - VPN на всей системе
   - Или HTTP прокси в настройках программы:

В config.json:
```json
{
  "http_proxy": "ip:port"  // например: "69.75.172.51:8080"
}
```

Или через переменную окружения:

Windows (PowerShell):
```powershell
$env:HTTP_PROXY="69.75.172.51:8080"
```

Linux/macOS:
```bash
export HTTP_PROXY="69.75.172.51:8080"
```

Прокси/VPN должен:
1. Быть расположен в той же стране, где зарегистрирован аккаунт Spotify
2. Поддерживать HTTPS трафик
3. Быть достаточно быстрым для стриминга музыки
