package main

import (
	"fmt"
	"idm/inner/database"
	"log"
)

func main() {
	fmt.Println("Hello, Go.")

	// Подключаемся к БД
	db := database.ConnectDb()
	defer db.Close()

	log.Println("DB connected")
}
