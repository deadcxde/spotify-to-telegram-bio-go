package spotify

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	spotify "github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2"
)

type Client struct {
	client     *spotify.Client
	auth       *spotifyauth.Authenticator
	httpClient *http.Client
	token      *oauth2.Token
	tokenFile  string
}

const redirectURI = "http://localhost:8080/callback"

func NewClient(clientID, clientSecret string) (*Client, error) {
	// Получаем путь к файлу токена
	tokenFile := "spotify_token"
	if exePath, err := os.Executable(); err == nil {
		tokenFile = filepath.Join(filepath.Dir(exePath), "spotify_token")
	}

	// Пытаемся загрузить существующий токен
	var token *oauth2.Token
	if data, err := os.ReadFile(tokenFile); err == nil {
		if err := json.Unmarshal(data, &token); err == nil {
			log.Println("Загружен существующий токен Spotify")
		}
	}

	httpClient := http.DefaultClient
	auth := spotifyauth.New(
		spotifyauth.WithRedirectURL(redirectURI),
		spotifyauth.WithScopes(
			spotifyauth.ScopeUserReadCurrentlyPlaying,
			spotifyauth.ScopeUserReadPlaybackState,
		),
		spotifyauth.WithClientID(clientID),
		spotifyauth.WithClientSecret(clientSecret),
	)

	return &Client{
		auth:       auth,
		httpClient: httpClient,
		token:      token,
		tokenFile:  tokenFile,
	}, nil
}

func (c *Client) saveToken() error {
	if c.token == nil {
		return nil
	}

	data, err := json.Marshal(c.token)
	if err != nil {
		return fmt.Errorf("failed to marshal token: %w", err)
	}

	if err := os.WriteFile(c.tokenFile, data, 0600); err != nil {
		return fmt.Errorf("failed to save token: %w", err)
	}

	log.Println("Токен Spotify сохранен")
	return nil
}

func (c *Client) Authenticate() error {
	// Если есть валидный токен, используем его
	if c.token != nil && c.token.Valid() {
		client := spotify.New(c.auth.Client(context.Background(), c.token))
		c.client = client
		log.Println("Использован существующий токен Spotify")
		return nil
	}

	ch := make(chan *spotify.Client)
	state := "some-random-state-123"
	serverReady := make(chan bool)
	var serverError error

	mux := http.NewServeMux()
	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Получен callback запрос: %s", r.URL.String())

		code := r.URL.Query().Get("code")
		if code == "" {
			log.Printf("Код авторизации отсутствует в запросе")
			errMsg := r.URL.Query().Get("error")
			if errMsg != "" {
				log.Printf("Получена ошибка от Spotify: %s", errMsg)
			}
			ch <- nil
			return
		}
		log.Printf("Получен код авторизации: %s", code)

		tok, err := c.auth.Token(r.Context(), state, r)
		if err != nil {
			log.Printf("Ошибка получения токена: %v", err)
			http.Error(w, "Не удалось получить токен", http.StatusForbidden)
			ch <- nil
			return
		}
		log.Printf("Токен успешно получен")

		// Сохраняем токен
		c.token = tok
		if err := c.saveToken(); err != nil {
			log.Printf("Ошибка сохранения токена: %v", err)
		}

		// Создаем клиента
		client := spotify.New(c.auth.Client(r.Context(), tok))
		log.Printf("Spotify клиент создан")

		// Проверяем работоспособность клиента
		user, err := client.CurrentUser(r.Context())
		if err != nil {
			log.Printf("Ошибка проверки клиента: %v", err)
			ch <- nil
			return
		}
		log.Printf("Успешная авторизация пользователя: %s", user.ID)

		// Отправляем успешный ответ пользователю
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(`
			<html>
				<head>
					<meta charset="utf-8">
					<title>Spotify Auth</title>
					<style>
						body { font-family: Arial, sans-serif; text-align: center; margin-top: 50px; }
						h1 { color: #1DB954; }
					</style>
				</head>
				<body>
					<h1>Успешная авторизация!</h1>
					<p>Можете закрыть это окно.</p>
				</body>
			</html>
		`))

		log.Printf("Отправляем клиента в канал")
		ch <- client
		log.Printf("Клиент отправлен в канал")

		go func() {
			log.Printf("Останавливаем сервер")
			if err := server.Shutdown(context.Background()); err != nil {
				log.Printf("Ошибка при остановке сервера: %v", err)
			}
			log.Printf("Сервер остановлен")
		}()
	})

	go func() {
		log.Printf("Запуск сервера на порту 8080...")
		serverReady <- true
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Printf("Ошибка сервера: %v", err)
			serverError = err
		}
	}()

	<-serverReady
	if serverError != nil {
		return fmt.Errorf("ошибка запуска сервера: %w", serverError)
	}

	url := c.auth.AuthURL(state)
	log.Printf("Перейдите по ссылке для авторизации в Spotify: %s", url)

	client := <-ch
	if client == nil {
		return fmt.Errorf("не удалось получить Spotify клиент")
	}

	log.Printf("Клиент получен из канала")
	c.client = client
	log.Printf("Авторизация завершена успешно")
	return nil
}

func (c *Client) GetCurrentlyPlaying(ctx context.Context) (*spotify.CurrentlyPlaying, error) {
	return c.client.PlayerCurrentlyPlaying(ctx)
}
