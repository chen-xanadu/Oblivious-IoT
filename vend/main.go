package main

import (
	"Oblivious-IoT/config"
	"Oblivious-IoT/helper"
	pb "Oblivious-IoT/message"
	"fmt"
	"google.golang.org/grpc"
	"io"
	"math/rand"
	"net"
	"sync"
	"time"
)

type shuffleServer struct {
	pb.UnimplementedShuffleServerServer
}

func (s *shuffleServer) Shuffle(stream pb.ShuffleServer_ShuffleServer) error {
	i := 0
	var responses [config.NumCommands][]byte
	var wg sync.WaitGroup

	sk := helper.ReadSk(config.VendorSkFile)

	perm := rand.Perm(config.NumCommands)

	for {
		request, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		wg.Add(1)
		go func(i int, data []byte) {
			defer wg.Done()
			responses[i] = helper.HybridDecrypt(request.Data, sk)
		}(i, request.Data)

		i += 1

		//fmt.Println(request.Data);
	}

	wg.Wait()
	fmt.Println(time.Now().UnixMilli())

	for _, j := range perm {
		//fmt.Println(i, j)
		err := stream.Send(&pb.ShuffleResponse{Data: responses[j]})
		if err != nil {
			return nil
		}
	}

	return nil
}

func main() {
	lis, err := net.Listen("tcp", "0.0.0.0:50051")
	if err != nil {
		fmt.Printf("failed to listen: %v", err)
	}

	var opts []grpc.ServerOption

	grpcServer := grpc.NewServer(opts...)

	pb.RegisterShuffleServerServer(grpcServer, &shuffleServer{})

	err = grpcServer.Serve(lis)
	if err != nil {
		fmt.Printf("failed to run grpc server: %v", err)
	}
}
