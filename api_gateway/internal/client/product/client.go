package product

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	grpclog "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	grpcretry "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/retry"
	productv1 "github.com/stpnv0/protos/gen/go/product"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	api productv1.ProductClient
	log *slog.Logger
}

func New(ctx context.Context, log *slog.Logger, addr string, timeout time.Duration, retriesCount int) (*Client, error) {
	const op = "product.New"

	retryOpts := []grpcretry.CallOption{
		grpcretry.WithCodes(codes.Aborted, codes.DeadlineExceeded, codes.Unavailable),
		grpcretry.WithMax(uint(retriesCount)),
		grpcretry.WithPerRetryTimeout(timeout),
	}

	logOpts := []grpclog.Option{
		grpclog.WithLogOnEvents(grpclog.PayloadReceived, grpclog.PayloadSent),
	}

	cc, err := grpc.DialContext(ctx, addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(
			grpclog.UnaryClientInterceptor(InterceptorLogger(log), logOpts...),
			grpcretry.UnaryClientInterceptor(retryOpts...),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Client{
		api: productv1.NewProductClient(cc),
		log: log,
	}, nil
}

func InterceptorLogger(l *slog.Logger) grpclog.Logger {
	return grpclog.LoggerFunc(func(ctx context.Context, level grpclog.Level, msg string, fields ...any) {
		l.Log(ctx, slog.Level(level), msg, fields...)
	})
}

func (c *Client) GetAllSneakers(ctx context.Context, limit, offset uint64) ([]*productv1.Sneaker, error) {
	const op = "product.GetAllSneakers"

	req := &productv1.GetAllSneakersRequest{
		Limit:  limit,
		Offset: offset,
	}
	resp, err := c.api.GetAllSneakers(ctx, req)
	if err != nil {
		c.log.Error("failed to get all sneakers", slog.String("error", err.Error()))
		return nil, err
	}
	return resp.GetSneakers(), nil
}

func (c *Client) GetSneakerByID(ctx context.Context, id int64) (*productv1.Sneaker, error) {
	const op = "product.GetSneakerByID"

	req := &productv1.GetSneakerByIDRequest{
		Id: id,
	}
	resp, err := c.api.GetSneakerByID(ctx, req)
	if err != nil {
		c.log.Error("failed to get sneaker by id", slog.String("error", err.Error()))
		return nil, err
	}
	return resp, nil
}

func (c *Client) AddSneaker(ctx context.Context, title string, price float32) (*productv1.Sneaker, error) {
	const op = "product.AddSneaker"

	req := &productv1.AddSneakerRequest{
		Title: title,
		Price: price,
	}
	resp, err := c.api.AddSneaker(ctx, req)
	if err != nil {
		c.log.Error("failed to add sneaker", slog.String("error", err.Error()))
		return nil, err
	}

	return resp, nil
}

func (c *Client) DeleteSneaker(ctx context.Context, id int64) error {
	const op = "product.DeleteSneaker"

	req := &productv1.DeleteSneakerRequest{
		Id: id,
	}
	_, err := c.api.DeleteSneaker(ctx, req)
	if err != nil {
		c.log.Error("failed to delete sneaker", slog.String("error", err.Error()))
		return err
	}

	return nil
}

func (c *Client) GenerateUploadURL(ctx context.Context, originalFilename, contentType string) (*productv1.GenerateUploadURLResponse, error) {
	const op = "product.GenerateUploadURL"

	req := &productv1.GenerateUploadURLRequest{
		OriginalFilename: originalFilename,
		ContentType:      contentType,
	}

	resp, err := c.api.GenerateUploadURL(ctx, req)
	if err != nil {
		c.log.Error("failed to generate upload URL", slog.String("error", err.Error()))
		return nil, err
	}

	return resp, nil
}

func (c *Client) UpdateProductImage(ctx context.Context, productID int64, imageKey string) error {
	const op = "product.UpdateProductImage"

	req := &productv1.UpdateProductImageRequest{
		ProductId: productID,
		ImageKey:  imageKey,
	}
	_, err := c.api.UpdateProductImage(ctx, req)
	if err != nil {
		c.log.Error("failed to update product image", slog.String("error", err.Error()))
		return err
	}
	return nil
}

func (c *Client) GetSneakersByIDs(ctx context.Context, ids []int64) ([]*productv1.Sneaker, error) {
	const op = "product.GetSneakersByIDs"

	req := &productv1.GetSneakersByIDsRequest{Ids: ids}
	resp, err := c.api.GetSneakersByIDs(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("GetSneakersByIDs gRPC call failed: %w", err)
	}
	return resp.Sneakers, nil
}
