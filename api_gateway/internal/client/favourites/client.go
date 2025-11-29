package favourites

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	favv1 "github.com/stpnv0/protos/gen/go/favourites"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

type Client struct {
	api favv1.FavouritesServiceClient
	log *slog.Logger
}

func New(
	ctx context.Context,
	log *slog.Logger,
	addr string,
	timeout time.Duration,
	retriesCount int,
) (*Client, error) {
	const op = "favourites.grpc.New"

	retryOpts := []grpc_retry.CallOption{
		grpc_retry.WithCodes(codes.NotFound, codes.Aborted, codes.DeadlineExceeded),
		grpc_retry.WithMax(uint(retriesCount)),
		grpc_retry.WithPerRetryTimeout(timeout),
	}

	logOpts := []grpc_retry.CallOption{
		grpc_retry.WithCodes(codes.NotFound, codes.Aborted, codes.DeadlineExceeded),
	}

	cc, err := grpc.DialContext(ctx, addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(grpc_middleware.ChainUnaryClient(
			grpc_retry.UnaryClientInterceptor(retryOpts...),
			grpc_retry.UnaryClientInterceptor(logOpts...),
		)),
	)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Client{
		api: favv1.NewFavouritesServiceClient(cc),
		log: log,
	}, nil
}

func (c *Client) AddToFavourites(ctx context.Context, userID, sneakerID int) error {
	const op = "favourites.grpc.AddToFavourites"

	md := metadata.New(map[string]string{
		"user_id": fmt.Sprintf("%d", userID),
	})
	ctx = metadata.NewOutgoingContext(ctx, md)

	_, err := c.api.AddToFavourites(ctx, &favv1.AddToFavouritesRequest{
		UserId:    int32(userID),
		SneakerId: int32(sneakerID),
	})
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (c *Client) RemoveFromFavourites(ctx context.Context, userID, sneakerID int) error {
	const op = "favourites.grpc.RemoveFromFavourites"

	md := metadata.New(map[string]string{
		"user_id": fmt.Sprintf("%d", userID),
	})
	ctx = metadata.NewOutgoingContext(ctx, md)

	_, err := c.api.RemoveFromFavourites(ctx, &favv1.RemoveFromFavouritesRequest{
		UserId:    int32(userID),
		SneakerId: int32(sneakerID),
	})
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (c *Client) GetFavourites(ctx context.Context, userID int) ([]*favv1.FavouriteItem, error) {
	const op = "favourites.grpc.GetFavourites"

	md := metadata.New(map[string]string{
		"user_id": fmt.Sprintf("%d", userID),
	})
	ctx = metadata.NewOutgoingContext(ctx, md)

	resp, err := c.api.GetFavourites(ctx, &favv1.GetFavouritesRequest{
		UserId: int32(userID),
	})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return resp.Items, nil
}

func (c *Client) IsFavourite(ctx context.Context, userID, sneakerID int) (bool, error) {
	const op = "favourites.grpc.IsFavourite"

	md := metadata.New(map[string]string{
		"user_id": fmt.Sprintf("%d", userID),
	})
	ctx = metadata.NewOutgoingContext(ctx, md)

	resp, err := c.api.IsFavourite(ctx, &favv1.IsFavouriteRequest{
		UserId:    int32(userID),
		SneakerId: int32(sneakerID),
	})
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}

	return resp.IsFavourite, nil
}
