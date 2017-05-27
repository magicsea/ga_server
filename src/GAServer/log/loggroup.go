package log

import "log"

type LogGroup struct {
	loggers []*Logger
}

var gGroup *LogGroup

//var gLogger *Logger

func init() {
	gLogger, _ := New("debug", "", "", log.LstdFlags)
	gGroup = new(LogGroup)
	gGroup.loggers = append(gGroup.loggers, gLogger)
}
func NewLogGroup(strLevel string, pathname string, isStdout bool, flag int) error {

	if pathname != "" {
		debug, err := New("debug", "debug_", pathname, flag)
		if err != nil {
			return err
		}
		info, err := New("info", "info_", pathname, flag)
		if err != nil {
			return err
		}
		erro, err := New("error", "error_", pathname, flag)
		if err != nil {
			return err
		}
		fatal, err := New("fatal", "fatal_", pathname, flag)
		if err != nil {
			return err
		}
		gGroup = &LogGroup{[]*Logger{debug, info, erro, fatal}}
	} else {
		gGroup = &LogGroup{}
	}

	if isStdout {
		gLogger, _ := New(strLevel, "", "", log.LstdFlags)
		gGroup.loggers = append(gGroup.loggers, gLogger)
	}
	return nil
}

func Debug(format string, a ...interface{}) {
	for _, v := range gGroup.loggers {
		v.Debug(format, a...)
	}
}

func Info(format string, a ...interface{}) {
	for _, v := range gGroup.loggers {
		v.Info(format, a...)
	}
}

func Println(format string, a ...interface{}) {
	//str := fmt.Sprint(a)
	for _, v := range gGroup.loggers {
		v.Info(format, a...)
	}

}

func Error(format string, a ...interface{}) {
	//fmt.Println("Error  ", len(gGroup.loggers), gGroup.loggers)
	for _, v := range gGroup.loggers {
		v.Error(format, a...)
	}
}

func Fatal(format string, a ...interface{}) {
	for _, v := range gGroup.loggers {
		v.Fatal(format, a...)
	}
}

func Close() {
	for _, v := range gGroup.loggers {
		v.Close()
	}
}
