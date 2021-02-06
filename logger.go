package fmp

import (
	"os"

	. "github.com/phachon/go-logger"
)

var l = newLogger()

func mkdir(dir string) error {
	_, er := os.Stat(dir)
	b := er == nil || os.IsExist(er)
	if !b {
		if err := os.MkdirAll(dir, 0775); err != nil {
			if os.IsPermission(err) {
				return err
			}
		}
	}
	return nil
}

func newLogger() *Logger {
	l := NewLogger()
	_ = l.Detach("console")
	format := "[%timestamp_format%] [%level_string%] [%file%:%line%]\n-->> %body%\n"

	// 命令行输出配置
	consoleConfig := &ConsoleConfig{
		Color:      true,   // 命令行输出字符串是否显示颜色
		JsonFormat: false,  // 命令行输出字符串是否格式化
		Format:     format, // 如果输出的不是 json 字符串，JsonFormat: false, 自定义输出的格式
	}
	// 添加 console 为 logger 的一个输出
	_ = l.Attach("console", LOGGER_LEVEL_DEBUG, consoleConfig)

	// 文件输出配置
	logDir := "logs"
	if err := mkdir(logDir); err != nil {
		panic(err)
	}
	fileConfig := &FileConfig{
		// 如果要将单独的日志分离为文件，请配置LealFrimeNem参数。
		LevelFileName: map[int]string{
			l.LoggerLevel("info"):    logDir + "/info.log",    // Error 级别日志被写入 error .log 文件
			l.LoggerLevel("warning"): logDir + "/warning.log", // Error 级别日志被写入 error .log 文件
			l.LoggerLevel("error"):   logDir + "/error.log",   // Error 级别日志被写入 error .log 文件
		},
		MaxSize:    1024 * 1024, // 文件最大值（KB），默认值0不限
		MaxLine:    100000,      // 文件最大行数，默认 0 不限制
		DateSlice:  "d",         // 文件根据日期切分， 支持 "Y" (年), "m" (月), "d" (日), "H" (时), 默认 "no"， 不切分
		JsonFormat: false,       // 写入文件的数据是否 json 格式化
		Format:     format,      // 如果写入文件的数据不 json 格式化，自定义日志格式
	}
	// 添加 file 为 logger 的一个输出
	_ = l.Attach("file", LOGGER_LEVEL_DEBUG, fileConfig)

	return l
}
