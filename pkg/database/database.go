package database

type Database interface {
	Connect(dsn string) error
	Close() error
}
