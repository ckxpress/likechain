package handlers

import (
	"reflect"

	"github.com/likecoin/likechain/abci/context"
	"github.com/likecoin/likechain/abci/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

func checkRegister(context *context.Context, rawTx *types.Transaction) abci.ResponseCheckTx {
	tx := rawTx.GetRegisterTx()
	_ = tx.Addr
	return abci.ResponseCheckTx{} // TODO
}

func deliverRegister(context *context.Context, rawTx *types.Transaction) abci.ResponseDeliverTx {
	tx := rawTx.GetRegisterTx()
	_ = tx.Addr
	return abci.ResponseDeliverTx{} // TODO
}

func init() {
	t := reflect.TypeOf((*types.Transaction_RegisterTx)(nil))
	registerCheckTxHandler(t, checkRegister)
	registerDeliverTxHandler(t, deliverRegister)
}