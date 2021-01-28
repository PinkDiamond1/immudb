/*
Copyright 2019-2020 vChain, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package store

import (
	"crypto/sha256"
)

type TxReader struct {
	InitialTxID uint64
	Desc        bool

	CurrTxID uint64
	CurrAlh  [sha256.Size]byte

	st  *ImmuStore
	_tx *Tx
}

func (s *ImmuStore) NewTxReader(initialTxID uint64, desc bool, tx *Tx) (*TxReader, error) {
	if s.closed {
		return nil, ErrAlreadyClosed
	}

	if initialTxID == 0 {
		return nil, ErrIllegalArguments
	}

	if tx == nil {
		return nil, ErrIllegalArguments
	}

	return &TxReader{
		InitialTxID: initialTxID,
		Desc:        desc,
		CurrTxID:    initialTxID,
		st:          s,
		_tx:         tx,
	}, nil
}

func (txr *TxReader) Read() (*Tx, error) {
	if txr.CurrTxID == 0 {
		return nil, ErrNoMoreEntries
	}

	err := txr.st.ReadTx(txr.CurrTxID, txr._tx)
	if err == ErrTxNotFound {
		return nil, ErrNoMoreEntries
	}
	if err != nil {
		return nil, err
	}

	if txr.InitialTxID != txr.CurrTxID {
		if txr.Desc && txr.CurrAlh != txr._tx.Alh {
			return nil, ErrorCorruptedTxData
		}

		if !txr.Desc && txr.CurrAlh != txr._tx.PrevAlh {
			return nil, ErrorCorruptedTxData
		}
	}

	if txr.Desc {
		txr.CurrTxID--
		txr.CurrAlh = txr._tx.PrevAlh
	} else {
		txr.CurrTxID++
		txr.CurrAlh = txr._tx.Alh
	}

	return txr._tx, nil
}
