package main

import (
	"flag"
    "os"
    "log"
    "encoding/csv"

	delegationsModule "github.com/brianosaurus/challenge1/delegations"
	validatorsModule "github.com/brianosaurus/challenge1/validators"

)

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
	validatorsFile, err := os.OpenFile(validatorOutputFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer validatorsFile.Close()
	validatorsModule.WriteValidators(validators, csv.NewWriter(validatorsFile))

    delegationResponses, err := delegationsModule.GetDelegationResponses(node, validators)
	if err != nil {
		log.Fatal(err)
	}

    delegationsMap := delegationsModule.GetDelegationsWithTotalBalance(delegationResponses)

	delegationsFile, err := os.OpenFile(delegationsOutputFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer delegationsFile.Close()
    delegationsModule.WriteDelegations(delegationsMap, delegationResponses, csv.NewWriter(delegationsFile))

	multipleDelegationsFile, err := os.OpenFile(multipleDelegationsOutputFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer multipleDelegationsFile.Close()
    delegationsModule.WriteMultipleDelegations(delegationsMap, delegationResponses, csv.NewWriter(multipleDelegationsFile))
}
