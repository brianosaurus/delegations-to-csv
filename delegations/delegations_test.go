package delegations

import (
	"encoding/json"
	// "strings"
	// big "math/big"

	context "context"
	"testing"

	query "github.com/cosmos/cosmos-sdk/types/query"
	grpc1 "github.com/gogo/protobuf/grpc"
	"google.golang.org/grpc"

	delegationTypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/assert"
)

const (
	DELEGATION_RESPONSES = 
`[
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
	}
]`

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

type ClientConn struct {
	grpc.ClientConn
}

func (c *ClientConn) Close() error {
	tt.Log("stubbed out ClientConn.Close")
	return nil
}

type queryClient struct {
	delegationTypes.QueryClient
}

// func (c *queryClient) Delegation(ctx context.Context, in *delegationTypes.QueryDelegationRequest, opts ...grpc.CallOption) (*delegationTypes.QueryDelegationResponse, error) {
// 	return nil, nil
// }

func (q *queryClient) ValidatorDelegations(ctx context.Context, in *delegationTypes.QueryValidatorDelegationsRequest, 
	opts ...grpc.CallOption) (*delegationTypes.QueryValidatorDelegationsResponse, error) {
	tt.Log("stubbed out ValidatorDelegations")
	var responses []delegationTypes.DelegationResponse
	err := json.Unmarshal([]byte(DELEGATION_RESPONSES), &responses)
	if err != nil {
		tt.Error(err)
	}

	response := delegationTypes.QueryValidatorDelegationsResponse{
		DelegationResponses: responses,
		Pagination: &query.PageResponse{
			Total: 1,
		},
	}

	return &response, nil
}

// Assert *ClientConn implements ClientConnInterface.
	
func stubDelegationResponses(t *testing.T) {
	t.Log("stub method called")
	// stub out the grpc connection for testing
	GrpcDial = func(node string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
		return nil, nil
	}

	DelegationTypesNewQueryClient = func(conn grpc1.ClientConn) delegationTypes.QueryClient {
		t.Log("stubbed out DelegationTypesNewQueryClient")
		client := &queryClient{}
		return client
	}
}

func TestGetDelegationResponses(t *testing.T) {
	tt = t
	ttt = t
	stubDelegationResponses(t)
	// cdc := codec.NewLegacyAmino()
	validators := make(delegationTypes.Validators, 1)
	// cdc.UnmarshalJSON([]byte(validators), &VALIDATORS)

	err := json.Unmarshal([]byte(VALIDATORS), &validators)
	if err != nil {
		t.Error(err)
	}

	delegationResponses, err := GetDelegationResponses("node value not needed", &validators)
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, 1, len(validators))
	assert.Equal(t, 1, len(*delegationResponses))
}