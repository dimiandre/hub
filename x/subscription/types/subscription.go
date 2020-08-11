package types

import (
	"fmt"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	hub "github.com/sentinel-official/hub/types"
)

type Subscription struct {
	ID      uint64         `json:"id"`
	Address sdk.AccAddress `json:"address"`

	Plan   uint64    `json:"plan,omitempty"`
	Expiry time.Time `json:"expiry,omitempty"`

	Node    hub.NodeAddress `json:"node,omitempty"`
	Price   sdk.Coin        `json:"price,omitempty"`
	Deposit sdk.Coin        `json:"deposit,omitempty"`

	Free     hub.Bandwidth `json:"free"`
	Status   hub.Status    `json:"status"`
	StatusAt time.Time     `json:"status_at"`
}

func (s Subscription) String() string {
	if s.Plan == 0 {
		return fmt.Sprintf(strings.TrimSpace(`
ID:        %d
Address:   %s
Node:      %s
Price:     %s
Deposit:   %s
Free:      %s
Status:    %s
Status at: %s
`), s.ID, s.Address, s.Node, s.Price, s.Deposit, s.Free, s.Status, s.StatusAt)
	}

	return fmt.Sprintf(strings.TrimSpace(`
ID:        %d
Address:   %s
Plan:      %d
Expiry:    %s
Free:      %s
Status:    %s
Status at: %s
`), s.ID, s.Address, s.Plan, s.Expiry, s.Free, s.Status, s.StatusAt)
}

func (s Subscription) Amount(consumed hub.Bandwidth) sdk.Coin {
	amount := consumed.
		CeilTo(hub.Gigabyte.Quo(s.Price.Amount)).
		Sum().
		Mul(s.Price.Amount).
		Quo(hub.Gigabyte)

	coin := sdk.NewCoin(s.Price.Denom, amount)
	if s.Deposit.IsLT(coin) {
		return s.Deposit
	}

	return coin
}

func (s Subscription) Validate() error {
	if s.ID == 0 {
		return fmt.Errorf("id should not be zero")
	}
	if s.Address == nil || s.Address.Empty() {
		return fmt.Errorf("address should not nil and empty")
	}

	if s.Plan == 0 {
		if s.Node == nil || s.Node.Empty() {
			return fmt.Errorf("node should not be nil and empty")
		}
		if !s.Price.IsValid() {
			return fmt.Errorf("price should be valid")
		}
		if !s.Deposit.IsValid() {
			return fmt.Errorf("deposit should be valid")
		}
	} else {
		if s.Expiry.IsZero() {
			return fmt.Errorf("expiry should not be zero")
		}
	}

	if s.Free.IsAnyNegative() {
		return fmt.Errorf("free should not be negative")
	}
	if !s.Status.Equal(hub.StatusActive) && !s.Status.Equal(hub.StatusInactive) {
		return fmt.Errorf("status should be either active or inactive")
	}
	if s.StatusAt.IsZero() {
		return fmt.Errorf("status_at should not be zero")
	}

	return nil
}

type Subscriptions []Subscription
