package validators

import (
    "context"
    "fmt"
    "sort"
    "encoding/csv"
    "os"
    "log"

    "google.golang.org/grpc"

    "github.com/cosmos/cosmos-sdk/codec"
		sdk "github.com/cosmos/cosmos-sdk/types"
    validatorTypes "github.com/cosmos/cosmos-sdk/x/staking/types"
    queryTypes "github.com/cosmos/cosmos-sdk/types/query"
)

type DelegationValidators struct {
	Delegation validatorTypes.Delegation
	ValidatorAddresses []string
}

func QueryValidators(node string) (validatorTypes.ValidatorsByVotingPower, error) {
	fmt.Println("Getting validators")
    var validators = make([]validatorTypes.Validator, 0)

    // Create a connection to the gRPC server.
    grpcConn, err := grpc.Dial(
        node, // your gRPC server address.
        grpc.WithInsecure(), // The Cosmos SDK doesn't support any transport security mechanism. 
        // This instantiates a general gRPC codec which handles proto bytes. We pass in a nil interface registry
        // if the request/response types contain interface instead of 'nil' you should pass the application specific codec.
        grpc.WithDefaultCallOptions(grpc.ForceCodec(codec.NewProtoCodec(nil).GRPCCodec())),
    )
    if err != nil {
        return validators, err
    }
    defer grpcConn.Close()

    // This creates a gRPC client to query the x/bank service.
    validatorsClient := validatorTypes.NewQueryClient(grpcConn)
    validatorsResult, err := validatorsClient.Validators(
        context.Background(),
        &validatorTypes.QueryValidatorsRequest{Pagination: &queryTypes.PageRequest{Limit: 100}},
    )
    if err != nil {
        return validators, err
    }

    validators = append(validators, validatorsResult.GetValidators()...)

    for validatorsResult.Pagination.NextKey != nil {
        validatorsResult, err = validatorsClient.Validators(
            context.Background(),
            &validatorTypes.QueryValidatorsRequest{Pagination: &queryTypes.PageRequest{Limit: 100, Key: validatorsResult.Pagination.NextKey}},
        )
        if err != nil {
            return validators, err
        }
        validators = append(validators, validatorsResult.GetValidators()...)
    }

    sort.SliceStable(validators, func(i, j int) bool {
		return validatorTypes.ValidatorsByVotingPower(validators).Less(i, j, sdk.DefaultPowerReduction)
	})

    return validators, nil
}


// convert validators to csv
func ValidatorsToCSV(validators validatorTypes.ValidatorsByVotingPower, outputFile string) {
		// open file and overwrite if exists
		file, err := os.OpenFile(outputFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()

    writer := csv.NewWriter(file)

    // convert json to csv
    writer.Write([]string{ "moniker", "voting_power", "self_delegation", "total_delegation" })

    for _, validator := range validators {
        strValidator := make([]string, 0)
        strValidator = append(strValidator, validator.Description.Moniker)
        strValidator = append(strValidator, fmt.Sprint(validator.ConsensusPower(sdk.DefaultPowerReduction)))
        strValidator = append(strValidator, validator.MinSelfDelegation.String())
        strValidator = append(strValidator, validator.DelegatorShares.String())

        writer.Write(strValidator)
    }

    writer.Flush()

	if err := writer.Error(); err != nil {
		log.Fatal(err)
	}
}

func QueryDelegators(node string, validators validatorTypes.ValidatorsByVotingPower) (validatorTypes.Delegations, 
	map[string]DelegationValidators, error) {
	var delegations = make([]validatorTypes.Delegation, 0)
	var delegationValidators = make(map[string]DelegationValidators)

	// Create a connection to the gRPC server.
	grpcConn, err := grpc.Dial(
			node, // your gRPC server address.
			grpc.WithInsecure(), // The Cosmos SDK doesn't support any transport security mechanism. 
			// This instantiates a general gRPC codec which handles proto bytes. We pass in a nil interface registry
			// if the request/response types contain interface instead of 'nil' you should pass the application specific codec.
			grpc.WithDefaultCallOptions(grpc.ForceCodec(codec.NewProtoCodec(nil).GRPCCodec())),
	)
	if err != nil {
			return delegations, delegationValidators, err
	}
	defer grpcConn.Close()

	for _, validator := range validators {
		fmt.Println("Getting delegators for validator: ", validator.Description.Moniker)

		// This creates a gRPC client to query the x/bank service.
		delegationsClient := validatorTypes.NewQueryClient(grpcConn)
		delegationsResult, err := delegationsClient.ValidatorDelegations(
				context.Background(),
				&validatorTypes.QueryValidatorDelegationsRequest{ValidatorAddr: validator.OperatorAddress, 
					Pagination: &queryTypes.PageRequest{Limit: 100}},
		)

		if err != nil {
			return delegations, delegationValidators, err
		}

		for delegationsResult.Pagination.NextKey != nil {
			delegationsResult, err = delegationsClient.ValidatorDelegations(
					context.Background(),
					&validatorTypes.QueryValidatorDelegationsRequest{ValidatorAddr: validator.OperatorAddress,
						Pagination: &queryTypes.PageRequest{Limit: 100, Key: delegationsResult.Pagination.NextKey}},
			)
			if err != nil {
				return delegations, delegationValidators, err
			}

			delegationResponses := delegationsResult.GetDelegationResponses()

			for _, delegationResponse := range delegationResponses {
				delegation := delegationResponse.Delegation

				// if the delegator is already in the map, then we need to add the delegation amount to the existing delegation
				if val, ok := delegationValidators[delegation.DelegatorAddress]; ok {
					val.Delegation.Shares.Add(delegation.Shares)
					delegationValidators[delegation.Delegation.DelegatorAddress].Delegation = val.Delegation
				} else {
					delegationValidators[delegation.DelegatorAddress] = delegationResponse.Delegation
				}
			}
		}
	}

	for _, delegation := range delegationValidators {
		delegations = append(delegations, delegation.Delegation)
	}

	sort.Slice(delegations, func(i, j int) bool {
		return delegations[i].Shares.GT(delegations[j].Shares) 
	})

	return delegations, delegationValidators, err
}
