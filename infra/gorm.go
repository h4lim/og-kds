package infra

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/pkg/errors"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gen"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var GormDB *gorm.DB

type GormContext struct {
	Driver   string
	Port     string
	Host     string
	Username string
	Password string
	DBName   string
}

type Gorm interface {
	Open() (*gorm.DB, *error)
	UseGormGen() (*gen.Generator, *error)
}

const (
	POSGRES_CONFIG    = "host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Jakarta"
	MYSQL_CONFIG      = "%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local"
	MYSQL_DRIVER      = "mysql"
	POSTGRESQL_DRIVER = "postgres"
)

func NewGormDB(model GormContext) Gorm {
	return GormContext{
		Driver:   model.Driver,
		Port:     model.Port,
		Host:     model.Host,
		Username: model.Username,
		Password: model.Password,
		DBName:   model.DBName,
	}
}

func (g GormContext) Open() (*gorm.DB, *error) {

	db, err := g.openDB()
	if err != nil {
		return nil, err
	}

	GormDB = db

	return db, nil
}

func (g GormContext) UseGormGen() (*gen.Generator, *error) {

	db, err := g.openDB()
	if err != nil {
		return nil, err
	}

	dbGen := gen.NewGenerator(gen.Config{
		OutPath: "./query",
		Mode:    gen.WithoutContext | gen.WithDefaultQuery | gen.WithQueryInterface,
	})

	dbGen.UseDB(db)

	return dbGen, nil
}

func (g GormContext) openDB() (*gorm.DB, *error) {

	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Millisecond,
			LogLevel:                  logger.Silent,
			IgnoreRecordNotFoundError: true,
			ParameterizedQueries:      true,
			Colorful:                  true,
		},
	)

	config := gorm.Config{
		Logger: newLogger,
	}

	switch strings.ToLower(g.Driver) {
	case MYSQL_DRIVER:
		connectionUrl := fmt.Sprintf(
			MYSQL_CONFIG, g.Username, g.Password, g.Host, g.Port, g.DBName,
		)
		db, err := gorm.Open(mysql.Open(connectionUrl), &config)

		if err != nil {
			return nil, &err
		}

		return db, nil
	case POSTGRESQL_DRIVER:
		connectionUrl := fmt.Sprintf(
			POSGRES_CONFIG, g.Host, g.Username, g.Password, g.DBName, g.Port,
		)
		db, err := gorm.Open(postgres.Open(connectionUrl), &config)
		if err != nil {
			return nil, &err
		}

		return db, nil
	default:
		newError := errors.New("invalid db driver")
		return nil, &newError
	}
}
