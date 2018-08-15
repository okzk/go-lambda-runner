package main

import (
	"errors"
	"fmt"
	"github.com/aws/aws-lambda-go/lambda/messages"
	"golang.org/x/crypto/ssh/terminal"
	"io/ioutil"
	"net/rpc"
	"os"
	"os/exec"
	"syscall"
	"time"
)

func main() {
	if os.Getenv("_LAMBDA_SERVER_PORT") == "" {
		os.Setenv("_LAMBDA_SERVER_PORT", "39999")
	}

	data := []byte("{}")
	if !terminal.IsTerminal(0) {
		var err error
		data, err = ioutil.ReadAll(os.Stdin)
		if err != nil {
			panic(err)
		}
	}

	if err := invoke(data); err != nil {
		panic(err)
	}
}

func invoke(data []byte) error {
	if len(os.Args) <= 1 {
		return errors.New("missing a command")
	}
	cmd := exec.Command(os.Args[1], os.Args[2:]...)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Start()
	if err != nil {
		return err
	}
	defer func() {
		syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		cmd.Wait()
	}()

	client, err := connect()
	if err != nil {
		return err
	}
	defer client.Close()

	req := &messages.InvokeRequest{
		Payload:            data,
		RequestId:          "1",
		XAmznTraceId:       "1",
		Deadline:           messages.InvokeRequest_Timestamp{Seconds: 300, Nanos: 0},
		InvokedFunctionArn: "arn:aws:lambda:ap-northeast-1:000000000000:function:test",
	}
	var response *messages.InvokeResponse
	err = client.Call("Function.Invoke", req, &response)

	if err != nil {
		return err
	}
	if response.Error != nil {
		fmt.Fprintf(os.Stderr, "Error.Type: %s\n", response.Error.Type)
		fmt.Fprintf(os.Stderr, "Error.Message: %s\n", response.Error.Message)
		fmt.Fprintf(os.Stderr, "Error.ShouldExit: %v\n", response.Error.ShouldExit)
		fmt.Fprintf(os.Stderr, "Error.StackTrace: %v\n", response.Error.StackTrace)
	}
	fmt.Println(string(response.Payload))
	return nil
}

func connect() (client *rpc.Client, err error) {
	addr := fmt.Sprintf("localhost:%s", os.Getenv("_LAMBDA_SERVER_PORT"))
	for i := 0; i < 32; i++ {
		time.Sleep(time.Millisecond * 100)
		client, err = rpc.Dial("tcp", addr)
		if err == nil {
			return
		}
	}
	return
}
