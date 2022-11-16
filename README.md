# challenge1

## Running the challenge

To run the code 
```sh
make; ./getData
```

getData takes several arguments who have defaults. They are listed here
```sh
./getData -h
Usage of ./getData:
  -delegationsFile string
    	the output file for the delegations csv (default "delegations.csv")
  -multipleDelegationsFile string
    	the output csv file for the delegations who delegated to more than one validator (default "multipleDelegations.csv")
  -node string
    	the node to query (default "grpc.osmosis.zone:9090")
  -validatorFile string
    	the output file for the validators csv (default "validators.csv")
```

getData will overwrite the output files on subsequent runs (for convenience).

## Output file formats

### validators.csv

The columns are labeled on the first line of the csv: moniker, voting_power, self_delegation, and total_delegation.

```csv
moniker,voting_power,self_delegation,total_delegation
Inotel,5954186,1,5954186272952.000000000000000000
Provalidator,5919132,1,5919132600191.000000000000000000
SG-1,4633419,1,4633419291280.000000000000000000
DACM,4172534,90000000,4172534401353.000000000000000000
strangelove-ventures,3450774,1,3450774672524.000000000000000000
Imperator.co,2872057,1,2872057393075.000000000000000000
Parakeet,2845181,1,2845181639728.000000000000000000
Stakecito,2651033,1,2651033528949.000000000000000000
wosmongton,2445932,1,2445932846498.000000000000000000
```

### delegations.csv

This file can be very large but it lists the totals for each delegator across all validators. The colums are labeled on the first line of the csv: delegator, voting_power.

```csv
moniker,voting_power,self_delegation,total_delegation
Inotel,5954186,1,5954186272952.000000000000000000
Provalidator,5919132,1,5919132600191.000000000000000000
SG-1,4633419,1,4633419291280.000000000000000000
DACM,4172534,90000000,4172534401353.000000000000000000
strangelove-ventures,3450774,1,3450774672524.000000000000000000
Imperator.co,2872057,1,2872057393075.000000000000000000
Parakeet,2845181,1,2845181639728.000000000000000000
Stakecito,2651033,1,2651033528949.000000000000000000
wosmongton,2445932,1,2445932846498.000000000000000000
```

### multipleDelegations.csv

This file lists the multiple delegations for each delegator that has delegated tokens to more than one validator. The colums are labeled on the first line of the csv: delegator, validator, bonded_tokens. Bonded tokens is zero if the validator is unbonded.

```csv
delegator,validator,bonded_tokens
osmo1kpn0v2rz54aljzdyflxhfd686kazfkjh7u0qg0,osmovaloper1z89utvygweg5l56fsk8ak7t6hh88fd0axx2fya,10
osmo1kpn0v2rz54aljzdyflxhfd686kazfkjh7u0qg0,osmovaloper1gy0nyn2hscxxayj2pdyu8axmfvv75nnvhc079s,10
osmo1kpn0v2rz54aljzdyflxhfd686kazfkjh7u0qg0,osmovaloper196ax4vc0lwpxndu9dyhvca7jhxp70rmcmmarz7,10
osmo1kpn0v2rz54aljzdyflxhfd686kazfkjh7u0qg0,osmovaloper1q6xc9x054z9y7ll7k740j2cvdsllsfhs5rxyaj,10
osmo1kpn0v2rz54aljzdyflxhfd686kazfkjh7u0qg0,osmovaloper1r2u5q6t6w0wssrk6l66n3t2q3dw2uqny4gj2e3,10
osmo1kpn0v2rz54aljzdyflxhfd686kazfkjh7u0qg0,osmovaloper1t8qckan2yrygq7kl9apwhzfalwzgc2429p8f0s,10
osmo1kpn0v2rz54aljzdyflxhfd686kazfkjh7u0qg0,osmovaloper1prmsfrgvla0u8x3kwc8k0mcqqve3h8y73d37nm,10
osmo1kpn0v2rz54aljzdyflxhfd686kazfkjh7u0qg0,osmovaloper12rzd5qr2wmpseypvkjl0spusts0eruw2g35lkn,10
osmo1kpn0v2rz54aljzdyflxhfd686kazfkjh7u0qg0,osmovaloper1thsw3n94lzxy0knhss9n554zqp4dnfzx78j7sq,10
```
