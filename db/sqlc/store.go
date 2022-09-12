package db

import (
	"context"
	"database/sql"
	"fmt"
)

// Store provides all functions to execute db queries and transactions
type Store struct {
	*Queries
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{
		db:      db,
		Queries: New(db),
	}
}

// execTx executes a function within a database transaction
func (s *Store) execTx(ctx context.Context, fn func(*Queries) error) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	q := New(tx)
	err = fn(q)
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("tx err: %v, rb err: %v", err, rbErr)
		}
		return err
	}

	return tx.Commit()
}

// TransferTxParams contains the input parameters of the transfer transaction
type TransferTxParams struct {
	FromAccountID int64 `json:"from_account_id"`
	ToAccountID   int64 `json:"to_account_id"`
	Amount        int64 `json:"amount"`
}

// TransferTxResult is the result of the transfer transaction
type TransferTxResult struct {
	Transfer    Transfer `json:"transfer"`
	FromAccount Account  `json:"from_account"`
	ToAccount   Account  `json:"to_account"`
	FromEntry   Entry    `json:"from_entry"`
	ToEntry     Entry    `json:"to_entry"`
}

var TxKey = struct{}{}

// TransferTx performs a money transfer from one account to the other.
// It creates a transfer record, add account entries, and update account's balance within a single database transaction
func (s *Store) TransferTx(ctx context.Context, arg TransferTxParams) (TransferTxResult, error) {
	var result TransferTxResult

	err := s.execTx(ctx, func(q *Queries) error {
		var err error

		result.Transfer, err = q.CreateTransfer(ctx, CreateTransferParams(arg))
		if err != nil {
			return err
		}

		result.FromEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: arg.FromAccountID,
			Amount:    -arg.Amount,
		})
		if err != nil {
			return err
		}

		result.ToEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: arg.ToAccountID,
			Amount:    arg.Amount,
		})
		if err != nil {
			return err
		}

		// 用 SELECT ... FOR UPDATE 不行，要用 SELECT ... FOR　NO kEY UPDATE (类似 MySQL 的 LOCK IN SHARE MODE)，因为
		// 1. INSERT 对 transfers 表插入数据会 对 accounts 表的对应数据上 S 锁，两个事务都上了 S 锁 （都插入了 transfers 数据）
		// 因为他们之间有外键，为了避免插入 transfers 的事务还没 commit，别的事务把 account 的 ID 改了，造成数据一致性的问题
		// 3. 两个执行到 SELECT ... FOR UPDATE，需要获取 X 锁，都需要对方释放 S 锁，造成死锁
		//
		// 以上还是有问题，试想：tx1 从 account1 转载到 account2，tx2 同时从 account2 转账到 account1:
		// 1. tx1 update account1：减 balance
		// 2. tx2 update account2: 减 balance
		// 3. tx1 update account2: 加 balance，要等待 tx2 释放 account2 的 X 锁
		// 4. tx2 update account1: 加 balance，要等待 tx1 释放 account1 的 X 锁，进入死锁
		// 因此执行顺序很重要：如果 tx1 和 tx2 都先操作 account1 再操作 account2 就没问题了
		// 比如此处可改为先更新 ID 小的 account
		if arg.FromAccountID < arg.ToAccountID {
			result.FromAccount, result.ToAccount, err = addMoney(ctx, q, arg.FromAccountID, -arg.Amount, arg.ToAccountID, arg.Amount)
		} else {
			result.ToAccount, result.FromAccount, err = addMoney(ctx, q, arg.ToAccountID, arg.Amount, arg.FromAccountID, -arg.Amount)
		}
		if err != nil {
			return err
		}

		return nil
	})

	return result, err
}

func addMoney(
	ctx context.Context,
	q *Queries,
	accountID1 int64,
	amount1 int64,
	accountID2 int64,
	amount2 int64,
) (account1, account2 Account, err error) {
	account1, err = q.AddAccountBalance(ctx, AddAccountBalanceParams{
		ID:     accountID1,
		Amount: amount1,
	})
	if err != nil {
		return
	}

	account2, err = q.AddAccountBalance(ctx, AddAccountBalanceParams{
		ID:     accountID2,
		Amount: amount2,
	})
	return
}
