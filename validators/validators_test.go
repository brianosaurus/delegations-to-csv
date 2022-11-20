package validators

import (
	"encoding/json"
	"time"

	context "context"
	"testing"

	query "github.com/cosmos/cosmos-sdk/types/query"
	grpc1 "github.com/gogo/protobuf/grpc"
	"google.golang.org/grpc"

	stakingTypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	validatorTypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/assert"
)

const (
	VALIDATORS = 
`[
	{
		"operator_address":"osmovaloper1z89utvygweg5l56fsk8ak7t6hh88fd0axx2fya",
		"consensus_pubkey":null,
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
		Pagination: &query.PageResponse{
			Total: 1,
		},
	}

	return &response, nil
}

func stubValidatorResponses(t *testing.T) {
	// stub out the grpc connection for testing
	GrpcDial = func(node string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
		return nil, nil
	}

	ValidatorTypesNewQueryClient = func(conn grpc1.ClientConn) validatorTypes.QueryClient {
		client := &queryClient{}
		return client
	}
}

func TestGetValidatorResponses(t *testing.T) {
	tt = t
	stubValidatorResponses(t)

	validators, err := GetValidators("node value not needed")
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, 1, len(*validators))


	// test first validator
	validator := (*validators)[0]
	assert.Equal(t, "osmovaloper1z89utvygweg5l56fsk8ak7t6hh88fd0axx2fya", validator.OperatorAddress)
	assert.Equal(t, "Inotel", validator.Description.Moniker)
	assert.Equal(t, "975D494265B1AC25", validator.Description.Identity)
	assert.Equal(t, "https://inotel.ro", validator.Description.Website)
	assert.Equal(t, "We do staking for a living", validator.Description.Details)
	assert.Equal(t, "0.050000000000000000", validator.Commission.CommissionRates.Rate.String())
	assert.Equal(t, "0.300000000000000000", validator.Commission.CommissionRates.MaxRate.String())
	assert.Equal(t, "0.300000000000000000", validator.Commission.CommissionRates.MaxChangeRate.String())
	assert.Equal(t, "2022-03-29T11:54:26Z", validator.Commission.UpdateTime.Format(time.RFC3339))
	assert.Equal(t, "1", validator.MinSelfDelegation.String())
	assert.Equal(t, "5956506193276", validator.Tokens.String())
	assert.Equal(t, "5956506193276.000000000000000000", validator.DelegatorShares.String())
	assert.Equal(t, stakingTypes.Bonded, validator.Status)
}
