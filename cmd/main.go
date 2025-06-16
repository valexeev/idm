package main

import (
	"context"
	"fmt"
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
)

func main() {
	// 1. Читаем конфигурацию из .env файла или переменных окружения
	var cfg = common.GetConfig(".env")

	// 2. Создаём подключение к базе данных
	var db = database.ConnectDbWithCfg(cfg)

	// 3. Собираем все компоненты приложения (ручная сборка зависимостей)
	var server = build(db, cfg)

	// 4. Запускаем веб-сервер в отдельной горутине
	go func() {
		var err = server.App.Listen(":8080")
		if err != nil {
			fmt.Printf("http server error: %s\n", err)
		}
	}()

	// 5. Создаем канал для ожидания сигнала завершения работы сервера
	var shutdownComplete = make(chan struct{})

	// 6. Запускаем gracefulShutdown в отдельной горутине
	go gracefulShutdown(server, db, shutdownComplete)

	// 7. Ожидаем сигнал от горутины gracefulShutdown, что сервер завершил работу
	<-shutdownComplete
	fmt.Println("Graceful shutdown complete.")
}

// gracefulShutdown - функция "элегантного" завершения работы сервера по сигналу от операционной системы
func gracefulShutdown(server *web.Server, db *sqlx.DB, shutdownComplete chan struct{}) {
	// Уведомить основную горутину о завершении работы
	defer close(shutdownComplete)

	// Создаём контекст, который слушает сигналы прерывания от операционной системы
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT)
	defer stop()

	// Слушаем сигнал прерывания от операционной системы
	<-ctx.Done()
	fmt.Println("shutting down gracefully, press Ctrl+C again to force")

	// Завершаем работу веб-сервера
	if err := server.App.Shutdown(); err != nil {
		fmt.Printf("Server forced to shutdown with error: %v\n", err)
	}

	// Закрываем соединение с базой данных
	if err := db.Close(); err != nil {
		fmt.Printf("error closing db: %v\n", err)
	}

	fmt.Println("Server exiting")
}

// build - главная функция сборки приложения (ручная инъекция зависимостей)
// Здесь мы создаём все компоненты и связываем их друг с другом как матрёшки
func build(db *sqlx.DB, cfg common.Config) *web.Server {
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

	// 3.3 Создаём контроллер, передавая в него сервер и сервис
	var employeeController = employee.NewController(server, employeeService)

	// 3.4 Регистрируем маршруты контроллера
	employeeController.RegisterRoutes()

	//  4. СБОРКА МОДУЛЯ ROLE
	// 4.1 Создаём репозиторий для работы с БД
	var roleRepo = role.NewRepository(db)

	// 4.2 Создаём сервис, передавая в него репозиторий
	var roleService = role.NewService(roleRepo)

	// 4.3 Создаём контроллер, передавая в него сервер и сервис
	var roleController = role.NewController(server, roleService)

	// 4.4 Регистрируем маршруты контроллера
	roleController.RegisterRoutes()

	// 5. СБОРКА МОДУЛЯ INFO (информация о приложении)
	// 5.1 Создаём контроллер, передавая сервер, конфиг и БД напрямую
	var infoController = info.NewController(server, cfg, db)

	// 5.2 Регистрируем маршруты контроллера
	infoController.RegisterRoutes()

	//  6. ВОЗВРАЩАЕМ СОБРАННЫЙ СЕРВЕР
	return server
}
