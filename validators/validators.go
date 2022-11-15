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

func QueryValidators(node string) (validatorTypes.ValidatorsByVotingPower, error) {
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
