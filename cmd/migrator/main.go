// package main

// import (
// 	"errors"
// 	"flag"
// 	"fmt"

// 	// Библиотека для миграций
// 	"github.com/golang-migrate/migrate/v4"
// 	// Драйвер для выполнения миграций SQLite 3
// 	_ "github.com/golang-migrate/migrate/v4/database/sqlite3"
// 	// Драйвер для получения миграций из файлов
// 	_ "github.com/golang-migrate/migrate/v4/source/file"
// )

// func main() {
// 	var storagePath, migrationsPath, migrationsTable string

// 	// Получаем необходимые значения из флагов запуска

// 	// Путь до файла БД.
// 	// Его достаточно, т.к. мы используем SQLite, другие креды не нужны.
// 	flag.StringVar(&storagePath, "storage-path", "", "path to storage")
// 	// Путь до папки с миграциями.
// 	flag.StringVar(&migrationsPath, "migrations-path", "", "path to migrations")
// 	// Таблица, в которой будет храниться информация о миграциях. Она нужна
// 	// для того, чтобы понимать, какие миграции уже применены, а какие нет.
// 	// Дефолтное значение - 'migrations'.
// 	flag.StringVar(&migrationsTable, "migrations-table", "migrations", "name of migrations table")
// 	flag.Parse() // Выполняем парсинг флагов

// 	// Валидация параметров
// 	if storagePath == "" {
// 		// Простейший способ обработки ошибки :)
// 		// При необходимости, можете выбрать более подходящий вариант.
// 		// Меня паника пока устраивает, поскольку это вспомогательная утилита.
// 		panic("storage-path is required")
// 	}
// 	if migrationsPath == "" {
// 		panic("migrations-path is required")
// 	}

// 	// Создаем объект мигратора, передав креды нашей БД
// 	m, err := migrate.New(
// 		"file://"+migrationsPath,
// 		fmt.Sprintf("sqlite3://%s?x-migrations-table=%s", storagePath, migrationsTable),
// 	)
// 	if err != nil {
// 		panic(err)
// 	}

// 	// Выполняем миграции до последней версии
// 	if err := m.Up(); err != nil {
// 		if errors.Is(err, migrate.ErrNoChange) {
// 			fmt.Println("no migrations to apply")

// 			return
// 		}

// 		panic(err)
// 	}
// }

package main

import (
	"errors"
	"flag"
	"fmt"

	// Библиотека для миграций
	"github.com/golang-migrate/migrate/v4"
	// Драйвер для выполнения миграций PostgreSQL
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	// Драйвер для получения миграций из файлов
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	var (
		dbHost          string
		dbPort          string
		dbUser          string
		dbPassword      string
		dbName          string
		sslMode         string
		migrationsPath  string
		migrationsTable string
	)

	// Получаем параметры подключения к PostgreSQL
	flag.StringVar(&dbHost, "db-host", "localhost", "database host")
	flag.StringVar(&dbPort, "db-port", "5432", "database port")
	flag.StringVar(&dbUser, "db-user", "", "database user")
	flag.StringVar(&dbPassword, "db-password", "", "database password")
	flag.StringVar(&dbName, "db-name", "", "database name")
	flag.StringVar(&sslMode, "ssl-mode", "disable", "SSL mode (disable, require, verify-full, etc.)")
	flag.StringVar(&migrationsPath, "migrations-path", "", "path to migrations")
	flag.StringVar(&migrationsTable, "migrations-table", "migrations", "name of migrations table")
	flag.Parse()

	// Валидация параметров
	if dbUser == "" {
		panic("db-user is required")
	}
	if dbName == "" {
		panic("db-name is required")
	}
	if migrationsPath == "" {
		panic("migrations-path is required")
	}

	// Формируем DSN (Data Source Name) для PostgreSQL
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s&x-migrations-table=%s",
		dbUser,
		dbPassword,
		dbHost,
		dbPort,
		dbName,
		sslMode,
		migrationsTable,
	)

	// Создаем объект мигратора
	m, err := migrate.New(
		"file://"+migrationsPath,
		dsn,
	)
	if err != nil {
		panic(err)
	}

	// Выполняем миграции до последней версии
	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			fmt.Println("no migrations to apply")
			return
		}

		panic(err)
	}

	fmt.Println("migrations applied successfully")
}
