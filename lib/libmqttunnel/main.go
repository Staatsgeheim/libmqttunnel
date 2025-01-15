package main

import (
	"context"
	"C"
	"github.com/Staatsgeheim/libmqttunnel"
	"go.uber.org/zap"
)

//export StartTunnelMem
func StartTunnelMem(configBuffer *C.char, control *C.char, logLevel C.int) C.int {
	logger := setupLog(C.int(logLevel))
	defer logger.Sync()
	undo := zap.ReplaceGlobals(logger)
	defer undo()

	config := []byte(C.GoString(configBuffer))
	conf, err := libmqttunnel.ParseConfig(config)
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

//export StartTunnel
func StartTunnel(config *C.char, control *C.char, logLevel C.int) C.int {
	logger := setupLog(C.int(logLevel))
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

//export ConnectTunnelMem
func ConnectTunnelMem(configBuffer *C.char, control *C.char, local C.int, remote C.int, logLevel C.int) C.int {
	logger := setupLog(C.int(logLevel))
	defer logger.Sync()
	undo := zap.ReplaceGlobals(logger)
	defer undo()

	config := []byte(C.GoString(configBuffer))
	conf, err := libmqttunnel.ParseConfig(config)
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

//export ConnectTunnel
func ConnectTunnel(config *C.char, control *C.char, local C.int, remote C.int, logLevel C.int) C.int {
	logger := setupLog(C.int(logLevel))
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