package rest

import (
	"net/http"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/rest"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys"
	csdkTypes "github.com/cosmos/cosmos-sdk/types"

	"github.com/ironman0x7b2/sentinel-sdk/x/vpn"
)

type msgInitSession struct {
	BaseReq      rest.BaseReq `json:"base_req"`
	AmountToLock string       `json:"amount_to_lock"`
	NodeID       string       `json:"node_id"`
}

func initSessionHandlerFunc(cliCtx context.CLIContext, cdc *codec.Codec, kb keys.Keybase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req msgInitSession

		if !rest.ReadRESTReq(w, r, cdc, &req) {
			return
		}

		req.BaseReq = req.BaseReq.Sanitize()
		if !req.BaseReq.ValidateBasic(w) {
			return
		}

		cliCtx.WithGenerateOnly(req.BaseReq.GenerateOnly).WithSimulation(req.BaseReq.Simulate)

		info, err := kb.Get(req.BaseReq.From)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		nodeID := vpn.NewNodeID(req.NodeID)

		amountToLock, err := csdkTypes.ParseCoin(req.AmountToLock)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		msg := vpn.NewMsgInitSession(info.GetAddress(), nodeID, amountToLock)
		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		rest.CompleteAndBroadcastTxREST(w, r, cliCtx, req.BaseReq, []csdkTypes.Msg{msg}, cdc)
	}
}
