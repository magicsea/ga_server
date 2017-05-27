package util

import (
	"runtime"

	"GAServer/log"
	"bytes"
	"fmt"
	//"github.com/davecgh/go-spew/spew"
)

// 产生panic时的调用栈打印
func PrintPanicStack(extras ...interface{}) {
	var buff bytes.Buffer
	var haveErr = false
	if x := recover(); x != nil {
		haveErr = true
		buff.WriteString(fmt.Sprintf("dump:%v\n", x))
		//log.Error("dump:%v", x)
		i := 0
		funcName, file, line, ok := runtime.Caller(i)
		for ok {
			buff.WriteString(fmt.Sprintf("frame %v:[func:%v,file:%v,line:%v]\n", i, runtime.FuncForPC(funcName).Name(), file, line))
			//log.Error("frame %v:[func:%v,file:%v,line:%v]\n", i, runtime.FuncForPC(funcName).Name(), file, line)
			i++
			funcName, file, line, ok = runtime.Caller(i)
		}

		//for k := range extras {
		//	buff.WriteString(fmt.Sprintf("EXRAS#%v DATA:%v\n", k, spew.Sdump(extras[k])))
		//}
	}
	if haveErr {
		log.Error(buff.String())
	}

}
