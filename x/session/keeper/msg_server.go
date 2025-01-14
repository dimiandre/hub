package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	hubtypes "github.com/sentinel-official/hub/types"
	"github.com/sentinel-official/hub/x/session/types"
)

var (
	_ types.MsgServiceServer = (*msgServer)(nil)
)

type msgServer struct {
	Keeper
}

func NewMsgServiceServer(keeper Keeper) types.MsgServiceServer {
	return &msgServer{Keeper: keeper}
}

func isAuthorized(ctx sdk.Context, k Keeper, plan, subscription uint64, node hubtypes.NodeAddress) bool {
	if plan == 0 {
		return k.HasSubscriptionForNode(ctx, node, subscription)
	}

	return k.HasNodeForPlan(ctx, plan, node)
}

func (k *msgServer) MsgUpsert(c context.Context, msg *types.MsgUpsertRequest) (*types.MsgUpsertResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	msgFrom, err := hubtypes.NodeAddressFromBech32(msg.Proof.Node)
	if err != nil {
		return nil, err
	}

	subscription, found := k.GetSubscription(ctx, msg.Proof.Subscription)
	if !found {
		return nil, types.ErrorSubscriptionDoesNotExit
	}
	if subscription.Status.Equal(hubtypes.StatusInactive) {
		return nil, types.ErrorInvalidSubscriptionStatus
	}

	msgProofNode, err := hubtypes.NodeAddressFromBech32(msg.Proof.Node)
	if err != nil {
		return nil, err
	}

	if !isAuthorized(ctx, k.Keeper, subscription.Plan, subscription.Id, msgProofNode) {
		return nil, types.ErrorUnauthorized
	}

	msgAddress, err := sdk.AccAddressFromBech32(msg.Address)
	if err != nil {
		return nil, err
	}

	if !k.HasQuota(ctx, subscription.Id, msgAddress) {
		return nil, types.ErrorQuotaDoesNotExist
	}

	if k.ProofVerificationEnabled(ctx) {
		channel := k.GetChannel(ctx, msgAddress, msg.Proof.Subscription, msgProofNode)
		if msg.Proof.Channel != channel {
			return nil, types.ErrorInvalidChannel
		}

		if err := k.VerifyProof(ctx, msgAddress, msg.Proof, msg.Signature); err != nil {
			return nil, types.ErrorFailedToVerifyProof
		}
	}

	session, found := k.GetActiveSessionForAddress(ctx, msgAddress, subscription.Id, msgProofNode)
	if found {
		k.DeleteActiveSessionAt(ctx, session.StatusAt, session.Id)
	}

	if !found {
		count := k.GetCount(ctx)
		session = types.Session{
			Id:           count + 1,
			Subscription: subscription.Id,
			Node:         msg.Proof.Node,
			Address:      msg.Address,
			Duration:     0,
			Bandwidth:    hubtypes.NewBandwidthFromInt64(0, 0),
			Status:       hubtypes.StatusActive,
			StatusAt:     ctx.BlockTime(),
		}

		var (
			sessionAddress = session.GetAddress()
			sessionNode    = session.GetNode()
		)

		k.SetCount(ctx, count+1)
		ctx.EventManager().EmitTypedEvent(
			&types.EventSetSessionCount{
				Count: count + 1,
			},
		)

		k.SetSessionForSubscription(ctx, session.Subscription, session.Id)
		k.SetSessionForNode(ctx, sessionNode, session.Id)
		k.SetSessionForAddress(ctx, sessionAddress, session.Id)
		k.SetActiveSessionForAddress(ctx, sessionAddress, session.Subscription, sessionNode, session.Id)
		ctx.EventManager().EmitTypedEvent(
			&types.EventAddSession{
				From:         sdk.AccAddress(msgFrom.Bytes()).String(),
				Channel:      msg.Proof.Channel,
				Subscription: session.Id,
				Node:         session.Node,
				Duration:     session.Duration,
				Bandwidth:    session.Bandwidth,
				Address:      session.Address,
				Signature:    msg.Signature,
			},
		)
	}

	session.Duration = msg.Proof.Duration
	session.Bandwidth = msg.Proof.Bandwidth
	session.Status = hubtypes.StatusActive
	session.StatusAt = ctx.BlockTime()

	k.SetSession(ctx, session)
	k.SetActiveSessionAt(ctx, session.StatusAt, session.Id)
	ctx.EventManager().EmitTypedEvent(
		&types.EventUpdateSession{
			From:         sdk.AccAddress(msgFrom.Bytes()).String(),
			Channel:      msg.Proof.Channel,
			Subscription: session.Subscription,
			Node:         session.Node,
			Duration:     session.Duration,
			Bandwidth:    session.Bandwidth,
			Address:      session.Address,
			Signature:    msg.Signature,
		},
	)

	ctx.EventManager().EmitTypedEvent(&types.EventModuleName)
	return &types.MsgUpsertResponse{}, nil
}
