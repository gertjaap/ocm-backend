package processor

import (
	"bytes"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/wire"
	"github.com/gertjaap/ocm-backend/logging"
)

type Processor struct {
	rpc              *rpcclient.Client
	db               *sql.DB
	Difficulty       float64
	TipHeight        int64
	BackendTipHeight int64
}

func NewProcessor(rpc *rpcclient.Client, db *sql.DB) (*Processor, error) {
	return &Processor{rpc: rpc, db: db}, nil
}

func (p *Processor) ProcessLoop() {
	startHeightStr := os.Getenv("OCM_BACKEND_STARTHEIGHT")
	startHeight := int64(-1)
	if startHeightStr != "" {
		startHeight, _ = strconv.ParseInt(startHeightStr, 10, 64)
	}

	apiOnly := (os.Getenv("OCM_BACKEND_APIONLY") == "1")

	var height int64
	for {
		err := p.db.QueryRow("SELECT height FROM blocks ORDER BY height DESC limit 1").Scan(&height)
		if err != nil {
			if err == sql.ErrNoRows {
				height = startHeight
				break
			}
			logging.Errorf("Error getting last processed height: %v", err)
			time.Sleep(time.Second * 5)
			continue
		}
		break
	}
	caughtUp := false
	catchUpStartHeight := height
	// monitor for tip changes
	for {
		p.BackendTipHeight, _ = p.rpc.GetBlockCount()
		if (height+1)%100 == 0 || (!caughtUp && height == catchUpStartHeight) {
			logging.Infof("Querying block %d", height+1)
		} else {
			logging.Debugf("Querying block %d", height+1)
		}

		if p.BackendTipHeight >= height+1 {

			start := time.Now()
			hash, err := p.rpc.GetBlockHash(height + 1)
			logging.Debugf("GetBlockHash: %d us", time.Now().Sub(start).Microseconds())
			if err != nil {
				if strings.Contains(err.Error(), "-8: Block height out of range") {

					// All caught up!
					if !caughtUp {
						logging.Infof("Block %d not there yet. All caught up!", height+1)
						caughtUp = true
					}
					time.Sleep(time.Second * 1)
					continue
				}
				logging.Warnf("Unable to get block at height %d: %v, retrying in 5 seconds", height+1, err)
				time.Sleep(time.Second * 5)
				continue
			}

			if (height+1)%100 == 0 || caughtUp {
				logging.Infof("Processing block %d", height+1)
			} else {
				logging.Debugf("Processing block %d", height+1)
			}

			start = time.Now()
			hdr, err := p.rpc.GetBlockHeader(hash)
			logging.Debugf("GetBlockHeader: %d us", time.Now().Sub(start).Microseconds())
			if err != nil {
				logging.Warnf("Unable to get block header for %s: %v, retrying in 5 seconds", hash.String(), err)
				time.Sleep(time.Second * 1)
				continue
			}

			if apiOnly == false {
				reorg := false
				if height > startHeight {
					var b []byte
					err = p.db.QueryRow("SELECT hash FROM blocks WHERE height=$1", height).Scan(&b)
					if err != nil {
						panic(err)
					}
					h, err := chainhash.NewHash(b)
					if err != nil {
						panic(err)
					}
					if !hdr.PrevBlock.IsEqual(h) {
						reorg = true
					}
				}

				if reorg {
					// Reorg - delete all transactions for that block and reset height
				} else {
					// Normal - process
					start = time.Now()
					blk, err := p.rpc.GetBlock(hash)
					logging.Debugf("GetBlock: %d us", time.Now().Sub(start).Microseconds())
					if err != nil {
						logging.Warnf("Unable to get block %s: %v, retrying in 5 seconds", hash.String(), err)
						time.Sleep(time.Second * 1)
						continue
					}

					// Start batch
					tx, err := p.db.Begin()

					var blockID int64
					bh := blk.BlockHash()

					err = tx.QueryRow("INSERT INTO blocks(hash, height) VALUES ($1,$2) RETURNING id", (&bh).CloneBytes(), height+1).Scan(&blockID)
					if err != nil {
						logging.Warnf("Unable to insert block: %v", err)
						return
					}

					start = time.Now()
					txIDs, err := p.GetTransactionIDsForBlock(tx, blockID, blk)
					if err != nil {
						logging.Warnf("Unable to query txids for block: %v", height+1, err)
						return
					}
					logging.Debugf("GetTransactionIDsForBlock: %d us", time.Now().Sub(start).Microseconds())

					start = time.Now()
					scriptIDs, err := p.GetScriptIDsForBlock(tx, blk)
					if err != nil {
						logging.Warnf("Unable to query script ids for block: %v", height+1, err)
						return
					}
					logging.Debugf("GetScriptIDsForBlock: %d us", time.Now().Sub(start).Microseconds())

					for i, t := range blk.Transactions {
						start = time.Now()
						err = p.processTransaction(tx, blockID, i, txIDs, scriptIDs, t)
						logging.Debugf("Process TX: %d us", time.Now().Sub(start).Microseconds())
						if err != nil {
							logging.Warnf("Unable to process transaction %v: %v", t.TxHash(), err)
							return
						}
					}

					err = tx.Commit()
					if err != nil {
						logging.Warnf("Unable to commit block: %v", err)
						return
					}
					logging.Debugf("Processed block %d", height+1)
				}
			}

			p.TipHeight = height + 1
			p.Difficulty = p.BitsToDiff(hdr.Bits)
			height++
		} else {
			time.Sleep(time.Second * 1)
		}
	}
}

func (p *Processor) BitsToDiff(bits uint32) float64 {
	shift := (bits >> 24) & 0xff
	diff := float64(0x0000ffff) / float64(bits&0x00ffffff)

	for shift < 29 {
		diff *= 256.0
		shift++
	}
	for shift > 29 {
		diff /= 256.0
		shift--
	}

	return diff
}

func (p *Processor) processTransaction(trx *sql.Tx, blockID int64, seq int, txIDs map[string]int64, scriptIDs map[string]int64, tx *wire.MsgTx) error {
	txHash := tx.TxHash()
	transID, ok := txIDs[hex.EncodeToString(txHash.CloneBytes())]
	if !ok {
		return errors.New("Transaction ID was not inserted")
	}

	err := p.MarkOutputsSpent(trx, transID, tx, txIDs)
	if err != nil {
		return err
	}

	return p.CreateOutputs(trx, transID, tx, scriptIDs)
}

func (p *Processor) CreateOutputs(trx *sql.Tx, transID int64, tx *wire.MsgTx, scriptIDs map[string]int64) error {
	var sqlParamBuf bytes.Buffer
	sqlParams := make([]interface{}, 0)
	sql := "INSERT INTO outputs(script_id, created_in_tx, vout, value, coinbase) VALUES %s ON CONFLICT (created_in_tx, vout) DO NOTHING"
	for idx, o := range tx.TxOut {
		scriptID, ok := scriptIDs[hex.EncodeToString(o.PkScript)]
		if !ok {
			panic("Did not find script in built array - this should never happen")
		}
		if idx > 0 {
			io.WriteString(&sqlParamBuf, ",")
		}
		io.WriteString(&sqlParamBuf, "(")
		for c := 1; c <= 5; c++ {
			if c > 1 {
				io.WriteString(&sqlParamBuf, ",")
			}
			io.WriteString(&sqlParamBuf, fmt.Sprintf("$%d", (idx*5)+c))
		}
		io.WriteString(&sqlParamBuf, ")")
		isCoinbase := p.IsCoinbase(tx)
		sqlParams = append(sqlParams, scriptID, transID, idx, o.Value, isCoinbase)
	}
	start := time.Now()
	_, err := trx.Exec(fmt.Sprintf(sql, string(sqlParamBuf.Bytes())), sqlParams...)
	if err != nil {
		return fmt.Errorf("Error inserting outputs: %v", err)
	}
	logging.Debugf("Insert outputs: %d us", time.Now().Sub(start).Microseconds())
	return nil
}

func (p *Processor) MarkOutputsSpent(trx *sql.Tx, transID int64, tx *wire.MsgTx, txIDs map[string]int64) error {
	var sqlParamBuf bytes.Buffer
	sqlParams := []interface{}{transID}
	sql := "UPDATE outputs SET spent_in_tx=$1 WHERE (created_in_tx,vout) IN (%s)"
	idx := 2

	for _, i := range tx.TxIn {
		if idx > 2 {
			io.WriteString(&sqlParamBuf, ",")
		}
		io.WriteString(&sqlParamBuf, fmt.Sprintf("($%d,$%d)", idx, idx+1))
		spentTx, ok := txIDs[hex.EncodeToString((&i.PreviousOutPoint.Hash).CloneBytes())]
		if !ok {
			return errors.New("Transaction ID was not found")
		}
		sqlParams = append(sqlParams, spentTx, i.PreviousOutPoint.Index)
		idx += 2
	}
	sql = fmt.Sprintf(sql, string(sqlParamBuf.Bytes()))
	start := time.Now()
	_, err := trx.Exec(sql, sqlParams...)
	if err != nil {
		return fmt.Errorf("Error updating spent outputs: %v", err)
	}
	logging.Debugf("Update spent outputs: %d us", time.Now().Sub(start).Microseconds())
	return nil
}

func (p *Processor) GetScriptIDsForBlock(trx *sql.Tx, blk *wire.MsgBlock) (map[string]int64, error) {
	result := map[string]int64{}
	var sqlParamBuf bytes.Buffer
	var sqlParamBuf2 bytes.Buffer
	sqlParams := make([]interface{}, 0)
	sql := "INSERT INTO scripts(script) VALUES %s ON CONFLICT(script) DO NOTHING"
	sql2 := "SELECT id, script FROM scripts WHERE script in (%s)"
	idx := 0
	for _, tx := range blk.Transactions {
		for _, o := range tx.TxOut {
			idx++
			if idx > 1 {
				io.WriteString(&sqlParamBuf, ",")
				io.WriteString(&sqlParamBuf2, ",")
			}
			io.WriteString(&sqlParamBuf, fmt.Sprintf("($%d)", idx))
			io.WriteString(&sqlParamBuf2, fmt.Sprintf("$%d", idx))
			sqlParams = append(sqlParams, o.PkScript)
		}
	}
	sql = fmt.Sprintf(sql, string(sqlParamBuf.Bytes()))
	sql2 = fmt.Sprintf(sql2, string(sqlParamBuf2.Bytes()))
	_, err := trx.Exec(sql, sqlParams...)
	if err != nil {
		return nil, err
	}
	rows, err := trx.Query(sql2, sqlParams...)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var hash []byte
		var id int64
		err := rows.Scan(&id, &hash)
		if err == nil {
			result[hex.EncodeToString(hash)] = id
		}
	}
	return result, nil
}

func (p *Processor) GetTransactionIDsForBlock(trx *sql.Tx, blockID int64, blk *wire.MsgBlock) (map[string]int64, error) {
	transactionsCreatedInBlock := make([]*chainhash.Hash, 0)
	transactionsSpentInBlock := make([]*chainhash.Hash, 0)
	for _, tx := range blk.Transactions {
		for _, i := range tx.TxIn {
			transactionsSpentInBlock = append(transactionsSpentInBlock, &i.PreviousOutPoint.Hash)
		}
		txHash := tx.TxHash()
		transactionsCreatedInBlock = append(transactionsCreatedInBlock, &txHash)
	}
	err := p.EnsureTransactionsInserted(trx, append(transactionsCreatedInBlock, transactionsSpentInBlock...))
	if err != nil {
		return map[string]int64{}, fmt.Errorf("Error occured during EnsureTransactionsInserted: %v", err)
	}
	err = p.SetBlockIDForTransactions(trx, blockID, transactionsCreatedInBlock)
	if err != nil {
		return map[string]int64{}, fmt.Errorf("Error occured during SetBlockIDForTransactions: %v", err)
	}

	results, err := p.QueryTransactionIDs(trx, append(transactionsCreatedInBlock, transactionsSpentInBlock...))
	if err != nil {
		return map[string]int64{}, fmt.Errorf("Error occured during QueryTransactionIDs: %v", err)
	}
	return results, nil
}

func (p *Processor) SetBlockIDForTransactions(trx *sql.Tx, blockID int64, hashes []*chainhash.Hash) error {
	var sqlParamBuf bytes.Buffer
	sqlParams := []interface{}{blockID}
	sql := "UPDATE transactions SET block_id=$1 WHERE hash in (%s)"
	for idx, h := range hashes {
		if idx > 0 {
			io.WriteString(&sqlParamBuf, ",")
		}
		io.WriteString(&sqlParamBuf, fmt.Sprintf("$%d", idx+2))
		sqlParams = append(sqlParams, h.CloneBytes())
	}
	sql = fmt.Sprintf(sql, string(sqlParamBuf.Bytes()))
	_, err := trx.Exec(sql, sqlParams...)
	return err
}

func (p *Processor) EnsureTransactionsInserted(trx *sql.Tx, hashes []*chainhash.Hash) error {
	var sqlParamBuf bytes.Buffer
	sqlParams := make([]interface{}, 0)
	sql := "INSERT INTO transactions(hash) VALUES %s ON CONFLICT(hash) DO NOTHING"
	for idx, h := range hashes {
		if idx > 0 {
			io.WriteString(&sqlParamBuf, ",")
		}
		io.WriteString(&sqlParamBuf, fmt.Sprintf("($%d)", idx+1))
		sqlParams = append(sqlParams, h.CloneBytes())
	}
	sql = fmt.Sprintf(sql, string(sqlParamBuf.Bytes()))
	_, err := trx.Exec(sql, sqlParams...)
	return err
}

func (p *Processor) QueryTransactionIDs(trx *sql.Tx, hashes []*chainhash.Hash) (map[string]int64, error) {
	result := map[string]int64{}
	var sqlParamBuf bytes.Buffer
	sqlParams := make([]interface{}, 0)
	sql := "SELECT id, hash FROM transactions WHERE hash in (%s)"
	for idx, h := range hashes {
		if idx > 0 {
			io.WriteString(&sqlParamBuf, ",")
		}
		io.WriteString(&sqlParamBuf, fmt.Sprintf("$%d", idx+1))
		sqlParams = append(sqlParams, h.CloneBytes())
	}
	sql = fmt.Sprintf(sql, string(sqlParamBuf.Bytes()))
	rows, err := trx.Query(sql, sqlParams...)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var hash []byte
		var id int64
		err := rows.Scan(&id, &hash)
		if err == nil {
			result[hex.EncodeToString(hash)] = id
		}
	}
	return result, nil
}

func (p *Processor) IsCoinbase(tx *wire.MsgTx) bool {
	if tx.TxIn[0].PreviousOutPoint.Index != 0xFFFFFFFF {
		return false
	}
	if (&tx.TxIn[0].PreviousOutPoint.Hash).String() != "0000000000000000000000000000000000000000000000000000000000000000" {
		return false
	}
	return true
}
