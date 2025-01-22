package telegram

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/tg"
)

type Client struct {
	client      *telegram.Client
	api         *tg.Client
	originalBio string
	phone       string
	ctx         context.Context
	cancel      context.CancelFunc
}

type noSignUp struct {
	phone string
}

func (n *noSignUp) SignUp(ctx context.Context) (auth.UserInfo, error) {
	return auth.UserInfo{}, fmt.Errorf("не поддерживается регистрация новых пользователей")
}

func (n *noSignUp) AcceptTermsOfService(ctx context.Context, tos tg.HelpTermsOfService) error {
	return fmt.Errorf("не поддерживается принятие условий использования")
}

func (n *noSignUp) Phone(_ context.Context) (string, error) {
	return n.phone, nil
}

func (n *noSignUp) Password(_ context.Context) (string, error) {
	return "", nil
}

func (n *noSignUp) Code(_ context.Context, _ *tg.AuthSentCode) (string, error) {
	var code string
	fmt.Print("Введите код из Telegram: ")
	fmt.Scanln(&code)
	return code, nil
}

func NewClient(apiID string, apiHash string, phone string) (*Client, error) {
	// Конвертируем apiID из строки в int
	var appID int
	_, err := fmt.Sscanf(apiID, "%d", &appID)
	if err != nil {
		return nil, fmt.Errorf("invalid API ID format: %w", err)
	}

	// Получаем путь к файлу сессии
	sessionFile := "telegram_session"
	if exePath, err := os.Executable(); err == nil {
		sessionFile = filepath.Join(filepath.Dir(exePath), "telegram_session")
	}

	// Создаем клиент с сохранением сессии
	client := telegram.NewClient(appID, apiHash, telegram.Options{
		SessionStorage: &telegram.FileSessionStorage{
			Path: sessionFile,
		},
	})

	// Создаем контекст с отменой для клиента
	ctx, cancel := context.WithCancel(context.Background())

	return &Client{
		client: client,
		phone:  phone,
		ctx:    ctx,
		cancel: cancel,
	}, nil
}

func (c *Client) Authenticate(ctx context.Context) error {
	log.Println("Начинаем авторизацию в Telegram...")

	// Канал для получения результата авторизации
	authDone := make(chan error, 1)

	// Запускаем клиент в отдельной горутине
	go func() {
		err := c.client.Run(c.ctx, func(ctx context.Context) error {
			// Канал для завершения авторизации
			done := make(chan error, 1)

			go func() {
				flow := auth.NewFlow(
					&noSignUp{phone: c.phone},
					auth.SendCodeOptions{},
				)

				if err := c.client.Auth().IfNecessary(ctx, flow); err != nil {
					done <- fmt.Errorf("ошибка авторизации: %w", err)
					return
				}

				log.Println("Авторизация успешна, получаем API клиент...")
				c.api = c.client.API()

				log.Println("Получаем информацию о пользователе...")
				user, err := c.api.UsersGetFullUser(ctx, &tg.InputUserSelf{})
				if err != nil {
					done <- fmt.Errorf("ошибка получения информации о пользователе: %w", err)
					return
				}

				c.originalBio = user.FullUser.About
				log.Printf("Авторизация завершена, текущая биография: %q", c.originalBio)

				// Сигнализируем об успешной авторизации
				done <- nil
			}()

			// Ждем завершения авторизации или отмены контекста
			select {
			case err := <-done:
				if err != nil {
					return err
				}
				authDone <- nil
			case <-ctx.Done():
				return ctx.Err()
			}

			// Держим соединение открытым
			<-ctx.Done()
			return ctx.Err()
		})

		if err != nil && err != context.Canceled {
			authDone <- fmt.Errorf("ошибка работы клиента: %w", err)
		}
	}()

	// Ждем завершения авторизации
	select {
	case err := <-authDone:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (c *Client) UpdateBio(ctx context.Context, bio string) error {
	_, err := c.api.AccountUpdateProfile(ctx, &tg.AccountUpdateProfileRequest{
		About: bio,
	})
	return err
}

func (c *Client) RestoreOriginalBio(ctx context.Context) error {
	return c.UpdateBio(ctx, c.originalBio)
}

func (c *Client) GetOriginalBio() string {
	return c.originalBio
}

func (c *Client) Close() error {
	if c.cancel != nil {
		c.cancel()
	}
	return nil
}
