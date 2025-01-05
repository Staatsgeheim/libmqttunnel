package main

import (
	"context"
	"C"
	"github.com/Staatsgeheim/libmqttunnel"
	"go.uber.org/zap"
)

//export StartTunnel
func StartTunnel(config *C.char, control *C.char, debug C.int) C.int {
	logger := setupLog(C.int(debug) == 1)
	defer logger.Sync()
	undo := zap.ReplaceGlobals(logger)
	defer undo()

	conf, err := libmqttunnel.ReadConfig(C.GoString(config))
	if err != nil {
		return C.int(-1)
	}
	// overwrite control topic if specified.
	strControl := C.GoString(control)
	if strControl != "" {
		conf.Control = strControl
	}

	mqt, err := libmqttunnel.NewMQTunnel(conf)
	if err != nil {
		return C.int(-1)
	}

	local := 0
	remote := 0

	ctx := context.Background()
	mqt.Start(ctx, local, remote)
	return C.int(0)
}

//export ConnectTunnel
func ConnectTunnel(config *C.char, control *C.char, local C.int, remote C.int, debug C.int) C.int {
	logger := setupLog(C.int(debug) == 1)
	defer logger.Sync()
	undo := zap.ReplaceGlobals(logger)
	defer undo()

	conf, err := libmqttunnel.ReadConfig(C.GoString(config))
	if err != nil {
		return C.int(-1)
	}
	// overwrite control topic if specified.
	strControl := C.GoString(control)
	if strControl != "" {
		conf.Control = strControl
	}

	mqt, err := libmqttunnel.NewMQTunnel(conf)
	if err != nil {
		return C.int(-1)
	}
	
	ctx := context.Background()
	mqt.Start(ctx, int(local), int(remote))
	return C.int(0) 
}

func main() {}