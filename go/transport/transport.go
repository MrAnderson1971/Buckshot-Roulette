package transport

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"sync/atomic"
	"time"
)

const (
	DISCOVERY_PORT = 0x60D
	PORT           = 0xDEA
	BUFFER_SIZE    = 1024
)

var GameOver = make(chan string)

type DiscoveryBroadcast struct {
	name string
	stop atomic.Bool
}

type RPCRequest struct {
	Method string `json:"method"`
	Params []byte `json:"params"`
}

type RPCResponse struct {
	Result []byte `json:"result"`
	Error  string `json:"error"`
}

func Register(name string, stub func([]byte) ([]byte, error)) {
	methods[name] = stub
}

func ClientStub[R any](funcName string, args any) (result R) {
	argData, err := json.Marshal(args)
	if err != nil {
		panic(err)
	}

	Api(RPCRequest{funcName, argData}, address)
	return
}

func ServerStub[T any](argData []byte, method func(T) any) (resData []byte, err error) {
	var args T
	json.Unmarshal(argData, &args)
	res := method(args)
	resData, err = json.Marshal(res)
	return
}

func (d *DiscoveryBroadcast) Start(name string, port int) {
	d.name = name
	d.stop.Store(false)
	go func() {
		conn, err := net.Dial("udp", fmt.Sprintf("255.255.255.255:%d", DISCOVERY_PORT))
		if err != nil {
			panic(fmt.Sprintf("Error %s", err))
		}
		defer conn.Close()
		for !d.stop.Load() {
			message := fmt.Sprintf("BUCKSHOT_ROULETTE:%s:%d\n", d.name, port)
			_, err = conn.Write([]byte(message))
			if err != nil {
				panic(fmt.Sprintf("Error %s", err))
			}
			time.Sleep(2 * time.Second)
		}
	}()
}

func (d *DiscoveryBroadcast) Close() {
	d.stop.Store(true)
}

func DiscoverHost() (ip string, port string, hostName string) {
	addr := net.UDPAddr{
		Port: DISCOVERY_PORT,
		IP:   net.IPv4zero,
	}
	udpConn, err := net.ListenUDP("udp4", &addr)
	if err != nil {
		panic(err)
	}
	defer udpConn.Close()
	err = udpConn.SetReadDeadline(time.Now().Add(10 * time.Second))
	if err != nil {
		panic(err)
	}

	buffer := make([]byte, BUFFER_SIZE)

	err = udpConn.SetReadDeadline(time.Now().Add(10 * time.Second))
	if err != nil {
		panic(err)
	}
	for {
		n, addr, err := udpConn.ReadFromUDP(buffer)
		if err != nil {
			panic(err)
		}
		message := string(buffer[:n])
		if strings.HasPrefix(message, "BUCKSHOT_ROULETTE:") {
			parts := strings.Split(message, ":")
			if len(parts) == 3 {
				hostName = parts[1]
				port = parts[2]
				ip = addr.IP.String()
				fmt.Printf("Discovered game hosted by %s at %s:%s\n", hostName, addr.IP, port)
				return
			}
		}
	}
}

// Initiate new call; return result
func Call(payload *bytes.Buffer, to net.Addr) (result *bytes.Buffer) {
	client := &http.Client{Timeout: 5 * time.Second}
	body, err := client.Post("http://"+to.String(), "application/json", payload)
	if err != nil {
		panic(err)
	}
	defer body.Body.Close()

	if body.StatusCode != http.StatusOK {
		panic(errors.New(body.Status))
	}

	result = &bytes.Buffer{}
	if _, err = io.Copy(result, body.Body); err != nil {
		panic(err)
	}
	return
}

func createError(message string) []byte {
	fmt.Println("Error: " + message)
	buf := bytes.NewBuffer(nil)
	json.NewEncoder(buf).Encode(RPCResponse{Error: message})
	return buf.Bytes()
}

func handleCall(msg *bytes.Buffer) []byte {
	var request RPCRequest
	if err := json.NewDecoder(msg).Decode(&request); err != nil {
		return createError(err.Error())
	}

	method, exists := methods[request.Method]
	if !exists {
		return createError("Invalid method: " + request.Method)
	}
	output, err := method(request.Params)
	if err != nil {
		return createError(err.Error())
	}
	buf := bytes.NewBuffer(nil)
	if err = json.NewEncoder(buf).Encode(RPCResponse{output, ""}); err != nil {
		return createError(err.Error())
	}
	return buf.Bytes()
}

// Start listening for incoming calls
func Listen(ctx context.Context, listener net.Listener) {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		var req bytes.Buffer
		if _, err := io.Copy(&req, r.Body); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		output := handleCall(&req)
		if _, err := w.Write(output); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	fmt.Println("Listening on " + listener.Addr().String())

	server := &http.Server{}
	go func() {
		if err := server.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			panic(err)
		}
	}()

	<-ctx.Done()
	server.Shutdown(context.Background())
}

func Api(request RPCRequest, address string) []byte {
	var buf2 bytes.Buffer
	err := json.NewEncoder(&buf2).Encode(request)
	if err != nil {
		panic(err)
	}

	addr, err := net.ResolveTCPAddr("tcp", address)
	if err != nil {
		panic(err)
	}

	result := Call(&buf2, addr)

	var response RPCResponse
	err = json.NewDecoder(result).Decode(&response)
	if err != nil {
		panic(err)
	}

	if response.Error != "" {
		panic(errors.New(response.Error))
	}

	return response.Result
}

func Bind(addr string) {
	fmt.Println("Other IP is: " + addr)
	address = addr
}

var methods = make(map[string]func([]byte) ([]byte, error))
var address string
