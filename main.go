package main

import (
	"encoding/csv"
	"flag"
	"log"
	"os"
	"sort"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	delegationsModule "github.com/brianosaurus/challenge1/delegations"
	validatorsModule "github.com/brianosaurus/challenge1/validators"

	validatorTypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	delegationTypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// writes delegations sorted by voting power to a csv file
func WriteDelegations(delegationsMap *delegationsModule.DelegationsWithTotalBalance, writer *csv.Writer) {
	fmt.Println("Writing delegations to csv file")

	delegations := make([]string, 0)

	for _, delegationWithTotalBalance := range *delegationsMap {
		if len(delegationWithTotalBalance.DelegationResponses) == 0 {
			continue
		}

		address := delegationWithTotalBalance.DelegationResponses[0].Delegation.DelegatorAddress
		delegations = append(delegations, address)
	}

	sort.Slice(delegations, func(i, j int) bool {
		return (*delegationsMap)[delegations[i]].TotalBalance.Cmp(
			(*delegationsMap)[delegations[j]].TotalBalance) == 1
		})

	writer.Write([]string{"delegator", "voting_power"})

	for key, delegation := range *delegationsMap {
		strDelegaton := make([]string, 0)
		strDelegaton = append(strDelegaton, key)
		strDelegaton = append(strDelegaton, delegation.TotalBalance.String())

		writer.Write(strDelegaton)
	}

	writer.Flush()

	if err := writer.Error(); err != nil {
		log.Fatal(err)
	}
}

// writes delegations who are delegated to multiple validators
func WriteMultipleDelegations(validators *delegationTypes.Validators, delegationsMap *delegationsModule.DelegationsWithTotalBalance, 
	delegationResponses *delegationTypes.DelegationResponses,
	writer *csv.Writer,
) {
	fmt.Println("Writing multiple delegations to csv file")

	validatorsMap := make(map[string]delegationTypes.Validator)
	for _, validator := range *validators {
		validatorsMap[validator.OperatorAddress] = validator
	}

	writer.Write([]string{"delegator", "validator", "bonded_tokens"})

	for _, delegationWithTotalBalance := range *delegationsMap {
		delegationResponses := delegationWithTotalBalance.DelegationResponses

		if len(delegationResponses) > 1 {
			for _, delegationResponse := range delegationWithTotalBalance.DelegationResponses {
				strDelegaton := make([]string, 0)
				strDelegaton = append(strDelegaton, delegationResponse.Delegation.DelegatorAddress)
				strDelegaton = append(strDelegaton, delegationResponse.Delegation.ValidatorAddress)

				// this is to get the delegator's bonded tokens. If a delegator is unbonding or redelegating
				// the tokens are still bonded unless the validator itself is unbonded. In that case the delegator
				// is by default unbonded. There is an issue here which is noted in this github issue:
				// https://github.com/cosmos/cosmos-sdk/issues/11350
				if validator, ok := validatorsMap[delegationResponse.Delegation.ValidatorAddress]; ok &&
					validator.Status == delegationTypes.Bonded {
					strDelegaton = append(strDelegaton, delegationResponse.Balance.Amount.String())
				} else {
					strDelegaton = append(strDelegaton, "0")
				}

				writer.Write(strDelegaton)
			}
		}
	}

	writer.Flush()

	if err := writer.Error(); err != nil {
		log.Fatal(err)
	}
}

// writes validators sorted by voting power to a csv file
func WriteValidators(validators *validatorTypes.Validators, writer *csv.Writer) {
	fmt.Println("Writing validators")

	sort.SliceStable(*validators, func(i, j int) bool {
		return validatorTypes.ValidatorsByVotingPower(*validators).Less(i, j, sdk.DefaultPowerReduction)
	})

	// write headers
	writer.Write([]string{"moniker", "voting_power", "self_delegation", "total_delegation"})

	for _, validator := range *validators {
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

func main() {
	var node string
	var validatorOutputFile string
	var delegationsOutputFile string
	var multipleDelegationsOutputFile string
	flag.StringVar(&node, "node", "grpc.osmosis.zone:9090", "the node to query")
	flag.StringVar(&validatorOutputFile, "validatorFile", "validators.csv", "the output file for the validators csv")
	flag.StringVar(&delegationsOutputFile, "delegationsFile", "delegations.csv", "the output file for the delegations csv")
	flag.StringVar(&multipleDelegationsOutputFile, "multipleDelegationsFile", "multipleDelegations.csv",
		"the output csv file for the delegations who delegated to more than one validator")
	flag.Parse()

	validators, err := validatorsModule.GetValidators(node)
	if err != nil {
		panic(err)
	}

	// open validators output csv and overwrite if exists
	validatorsFile, err := os.OpenFile(validatorOutputFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		log.Fatal(err)
	}
	defer validatorsFile.Close()
	WriteValidators(validators, csv.NewWriter(validatorsFile))

	delegationResponses, err := delegationsModule.GetDelegationResponses(node, validators)
	if err != nil {
		log.Fatal(err)
	}

	delegationsMap := delegationsModule.GetDelegationsWithTotalBalance(delegationResponses)

	delegationsFile, err := os.OpenFile(delegationsOutputFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		log.Fatal(err)
	}
	defer delegationsFile.Close()
	WriteDelegations(delegationsMap, csv.NewWriter(delegationsFile))

	multipleDelegationsFile, err := os.OpenFile(multipleDelegationsOutputFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		log.Fatal(err)
	}
	defer multipleDelegationsFile.Close()
	WriteMultipleDelegations(validators, delegationsMap, delegationResponses, csv.NewWriter(multipleDelegationsFile))
}
