package main

import (
	"context"
	"crypto/tls"
	"idm/docs"
	"idm/inner/common"
	"idm/inner/common/validator"
	"idm/inner/database"
	"idm/inner/employee"
	"idm/inner/info"
	"idm/inner/role"
	"idm/inner/web"
	"os/signal"
	"syscall"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

// @title IDM API documentation
// @BasePath /api/v1/

func main() {
	// 1. Читаем конфигурацию из .env файла или переменных окружения
	var cfg = common.GetConfig(".env")

	// Переопределяем версию приложения, которая будет отображаться в swagger UI.
	// Пакет docs и структура SwaggerInfo в нём появятся поле генерации документации (см. далее).
	docs.SwaggerInfo.Version = cfg.AppVersion

	// 2. Создаём логгер
	var logger = common.NewLogger(cfg)
	// Отложенный вызов записи сообщений из буфера в лог. Необходимо вызывать перед выходом из приложения
	defer func() { _ = logger.Sync() }()

	// 3. Создаём подключение к базе данных
	var db = database.ConnectDbWithCfg(cfg)

	// 4. Собираем все компоненты приложения (ручная сборка зависимостей)
	var server = build(db, cfg, logger)

	// 5. Запускаем веб-сервер в отдельной горутине
	go func() {
		// загружаем сертификаты
		cer, err := tls.LoadX509KeyPair(cfg.SslSert, cfg.SslKey)
		if err != nil {
			logger.Panic("failed certificate loading: %s", zap.Error(err))
		}
		// создаём конфигурацию TLS сервера
		tlsConfig := &tls.Config{Certificates: []tls.Certificate{cer}}
		// создаём слушателя https соединения
		ln, err := tls.Listen("tcp", ":8080", tlsConfig)
		if err != nil {
			logger.Panic("failed TLS listener creating: %s", zap.Error(err))
		}
		if err := server.App.Listener(ln); err != nil {
			// паникуем через метод логгера
			logger.Panic("http server error", zap.Error(err))
		}
	}()

	// 6. Создаем канал для ожидания сигнала завершения работы сервера
	var shutdownComplete = make(chan struct{})

	// 7. Запускаем gracefulShutdown в отдельной горутине
	go gracefulShutdown(server, db, logger, shutdownComplete)

	// 8. Ожидаем сигнал от горутины gracefulShutdown, что сервер завершил работу
	<-shutdownComplete
	// все события логируем через общий логгер
	logger.Info("graceful shutdown complete")
}

// gracefulShutdown - функция "элегантного" завершения работы сервера по сигналу от операционной системы
func gracefulShutdown(server *web.Server, db *sqlx.DB, logger *common.Logger, shutdownComplete chan struct{}) {
	// Уведомить основную горутину о завершении работы
	defer close(shutdownComplete)

	// Создаём контекст, который слушает сигналы прерывания от операционной системы
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT)
	defer stop()

	// Слушаем сигнал прерывания от операционной системы
	<-ctx.Done()
	// заменили отладочную печать на логирование
	logger.Info("shutting down gracefully, press Ctrl+C again to force")

	// Завершаем работу веб-сервера
	if err := server.App.Shutdown(); err != nil {
		// Запись ошибки в лог
		logger.Error("Server forced to shutdown with error", zap.Error(err))
	}

	// Закрываем соединение с базой данных
	if err := db.Close(); err != nil {
		logger.Error("error closing db", zap.Error(err))
	}

	logger.Info("Server exiting")
}

// build - главная функция сборки приложения (ручная инъекция зависимостей)
// Здесь мы создаём все компоненты и связываем их друг с другом как матрёшки
// передаём сюда логгер и конфиг
func build(db *sqlx.DB, cfg common.Config, logger *common.Logger) *web.Server {
	//  1. СОЗДАЁМ ВЕБ-СЕРВЕР (самая большая "матрёшка")
	var server = web.NewServer()

	//  2. СОЗДАЁМ ОБЩИЕ КОМПОНЕНТЫ
	// Валидатор для проверки входящих данных
	var vld = validator.New()

	//  3. СБОРКА МОДУЛЯ EMPLOYEE
	// 3.1 Создаём репозиторий для работы с БД
	var employeeRepo = employee.NewRepository(db)

	// 3.2 Создаём сервис, передавая в него репозиторий и валидатор
	var employeeService = employee.NewService(employeeRepo, vld)

	// 3.3 Создаём контроллер, передавая в него сервер, сервис и логгер
	var employeeController = employee.NewController(server, employeeService, logger)

	// 3.4 Регистрируем маршруты контроллера
	employeeController.RegisterRoutes()

	//  4. СБОРКА МОДУЛЯ ROLE
	// 4.1 Создаём репозиторий для работы с БД
	var roleRepo = role.NewRepository(db)

	// 4.2 Создаём сервис, передавая в него репозиторий
	var roleService = role.NewService(roleRepo, vld)

	// 4.3 Создаём контроллер, передавая в него сервер и сервис
	// Если в roleController тоже нужен логгер, то добавьте его в NewController
	var roleController = role.NewController(server, roleService)

	// 4.4 Регистрируем маршруты контроллера
	roleController.RegisterRoutes()

	// 5. СБОРКА МОДУЛЯ INFO (информация о приложении)
	// 5.1 Создаём контроллер, передавая сервер, конфиг и БД напрямую
	// Если в infoController тоже нужен логгер, то добавьте его в NewController
	var infoController = info.NewController(server, cfg, db)

	// 5.2 Регистрируем маршруты контроллера
	infoController.RegisterRoutes()

	//  6. ВОЗВРАЩАЕМ СОБРАННЫЙ СЕРВЕР
	return server
}
