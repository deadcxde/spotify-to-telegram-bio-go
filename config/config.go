package config

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	SpotifyClientID     string `json:"spotify_client_id"`
	SpotifyClientSecret string `json:"spotify_client_secret"`
	TelegramAPIID       string `json:"telegram_api_id"`
	TelegramAPIHash     string `json:"telegram_api_hash"`
	TelegramPhoneNumber string `json:"telegram_phone"`
	UpdateInterval      int    `json:"update_interval"` // в секундах
	HttpProxy           string `json:"http_proxy,omitempty"`
}

const configFileName = "config.json"

func getConfigPath() string {
	// Получаем путь к исполняемому файлу
	execPath, err := os.Executable()
	if err != nil {
		return configFileName
	}
	// Возвращаем путь к config.json в той же директории
	return filepath.Join(filepath.Dir(execPath), configFileName)
}

func Load() (*Config, error) {
	configPath := getConfigPath()

	// Проверяем существование файла
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return promptAndSaveConfig()
	}

	// Читаем существующий конфиг
	file, err := os.Open(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	var config Config
	if err := json.NewDecoder(file).Decode(&config); err != nil {
		return nil, fmt.Errorf("failed to decode config: %w", err)
	}

	return &config, nil
}

func promptAndSaveConfig() (*Config, error) {
	reader := bufio.NewReader(os.Stdin)

	config := &Config{
		UpdateInterval: 45, // Стандартное значение
	}

	fmt.Println("Конфигурационный файл не найден. Пожалуйста, введите следующие данные:")

	config.SpotifyClientID = prompt(reader, "Spotify Client ID")
	config.SpotifyClientSecret = prompt(reader, "Spotify Client Secret")
	config.TelegramAPIID = prompt(reader, "Telegram API ID")
	config.TelegramAPIHash = prompt(reader, "Telegram API Hash")
	config.TelegramPhoneNumber = prompt(reader, "Telegram Phone Number (в международном формате, например: +79001234567)")

	fmt.Print("HTTP Proxy (оставьте пустым, если не требуется): ")
	proxy, _ := reader.ReadString('\n')
	proxy = strings.TrimSpace(proxy)
	if proxy != "" {
		config.HttpProxy = proxy
	}

	// Создаем директорию для конфига если её нет
	configPath := getConfigPath()
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	// Сохраняем конфиг
	file, err := os.Create(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create config file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(config); err != nil {
		return nil, fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("Конфигурация сохранена в %s\n", configPath)
	return config, nil
}

func prompt(reader *bufio.Reader, name string) string {
	for {
		fmt.Printf("%s: ", name)
		value, _ := reader.ReadString('\n')
		value = strings.TrimSpace(value)

		if value != "" {
			return value
		}
		fmt.Println("Значение не может быть пустым, попробуйте снова")
	}
}
