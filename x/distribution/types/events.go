package types

// distribution module event types
const (
	EventTypeSetWithdrawAddress = "set_withdraw_address"
	EventTypeRewards            = "rewards"
	EventTypeCommission         = "commission"
	EventTypeWithdrawRewards    = "withdraw_rewards"
	EventTypeWithdrawCommission = "withdraw_commission"
	EventTypeWithdrawTeamCommission = "withdraw_team_commission"
	EventTypeProposerReward     = "proposer_reward"

	AttributeKeyWithdrawAddress = "withdraw_address"
	AttributeKeyValidator       = "validator"
	AttributeKeyRecommandersRewards = "recommanders_rewards"

	AttributeValueCategory = ModuleName
)
