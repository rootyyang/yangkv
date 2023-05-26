package log

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

func init() {
	registerLog("zaplumberjack", new(ZapLumberjack))
}

var logLevel2zapLevel = [3]zapcore.Level{zap.DebugLevel, zap.WarnLevel, zap.ErrorLevel}

type ZapLumberjack struct {
	zapCore          zapcore.Core
	zapAtomicLevel   zap.AtomicLevel
	zapLogger        *zap.Logger
	zapSugaredLogger *zap.SugaredLogger
}

func (zl *ZapLumberjack) Start(pFileName string, pMaxSizeMB int, pMaxBackupFileNum int, pMaxAgeDay int, pCompress bool) error {
	hook := lumberjack.Logger{
		Filename:   pFileName,
		MaxSize:    pMaxSizeMB,
		MaxBackups: pMaxBackupFileNum,
		MaxAge:     pMaxAgeDay,
		Compress:   pCompress,
	}
	zl.zapAtomicLevel = zap.NewAtomicLevel()
	zl.zapAtomicLevel.SetLevel(zapcore.ErrorLevel)
	encoderConfig := zap.NewProductionEncoderConfig()
	zl.zapCore = zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.AddSync(&hook),
		zl.zapAtomicLevel,
	)
	zl.zapLogger = zap.New(zl.zapCore, zap.AddCaller(), zap.AddCallerSkip(1))
	zl.zapSugaredLogger = zl.zapLogger.Sugar()
	return nil
}

func (zl *ZapLumberjack) Stop() error {
	zl.zapSugaredLogger.Sync()
	return nil
}
func (zl *ZapLumberjack) Sync() error {
	zl.zapSugaredLogger.Sync()
	return nil
}

func (zl *ZapLumberjack) SetLevel(pLevel int) error {
	if pLevel > len(logLevel2zapLevel)-1 || pLevel < 0 {
		return fmt.Errorf("Param Error")
	}
	zl.zapAtomicLevel.SetLevel(logLevel2zapLevel[pLevel])
	return nil
}

func (zl *ZapLumberjack) Debugf(template string, args ...interface{}) {
	zl.zapSugaredLogger.Debugf(template, args...)
}

func (zl *ZapLumberjack) Warnf(template string, args ...interface{}) {
	zl.zapSugaredLogger.Warnf(template, args...)
}

func (zl *ZapLumberjack) Errorf(template string, args ...interface{}) {
	zl.zapSugaredLogger.Errorf(template, args...)
}
