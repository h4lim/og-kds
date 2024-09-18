package http

import (
	"fmt"
	"os"
	"time"

	"github.com/h4lim/og-kds/infra"
)

type OptConfigModel struct {
	SqlLogs        bool
	RequestIdAlias string
}

type SqlLog struct {
	ID           uint `gorm:"primarykey" swaggerignore:"true"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
	RequestID    string `db:"request_id"`
	ResponseID   string `db:"response_id"`
	Step         int    `db:"step"`
	Code         string `db:"code"`
	Message      string `db:"message"`
	FunctionName string `db:"function_name"`
	Data         string `db:"data"`
	Duration     string `db:"duration"`
	Tracer       string `db:"tracer"`
}

func SetOptionalConfig(config OptConfigModel) {

	if infra.GormDB != nil && config.SqlLogs {
		if err := infra.GormDB.AutoMigrate(&SqlLog{}); err != nil {
			fmt.Println("error db migrate", err)
			os.Exit(1)
		}
	}

	OptConfig = config
}
