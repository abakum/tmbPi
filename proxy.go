package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	tg "github.com/mymmrac/telego"
	"golang.org/x/net/proxy"
)

// CreateBotWithProxy создает бота с поддержкой прокси из переменных окружения
func CreateBotWithProxy(token string) (*tg.Bot, error) {
	// Проверяем наличие токена
	if token == "" {
		return nil, Errorf(dic.add(ul,
			"en:bot token not specified",
			"ru:токен бота не указан",
		))
	}

	// Проверяем переменные окружения для прокси
	proxyEnv := os.Getenv("TMB_PROXY")
	if proxyEnv == "" {
		proxyEnv = os.Getenv("ALL_PROXY")
	}
	if proxyEnv == "" {
		proxyEnv = os.Getenv("SOCKS5_PROXY")
	}

	// Если прокси не указан, создаем обычного бота
	if proxyEnv == "" {
		fmt.Printf("%s\n", dic.add(ul,
			"en:Proxy not specified, creating bot without proxy",
			"ru:Прокси не указан, создаю бота без прокси",
		))
		return tg.NewBot(token, tg.WithLogger(tg.Logger(Logger{})))
	}

	// Парсим строку прокси (ожидается формат: socks5://127.0.0.1:1080)
	fmt.Printf("%s\n", fmt.Sprintf(dic.add(ul,
		"en:Using proxy: %s",
		"ru:Использую прокси: %s",
	), proxyEnv))

	// Создаем HTTP клиент с прокси
	httpClient, err := createHTTPClientWithProxy(proxyEnv)
	if err != nil {
		return nil, Errorf(dic.add(ul,
			"en:failed to create HTTP client with proxy: %w",
			"ru:ошибка создания HTTP клиента с прокси: %w",
		), err)
	}

	// Создаем бота с кастомным HTTP клиентом
	return tg.NewBot(token,
		tg.WithHTTPClient(httpClient),
		tg.WithLogger(tg.Logger(Logger{})),
	)
}

// createHTTPClientWithProxy создает HTTP клиент с поддержкой SOCKS5 прокси
func createHTTPClientWithProxy(proxyStr string) (*http.Client, error) {
	// Удаляем префикс socks5:// если есть
	proxyStr = strings.TrimPrefix(proxyStr, "socks5://")

	// Разделяем хост и порт с использованием стандартной функции
	host, portStr, err := net.SplitHostPort(proxyStr)
	port := "1080" // порт по умолчанию для SOCKS5
	if err != nil {
		// Если порт не указан, используем хост как есть с портом по умолчанию
		// Убираем квадратные скобки для IPv6 адресов
		host = strings.Trim(proxyStr, "[]")
	} else {
		port = portStr // используем указанный порт
	}
	if host == "" {
		return nil, Errorf(dic.add(ul,
			"en:host not specified",
			"ru:хост не указан",
		))
	}

	// Проверяем порт
	portNum, err := strconv.Atoi(port)
	if err != nil || portNum < 1 || portNum > 65535 {
		return nil, Errorf(dic.add(ul,
			"en:invalid port: %s",
			"ru:неверный порт: %s",
		), port)
	}

	// Создаем SOCKS5 dialer
	dialer, err := proxy.SOCKS5("tcp", net.JoinHostPort(host, port), nil, proxy.Direct)
	if err != nil {
		return nil, Errorf(dic.add(ul,
			"en:failed to create SOCKS5 dialer: %w",
			"ru:ошибка создания SOCKS5 dialer: %w",
		), err)
	}

	// Создаем транспорт с SOCKS5 dialer
	transport := &http.Transport{
		Dial: dialer.Dial,
		// Таймауты для надежности
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			return dialer.Dial(network, addr)
		},
		MaxIdleConns:       100,
		IdleConnTimeout:    90 * time.Second,
		DisableCompression: true,
	}

	// Создаем HTTP клиент
	client := &http.Client{
		Transport: transport,
		Timeout:   30 * time.Second,
	}

	return client, nil
}
