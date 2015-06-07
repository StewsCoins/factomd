// Copyright (c) 2013-2014 Conformal Systems LLC.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package blockchain

import (
	"fmt"
	"math/big"

	"github.com/FactomProject/FactomCode/util"
	"github.com/FactomProject/btcd/database"
	"github.com/FactomProject/btcd/wire"
	"github.com/FactomProject/btcutil"

	"github.com/davecgh/go-spew/spew"
)

const (
	// MaxSigOpsPerBlock is the maximum number of signature operations
	// allowed for a block.  It is a fraction of the max block payload size.
	MaxSigOpsPerBlock = wire.MaxBlockPayload / 50

	// lockTimeThreshold is the number below which a lock time is
	// interpreted to be a block number.  Since an average of one block
	// is generated per 10 minutes, this allows blocks for about 9,512
	// years.  However, if the field is interpreted as a timestamp, given
	// the lock time is a uint32, the max is sometime around 2106.
	//	lockTimeThreshold uint32 = 5e8 // Tue Nov 5 00:53:20 1985 UTC
	lockTimeThreshold int64 = 5e8 // Tue Nov 5 00:53:20 1985 UTC  // FIXME TODO

	/*
		// MaxTimeOffsetSeconds is the maximum number of seconds a block time
		// is allowed to be ahead of the current time.  This is currently 2
		// hours.
		MaxTimeOffsetSeconds = 2 * 60 * 60

			// MinCoinbaseScriptLen is the minimum length a coinbase script can be.
			MinCoinbaseScriptLen = 2

			// MaxCoinbaseScriptLen is the maximum length a coinbase script can be.
			MaxCoinbaseScriptLen = 100
	*/

	// medianTimeBlocks is the number of previous blocks which should be
	// used to calculate the median time used to validate block timestamps.
	medianTimeBlocks = 11

	// serializedHeightVersion is the block version which changed block
	// coinbases to start with the serialized block height.
	serializedHeightVersion = 2

	// baseSubsidy is the starting subsidy amount for mined blocks.  This
	// value is halved every SubsidyHalvingInterval blocks.
	//	baseSubsidy = 50 * btcutil.SatoshiPerBitcoin
	//	BaseSubsidy = 10000000 // Phase 1 Factom release, real small subsidy
	BaseSubsidy = 500000000 // Phase 1 Factom release, small subsidy test only: 5 BTC

	// CoinbaseMaturity is the number of blocks required before newly
	// mined bitcoins (coinbase transactions) can be spent.
	//	CoinbaseMaturity = 100
	CoinbaseMaturity = 1
)

var (
	// coinbaseMaturity is the internal variable used for validating the
	// spending of coinbase outputs.  A variable rather than the exported
	// constant is used because the tests need the ability to modify it.
	coinbaseMaturity = int64(CoinbaseMaturity)

	// zeroHash is the zero value for a wire.ShaHash and is defined as
	// a package level variable to avoid the need to create a new instance
	// every time a check is needed.
	zeroHash = &wire.ShaHash{}

	// block91842Hash is one of the two nodes which violate the rules
	// set forth in BIP0030.  It is defined as a package level variable to
	// avoid the need to create a new instance every time a check is needed.
	block91842Hash = newShaHashFromStr("00000000000a4d0a398161ffc163c503763b1f4360639393e0e4c8e300e0caec")

	// block91880Hash is one of the two nodes which violate the rules
	// set forth in BIP0030.  It is defined as a package level variable to
	// avoid the need to create a new instance every time a check is needed.
	block91880Hash = newShaHashFromStr("00000000000743f190a18c5577a3c2d2a1f610ae9601ac046a38084ccb7cd721")
)

// isNullOutpoint determines whether or not a previous transaction output point
// is set.
func isNullOutpoint(outpoint *wire.OutPoint) bool {
	util.Trace()
	//	if outpoint.Index == math.MaxUint32 && outpoint.Hash.IsEqual(zeroHash) {
	if outpoint.Hash.IsEqual(zeroHash) {
		return true
	}
	return false
}

// IsCoinBase determines whether or not a transaction is a coinbase.  A coinbase
// is a special transaction created by miners that has no inputs.  This is
// represented in the block chain by a transaction with a single input that has
// a previous output transaction index set to the maximum value along with a
// zero hash.
func IsCoinBase(tx *btcutil.Tx) bool {
	//	util.Trace()

	msgTx := tx.MsgTx()

	//	fmt.Println("tx=", spew.Sdump(tx))

	// A coin base must only have one transaction input.
	if len(msgTx.TxIn) != 1 {
		return false
	}
	//	util.Trace()

	// A coin base must not have any EntryCredit outputs.
	if len(msgTx.ECOut) != 0 {
		return false
	}

	/* TODO: RE-ENABLE
	//	util.Trace()
	// A coin base must not have any RCD reveals
	if len(msgTx.RCDreveal) != 0 {
		return false
	}
	*/

	prevOut := msgTx.TxIn[0].PreviousOutPoint
	fmt.Println("prevOut=", spew.Sdump(prevOut))

	if !prevOut.Hash.IsEqual(zeroHash) {
		util.Trace("WARNING: not zeroHash !")
		return false
	}
	//	util.Trace()

	return true
}

// IsFinalizedTransaction determines whether or not a transaction is finalized.
// TODO: pay attention to locktime in some way.
func IsFinalizedTransaction(tx *btcutil.Tx, blockHeight int64) bool {

	return true
}

// isBIP0030Node returns whether or not the passed node represents one of the
// two blocks that violate the BIP0030 rule which prevents transactions from
// overwriting old ones.
func isBIP0030Node(node *blockNode) bool {
	return true
	/*
		if node.height == 91842 && node.hash.IsEqual(block91842Hash) {
			return true
		}

		if node.height == 91880 && node.hash.IsEqual(block91880Hash) {
			return true
		}

		return false
	*/
}

/*
// CalcBlockSubsidy returns the subsidy amount a block at the provided height
// should have. This is mainly used for determining how much the coinbase for
// newly generated blocks awards as well as validating the coinbase for blocks
// has the expected value.
//
// The subsidy is halved every SubsidyHalvingInterval blocks.  Mathematically
// this is: baseSubsidy / 2^(height/subsidyHalvingInterval)
//
// At the target block generation rate for the main network, this is
// approximately every 4 years.
func CalcBlockSubsidy(height int64, chainParams *chaincfg.Params) int64 {
	if chainParams.SubsidyHalvingInterval == 0 {
		return baseSubsidy
	}

	// Equivalent to: baseSubsidy / 2^(height/subsidyHalvingInterval)
	return baseSubsidy >> uint(height/int64(chainParams.SubsidyHalvingInterval))
}
*/

// CheckTransactionSanity performs some preliminary checks on a transaction to
// ensure it is sane.  These checks are context free.
func CheckTransactionSanity(tx *btcutil.Tx) error {
	// A transaction must have at least one input.
	msgTx := tx.MsgTx()
	if len(msgTx.TxIn) == 0 {
		return ruleError(ErrNoTxInputs, "transaction has no inputs")
	}

	// A transaction must have at least one output.
	if len(msgTx.TxOut) == 0 {
		return ruleError(ErrNoTxOutputs, "transaction has no outputs")
	}

	// A transaction must not exceed the maximum allowed block payload when
	// serialized.
	serializedTxSize := tx.MsgTx().SerializeSize()
	if serializedTxSize > wire.MaxBlockPayload {
		str := fmt.Sprintf("serialized transaction is too big - got "+
			"%d, max %d", serializedTxSize, wire.MaxBlockPayload)
		return ruleError(ErrTxTooBig, str)
	}

	// Ensure the transaction amounts are in range.  Each transaction
	// output must not be negative or more than the max allowed per
	// transaction.  Also, the total of all outputs must abide by the same
	// restrictions.  All amounts in a transaction are in a unit value known
	// as a satoshi.  One bitcoin is a quantity of satoshi as defined by the
	// SatoshiPerBitcoin constant.
	var totalSatoshi int64
	for _, txOut := range msgTx.TxOut {
		satoshi := txOut.Value
		if satoshi < 0 {
			str := fmt.Sprintf("transaction output has negative "+
				"value of %v", satoshi)
			return ruleError(ErrBadTxOutValue, str)
		}
		if satoshi > btcutil.MaxSatoshi {
			str := fmt.Sprintf("transaction output value of %v is "+
				"higher than max allowed value of %v", satoshi,
				btcutil.MaxSatoshi)
			return ruleError(ErrBadTxOutValue, str)
		}

		// TODO(davec): No need to check < 0 here as satoshi is
		// guaranteed to be positive per the above check.  Also need
		// to add overflow checks.
		totalSatoshi += satoshi
		if totalSatoshi < 0 {
			str := fmt.Sprintf("total value of all transaction "+
				"outputs has negative value of %v", totalSatoshi)
			return ruleError(ErrBadTxOutValue, str)
		}
		if totalSatoshi > btcutil.MaxSatoshi {
			str := fmt.Sprintf("total value of all transaction "+
				"outputs is %v which is higher than max "+
				"allowed value of %v", totalSatoshi,
				btcutil.MaxSatoshi)
			return ruleError(ErrBadTxOutValue, str)
		}
	}

	// Check for duplicate transaction inputs.
	existingTxOut := make(map[wire.OutPoint]struct{})
	for _, txIn := range msgTx.TxIn {
		if _, exists := existingTxOut[txIn.PreviousOutPoint]; exists {
			return ruleError(ErrDuplicateTxInputs, "transaction "+
				"contains duplicate inputs")
		}
		existingTxOut[txIn.PreviousOutPoint] = struct{}{}
	}

	// Coinbase script length must be between min and max length.
	if IsCoinBase(tx) {
		/*
			slen := len(msgTx.TxIn[0].SignatureScript)
			if slen < MinCoinbaseScriptLen || slen > MaxCoinbaseScriptLen {
				str := fmt.Sprintf("coinbase transaction script length "+
					"of %d is out of range (min: %d, max: %d)",
					slen, MinCoinbaseScriptLen, MaxCoinbaseScriptLen)
				return ruleError(ErrBadCoinbaseScriptLen, str)
			}
		*/
	} else {
		// Previous transaction outputs referenced by the inputs to this
		// transaction must not be null.
		for _, txIn := range msgTx.TxIn {
			prevOut := &txIn.PreviousOutPoint
			if isNullOutpoint(prevOut) {
				return ruleError(ErrBadTxInput, "transaction "+
					"input refers to previous output that "+
					"is null")
			}
		}
	}

	return nil
}

/*
// checkProofOfWork ensures the block header bits which indicate the target
// difficulty is in min/max range and that the block hash is less than the
// target difficulty as claimed.
//
//
// The flags modify the behavior of this function as follows:
//  - BFNoPoWCheck: The check to ensure the block hash is less than the target
//    difficulty is not performed.
func checkProofOfWork(block *btcutil.Block, powLimit *big.Int, flags BehaviorFlags) error {
	// The target difficulty must be larger than zero.
	target := CompactToBig(block.MsgBlock().Header.Bits)
	if target.Sign() <= 0 {
		str := fmt.Sprintf("block target difficulty of %064x is too low",
			target)
		return ruleError(ErrUnexpectedDifficulty, str)
	}

	// The target difficulty must be less than the maximum allowed.
	if target.Cmp(powLimit) > 0 {
		str := fmt.Sprintf("block target difficulty of %064x is "+
			"higher than max of %064x", target, powLimit)
		return ruleError(ErrUnexpectedDifficulty, str)
	}

	// The block hash must be less than the claimed target unless the flag
	// to avoid proof of work checks is set.
	if flags&BFNoPoWCheck != BFNoPoWCheck {
		// The block hash must be less than the claimed target.
		blockHash, err := block.Sha()
		if err != nil {
			return err
		}
		hashNum := ShaHashToBig(blockHash)
		if hashNum.Cmp(target) > 0 {
			str := fmt.Sprintf("block hash of %064x is higher than "+
				"expected max of %064x", hashNum, target)
			return ruleError(ErrHighHash, str)
		}
	}

	return nil
}

// CheckProofOfWork ensures the block header bits which indicate the target
// difficulty is in min/max range and that the block hash is less than the
// target difficulty as claimed.
func CheckProofOfWork(block *btcutil.Block, powLimit *big.Int) error {
	return checkProofOfWork(block, powLimit, BFNone)
}
*/

// CountSigOps returns the number of signature operations for all transaction
// input and output scripts in the provided transaction.  This uses the
// quicker, but imprecise, signature operation counting mechanism from
// txscript.
func CountSigOps(tx *btcutil.Tx) int {
	totalSigOps := 0
	util.Trace("NOT IMPLEMENTED --- NEEDED; ignored for now !!!!")
	/*
		msgTx := tx.MsgTx()
			// Accumulate the number of signature operations in all transaction
			// inputs.
			for _, txIn := range msgTx.TxIn {
				numSigOps := txscript.GetSigOpCount(txIn.SignatureScript)
				totalSigOps += numSigOps
			}

			// Accumulate the number of signature operations in all transaction
			// outputs.
			for _, txOut := range msgTx.TxOut {
				numSigOps := txscript.GetSigOpCount(txOut.PkScript)
				totalSigOps += numSigOps
			}
	*/

	return totalSigOps
}

// CountP2SHSigOps returns the number of signature operations for all input
// transactions which are of the pay-to-script-hash type.  This uses the
// precise, signature operation counting mechanism from the script engine which
// requires access to the input transaction scripts.
func CountP2SHSigOps(tx *btcutil.Tx, isCoinBaseTx bool, txStore TxStore) (int, error) {
	// Coinbase transactions have no interesting inputs.
	if isCoinBaseTx {
		return 0, nil
	}

	// Accumulate the number of signature operations in all transaction
	// inputs.
	msgTx := tx.MsgTx()
	totalSigOps := 0
	for _, txIn := range msgTx.TxIn {
		// Ensure the referenced input transaction is available.
		txInHash := &txIn.PreviousOutPoint.Hash
		originTx, exists := txStore[*txInHash]
		if !exists || originTx.Err != nil || originTx.Tx == nil {
			str := fmt.Sprintf("unable to find input transaction "+
				"%v referenced from transaction %v", txInHash,
				tx.Sha())
			return 0, ruleError(ErrMissingTx, str)
		}
		originMsgTx := originTx.Tx.MsgTx()

		// Ensure the output index in the referenced transaction is
		// available.
		originTxIndex := txIn.PreviousOutPoint.Index
		if originTxIndex >= uint32(len(originMsgTx.TxOut)) {
			str := fmt.Sprintf("out of bounds input index %d in "+
				"transaction %v referenced from transaction %v",
				originTxIndex, txInHash, tx.Sha())
			return 0, ruleError(ErrBadTxInput, str)
		}

		/*
			// We're only interested in pay-to-script-hash types, so skip
			// this input if it's not one.
			pkScript := originMsgTx.TxOut[originTxIndex].PkScript
			if !txscript.IsPayToScriptHash(pkScript) {
				continue
			}

				// Count the precise number of signature operations in the
				// referenced public key script.
				sigScript := txIn.SignatureScript
				numSigOps := txscript.GetPreciseSigOpCount(sigScript, pkScript,
					true)
		*/

		// We could potentially overflow the accumulator so check for
		// overflow.
		lastSigOps := totalSigOps
		//		totalSigOps += numSigOps
		if totalSigOps < lastSigOps {
			str := fmt.Sprintf("the public key script from "+
				"output index %d in transaction %v contains "+
				"too many signature operations - overflow",
				originTxIndex, txInHash)
			return 0, ruleError(ErrTooManySigOps, str)
		}
	}

	return totalSigOps, nil
}

// checkBlockSanity performs some preliminary checks on a block to ensure it is
// sane before continuing with block processing.  These checks are context free.
//
// The flags do not modify the behavior of this function directly, however they
// are needed to pass along to checkProofOfWork.
func checkBlockSanity(block *btcutil.Block, powLimit *big.Int, timeSource MedianTimeSource, flags BehaviorFlags) error {
	util.Trace()

	// A block must have at least one transaction.
	msgBlock := block.MsgBlock()
	numTx := len(msgBlock.Transactions)
	if numTx == 0 {
		return ruleError(ErrNoTransactions, "block does not contain "+
			"any transactions")
	}

	// A block must not have more transactions than the max block payload.
	if numTx > wire.MaxBlockPayload {
		str := fmt.Sprintf("block contains too many transactions - "+
			"got %d, max %d", numTx, wire.MaxBlockPayload)
		return ruleError(ErrTooManyTransactions, str)
	}

	// A block must not exceed the maximum allowed block payload when
	// serialized.
	serializedSize := msgBlock.SerializeSize()
	if serializedSize > wire.MaxBlockPayload {
		str := fmt.Sprintf("serialized block is too big - got %d, "+
			"max %d", serializedSize, wire.MaxBlockPayload)
		return ruleError(ErrBlockTooBig, str)
	}

	/*
		// Ensure the proof of work bits in the block header is in min/max range
		// and the block hash is less than the target value described by the
		// bits.
		err := checkProofOfWork(block, powLimit, flags)
		if err != nil {
			return err
		}
	*/

	util.Trace()
	// The first transaction in a block must be a coinbase.
	transactions := block.Transactions()
	if !IsCoinBase(transactions[0]) {
		return ruleError(ErrFirstTxNotCoinbase, "first transaction in "+
			"block is not a coinbase")
	}
	util.Trace()

	// A block must not have more than one coinbase.
	for i, tx := range transactions[1:] {
		if IsCoinBase(tx) {
			str := fmt.Sprintf("block contains second coinbase at "+
				"index %d", i)
			return ruleError(ErrMultipleCoinbases, str)
		}
	}

	// Do some preliminary checks on each transaction to ensure they are
	// sane before continuing.
	for _, tx := range transactions {
		err := CheckTransactionSanity(tx)
		if err != nil {
			return err
		}
	}

	// Build merkle tree and ensure the calculated merkle root matches the
	// entry in the block header.  This also has the effect of caching all
	// of the transaction hashes in the block to speed up future hash
	// checks.  Bitcoind builds the tree here and checks the merkle root
	// after the following checks, but there is no reason not to check the
	// merkle root matches here.

	header := &block.MsgBlock().Header
	merkles := BuildMerkleTreeStore(block.Transactions())
	calculatedMerkleRoot := merkles[len(merkles)-1]
	if !header.MerkleRoot.IsEqual(calculatedMerkleRoot) {
		str := fmt.Sprintf("block merkle root is invalid - block "+
			"header indicates %v, but calculated value is %v",
			header.MerkleRoot, calculatedMerkleRoot)
		util.Trace("ERROR merkle")
		return ruleError(ErrBadMerkleRoot, str)
	}

	util.Trace()
	// Check for duplicate transactions.  This check will be fairly quick
	// since the transaction hashes are already cached due to building the
	// merkle tree above.
	existingTxHashes := make(map[wire.ShaHash]struct{})
	for _, tx := range transactions {
		hash := tx.Sha()
		if _, exists := existingTxHashes[*hash]; exists {
			str := fmt.Sprintf("block contains duplicate "+
				"transaction %v", hash)
			return ruleError(ErrDuplicateTx, str)
		}
		existingTxHashes[*hash] = struct{}{}
	}
	util.Trace()

	// The number of signature operations must be less than the maximum
	// allowed per block.
	totalSigOps := 0
	for _, tx := range transactions {
		// We could potentially overflow the accumulator so check for
		// overflow.
		lastSigOps := totalSigOps
		totalSigOps += CountSigOps(tx)
		if totalSigOps < lastSigOps || totalSigOps > MaxSigOpsPerBlock {
			str := fmt.Sprintf("block contains too many signature "+
				"operations - got %v, max %v", totalSigOps,
				MaxSigOpsPerBlock)
			return ruleError(ErrTooManySigOps, str)
		}
	}
	util.Trace()

	return nil
}

// CheckBlockSanity performs some preliminary checks on a block to ensure it is
// sane before continuing with block processing.  These checks are context free.
func CheckBlockSanity(block *btcutil.Block, powLimit *big.Int, timeSource MedianTimeSource) error {
	return checkBlockSanity(block, powLimit, timeSource, BFNone)
}

/*
// checkSerializedHeight checks if the signature script in the passed
// transaction starts with the serialized block height of wantHeight.
func checkSerializedHeight(coinbaseTx *btcutil.Tx, wantHeight int64) error {
	sigScript := coinbaseTx.MsgTx().TxIn[0].SignatureScript
	if len(sigScript) < 1 {
		str := "the coinbase signature script for blocks of " +
			"version %d or greater must start with the " +
			"length of the serialized block height"
		str = fmt.Sprintf(str, serializedHeightVersion)
		return ruleError(ErrMissingCoinbaseHeight, str)
	}

	serializedLen := int(sigScript[0])
	if len(sigScript[1:]) < serializedLen {
		str := "the coinbase signature script for blocks of " +
			"version %d or greater must start with the " +
			"serialized block height"
		str = fmt.Sprintf(str, serializedLen)
		return ruleError(ErrMissingCoinbaseHeight, str)
	}

	serializedHeightBytes := make([]byte, 8, 8)
	copy(serializedHeightBytes, sigScript[1:serializedLen+1])
	serializedHeight := binary.LittleEndian.Uint64(serializedHeightBytes)
	if int64(serializedHeight) != wantHeight {
		str := fmt.Sprintf("the coinbase signature script serialized "+
			"block height is %d when %d was expected",
			serializedHeight, wantHeight)
		return ruleError(ErrBadCoinbaseHeight, str)
	}

	return nil
}
*/

// isTransactionSpent returns whether or not the provided transaction data
// describes a fully spent transaction.  A fully spent transaction is one where
// all outputs have been spent.
func isTransactionSpent(txD *TxData) bool {
	for _, isOutputSpent := range txD.Spent {
		if !isOutputSpent {
			return false
		}
	}
	return true
}

// checkBIP0030 ensures blocks do not contain duplicate transactions which
// 'overwrite' older transactions that are not fully spent.  This prevents an
// attack where a coinbase and all of its dependent transactions could be
// duplicated to effectively revert the overwritten transactions to a single
// confirmation thereby making them vulnerable to a double spend.
//
// For more details, see https://en.bitcoin.it/wiki/BIP_0030 and
// http://r6.ca/blog/20120206T005236Z.html.
func (b *BlockChain) checkBIP0030(node *blockNode, block *btcutil.Block) error {
	util.Trace()
	// Attempt to fetch duplicate transactions for all of the transactions
	// in this block from the point of view of the parent node.
	fetchSet := make(map[wire.ShaHash]struct{})
	for _, tx := range block.Transactions() {
		fetchSet[*tx.Sha()] = struct{}{}
	}
	txResults, err := b.fetchTxStore(node, fetchSet)
	if err != nil {
		return err
	}

	// Examine the resulting data about the requested transactions.
	for _, txD := range txResults {
		switch txD.Err {
		// A duplicate transaction was not found.  This is the most
		// common case.
		case database.ErrTxShaMissing:
			continue

		// A duplicate transaction was found.  This is only allowed if
		// the duplicate transaction is fully spent.
		case nil:
			if !isTransactionSpent(txD) {
				str := fmt.Sprintf("tried to overwrite "+
					"transaction %v at block height %d "+
					"that is not fully spent", txD.Hash,
					txD.BlockHeight)
				return ruleError(ErrOverwriteTx, str)
			}

		// Some other unexpected error occurred.  Return it now.
		default:
			return txD.Err
		}
	}

	return nil
}

// CheckTransactionInputs performs a series of checks on the inputs to a
// transaction to ensure they are valid.  An example of some of the checks
// include verifying all inputs exist, ensuring the coinbase seasoning
// requirements are met, detecting double spends, validating all values and fees
// are in the legal range and the total output amount doesn't exceed the input
// amount, and verifying the signatures to prove the spender was the owner of
// the bitcoins and therefore allowed to spend them.  As it checks the inputs,
// it also calculates the total fees for the transaction and returns that value.
func CheckTransactionInputs(tx *btcutil.Tx, txHeight int64, txStore TxStore) (int64, error) {
	// Coinbase transactions have no inputs.
	if IsCoinBase(tx) {
		return 0, nil
	}

	txHash := tx.Sha()
	var totalSatoshiIn int64
	for _, txIn := range tx.MsgTx().TxIn {
		// Ensure the input is available.
		txInHash := &txIn.PreviousOutPoint.Hash
		originTx, exists := txStore[*txInHash]
		if !exists || originTx.Err != nil || originTx.Tx == nil {
			str := fmt.Sprintf("unable to find input transaction "+
				"%v for transaction %v", txInHash, txHash)
			return 0, ruleError(ErrMissingTx, str)
		}

		// Ensure the transaction is not spending coins which have not
		// yet reached the required coinbase maturity.
		if IsCoinBase(originTx.Tx) {
			originHeight := originTx.BlockHeight
			blocksSincePrev := txHeight - originHeight
			if blocksSincePrev < coinbaseMaturity {
				str := fmt.Sprintf("tried to spend coinbase "+
					"transaction %v from height %v at "+
					"height %v before required maturity "+
					"of %v blocks", txInHash, originHeight,
					txHeight, coinbaseMaturity)
				return 0, ruleError(ErrImmatureSpend, str)
			}
		}

		// Ensure the transaction is not double spending coins.
		originTxIndex := txIn.PreviousOutPoint.Index
		if originTxIndex >= uint32(len(originTx.Spent)) {
			str := fmt.Sprintf("out of bounds input index %d in "+
				"transaction %v referenced from transaction %v",
				originTxIndex, txInHash, txHash)
			return 0, ruleError(ErrBadTxInput, str)
		}
		if originTx.Spent[originTxIndex] {
			str := fmt.Sprintf("transaction %v tried to double "+
				"spend output %v", txHash, txIn.PreviousOutPoint)
			return 0, ruleError(ErrDoubleSpend, str)
		}

		// Ensure the transaction amounts are in range.  Each of the
		// output values of the input transactions must not be negative
		// or more than the max allowed per transaction.  All amounts in
		// a transaction are in a unit value known as a satoshi.  One
		// bitcoin is a quantity of satoshi as defined by the
		// SatoshiPerBitcoin constant.
		originTxSatoshi := originTx.Tx.MsgTx().TxOut[originTxIndex].Value
		if originTxSatoshi < 0 {
			str := fmt.Sprintf("transaction output has negative "+
				"value of %v", originTxSatoshi)
			return 0, ruleError(ErrBadTxOutValue, str)
		}
		if originTxSatoshi > btcutil.MaxSatoshi {
			str := fmt.Sprintf("transaction output value of %v is "+
				"higher than max allowed value of %v",
				originTxSatoshi, btcutil.MaxSatoshi)
			return 0, ruleError(ErrBadTxOutValue, str)
		}

		// The total of all outputs must not be more than the max
		// allowed per transaction.  Also, we could potentially overflow
		// the accumulator so check for overflow.
		lastSatoshiIn := totalSatoshiIn
		totalSatoshiIn += originTxSatoshi
		if totalSatoshiIn < lastSatoshiIn ||
			totalSatoshiIn > btcutil.MaxSatoshi {
			str := fmt.Sprintf("total value of all transaction "+
				"inputs is %v which is higher than max "+
				"allowed value of %v", totalSatoshiIn,
				btcutil.MaxSatoshi)
			return 0, ruleError(ErrBadTxOutValue, str)
		}

		// Mark the referenced output as spent.
		originTx.Spent[originTxIndex] = true
	}

	// Calculate the total output amount for this transaction.  It is safe
	// to ignore overflow and out of range errors here because those error
	// conditions would have already been caught by checkTransactionSanity.
	var totalSatoshiOut int64
	for _, txOut := range tx.MsgTx().TxOut {
		totalSatoshiOut += txOut.Value
	}

	// Ensure the transaction does not spend more than its inputs.
	if totalSatoshiIn < totalSatoshiOut {
		str := fmt.Sprintf("total value of all transaction inputs for "+
			"transaction %v is %v which is less than the amount "+
			"spent of %v", txHash, totalSatoshiIn, totalSatoshiOut)
		return 0, ruleError(ErrSpendTooHigh, str)
	}

	// NOTE: bitcoind checks if the transaction fees are < 0 here, but that
	// is an impossible condition because of the check above that ensures
	// the inputs are >= the outputs.
	txFeeInSatoshi := totalSatoshiIn - totalSatoshiOut
	return txFeeInSatoshi, nil
}

// checkConnectBlock performs several checks to confirm connecting the passed
// block to the main chain (including whatever reorganization might be necessary
// to get this node to the main chain) does not violate any rules.
//
// The CheckConnectBlock function makes use of this function to perform the
// bulk of its work.  The only difference is this function accepts a node which
// may or may not require reorganization to connect it to the main chain whereas
// CheckConnectBlock creates a new node which specifically connects to the end
// of the current main chain and then calls this function with that node.
//
// See the comments for CheckConnectBlock for some examples of the type of
// checks performed by this function.
func (b *BlockChain) checkConnectBlock(node *blockNode, block *btcutil.Block) error {
	util.Trace()
	// If the side chain blocks end up in the database, a call to
	// CheckBlockSanity should be done here in case a previous version
	// allowed a block that is no longer valid.  However, since the
	// implementation only currently uses memory for the side chain blocks,
	// it isn't currently necessary.

	// The coinbase for the Genesis block is not spendable, so just return
	// now.
	util.Trace(fmt.Sprintf("node.hash= %v\n", node.hash.String()))
	util.Trace(fmt.Sprintf("Hard-Coded GenesisHash= %v\n", b.chainParams.GenesisHash.String()))

	if node.hash.IsEqual(b.chainParams.GenesisHash) && b.bestChain == nil {
		return nil
	}

	// BIP0030 added a rule to prevent blocks which contain duplicate
	// transactions that 'overwrite' older transactions which are not fully
	// spent.  See the documentation for checkBIP0030 for more details.
	//
	// There are two blocks in the chain which violate this
	// rule, so the check must be skipped for those blocks. The
	// isBIP0030Node function is used to determine if this block is one
	// of the two blocks that must be skipped.
	enforceBIP0030 := !isBIP0030Node(node)
	if enforceBIP0030 {
		err := b.checkBIP0030(node, block)
		if err != nil {
			return err
		}
	}

	// Request a map that contains all input transactions for the block from
	// the point of view of its position within the block chain.  These
	// transactions are needed for verification of things such as
	// transaction inputs, counting pay-to-script-hashes, and scripts.
	txInputStore, err := b.fetchInputTransactions(node, block)
	if err != nil {
		return err
	}

	// BIP0016 describes a pay-to-script-hash type that is considered a
	// "standard" type.  The rules for this BIP only apply to transactions
	// after the timestamp defined by txscript.Bip16Activation.  See
	// https://en.bitcoin.it/wiki/BIP_0016 for more details.
	enforceBIP0016 := false
	/*
		if node.timestamp.After(txscript.Bip16Activation) {
			enforceBIP0016 = true
		}
	*/

	// The number of signature operations must be less than the maximum
	// allowed per block.  Note that the preliminary sanity checks on a
	// block also include a check similar to this one, but this check
	// expands the count to include a precise count of pay-to-script-hash
	// signature operations in each of the input transaction public key
	// scripts.
	transactions := block.Transactions()
	totalSigOps := 0
	for i, tx := range transactions {
		numsigOps := CountSigOps(tx)
		if enforceBIP0016 {
			// Since the first (and only the first) transaction has
			// already been verified to be a coinbase transaction,
			// use i == 0 as an optimization for the flag to
			// countP2SHSigOps for whether or not the transaction is
			// a coinbase transaction rather than having to do a
			// full coinbase check again.
			numP2SHSigOps, err := CountP2SHSigOps(tx, i == 0,
				txInputStore)
			if err != nil {
				return err
			}
			numsigOps += numP2SHSigOps
		}

		// Check for overflow or going over the limits.  We have to do
		// this on every loop iteration to avoid overflow.
		lastSigops := totalSigOps
		totalSigOps += numsigOps
		if totalSigOps < lastSigops || totalSigOps > MaxSigOpsPerBlock {
			str := fmt.Sprintf("block contains too many "+
				"signature operations - got %v, max %v",
				totalSigOps, MaxSigOpsPerBlock)
			return ruleError(ErrTooManySigOps, str)
		}
	}

	// Perform several checks on the inputs for each transaction.  Also
	// accumulate the total fees.  This could technically be combined with
	// the loop above instead of running another loop over the transactions,
	// but by separating it we can avoid running the more expensive (though
	// still relatively cheap as compared to running the scripts) checks
	// against all the inputs when the signature operations are out of
	// bounds.
	var totalFees int64
	for _, tx := range transactions {
		txFee, err := CheckTransactionInputs(tx, node.height, txInputStore)
		if err != nil {
			return err
		}

		// Sum the total fees and ensure we don't overflow the
		// accumulator.
		lastTotalFees := totalFees
		totalFees += txFee
		if totalFees < lastTotalFees {
			return ruleError(ErrBadFees, "total fees for block "+
				"overflows accumulator")
		}
	}

	// The total output values of the coinbase transaction must not exceed
	// the expected subsidy value plus total transaction fees gained from
	// mining the block.  It is safe to ignore overflow and out of range
	// errors here because those error conditions would have already been
	// caught by checkTransactionSanity.
	var totalSatoshiOut int64
	for _, txOut := range transactions[0].MsgTx().TxOut {
		totalSatoshiOut += txOut.Value
	}

	//	expectedSatoshiOut := CalcBlockSubsidy(node.height, b.chainParams) +
	//	totalFees

	expectedSatoshiOut := int64(BaseSubsidy)

	if totalSatoshiOut > expectedSatoshiOut {
		str := fmt.Sprintf("coinbase transaction for block pays %v "+
			"which is more than expected value of %v",
			totalSatoshiOut, expectedSatoshiOut)
		return ruleError(ErrBadCoinbaseValue, str)
	}

	/*
		// Don't run scripts if this node is before the latest known good
		// checkpoint since the validity is verified via the checkpoints (all
		// transactions are included in the merkle root hash and any changes
		// will therefore be detected by the next checkpoint).  This is a huge
		// optimization because running the scripts is the most time consuming
		// portion of block handling.
		checkpoint := b.LatestCheckpoint()
		runScripts := !b.noVerify
		if checkpoint != nil && node.height <= checkpoint.Height {
		runScripts = false
		}

			// Now that the inexpensive checks are done and have passed, verify the
			// transactions are actually allowed to spend the coins by running the
			// expensive ECDSA signature check scripts.  Doing this last helps
			// prevent CPU exhaustion attacks.
			if runScripts {
				err := checkBlockScripts(block, txInputStore)
				if err != nil {
					return err
				}
			}
	*/

	return nil
}

// CheckConnectBlock performs several checks to confirm connecting the passed
// block to the main chain does not violate any rules.  An example of some of
// the checks performed are ensuring connecting the block would not cause any
// duplicate transaction hashes for old transactions that aren't already fully
// spent, double spends, exceeding the maximum allowed signature operations
// per block, invalid values in relation to the expected block subsidy, or fail
// transaction script validation.
//
// This function is NOT safe for concurrent access.
func (b *BlockChain) CheckConnectBlock(block *btcutil.Block) error {
	util.Trace()
	prevNode := b.bestChain
	blockSha, _ := block.Sha()
	newNode := newBlockNode(&block.MsgBlock().Header, blockSha, block.Height())
	if prevNode != nil {
		newNode.parent = prevNode
		//		newNode.workSum.Add(prevNode.workSum, newNode.workSum)
	}

	return b.checkConnectBlock(newNode, block)
}

///////////////////////////////////////////////////////////////////////////
/// The following functions are from difficulty.go & checkpoints.go
///////////////////////////////////////////////////////////////////////////

// newShaHashFromStr converts the passed big-endian hex string into a
// wire.ShaHash.  It only differs from the one available in wire in that
// it ignores the error since it will only (and must only) be called with
// hard-coded, and therefore known good, hashes.
func newShaHashFromStr(hexStr string) *wire.ShaHash {
	sha, _ := wire.NewShaHashFromStr(hexStr)
	return sha
}

// ShaHashToBig converts a wire.ShaHash into a big.Int that can be used to
// perform math comparisons.
func ShaHashToBig(hash *wire.ShaHash) *big.Int {
	// A ShaHash is in little-endian, but the big package wants the bytes
	// in big-endian.  Reverse them.  ShaHash.Bytes makes a copy, so it
	// is safe to modify the returned buffer.
	buf := hash.Bytes()
	blen := len(buf)
	for i := 0; i < blen/2; i++ {
		buf[i], buf[blen-1-i] = buf[blen-1-i], buf[i]
	}

	return new(big.Int).SetBytes(buf)
}

// CompactToBig converts a compact representation of a whole number N to an
// unsigned 32-bit number.  The representation is similar to IEEE754 floating
// point numbers.
//
// Like IEEE754 floating point, there are three basic components: the sign,
// the exponent, and the mantissa.  They are broken out as follows:
//
//	* the most significant 8 bits represent the unsigned base 256 exponent
// 	* bit 23 (the 24th bit) represents the sign bit
//	* the least significant 23 bits represent the mantissa
//
//	-------------------------------------------------
//	|   Exponent     |    Sign    |    Mantissa     |
//	-------------------------------------------------
//	| 8 bits [31-24] | 1 bit [23] | 23 bits [22-00] |
//	-------------------------------------------------
//
// The formula to calculate N is:
// 	N = (-1^sign) * mantissa * 256^(exponent-3)
//
// This compact form is only used in bitcoin to encode unsigned 256-bit numbers
// which represent difficulty targets, thus there really is not a need for a
// sign bit, but it is implemented here to stay consistent with bitcoind.
func CompactToBig(compact uint32) *big.Int {
	// Extract the mantissa, sign bit, and exponent.
	mantissa := compact & 0x007fffff
	isNegative := compact&0x00800000 != 0
	exponent := uint(compact >> 24)

	// Since the base for the exponent is 256, the exponent can be treated
	// as the number of bytes to represent the full 256-bit number.  So,
	// treat the exponent as the number of bytes and shift the mantissa
	// right or left accordingly.  This is equivalent to:
	// N = mantissa * 256^(exponent-3)
	var bn *big.Int
	if exponent <= 3 {
		mantissa >>= 8 * (3 - exponent)
		bn = big.NewInt(int64(mantissa))
	} else {
		bn = big.NewInt(int64(mantissa))
		bn.Lsh(bn, 8*(exponent-3))
	}

	// Make it negative if the sign bit is set.
	if isNegative {
		bn = bn.Neg(bn)
	}

	return bn
}

// BigToCompact converts a whole number N to a compact representation using
// an unsigned 32-bit number.  The compact representation only provides 23 bits
// of precision, so values larger than (2^23 - 1) only encode the most
// significant digits of the number.  See CompactToBig for details.
func BigToCompact(n *big.Int) uint32 {
	// No need to do any work if it's zero.
	if n.Sign() == 0 {
		return 0
	}

	// Since the base for the exponent is 256, the exponent can be treated
	// as the number of bytes.  So, shift the number right or left
	// accordingly.  This is equivalent to:
	// mantissa = mantissa / 256^(exponent-3)
	var mantissa uint32
	exponent := uint(len(n.Bytes()))
	if exponent <= 3 {
		mantissa = uint32(n.Bits()[0])
		mantissa <<= 8 * (3 - exponent)
	} else {
		// Use a copy to avoid modifying the caller's original number.
		tn := new(big.Int).Set(n)
		mantissa = uint32(tn.Rsh(tn, 8*(exponent-3)).Bits()[0])
	}

	// When the mantissa already has the sign bit set, the number is too
	// large to fit into the available 23-bits, so divide the number by 256
	// and increment the exponent accordingly.
	if mantissa&0x00800000 != 0 {
		mantissa >>= 8
		exponent++
	}

	// Pack the exponent, sign bit, and mantissa into an unsigned 32-bit
	// int and return it.
	compact := uint32(exponent<<24) | mantissa
	if n.Sign() < 0 {
		compact |= 0x00800000
	}
	return compact
}
