package handler

import (
	"context"
	"time"

	"github.com/google/uuid"
	grpc_opentracing "github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
	ot "github.com/opentracing/opentracing-go"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	pb "github.com/akiliLab/account/proto"
)

// AccountServiceServer : server for Account
type AccountServiceServer struct {
}

var (
	accounts []*pb.Account
)

// Account : Get accounts
func (s *AccountServiceServer) Account(ctx context.Context, req *pb.AccountRequest) (*pb.AccountResponse, error) {
	accounts = nil

	tmpAccount := pb.Account{
		Id:          uuid.New().String(),
		Description: "Peter Pan's Account",
		Created:     time.Now().Local().String(),
	}

	accounts = append(accounts, &tmpAccount)

	// CallGrpcService(ctx, "transcation:50051")
	// CallGrpcService(ctx, "balance:50051")

	return &pb.AccountResponse{
		Account: accounts,
	}, nil
}

// CallGrpcService : connect to gprc
func CallGrpcService(ctx context.Context, address string) {
	conn, err := createGRPCConn(ctx, address)
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	headersIn, _ := metadata.FromIncomingContext(ctx)
	log.Infof("headersIn: %s", headersIn)

	client := pb.NewAccountServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	ctx = metadata.NewOutgoingContext(context.Background(), headersIn)

	defer cancel()

	req := pb.AccountRequest{}
	account, err := client.Account(ctx, &req)
	log.Info(account.GetAccount())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}

	for _, account := range account.GetAccount() {
		accounts = append(accounts, account)
	}
}

func createGRPCConn(ctx context.Context, addr string) (*grpc.ClientConn, error) {
	//https://aspenmesh.io/2018/04/tracing-grpc-with-istio/
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithStreamInterceptor(
		grpc_opentracing.StreamClientInterceptor(
			grpc_opentracing.WithTracer(ot.GlobalTracer()))))
	opts = append(opts, grpc.WithUnaryInterceptor(
		grpc_opentracing.UnaryClientInterceptor(
			grpc_opentracing.WithTracer(ot.GlobalTracer()))))
	opts = append(opts, grpc.WithInsecure())
	conn, err := grpc.DialContext(ctx, addr, opts...)
	if err != nil {
		log.Fatalf("Failed to connect to application addr: ", err)
		return nil, err
	}
	return conn, nil
}
