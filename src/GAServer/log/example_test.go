package log_test

import (
	"GAServer/log"
	"testing"
)

func TestLog(t *testing.T) {
	name := "Leaf"
	log.NewLogGroup("debug", "ss", true, 0)

	log.Debug("My name is %v", name)
	log.Info("My name is %v", name)
	log.Error("My name is %v", name)
	//log.Fatal("My name is %v", name)

	/*
		logger, err := log.New("release", "", l.LstdFlags)
		if err != nil {
			return
		}
		defer logger.Close()

		logger.Debug("will not print")
		logger.Info("My name is %v", name)
	*/
	//log.Export(logger)

	//log.Debug("will not print")
	//log.Info("My name is %v", name)
}
