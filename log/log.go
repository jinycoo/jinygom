package log

import (
	"os"
	"fmt"
	"time"
	"runtime"
	"path/filepath"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"github.com/jinygo/utils"
	"reflect"
)

const (
	LevelDebug = "debug"
	LevelError = "error"
	LevelWarn  = "warn"

	Stderr  = "stderr"
	File    = "file"
	Console = "console"
)

var (
	JLog *JLogger
	encoderConfig zapcore.EncoderConfig
)

type JLogger struct {
	name        string
	development bool
	core        zapcore.Core
	errorOutput zapcore.WriteSyncer
	addCaller   bool
	addStack    zapcore.LevelEnabler
	callerSkip  int
}

type JLogConfig struct {
	Dev       bool               `json:"dev" yaml:"dev"`
	Level     string             `json:"level" yaml:"level"`
	Encoding  string             `json:"encoding" yaml:"encoding"`
	Encode    map[string]string  `json:"encode" yaml:"encode"`
	Key       map[string]string  `json:"key" yaml:"key"`
	OutPuts   []string           `json:"outputs" yaml:"outputs"`
	LogPath   string             `json:"path" yaml:"path"`
	LogFile   string             `json:"file" yaml:"file"`
	Format    string             `json:"format" yaml:"format"`
}

func DevConfig() *JLogConfig {
	return &JLogConfig{
		Dev:      true,
		Level:    LevelDebug,
		Encoding: Console,
		Encode:   map[string]string{"time": "", "level": "capital", "duration": "string", "caller": "short"},
		Key:      map[string]string{
			"name": "logger",
			"time": "time",
			"level": "level",
			"caller": "caller",
			"message": "msg",
			"stacktrace": "stacktrace",
		},
		OutPuts: []string{Stderr, File},
		LogPath: utils.RootDir(),
		LogFile: "log",
		Format: "2006-01-02",
	}
}

func New(logCfg *JLogConfig) {
	if logCfg == nil {
		logCfg = DevConfig()
	}
	initJLog(logCfg)
}

func initJLog(logConf *JLogConfig) {
	encoderConfig.NameKey       = logConf.Key["name"]
	encoderConfig.TimeKey       = logConf.Key["time"]
	encoderConfig.LevelKey      = logConf.Key["level"]
	encoderConfig.CallerKey     = logConf.Key["caller"]
	encoderConfig.MessageKey    = logConf.Key["message"]
	encoderConfig.StacktraceKey = logConf.Key["stacktrace"]

	encoderConfig.LineEnding = zapcore.DefaultLineEnding

	logConf.timeEncoder()
	logConf.lvlEncoder()
	logConf.durEncoder()
	logConf.callerEncoder()

	var lvl zap.AtomicLevel
	switch logConf.Level {
	case LevelDebug:
		lvl = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	case LevelWarn:
		lvl = zap.NewAtomicLevelAt(zapcore.WarnLevel)
	case LevelError:
		lvl = zap.NewAtomicLevelAt(zapcore.ErrorLevel)
	default:
		lvl = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	}

	var outputs []string
	for _, p := range logConf.OutPuts {
		if p == File {
			filename := fmt.Sprintf("%s_%s.log", logConf.LogFile, time.Now().Format(logConf.Format))
			logFile := filepath.Join(logConf.LogPath, filename)
			outputs = append(outputs, logFile)
		} else {
			outputs = append(outputs, p)
		}
	}

	sink, _, _ := zap.Open(outputs...)

	var cores []zapcore.Core
	switch logConf.Encoding {
	case Console:
		consoleEncoder := zapcore.NewConsoleEncoder(encoderConfig)
		cores = append(cores, zapcore.NewCore(consoleEncoder, sink, lvl))
	case "json":
		jsonEncoder := zapcore.NewJSONEncoder(encoderConfig)
		cores = append(cores, zapcore.NewCore(jsonEncoder, sink, lvl))
	}

	//errSink, _, err := zap.Open("stderr")
	//if err != nil {
	//	closeOut()
	//}

	JLog = &JLogger{
		name:       "jiny-log",
		core:        zapcore.NewTee(cores...),
		development: logConf.Dev,
		errorOutput: zapcore.Lock(os.Stderr),
		addStack:    zapcore.FatalLevel + 1,
		addCaller:   true,
	}
}

func (jlc *JLogConfig) lvlEncoder() {
	lvl := jlc.Encode["level"]
	switch lvl {
	case "capital":
		encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	case "capitalColor":
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	case "color":
		encoderConfig.EncodeLevel = zapcore.LowercaseColorLevelEncoder
	default:
		encoderConfig.EncodeLevel = zapcore.LowercaseLevelEncoder
	}
}

func (jlc *JLogConfig) timeEncoder() {
	encTime := jlc.Encode["time"]
	switch encTime {
	case "iso8601", "ISO8601":
		encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	case "millis":
		encoderConfig.EncodeTime = zapcore.EpochMillisTimeEncoder
	case "nanos":
		encoderConfig.EncodeTime = zapcore.EpochNanosTimeEncoder
	case "localtime":
		encoderConfig.EncodeTime = zapcore.EpochTimeEncoder
	case "unix":
		encoderConfig.EncodeTime = func(i time.Time, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendInt64(i.Local().Unix())
		}
	default:
		encoderConfig.EncodeTime = func(i time.Time, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString(i.Format("2006-01-02 15:04:05"))
		}
	}
}

func (jlc *JLogConfig) durEncoder() {
	dur := jlc.Encode["duration"]
	switch dur {
	case "string":
		encoderConfig.EncodeDuration = zapcore.StringDurationEncoder
	case "nanos":
		encoderConfig.EncodeDuration = zapcore.NanosDurationEncoder
	default:
		encoderConfig.EncodeDuration = zapcore.SecondsDurationEncoder
	}
}

func (jlc *JLogConfig) callerEncoder() {
	caller := jlc.Encode["caller"]
	switch caller {
	case "full":
		encoderConfig.EncodeCaller = zapcore.FullCallerEncoder
	default:
		encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
	}
}

func (jl *JLogger) check(lvl zapcore.Level, msg string) *zapcore.CheckedEntry {
	const callerSkipOffset = 2
	if JLog == nil {
		initJLog(nil)
	}
	ent := zapcore.Entry{
		LoggerName: JLog.name,
		Time:       time.Now().Local(),
		Level:      lvl,
		Message:    msg,
	}
	ce := JLog.core.Check(ent, nil)
	willWrite := ce != nil

	switch ent.Level {
	case zapcore.PanicLevel:
		ce = ce.Should(ent, zapcore.WriteThenPanic)
	case zapcore.FatalLevel:
		ce = ce.Should(ent, zapcore.WriteThenFatal)
	case zapcore.DPanicLevel:
		if JLog.development {
			ce = ce.Should(ent, zapcore.WriteThenPanic)
		}
	}

	if !willWrite {
		return ce
	}

	ce.ErrorOutput = JLog.errorOutput
	if JLog.addCaller {
		ce.Entry.Caller = zapcore.NewEntryCaller(runtime.Caller(JLog.callerSkip + callerSkipOffset))
		if !ce.Entry.Caller.Defined {
			fmt.Fprintf(JLog.errorOutput, "%v Logger.check error: failed to get caller\n", time.Now().Local())
			JLog.errorOutput.Sync()
		}
	}
	if JLog.addStack.Enabled(ce.Entry.Level) {
		ce.Entry.Stack = zap.Stack("").String
	}

	return ce
}

func Debug(details ...interface{}) {
	if ce := JLog.check(zapcore.DebugLevel, fmt.Sprint(details...)); ce != nil {
		ce.Write()
	}
}
func Info(details ...interface{}) {
	if ce := JLog.check(zapcore.InfoLevel, fmt.Sprint(details...)); ce != nil {
		ce.Write()
	}
}
func CInfo(details ...interface{}) {
	if len(details) > 2 {
		if ce := JLog.check(zapcore.InfoLevel, details[0].(string)); ce != nil {
			if reflect.TypeOf(details[1]).String() == "map[string]interface {}" {
				ce.Write(genFields(details[1].(map[string]interface{}))...)
			}
		}
	} else {
		Info(details...)
	}
}
func Warn(details ...interface{}) {
	if ce := JLog.check(zapcore.WarnLevel, fmt.Sprint(details...)); ce != nil {
		ce.Write()
		//dingMessage := make(map[string]interface{})
		//dingMessage["msgtype"] = "text"
		//dingMessage["text"] = map[string]string{
		//	"content": strings.Title(JLog.appName) + " - [" + time.Now().Format("2006-01-02 15:04:05") + "] :: " + fmt.Sprint(details...),
		//}
		//oapi.SendReq(dingMessage, "message/send", "POST", nil)
		//message.PostInfo("message/send", nil, dingMessage)
	}
}
func Error(details ...interface{}) {
	if ce := JLog.check(zapcore.ErrorLevel, fmt.Sprint(details...)); ce != nil {
		ce.Write()
	}
}
func DPanic(details ...interface{}) {
	if ce := JLog.check(zapcore.DPanicLevel, fmt.Sprint(details...)); ce != nil {
		ce.Write()
	}
}
func Panic(details ...interface{}) {
	if ce := JLog.check(zapcore.PanicLevel, fmt.Sprint(details...)); ce != nil {
		ce.Write()
	}
}
func Fatal(details ...interface{}) {
	if ce := JLog.check(zapcore.FatalLevel, fmt.Sprint(details...)); ce != nil {
		ce.Write()
	}
}
func Sync() error {
	return JLog.core.Sync()
}

func genFields(details map[string]interface{}) []zap.Field {
	var fields = make([]zap.Field, 0)
	for k, v := range details {
		switch v.(type) {
		case bool:
			fields = append(fields, zap.Bool(k, v.(bool)))
		case int8:
			fields = append(fields, zap.Int8(k, v.(int8)))
		case int, int32:
			fields = append(fields, zap.Int(k, v.(int)))
		case uint, uint32:
			fields = append(fields, zap.Uint(k, v.(uint)))
		case int64:
			fields = append(fields, zap.Int64(k, v.(int64)))
		case string:
			fields = append(fields, zap.String(k, v.(string)))
		default:
			fields = append(fields, zap.Reflect(k, v))
		}
	}
	return fields
}