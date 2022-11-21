package delegations

import (
	"context"
	"fmt"
	big "math/big"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/cosmos/cosmos-sdk/codec"

	queryTypes "github.com/cosmos/cosmos-sdk/types/query"
	delegationTypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// these are so we can mock out the grpc connection for testing
var (
	GrpcDial = grpc.Dial
	DelegationTypesNewQueryClient = delegationTypes.NewQueryClient
)

// this will hold not only the delegation responses for a delegation addr but also the total delegation amount
type DelegationResponsesWithTotalBalance struct {
	DelegationResponses delegationTypes.DelegationResponses
	TotalBalance        *big.Int
}

// a map to hold the delegations for each delegator by address
type DelegationsWithTotalBalance map[string]DelegationResponsesWithTotalBalance

// collect all delegation responses for all validators
// NOTE: This has to be done in sequentially because cosmos gRPC does not support parallel queries
// or batching. This is a known issue and will be fixed in the future.
// See: https://github.com/cosmos/cosmos-sdk/issues/8591 and
// https://github.com/terra-money/classic-core/issues/694 for discussions on this issue
func GetDelegationResponses(node string, validators *delegationTypes.Validators) (*delegationTypes.DelegationResponses, error) {
	delegationResponses := delegationTypes.DelegationResponses{}

	// Create a connection to the gRPC server.
	grpcConn, err := GrpcDial(
		node, // your gRPC server address.
		grpc.WithTransportCredentials(insecure.NewCredentials()), // The Cosmos SDK doesn't support any transport security mechanism.
		// This instantiates a general gRPC codec which handles proto bytes. We pass in a nil interface registry
		// if the request/response types contain interface instead of 'nil' you should pass the application specific codec.
		grpc.WithDefaultCallOptions(grpc.ForceCodec(codec.NewProtoCodec(nil).GRPCCodec())),
	)
	if err != nil {
		return &delegationResponses, err
	}

	// this is a hack to make testing work. I'm sure there is a better solution but I had to punt due to time
	if grpcConn != nil {
		defer grpcConn.Close()
	}
	delegationResponsesClient := DelegationTypesNewQueryClient(grpcConn)

	for _, validator := range *validators {
		fmt.Println("Getting delegation responses for validator:", validator.Description.Moniker)

		delegationResponsesResult, err := delegationResponsesClient.ValidatorDelegations(
			context.Background(),
			&delegationTypes.QueryValidatorDelegationsRequest{
				ValidatorAddr: validator.OperatorAddress,
				Pagination:    &queryTypes.PageRequest{Limit: 10000},
			},
		)
		if err != nil {
			return &delegationResponses, err
		}

		delegationResponses = append(delegationResponses, delegationResponsesResult.DelegationResponses...)

		for delegationResponsesResult.Pagination != nil && delegationResponsesResult.Pagination.NextKey != nil {
			delegationResponsesResult, err = delegationResponsesClient.ValidatorDelegations(
				context.Background(),
				&delegationTypes.QueryValidatorDelegationsRequest{
					ValidatorAddr: validator.OperatorAddress,
					Pagination:    &queryTypes.PageRequest{Limit: 10000, Key: delegationResponsesResult.Pagination.NextKey},
				},
			)
			if err != nil {
				return &delegationResponses, err
			}

			delegationResponses = append(delegationResponses, delegationResponsesResult.DelegationResponses...)
		}
	}

	return &delegationResponses, nil
}

func GetDelegationsWithTotalBalance(delegationResponses *delegationTypes.DelegationResponses) *DelegationsWithTotalBalance {
	fmt.Println("Collecting delegations")
	delegationsMap := make(DelegationsWithTotalBalance)

	for _, delegationResponse := range *delegationResponses {
		delegation := delegationResponse.Delegation

		// if the delegator is already in the map, then we need to add the delegation amount to the existing delegation
		if delegationWithBalance, ok := delegationsMap[delegation.DelegatorAddress]; ok {
			delegationWithBalance.TotalBalance = delegationWithBalance.TotalBalance.Add(
				delegationWithBalance.TotalBalance, delegationResponse.Balance.Amount.BigInt())
			delegationWithBalance.DelegationResponses = append(delegationWithBalance.DelegationResponses, delegationResponse)
			delegationsMap[delegation.DelegatorAddress] = delegationWithBalance
		} else {
			delegationWithBalance = DelegationResponsesWithTotalBalance{TotalBalance: delegationResponse.Balance.Amount.BigInt()}

			delegationWithBalance.DelegationResponses = append(delegationWithBalance.DelegationResponses, delegationResponse)
			delegationsMap[delegation.DelegatorAddress] = delegationWithBalance
		}
	}

	return &delegationsMap
}