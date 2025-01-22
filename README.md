# Spotify to Telegram Bio

Автоматически обновляет биографию в Telegram, показывая текущий трек из Spotify.

[Скачать]([https://developer.spotify.com/dashboard](https://github.com/deadcxde/spotify-to-telegram-bio-go/releases)):

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
   - В консоли появится запрос кода подтверждения от Telegram (придет в приложение)
   - Откроется браузер для авторизации в Spotify

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
  "update_interval": 45
}
```

- `update_interval`: интервал обновления статуса в секундах (по умолчанию 45)

## Использование

- Программа сохраняет сессии Telegram и Spotify, поэтому при перезапуске повторная авторизация не требуется
- Для остановки программы нажмите Ctrl+C - оригинальная биография будет восстановлена
- Биография обновляется каждые 45 секунд (можно изменить в config.json)

## Требования

- Для работы программы необходим VPN в стране, где доступен Spotify
- Аккаунт Spotify должен быть зарегистрирован в стране, где сервис доступен
- Telegram аккаунт должен быть активен и иметь доступ к API

## Устранение неполадок

1. Если Spotify не авторизуется:
   - Убедитесь, что VPN включен
   - Проверьте правильность Client ID и Client Secret
   - Проверьте, что добавлен правильный Redirect URI

2. Если Telegram не авторизуется:
   - Проверьте правильность API ID и API Hash
   - Убедитесь, что номер телефона введен в правильном формате
   - Проверьте, что код подтверждения вводится правильно

3. Если биография не обновляется:
   - Проверьте, что в Spotify играет музыка
   - Убедитесь, что VPN включен и работает
   - Проверьте логи на наличие ошибок
