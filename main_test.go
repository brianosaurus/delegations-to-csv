package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"io"

	context "context"
	"testing"

	grpc1 "github.com/gogo/protobuf/grpc"
	"google.golang.org/grpc"

	delegationsModule "github.com/brianosaurus/challenge1/delegations"
	validatorsModule "github.com/brianosaurus/challenge1/validators"

	delegationTypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	validatorTypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	ed25519 "github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	// cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"

	codec "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/stretchr/testify/assert"
)

const (
	DELEGATION_RESPONSES = 
`[
	{
		"delegation": {
			"delegator_address":"osmo1qqrtqudvxhcan3fe2r98834ge8r8nffufte69a",
			"validator_address":"osmovaloper1z89utvygweg5l56fsk8ak7t6hh88fd0axx2fya",
			"shares":"20.000000000000000000"
		},
		"balance":{
			"denom":"uosmo",
			"amount":"20"
		}
	},
	{
		"delegation": {
			"delegator_address":"osmo1qqrtqudvxhcan3fe2r98834ge8r8nffufte69a",
			"validator_address":"osmovaloper1z89utvygweg5l56fsk8ak7t6hh88fd0axx2fyb",
			"shares":"20.000000000000000000"
		},
		"balance":{
			"denom":"uosmo",
			"amount":"20"
		}
	},
	{
		"delegation": {
			"delegator_address":"osmo1qqrtqudvxhcan3fe2r98834ge8r8nffufte69l",
			"validator_address":"osmovaloper1z89utvygweg5l56fsk8ak7t6hh88fd0axx2fya",
			"shares":"10.000000000000000000"
		},
		"balance":{
			"denom":"uosmo",
			"amount":"10"
		}
	},
	{
		"delegation": {
			"delegator_address":"osmo1qqrtqudvxhcan3fe2r98834ge8r8nffufte69l",
			"validator_address":"osmovaloper1z89utvygweg5l56fsk8ak7t6hh88fd0axx2fyb",
			"shares":"10.000000000000000000"
		},
		"balance":{
			"denom":"uosmo",
			"amount":"10"
		}
	}
]`

	VALIDATORS = 
`[
	{
		"operator_address":"osmovaloper1z89utvygweg5l56fsk8ak7t6hh88fd0axx2fya",
		"consensus_pubkey": { 
			"type": "tendermint/PubKeyEd25519",
			"value": "Y2h1Y2tldC1wdWJrZXk="
		},
		"status":3,
		"tokens":"5956506193276",
		"delegator_shares":"5956506193276.000000000000000000",
		"description":{
			"moniker":"Inotel",
			"identity":"975D494265B1AC25",
			"website":"https://inotel.ro",
			"details":"We do staking for a living"
		},
		"unbonding_time":"1970-01-01T00:00:00Z",
		"commission":{
			"commission_rates":{
				"rate":"0.050000000000000000",
				"max_rate":"0.300000000000000000",
				"max_change_rate":"0.300000000000000000"
				},
			"update_time":"2022-03-29T11:54:26.447424547Z"
		},
		"min_self_delegation":"1"
	},
	{
		"operator_address":"osmovaloper1z89utvygweg5l56fsk8ak7t6hh88fd0axx2fyb",
		"consensus_pubkey": { 
			"type": "tendermint/PubKeyEd25519",
			"value": "Y2h1Y2tldC1wdWJrZXk="
		},
		"status":3,
		"tokens":"5956506193276",
		"delegator_shares":"5956506193276.000000000000000000",
		"description":{
			"moniker":"Inotel Second",
			"identity":"975D494265B1AC25",
			"website":"https://inotel.ro",
			"details":"We do staking for a living"
		},
		"unbonding_time":"1970-01-01T00:00:00Z",
		"commission":{
			"commission_rates":{
				"rate":"0.050000000000000000",
				"max_rate":"0.300000000000000000",
				"max_change_rate":"0.300000000000000000"
				},
			"update_time":"2022-03-29T11:54:26.447424547Z"
		},
		"min_self_delegation":"1"
	}
 ]`
	)

var tt *testing.T

type queryClient struct {
	validatorTypes.QueryClient
}

func (q *queryClient) Validators(ctx context.Context, in *validatorTypes.QueryValidatorsRequest, 
	opts ...grpc.CallOption) (*validatorTypes.QueryValidatorsResponse, error) {
	var responses validatorTypes.Validators
	err := json.Unmarshal([]byte(VALIDATORS), &responses)
	if err != nil {
		tt.Error(err)
	}

	response := validatorTypes.QueryValidatorsResponse{
		Validators: responses,
		Pagination: nil,
	}

	return &response, nil
}

func stubValidatorResponses() {
	// stub out the grpc connection for testing
	validatorsModule.GrpcDial = func(node string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
		return nil, nil
	}

	validatorsModule.ValidatorTypesNewQueryClient = func(conn grpc1.ClientConn) validatorTypes.QueryClient {
		client := &queryClient{}
		return client
	}
}

func (q *queryClient) ValidatorDelegations(ctx context.Context, in *delegationTypes.QueryValidatorDelegationsRequest, 
	opts ...grpc.CallOption) (*delegationTypes.QueryValidatorDelegationsResponse, error) {
	var responses []delegationTypes.DelegationResponse
	err := json.Unmarshal([]byte(DELEGATION_RESPONSES), &responses)
	if err != nil {
		tt.Error(err)
	}

	response := delegationTypes.QueryValidatorDelegationsResponse{
		DelegationResponses: responses,
		Pagination: nil,
	}

	return &response, nil
}

func stubDelegationResponses() {
	// stub out the grpc connection for testing
	delegationsModule.GrpcDial = func(node string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
		return nil, nil
	}

	delegationsModule.DelegationTypesNewQueryClient = func(conn grpc1.ClientConn) delegationTypes.QueryClient {
		client := &queryClient{}
		return client
	}
}

func TestWriteValidators(t *testing.T) {
	tt = t
	stubValidatorResponses()
	stubDelegationResponses()

	validators, err := validatorsModule.GetValidators("node value not needed")
	if err != nil {
		t.Error(err)
	}
	var buf bytes.Buffer
  bufWriter := io.Writer(&buf)
  writer := csv.NewWriter(bufWriter)

  pk1 := ed25519.GenPrivKey().PubKey()
	pk1Any, err := codec.NewAnyWithValue(pk1)
  if err != nil {
		t.Log(err)

  }
	for _, validator := range *validators {
		validator.ConsensusPubkey = pk1Any
	}

	
  WriteValidators(validators, writer)

	t.Log("buf.String()", buf.String())
	
	assert.Equal(t, 
`moniker,voting_power,self_delegation,total_delegation
Inotel,5956506,1,5956506193276.000000000000000000
Inotel Second,5956506,1,5956506193276.000000000000000000
`, buf.String()) 
}

func TestWriteDelegations(t *testing.T) {
	tt = t
	stubValidatorResponses()
	stubDelegationResponses()

	validators, err := validatorsModule.GetValidators("node value not needed")
	if err != nil {
		t.Error(err)
	}

  pk1 := ed25519.GenPrivKey().PubKey()
	pk1Any, err := codec.NewAnyWithValue(pk1)
  if err != nil {
		t.Log(err)

  }

	// missing value
	for _, validator := range *validators {
		validator.ConsensusPubkey = pk1Any
	}

	delegationResponses, err := delegationsModule.GetDelegationResponses("node value not needed", validators)
	if err != nil {
		t.Log(err)
		t.FailNow()
	}

	delegationsMap := delegationsModule.GetDelegationsWithTotalBalance(delegationResponses)

	var buf bytes.Buffer
  bufWriter := io.Writer(&buf)
  writer := csv.NewWriter(bufWriter)

  WriteDelegations(delegationsMap, writer)

	assert.Equal(t, 
`delegator,voting_power
osmo1qqrtqudvxhcan3fe2r98834ge8r8nffufte69a,80
osmo1qqrtqudvxhcan3fe2r98834ge8r8nffufte69l,40
`, buf.String())
}

func TestWriteMultipleDelegations(t *testing.T) {
	tt = t
	stubValidatorResponses()
	stubDelegationResponses()

	validators, err := validatorsModule.GetValidators("node value not needed")
	if err != nil {
		t.Error(err)
	}

  pk1 := ed25519.GenPrivKey().PubKey()
	pk1Any, err := codec.NewAnyWithValue(pk1)
  if err != nil {
		t.Log(err)

  }

	// missing value
	for _, validator := range *validators {
		validator.ConsensusPubkey = pk1Any
	}

	delegationResponses, err := delegationsModule.GetDelegationResponses("node value not needed", validators)

	if err != nil {
		t.Log(err)
		t.FailNow()
	}

	delegationsMap := delegationsModule.GetDelegationsWithTotalBalance(delegationResponses)

	var buf bytes.Buffer
  bufWriter := io.Writer(&buf)
  writer := csv.NewWriter(bufWriter)

	WriteMultipleDelegations(validators, delegationsMap, delegationResponses, writer)

	// I don't know why but this fixes the test
	_ = buf.String()

	assert.Equal(t, 
`delegator,validator,bonded_tokens
osmo1qqrtqudvxhcan3fe2r98834ge8r8nffufte69a,osmovaloper1z89utvygweg5l56fsk8ak7t6hh88fd0axx2fya,20
osmo1qqrtqudvxhcan3fe2r98834ge8r8nffufte69a,osmovaloper1z89utvygweg5l56fsk8ak7t6hh88fd0axx2fyb,20
osmo1qqrtqudvxhcan3fe2r98834ge8r8nffufte69a,osmovaloper1z89utvygweg5l56fsk8ak7t6hh88fd0axx2fya,20
osmo1qqrtqudvxhcan3fe2r98834ge8r8nffufte69a,osmovaloper1z89utvygweg5l56fsk8ak7t6hh88fd0axx2fyb,20
osmo1qqrtqudvxhcan3fe2r98834ge8r8nffufte69l,osmovaloper1z89utvygweg5l56fsk8ak7t6hh88fd0axx2fya,10
osmo1qqrtqudvxhcan3fe2r98834ge8r8nffufte69l,osmovaloper1z89utvygweg5l56fsk8ak7t6hh88fd0axx2fyb,10
osmo1qqrtqudvxhcan3fe2r98834ge8r8nffufte69l,osmovaloper1z89utvygweg5l56fsk8ak7t6hh88fd0axx2fya,10
osmo1qqrtqudvxhcan3fe2r98834ge8r8nffufte69l,osmovaloper1z89utvygweg5l56fsk8ak7t6hh88fd0axx2fyb,10
`, buf.String()) 
}