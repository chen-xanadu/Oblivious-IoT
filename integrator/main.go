package main

import (
	"Oblivious-IoT/config"
	"Oblivious-IoT/helper"
	pb "Oblivious-IoT/message"
	"Oblivious-IoT/user"
	"context"
	"flag"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"io"
	"strconv"
	"sync"
	"time"
)

func outsourceShuffle(client pb.ShuffleServerClient) {
	var requests [config.NumCommands][]byte
	for i := 0; i < config.NumCommands; i++ {
		requests[i] = user.GenerateUserRequest(strconv.Itoa(i))
	}

	start := time.Now()

	sk := helper.ReadSk(config.IntegratorSkFile)
	var database [config.NumCommands]*user.UserMessage

	ctx, cancel := context.WithTimeout(context.Background(), 1000*time.Second)
	defer cancel()

	stream, err := client.Shuffle(ctx)
	if err != nil {
		fmt.Printf("failed to start shuffle: %v", err)
	}

	// send data
	for i, data := range requests {
		err = stream.Send(&pb.ShuffleRequest{Data: data})
		if err != nil {
			fmt.Printf("failed to send %d -th data: %v", i, err)
		}
	}
	err = stream.CloseSend()
	if err != nil {
		fmt.Printf("failed to close send: %v", err)
		return
	}

	// recv data
	var wg sync.WaitGroup
	waitc := make(chan struct{})
	go func() {
		i := 0
		for {
			response, err := stream.Recv()
			if err == io.EOF {
				close(waitc)

				return
			}
			if err != nil {
				fmt.Printf("failed to receive %d -th data: %v", i, err)
			}

			//fmt.Println(i)
			wg.Add(1)
			go func(i int, data []byte) {
				defer wg.Done()
				rawMessage := helper.HybridDecrypt(response.Data, sk)
				var m user.UserMessage
				m.Deserialize(rawMessage)
				database[i] = &m
			}(i, response.Data)

			i += 1

		}
	}()
	<-waitc
	wg.Wait()

	duration := time.Since(start)
	fmt.Println(duration)

	//// decryption check

	//for i, m := range database {
	//	devSk := helper.ReadSk(config.DeviceSkFile)
	//
	//	rawCmd, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, devSk, m.Cmd, nil)
	//	if err != nil {
	//		panic(err)
	//	}
	//
	//	buf := bytes.NewBuffer(rawCmd)
	//	dec := gob.NewDecoder(buf)
	//	var cmd user.Command
	//	dec.Decode(&cmd)
	//	fmt.Println(i, cmd)
	//}
}

var (
	serverAddr = flag.String("addr", "localhost:50051", "vendor address host:port")
)

func main() {
	flag.Parse()
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))

	conn, err := grpc.Dial(*serverAddr, opts...)
	if err != nil {
		fmt.Printf("failed to dail: %v", err)
	}
	defer conn.Close()

	client := pb.NewShuffleServerClient(conn)

	outsourceShuffle(client)
}
