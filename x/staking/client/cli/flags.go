package cli

import (
	flag "github.com/spf13/pflag"

	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

const (
	FlagAddressValidator    = "validator"
	FlagAddressValidatorSrc = "addr-validator-source"
	FlagAddressValidatorDst = "addr-validator-dest"
	FlagPubKey              = "pubkey"
	FlagAmount              = "amount"
	FlagSharesAmount        = "shares-amount"
	FlagSharesFraction      = "shares-fraction"

	FlagMoniker         = "moniker"
	FlagEditMoniker     = "new-moniker"
	FlagIdentity        = "identity"
	FlagWebsite         = "website"
	FlagSecurityContact = "security-contact"
	FlagDetails         = "details"

	FlagCommissionRate          = "commission-rate"
	FlagCommissionMaxRate       = "commission-max-rate"
	FlagCommissionMaxChangeRate = "commission-max-change-rate"

	FlagRCommissionValidatorRate = "reallocated-commission-validator-rate"
	FlagRCommissionRecommandersRate = "reallocated-commission-recommanders-rate"
	FlagIncentiveDepth = "incentive-depth"
	FlagFileIncentiveClassRates = "incentive-class-rates-file"
	FlagAddressIncentiveTeam = "incentive-team"

	FlagAddressRecommander = "recommander"

	FlagMinSelfDelegation = "min-self-delegation"

	FlagGenesisFormat = "genesis-format"
	FlagNodeID        = "node-id"
	FlagIP            = "ip"
)

// common flagsets to add to various functions
var (
	fsShares       = flag.NewFlagSet("", flag.ContinueOnError)
	fsValidator    = flag.NewFlagSet("", flag.ContinueOnError)
	fsDelegation   = flag.NewFlagSet("", flag.ContinueOnError)
	fsRedelegation = flag.NewFlagSet("", flag.ContinueOnError)
)

func init() {
	fsShares.String(FlagSharesAmount, "", "Amount of source-shares to either unbond or redelegate as a positive integer or decimal")
	fsShares.String(FlagSharesFraction, "", "Fraction of source-shares to either unbond or redelegate as a positive integer or decimal >0 and <=1")
	fsValidator.String(FlagAddressValidator, "", "The Bech32 address of the validator")
	fsDelegation.String(FlagAddressRecommander, "", "The Bech32 address of the recommander")
	fsRedelegation.String(FlagAddressValidatorSrc, "", "The Bech32 address of the source validator")
	fsRedelegation.String(FlagAddressValidatorDst, "", "The Bech32 address of the destination validator")
}

// FlagSetCommissionCreate Returns the FlagSet used for commission create.
func FlagSetCommissionCreate() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)

	fs.String(FlagCommissionRate, "", "The initial commission rate percentage")
	fs.String(FlagCommissionMaxRate, "", "The maximum commission rate percentage")
	fs.String(FlagCommissionMaxChangeRate, "", "The maximum commission change rate percentage (per day)")

	return fs
}

// FlagSetRellocatedCommissionCreate Returns the FlagSet used for commission create.
func FlagSetRellocatedCommissionCreate() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)

	fs.String(FlagRCommissionValidatorRate, "", "The initial validator commission rate percentage")
	fs.String(FlagRCommissionRecommandersRate, "", "The initial all recommanders commission rate percentage")
	fs.Uint32(FlagIncentiveDepth, 0, "The incentive depth of current validator")
	fs.String(FlagFileIncentiveClassRates, "", "The every recommander rate percentage (per depth) in file")

	return fs
}

// FlagSetIncentiveCreate Returns the FlagSet for incentive create
func FlagSetIncentiveCreate() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)

	fs.String(FlagAddressIncentiveTeam, "", "The Bech32 address of the team")

	return fs
}

func flagSetIncentiveEdit() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)

	fs.String(FlagAddressIncentiveTeam, types.DoNotModifyDesc, "The (optional) Bech32 address of the team")

	return fs
}

// FlagSetMinSelfDelegation Returns the FlagSet used for minimum set delegation.
func FlagSetMinSelfDelegation() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)
	fs.String(FlagMinSelfDelegation, "", "The minimum self delegation required on the validator")
	return fs
}

// FlagSetAmount Returns the FlagSet for amount related operations.
func FlagSetAmount() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)
	fs.String(FlagAmount, "", "Amount of coins to bond")
	return fs
}

// FlagSetPublicKey Returns the flagset for Public Key related operations.
func FlagSetPublicKey() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)
	fs.String(FlagPubKey, "", "The validator's Protobuf JSON encoded public key")
	return fs
}

func flagSetDescriptionEdit() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)

	fs.String(FlagEditMoniker, types.DoNotModifyDesc, "The validator's name")
	fs.String(FlagIdentity, types.DoNotModifyDesc, "The (optional) identity signature (ex. UPort or Keybase)")
	fs.String(FlagWebsite, types.DoNotModifyDesc, "The validator's (optional) website")
	fs.String(FlagSecurityContact, types.DoNotModifyDesc, "The validator's (optional) security contact email")
	fs.String(FlagDetails, types.DoNotModifyDesc, "The validator's (optional) details")

	return fs
}

func flagSetCommissionUpdate() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)

	fs.String(FlagCommissionRate, "", "The new commission rate percentage")

	return fs
}

func flagSetDescriptionCreate() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)

	fs.String(FlagMoniker, "", "The validator's name")
	fs.String(FlagIdentity, "", "The optional identity signature (ex. UPort or Keybase)")
	fs.String(FlagWebsite, "", "The validator's (optional) website")
	fs.String(FlagSecurityContact, "", "The validator's (optional) security contact email")
	fs.String(FlagDetails, "", "The validator's (optional) details")

	return fs
}

// FlagSetRecommander Returns the FlagSet used for recommander set delegation
func FlagSetRecommander() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)

	fs.String(FlagAddressRecommander, "", "The recommander to the delegator")

	return fs
}
