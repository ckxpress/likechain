package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/likecoin/likechain/abci/txs"
	"github.com/likecoin/likechain/abci/types"
)

type simpleTransferJSON struct {
	Identity string    `json:"identity" binding:"required,identity"`
	To       string    `json:"to" binding:"required,identity"`
	Value    string    `json:"value" binding:"required,biginteger"`
	Remark   string    `json:"remark" binding:"required"`
	Nonce    int64     `json:"nonce" binding:"required,min=1"`
	Fee      string    `json:"fee" binding:"required,biginteger"`
	Sig      Signature `json:"sig" binding:"required"`
}

func postSimpleTransfer(c *gin.Context) {
	var json simpleTransferJSON
	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	value, ok := types.NewBigIntFromString(json.Value)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid transfer value"})
		return
	}

	fee, ok := types.NewBigIntFromString(json.Fee)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid transfer fee"})
		return
	}

	tx := txs.SimpleTransferTransaction{
		From:   types.NewIdentifier(json.Identity),
		To:     types.NewIdentifier(json.To),
		Value:  value,
		Remark: json.Remark,
		Nonce:  uint64(json.Nonce),
		Fee:    fee,
	}
	switch json.Sig.Type {
	case "eip712":
		tx.Sig = &txs.SimpleTransferEIP712Signature{EIP712Signature: txs.SigEIP712(json.Sig.Value)}
	default:
		tx.Sig = &txs.SimpleTransferJSONSignature{JSONSignature: txs.Sig(json.Sig.Value)}
	}
	data := txs.EncodeTx(&tx)

	result, err := tendermint.BroadcastTxCommit(data)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}

	if res := result.CheckTx; res.IsErr() {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":  res.Code,
			"error": res.Info,
		})
		return
	}

	if res := result.DeliverTx; res.IsErr() {
		c.JSON(http.StatusBadRequest, gin.H{
			"tx_hash": result.Hash,
			"code":    res.Code,
			"error":   res.Info,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"tx_hash": result.Hash,
	})
}
