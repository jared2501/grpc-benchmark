package grpc_benchmark

import (
	"context"
	"crypto/rand"
	"fmt"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"io"
	"net"
	"testing"
	"time"
)

func TestGrpcStream(t *testing.T) {
	s := &server{}
	grpcServer := grpc.NewServer()
	RegisterTestServer(grpcServer, s)

	l, err := net.Listen("tcp", "127.0.0.1:0")
	assert.Nil(t, err)
	port := l.Addr().(*net.TCPAddr).Port

	go grpcServer.Serve(l)

	conn, err := grpc.Dial(fmt.Sprintf("127.0.0.1:%d", port), grpc.WithInsecure())
	assert.Nil(t, err)
	client := NewTestClient(conn)

	bytes := make([]byte, 100)
	_, err = io.ReadFull(rand.Reader, bytes)
	assert.Nil(t, err)

	start := time.Now()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	strm, err := client.Stream(ctx)
	assert.Nil(t, err)
	for i := 0; i < 10000000; i++ {
		msg := Message{
			Bytes: bytes,
		}
		if err := strm.Send(&msg); err != nil {
			panic(err)
		}
	}
	_, err = strm.CloseAndRecv()
	assert.Nil(t, err)
	fmt.Printf("done! %s", time.Since(start).String())
}

func TestGrpcBulkStream(t *testing.T) {
	s := &server{}
	grpcServer := grpc.NewServer()
	RegisterTestServer(grpcServer, s)

	l, err := net.Listen("tcp", "127.0.0.1:0")
	assert.Nil(t, err)
	port := l.Addr().(*net.TCPAddr).Port

	go grpcServer.Serve(l)

	conn, err := grpc.Dial(fmt.Sprintf("127.0.0.1:%d", port), grpc.WithInsecure())
	assert.Nil(t, err)
	client := NewTestClient(conn)

	bytes := make([]byte, 100)
	_, err = io.ReadFull(rand.Reader, bytes)
	assert.Nil(t, err)

	start := time.Now()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	strm, err := client.BulkStream(ctx)
	assert.Nil(t, err)
	for i := 0; i < 1000; i++ {
		var msgs []*Message
		for j := 0; j < 10000; j++ {
			msgs = append(msgs, &Message{Bytes: bytes})
		}
		bulkMsg := &BulkMessage{
			Message: msgs,
		}
		if err := strm.Send(bulkMsg); err != nil {
			panic(err)
		}
	}
	_, err = strm.CloseAndRecv()
	assert.Nil(t, err)
	fmt.Printf("done! %s", time.Since(start).String())
}

type server struct{}

func (s *server) BulkStream(strm Test_BulkStreamServer) error {
	ctx := strm.Context()
	for {
		if _, err := strm.Recv(); err == io.EOF {
			break
		} else if err != nil {
			return err
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
	}
	return strm.SendAndClose(&empty.Empty{})
}

func (s *server) Stream(strm Test_StreamServer) error {
	ctx := strm.Context()
	for {
		if _, err := strm.Recv(); err == io.EOF {
			break
		} else if err != nil {
			return err
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
	}
	return strm.SendAndClose(&empty.Empty{})
}
