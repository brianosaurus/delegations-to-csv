package main

import (
    "flag"
    validatorsModule "github.com/brianosaurus/challenge1/validators"
)

func main() {
    var node string
    var validatorOutputFile string
    flag.StringVar(&node, "node", "grpc.osmosis.zone:9090", "the node to query")
    flag.StringVar(&validatorOutputFile, "validatorFile", "validators.csv", "the output file for the validators csv")
    flag.Parse()
    
	validators, err := validatorsModule.QueryValidators(node)
    if err != nil {
        panic(err)
    }
    validatorsModule.ValidatorsToCSV(validators, validatorOutputFile)
}