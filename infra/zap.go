package infra

import (
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var ZapLog *zap.Logger

type ZapModel struct {
	ServiceName string
	Mode        string
	OutputPath  string
}

type IZapConfig interface {
	ZapSetup() *error
}

func NewZapConfig(model ZapModel) IZapConfig {
	return ZapModel{
		ServiceName: model.ServiceName,
		Mode:        model.Mode,
		OutputPath:  model.OutputPath,
	}
}

func (z ZapModel) ZapSetup() *error {

	wd, err := os.Getwd()
	if err != nil {
		return &err
	}

	zapConfig := zap.NewDevelopmentConfig()

	outputPath := filepath.Join(wd, z.OutputPath, z.ServiceName+".log")
	zapConfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	zapConfig.OutputPaths = []string{"stdout", outputPath}

	zapLog, err := zapConfig.Build()
	if err != nil {
		return &err
	}

	ZapLog = zapLog

	return nil
}
