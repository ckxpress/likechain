package proposer

import (
	"crypto/ecdsa"

	tmRPC "github.com/tendermint/tendermint/rpc/client"

	"github.com/ethereum/go-ethereum/common"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/likecoin/likechain/abci/query"
	"github.com/likecoin/likechain/abci/state/deposit"
	"github.com/likecoin/likechain/abci/txs"
	"github.com/likecoin/likechain/abci/types"

	"github.com/likecoin/likechain/services/abi/token"
	"github.com/likecoin/likechain/services/eth"
	logger "github.com/likecoin/likechain/services/log"
)

var log = logger.L

func fillSig(tx *txs.DepositTransaction, privKey *ecdsa.PrivateKey) {
	tx.Proposal.Sort()
	jsonMap := tx.GenerateJSONMap()
	hash, err := jsonMap.Hash()
	if err != nil {
		panic(err)
	}
	sig, err := crypto.Sign(hash, privKey)
	if err != nil {
		panic(err)
	}
	sig[64] += 27
	jsonSig := txs.DepositJSONSignature{}
	copy(jsonSig.JSONSignature[:], sig)
	tx.Sig = &jsonSig
}

func propose(tmClient *tmRPC.HTTP, tmPrivKey *ecdsa.PrivateKey, blockNumber uint64, events []token.TokenTransfer) {
	if len(events) == 0 {
		return
	}
	log.
		WithField("block_number", blockNumber).
		Info("Proposing new proposal")
	ethAddr := crypto.PubkeyToAddress(tmPrivKey.PublicKey)
	addr, err := types.NewAddress(ethAddr[:])
	if err != nil {
		panic(err)
	}
	queryResult, err := tmClient.ABCIQuery("account_info", []byte(addr.String()))
	if err != nil {
		panic(err)
	}
	accInfo := query.GetAccountInfoRes(queryResult.Response.Value)
	if accInfo == nil {
		panic("Cannot parse account_info result")
	}
	log.
		WithField("nonce", accInfo.NextNonce).
		Debug("Got account info")
	inputs := make([]deposit.Input, 0, len(events))
	for _, e := range events {
		addr, err := types.NewAddress(e.From[:])
		if err != nil {
			panic(err)
		}
		inputs = append(inputs, deposit.Input{
			FromAddr: *addr,
			Value:    types.BigInt{Int: e.Value},
		})
	}
	tx := &txs.DepositTransaction{
		Proposer: addr,
		Proposal: deposit.Proposal{
			BlockNumber: blockNumber,
			Inputs:      inputs,
		},
		Nonce: accInfo.NextNonce,
	}
	fillSig(tx, tmPrivKey)
	rawTx := txs.EncodeTx(tx)
	log.
		WithField("raw_tx", common.Bytes2Hex(rawTx)).
		Debug("Broadcasting transaction onto LikeChain")
	_, err = tmClient.BroadcastTxCommit(rawTx)
	if err != nil {
		log.
			WithField("raw_tx", common.Bytes2Hex(rawTx)).
			WithError(err).
			Panic("Broadcast transaction onto LikeChain failed")
	}
	log.Info("Finished broadcasting transaction onto LikeChain")
}

// Run starts the subscription to the deposits on Ethereum into the relay contract and commits proposal onto LikeChain
func Run(tmClient *tmRPC.HTTP, ethClient *ethclient.Client, tokenAddr, relayAddr common.Address, tmPrivKey *ecdsa.PrivateKey, blockDelay uint64) {
	lastHeight := uint64(0) // TODO: load from DB
	eth.SubscribeHeader(ethClient, func(header *ethTypes.Header) bool {
		blockNumber := header.Number.Int64()
		if blockNumber <= 0 {
			return true
		}
		newHeight := uint64(blockNumber)
		if newHeight < blockDelay {
			return true
		}
		log.
			WithField("last_height", lastHeight).
			WithField("block_number", blockNumber).
			Info("Received new Ethereum block")
		for h := lastHeight; h <= newHeight-blockDelay; h++ {
			events := eth.GetTransfersFromBlock(ethClient, tokenAddr, relayAddr, h)
			if len(events) == 0 {
				continue
			}
			propose(tmClient, tmPrivKey, h, events)
		}
		lastHeight = newHeight
		return true
	})
}
