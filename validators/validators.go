package validators

import (
	"context"
	// "encoding/json"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/cosmos/cosmos-sdk/codec"
	queryTypes "github.com/cosmos/cosmos-sdk/types/query"
	validatorTypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// these are so we can mock out the grpc connection for testing
var (
	GrpcDial = grpc.Dial
	ValidatorTypesNewQueryClient = validatorTypes.NewQueryClient
)

// get all validators
func GetValidators(node string) (*validatorTypes.Validators, error) {
	fmt.Println("Getting validators")
	validators := make(validatorTypes.Validators, 0)

	// Create a connection to the gRPC server.
	grpcConn, err := GrpcDial(
		node, // your gRPC server address.
		grpc.WithTransportCredentials(insecure.NewCredentials()), // The Cosmos SDK doesn't support any transport security mechanism.
		// This instantiates a general gRPC codec which handles proto bytes. We pass in a nil interface registry
		// if the request/response types contain interface instead of 'nil' you should pass the application specific codec.
		grpc.WithDefaultCallOptions(grpc.ForceCodec(codec.NewProtoCodec(nil).GRPCCodec())),
	)
	if err != nil {
		return &validators, err
	}

	// this is a hack to make testing work. I'm sure there is a better solution but I had to punt due to time
	if grpcConn != nil {
		defer grpcConn.Close()
	}

	// This creates a gRPC client to query the x/bank service.
	validatorsClient := ValidatorTypesNewQueryClient(grpcConn)
	validatorsResult, err := validatorsClient.Validators(
		context.Background(),
		&validatorTypes.QueryValidatorsRequest{Pagination: &queryTypes.PageRequest{Limit: 1000}},
	)
	if err != nil {
		return &validators, err
	}

	validators = append(validators, validatorsResult.GetValidators()...)

	for validatorsResult.Pagination != nil && validatorsResult.Pagination.NextKey != nil && false {
		validatorsResult, err = validatorsClient.Validators(
			context.Background(),
			&validatorTypes.QueryValidatorsRequest{Pagination: &queryTypes.PageRequest{Limit: 100, Key: validatorsResult.Pagination.NextKey}},
		)
		if err != nil {
			return &validators, err
		}
		validators = append(validators, validatorsResult.GetValidators()...)
	}

	return &validators, nil
}

