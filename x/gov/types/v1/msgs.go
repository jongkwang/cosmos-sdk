package v1

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	sdktx "github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/x/auth/migrations/legacytx"
	"github.com/cosmos/cosmos-sdk/x/gov/codec"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

var (
	_, _, _, _, _, _, _ sdk.Msg                            = &MsgSubmitProposal{}, &MsgDeposit{}, &MsgVote{}, &MsgVoteWeighted{}, &MsgExecLegacyContent{}, &MsgUpdateParams{}, &MsgCancelProposal{}
	_, _, _, _, _, _, _ legacytx.LegacyMsg                 = &MsgSubmitProposal{}, &MsgDeposit{}, &MsgVote{}, &MsgVoteWeighted{}, &MsgExecLegacyContent{}, &MsgUpdateParams{}, &MsgCancelProposal{}
	_, _                codectypes.UnpackInterfacesMessage = &MsgSubmitProposal{}, &MsgExecLegacyContent{}
)

// NewMsgSubmitProposal creates a new MsgSubmitProposal.
func NewMsgSubmitProposal(
	messages []sdk.Msg,
	initialDeposit sdk.Coins,
	proposer, metadata, title, summary string,
	expedited bool,
) (*MsgSubmitProposal, error) {
	m := &MsgSubmitProposal{
		InitialDeposit: initialDeposit,
		Proposer:       proposer,
		Metadata:       metadata,
		Title:          title,
		Summary:        summary,
		Expedited:      expedited,
	}

	anys, err := sdktx.SetMsgs(messages)
	if err != nil {
		return nil, err
	}

	m.Messages = anys

	return m, nil
}

// GetMsgs unpacks m.Messages Any's into sdk.Msg's
func (m *MsgSubmitProposal) GetMsgs() ([]sdk.Msg, error) {
	return sdktx.GetMsgs(m.Messages, "sdk.MsgProposal")
}

// SetMsgs packs sdk.Msg's into m.Messages Any's
// NOTE: this will overwrite any existing messages
func (m *MsgSubmitProposal) SetMsgs(msgs []sdk.Msg) error {
	anys, err := sdktx.SetMsgs(msgs)
	if err != nil {
		return err
	}

	m.Messages = anys
	return nil
}

// ValidateBasic implements the sdk.Msg interface.
func (m MsgSubmitProposal) ValidateBasic() error {
	if m.Title == "" {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "proposal title cannot be empty")
	}
	if m.Summary == "" {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "proposal summary cannot be empty")
	}

	if _, err := sdk.AccAddressFromBech32(m.Proposer); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid proposer address: %s", err)
	}

	deposit := sdk.NewCoins(m.InitialDeposit...)
	if !deposit.IsValid() {
		return errorsmod.Wrap(sdkerrors.ErrInvalidCoins, deposit.String())
	}

	if deposit.IsAnyNegative() {
		return errorsmod.Wrap(sdkerrors.ErrInvalidCoins, deposit.String())
	}

	// Check that either metadata or Msgs length is non nil.
	if len(m.Messages) == 0 && len(m.Metadata) == 0 {
		return errorsmod.Wrap(types.ErrNoProposalMsgs, "either metadata or Msgs length must be non-nil")
	}

	msgs, err := m.GetMsgs()
	if err != nil {
		return err
	}

	for idx, msg := range msgs {
		if err := msg.ValidateBasic(); err != nil {
			return errorsmod.Wrap(types.ErrInvalidProposalMsg,
				fmt.Sprintf("msg: %d, err: %s", idx, err.Error()))
		}
	}

	return nil
}

// GetSignBytes returns the message bytes to sign over.
func (m MsgSubmitProposal) GetSignBytes() []byte {
	bz := codec.Amino.MustMarshalJSON(&m)
	return sdk.MustSortJSON(bz)
}

// GetSigners returns the expected signers for a MsgSubmitProposal.
func (m MsgSubmitProposal) GetSigners() []sdk.AccAddress {
	proposer, _ := sdk.AccAddressFromBech32(m.Proposer)
	return []sdk.AccAddress{proposer}
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (m MsgSubmitProposal) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	return sdktx.UnpackInterfaces(unpacker, m.Messages)
}

// NewMsgDeposit creates a new MsgDeposit instance
func NewMsgDeposit(depositor sdk.AccAddress, proposalID uint64, amount sdk.Coins) *MsgDeposit {
	return &MsgDeposit{proposalID, depositor.String(), amount}
}

// ValidateBasic implements the sdk.Msg interface.
func (msg MsgDeposit) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Depositor); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid depositor address: %s", err)
	}
	amount := sdk.NewCoins(msg.Amount...)
	if !amount.IsValid() {
		return errorsmod.Wrap(sdkerrors.ErrInvalidCoins, amount.String())
	}
	if amount.IsAnyNegative() {
		return errorsmod.Wrap(sdkerrors.ErrInvalidCoins, amount.String())
	}

	return nil
}

// GetSignBytes returns the message bytes to sign over.
func (msg MsgDeposit) GetSignBytes() []byte {
	bz := codec.Amino.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners returns the expected signers for a MsgDeposit.
func (msg MsgDeposit) GetSigners() []sdk.AccAddress {
	depositor, _ := sdk.AccAddressFromBech32(msg.Depositor)
	return []sdk.AccAddress{depositor}
}

// NewMsgVote creates a message to cast a vote on an active proposal
func NewMsgVote(voter sdk.AccAddress, proposalID uint64, option VoteOption, metadata string) *MsgVote {
	return &MsgVote{proposalID, voter.String(), option, metadata}
}

// ValidateBasic implements the sdk.Msg interface.
func (msg MsgVote) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Voter); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid voter address: %s", err)
	}
	if !ValidVoteOption(msg.Option) {
		return errorsmod.Wrap(types.ErrInvalidVote, msg.Option.String())
	}

	return nil
}

// GetSignBytes returns the message bytes to sign over.
func (msg MsgVote) GetSignBytes() []byte {
	bz := codec.Amino.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners returns the expected signers for a MsgVote.
func (msg MsgVote) GetSigners() []sdk.AccAddress {
	voter, _ := sdk.AccAddressFromBech32(msg.Voter)
	return []sdk.AccAddress{voter}
}

// NewMsgVoteWeighted creates a message to cast a vote on an active proposal
func NewMsgVoteWeighted(voter sdk.AccAddress, proposalID uint64, options WeightedVoteOptions, metadata string) *MsgVoteWeighted {
	return &MsgVoteWeighted{proposalID, voter.String(), options, metadata}
}

// ValidateBasic implements the sdk.Msg interface.
func (msg MsgVoteWeighted) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Voter); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid voter address: %s", err)
	}
	if len(msg.Options) == 0 {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, WeightedVoteOptions(msg.Options).String())
	}

	totalWeight := math.LegacyNewDec(0)
	usedOptions := make(map[VoteOption]bool)
	for _, option := range msg.Options {
		if !option.IsValid() {
			return errorsmod.Wrap(types.ErrInvalidVote, option.String())
		}
		weight, err := sdk.NewDecFromStr(option.Weight)
		if err != nil {
			return errorsmod.Wrapf(types.ErrInvalidVote, "Invalid weight: %s", err)
		}
		totalWeight = totalWeight.Add(weight)
		if usedOptions[option.Option] {
			return errorsmod.Wrap(types.ErrInvalidVote, "Duplicated vote option")
		}
		usedOptions[option.Option] = true
	}

	if totalWeight.GT(math.LegacyNewDec(1)) {
		return errorsmod.Wrap(types.ErrInvalidVote, "Total weight overflow 1.00")
	}

	if totalWeight.LT(math.LegacyNewDec(1)) {
		return errorsmod.Wrap(types.ErrInvalidVote, "Total weight lower than 1.00")
	}

	return nil
}

// GetSignBytes returns the message bytes to sign over.
func (msg MsgVoteWeighted) GetSignBytes() []byte {
	bz := codec.Amino.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners returns the expected signers for a MsgVoteWeighted.
func (msg MsgVoteWeighted) GetSigners() []sdk.AccAddress {
	voter, _ := sdk.AccAddressFromBech32(msg.Voter)
	return []sdk.AccAddress{voter}
}

// NewMsgExecLegacyContent creates a new MsgExecLegacyContent instance.
func NewMsgExecLegacyContent(content *codectypes.Any, authority string) *MsgExecLegacyContent {
	return &MsgExecLegacyContent{
		Content:   content,
		Authority: authority,
	}
}

// GetSignBytes returns the message bytes to sign over.
func (c MsgExecLegacyContent) GetSignBytes() []byte {
	bz := codec.Amino.MustMarshalJSON(&c)
	return sdk.MustSortJSON(bz)
}

// GetSigners returns the expected signers for a MsgExecLegacyContent.
func (c MsgExecLegacyContent) GetSigners() []sdk.AccAddress {
	authority, _ := sdk.AccAddressFromBech32(c.Authority)
	return []sdk.AccAddress{authority}
}

// ValidateBasic implements the sdk.Msg interface.
func (c MsgExecLegacyContent) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(c.Authority)
	if err != nil {
		return err
	}

	return nil
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (c MsgExecLegacyContent) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	var content v1beta1.Content
	return unpacker.UnpackAny(c.Content, &content)
}

// ValidateBasic implements the sdk.Msg interface.
func (msg MsgUpdateParams) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Authority); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid authority address: %s", err)
	}

	return msg.Params.ValidateBasic()
}

// GetSignBytes returns the message bytes to sign over.
func (msg MsgUpdateParams) GetSignBytes() []byte {
	bz := codec.Amino.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners returns the expected signers for a MsgUpdateParams.
func (msg MsgUpdateParams) GetSigners() []sdk.AccAddress {
	authority, _ := sdk.AccAddressFromBech32(msg.Authority)
	return []sdk.AccAddress{authority}
}

// NewMsgCancelProposal creates a new MsgCancelProposal instance.
func NewMsgCancelProposal(proposalID uint64, proposer string) *MsgCancelProposal {
	return &MsgCancelProposal{
		ProposalId: proposalID,
		Proposer:   proposer,
	}
}

// ValidateBasic implements Msg
func (msg MsgCancelProposal) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Proposer); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid proposer address: %s", err)
	}

	return nil
}

// GetSignBytes implements Msg
func (msg MsgCancelProposal) GetSignBytes() []byte {
	bz := codec.Amino.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners implements Msg
func (msg MsgCancelProposal) GetSigners() []sdk.AccAddress {
	proposer, _ := sdk.AccAddressFromBech32(msg.Proposer)
	return []sdk.AccAddress{proposer}
}
