package delegations

import (
	"context"
	"encoding/csv"
	"fmt"
	"log"
	big "math/big"
	"sort"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/cosmos/cosmos-sdk/codec"

	queryTypes "github.com/cosmos/cosmos-sdk/types/query"
	delegationTypes "github.com/cosmos/cosmos-sdk/x/staking/types"
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
	delegationResponses := make(delegationTypes.DelegationResponses, 0)

	// Create a connection to the gRPC server.
	grpcConn, err := grpc.Dial(
		node, // your gRPC server address.
		grpc.WithTransportCredentials(insecure.NewCredentials()), // The Cosmos SDK doesn't support any transport security mechanism.
		// This instantiates a general gRPC codec which handles proto bytes. We pass in a nil interface registry
		// if the request/response types contain interface instead of 'nil' you should pass the application specific codec.
		grpc.WithDefaultCallOptions(grpc.ForceCodec(codec.NewProtoCodec(nil).GRPCCodec())),
	)
	if err != nil {
		return &delegationResponses, err
	}
	defer grpcConn.Close()
	DelegationResponsesClient := delegationTypes.NewQueryClient(grpcConn)

	for _, validator := range *validators {
		fmt.Println("Getting delegation responses for validator:", validator.Description.Moniker)

		DelegationResponsesResult, err := DelegationResponsesClient.ValidatorDelegations(
			context.Background(),
			&delegationTypes.QueryValidatorDelegationsRequest{
				ValidatorAddr: validator.OperatorAddress,
				Pagination:    &queryTypes.PageRequest{Limit: 10000},
			},
		)
		if err != nil {
			return &delegationResponses, err
		}

		delegationResponses = append(delegationResponses, DelegationResponsesResult.DelegationResponses...)

		for DelegationResponsesResult.Pagination.NextKey != nil {
			DelegationResponsesResult, err = DelegationResponsesClient.ValidatorDelegations(
				context.Background(),
				&delegationTypes.QueryValidatorDelegationsRequest{
					ValidatorAddr: validator.OperatorAddress,
					Pagination:    &queryTypes.PageRequest{Limit: 10000, Key: DelegationResponsesResult.Pagination.NextKey},
				},
			)
			if err != nil {
				return &delegationResponses, err
			}

			delegationResponses = append(delegationResponses, DelegationResponsesResult.DelegationResponses...)
		}

		break
	}

	return &delegationResponses, nil
}

func GetDelegationsWithTotalBalance(delegationResponses *delegationTypes.DelegationResponses) *DelegationsWithTotalBalance {
	fmt.Println("Collecting delegations")
	delegationsMap := make(DelegationsWithTotalBalance)

	for _, delegationResponse := range *delegationResponses {
		delegation := delegationResponse.Delegation
		fmt.Println(delegation)

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

// writes delegations sorted by voting power to a csv file
func WriteDelegations(delegationsMap *DelegationsWithTotalBalance, delegationResponses *delegationTypes.DelegationResponses,
	writer *csv.Writer,
) {
	fmt.Println("Writing delegations to csv file")

	fmt.Println("Sorting delegations")
	sort.Slice(*delegationResponses, func(i, j int) bool {
		return (*delegationsMap)[(*delegationResponses)[i].Delegation.DelegatorAddress].TotalBalance.Cmp(
			(*delegationsMap)[(*delegationResponses)[j].Delegation.DelegatorAddress].TotalBalance) == 1
	})

	writer.Write([]string{"delegator", "voting_power"})

	fmt.Println("Final step, writing to csv")
	for _, delegationResponse := range *delegationResponses {
		strDelegaton := make([]string, 0)
		strDelegaton = append(strDelegaton, delegationResponse.Delegation.DelegatorAddress)
		strDelegaton = append(strDelegaton, fmt.Sprint((*delegationsMap)[delegationResponse.Delegation.DelegatorAddress].TotalBalance))

		writer.Write(strDelegaton)
	}

	writer.Flush()

	if err := writer.Error(); err != nil {
		log.Fatal(err)
	}
}

// writes delegations who are delegated to multiple validators
func WriteMultipleDelegations(delegationsMap *DelegationsWithTotalBalance, delegationResponses *delegationTypes.DelegationResponses,
	writer *csv.Writer,
) {
	fmt.Println("Writing multiple delegations to csv file")

	writer.Write([]string{"delegator", "validator", "bonded_token"})

	for _, delegationWithTotalBalance := range *delegationsMap {
		delegationResponses := delegationWithTotalBalance.DelegationResponses

		if len(delegationResponses) > 1 {
			for _, delegationResponse := range delegationWithTotalBalance.DelegationResponses {
				strDelegaton := make([]string, 0)
				strDelegaton = append(strDelegaton, delegationResponse.Delegation.DelegatorAddress)
				strDelegaton = append(strDelegaton, delegationResponse.Delegation.ValidatorAddress)
				strDelegaton = append(strDelegaton, delegationResponse.Balance.Denom)

				writer.Write(strDelegaton)
			}
		}
	}

	writer.Flush()

	if err := writer.Error(); err != nil {
		log.Fatal(err)
	}
}
