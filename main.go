package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"spotify-tg-bio/config"
	spotifyclient "spotify-tg-bio/spotify"
	telegramclient "spotify-tg-bio/telegram"
)

func main() {
	log.Println("Starting Spotify to Telegram Bio translator...")

	config, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Сначала инициализируем Telegram
	log.Println("Инициализация Telegram клиента...")
	telegramClient := initTelegramClient(config)
	if telegramClient == nil {
		log.Fatal("Failed to initialize Telegram client")
	}
	log.Println("Telegram клиент успешно инициализирован")

	defer func() {
		if err := telegramClient.Close(); err != nil {
			log.Printf("Ошибка закрытия Telegram клиента: %v", err)
		}
	}()

	// Затем инициализируем Spotify
	log.Println("Инициализация Spotify клиента...")
	spotifyClient := initSpotifyClient(config)
	if spotifyClient == nil {
		log.Fatal("Failed to initialize Spotify client")
	}
	log.Println("Spotify клиент успешно инициализирован")

	// Создаем канал для обработки сигналов завершения
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Создаем контекст с отменой
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Запускаем горутину для обработки сигналов
	go func() {
		sig := <-sigChan
		log.Printf("Получен сигнал: %v\n", sig)
		log.Println("Восстанавливаем оригинальную биографию...")

		if err := telegramClient.RestoreOriginalBio(context.Background()); err != nil {
			log.Printf("Ошибка при восстановлении биографии: %v\n", err)
		} else {
			log.Println("Биография успешно восстановлена")
		}

		cancel()
	}()

	log.Printf("Оригинальная биография: %q\n", telegramClient.GetOriginalBio())
	log.Println("Начинаем обновление статуса...")

	// Бесконечный цикл обновления статуса
	for {
		select {
		case <-ctx.Done():
			log.Println("Завершение работы...")
			return
		default:
			currentTrack := getCurrentTrack(spotifyClient)
			if err := updateTelegramBio(telegramClient, currentTrack); err != nil {
				log.Printf("Failed to update Telegram bio: %v", err)
			} else {
				log.Printf("Биография обновлена: %s", currentTrack)
			}
			time.Sleep(time.Duration(config.UpdateInterval) * time.Second)
		}
	}
}

func getCurrentTrack(client *spotifyclient.Client) string {
	ctx := context.Background()
	playing, err := client.GetCurrentlyPlaying(ctx)
	if err != nil {
		log.Printf("Error getting current track: %v", err)
		return "Не удалось получить текущий трек"
	}

	if playing == nil || !playing.Playing {
		log.Printf("Сейчас ничего не играет")
		return "Сейчас ничего не играет"
	}

	track := playing.Item
	log.Printf("Текущий трек: %s - %s", track.Name, track.Artists[0].Name)
	return "🎵 " + track.Name + " - " + track.Artists[0].Name
}

func updateTelegramBio(client *telegramclient.Client, status string) error {
	ctx := context.Background()
	return client.UpdateBio(ctx, status)
}

func initSpotifyClient(config *config.Config) *spotifyclient.Client {
	client, err := spotifyclient.NewClient(config.SpotifyClientID, config.SpotifyClientSecret)
	if err != nil {
		log.Printf("Failed to create Spotify client: %v", err)
		return nil
	}

	if err := client.Authenticate(); err != nil {
		log.Printf("Failed to authenticate Spotify client: %v", err)
		return nil
	}

	return client
}

func initTelegramClient(config *config.Config) *telegramclient.Client {
	ctx := context.Background()
	client, err := telegramclient.NewClient(
		config.TelegramAPIID,
		config.TelegramAPIHash,
		config.TelegramPhoneNumber,
	)
	if err != nil {
		log.Printf("Failed to create Telegram client: %v", err)
		return nil
	}

	log.Println("Авторизация в Telegram...")
	if err := client.Authenticate(ctx); err != nil {
		log.Printf("Failed to authenticate Telegram client: %v", err)
		return nil
	}
	log.Println("Telegram авторизация успешна")

	return client
}
