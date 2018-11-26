package vpn

import (
	"sort"

	"github.com/cosmos/cosmos-sdk/codec"
	csdkTypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"

	sdkTypes "github.com/ironman0x7b2/sentinel-sdk/types"
)

type Keeper struct {
	VPNStoreKey     csdkTypes.StoreKey
	SessionStoreKey csdkTypes.StoreKey
	cdc             *codec.Codec
	AccountKeeper   auth.AccountKeeper
}

func NewKeeper(cdc *codec.Codec, vpnKey csdkTypes.StoreKey, accountKeeper auth.AccountKeeper) Keeper {
	return Keeper{
		VPNStoreKey:   vpnKey,
		cdc:           cdc,
		AccountKeeper: accountKeeper,
	}
}

func (k Keeper) SetVPNDetails(ctx csdkTypes.Context, vpnID string, vpnDetails *sdkTypes.VPNDetails) {
	vpnStore := ctx.KVStore(k.VPNStoreKey)
	keyBytes, err := k.cdc.MarshalBinaryLengthPrefixed(vpnID)

	if err != nil {
		panic(err)
	}

	valueBytes, err := k.cdc.MarshalBinaryLengthPrefixed(vpnDetails)

	if err != nil {
		panic(err)
	}

	vpnStore.Set(keyBytes, valueBytes)
}

func (k Keeper) GetVPNDetails(ctx csdkTypes.Context, vpnID string) *sdkTypes.VPNDetails {
	store := ctx.KVStore(k.VPNStoreKey)
	keyBytes, err := k.cdc.MarshalBinaryLengthPrefixed(vpnID)

	if err != nil {
		panic(err)
	}

	valueBytes := store.Get(keyBytes)

	if valueBytes == nil {
		return nil
	}

	var vpnDetails sdkTypes.VPNDetails

	if err := k.cdc.UnmarshalBinaryLengthPrefixed(valueBytes, &vpnDetails); err != nil {
		panic(err)
	}

	return &vpnDetails
}

func (k Keeper) SetVPNStatus(ctx csdkTypes.Context, vpnID string, status string) {
	vpnDetails := k.GetVPNDetails(ctx, vpnID)
	vpnDetails.Status = status

	k.SetVPNDetails(ctx, vpnID, vpnDetails)
}

func (k Keeper) SetSessionDetails(ctx csdkTypes.Context, sessionID string, sessionDetails *sdkTypes.SessionDetails) {
	store := ctx.KVStore(k.SessionStoreKey)
	keyBytes, err := k.cdc.MarshalBinaryLengthPrefixed(sessionID)

	if err != nil {
		panic(err)
	}

	valueBytes, err := k.cdc.MarshalBinaryLengthPrefixed(sessionDetails)

	if err != nil {
		panic(err)
	}

	store.Set(keyBytes, valueBytes)
}

func (k Keeper) GetSessionDetails(ctx csdkTypes.Context, sessionID string) *sdkTypes.SessionDetails {
	store := ctx.KVStore(k.SessionStoreKey)
	keyBytes, err := k.cdc.MarshalBinaryLengthPrefixed(sessionID)

	if err != nil {
		panic(err)
	}

	valueBytes := store.Get(keyBytes)

	if valueBytes == nil {
		return nil
	}

	var sessionDetails sdkTypes.SessionDetails

	if err := k.cdc.UnmarshalBinaryLengthPrefixed(valueBytes, &sessionDetails); err != nil {
		panic(err)
	}

	return &sessionDetails
}

func (k Keeper) SetSessionStatus(ctx csdkTypes.Context, sessionID string, status bool) {
	sessionDetails := k.GetSessionDetails(ctx, sessionID)
	sessionDetails.Status = status
	blockTime := ctx.BlockHeader().Time.UTC()
	sessionDetails.StartTime = &blockTime

	k.SetSessionDetails(ctx, sessionID, sessionDetails)
}

func (k Keeper) SetActiveSessionIDs(ctx csdkTypes.Context, sessionIDs []string) {
	keyBytes, err := k.cdc.MarshalBinaryLengthPrefixed("ACTIVE_SESSION_IDS")

	if err != nil {
		panic(err)
	}

	sort.Strings(sessionIDs)
	valueBytes, err := k.cdc.MarshalBinaryLengthPrefixed(sessionIDs)

	if err != nil {
		panic(err)
	}

	store := ctx.KVStore(k.SessionStoreKey)
	store.Set(keyBytes, valueBytes)
}

func (k Keeper) GetActiveSessionIDs(ctx csdkTypes.Context) []string {
	store := ctx.KVStore(k.SessionStoreKey)
	valueBytes := store.Get([]byte("ACTIVE_SESSION_IDS"))

	var sessionIDs []string

	if err := k.cdc.UnmarshalBinaryLengthPrefixed(valueBytes, &sessionIDs); err != nil {
		panic(err)
	}

	return sessionIDs
}

func (k Keeper) AddActiveSession(ctx csdkTypes.Context, sessionID string) {
	sessionIDs := k.GetActiveSessionIDs(ctx)
	sessionIDs = append(sessionIDs, sessionID)
	k.SetActiveSessionIDs(ctx, sessionIDs)
}

func (k Keeper) RemoveActiveSession(ctx csdkTypes.Context, sessionID string) {
	oldSessionIDs := k.GetActiveSessionIDs(ctx)
	var sessionIDs []string

	for _, id := range oldSessionIDs {
		if id != sessionID {
			sessionIDs = append(sessionIDs, id)
		}
	}

	k.SetActiveSessionIDs(ctx, sessionIDs)
}

func (k Keeper) RemoveVPNDetails(ctx csdkTypes.Context, vpnID string) {
	store := ctx.KVStore(k.VPNStoreKey)

	vpnIDBytes := []byte(vpnID)
	store.Delete(vpnIDBytes)
}
