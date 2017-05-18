package module

import (
	"fmt"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/AsynkronIT/protoactor-go/remote"
)

//进程内Actor
type InprocActorObject struct {
	pid *actor.PID
}
