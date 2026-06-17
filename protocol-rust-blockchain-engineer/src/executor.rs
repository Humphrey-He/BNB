//! Transaction and block executor.

use crate::crypto::{merkle_root, transaction_hash};
use crate::error::{Error, Result};
use crate::types::{Block, Hash, Log, Receipt, SignedTransaction, ZERO_HASH};
use crate::StateDB;

/// Execution outcome after processing a block.
#[derive(Debug, Clone)]
pub struct ExecutionOutcome {
    /// Receipts for all transactions.
    pub receipts: Vec<Receipt>,
    /// Merkle root of transactions.
    pub tx_root: Hash,
    /// Merkle root of receipts.
    pub receipt_root: Hash,
    /// State root after execution.
    pub state_root: Hash,
    /// Total gas used in block.
    pub gas_used: u64,
}

pub struct Executor {
    state: StateDB,
    #[allow(dead_code)]
    gas_used: u64,
    #[allow(dead_code)]
    logs: Vec<Log>,
}

impl Executor {
    pub fn new(state: StateDB) -> Self {
        Self {
            state,
            gas_used: 0,
            logs: Vec::new(),
        }
    }

    pub fn state(&self) -> &StateDB {
        &self.state
    }

    pub fn execute_transaction(&mut self, tx: &SignedTransaction) -> Result<Receipt> {
        let tx_nonce = self.state.nonce(&tx.from);

        if tx.nonce != tx_nonce {
            return Err(Error::InvalidNonce {
                expected: tx_nonce,
                got: tx.nonce,
            });
        }

        let sender_balance = self.state.balance(&tx.from);
        if sender_balance < tx.value {
            return Err(Error::InsufficientBalance {
                have: sender_balance,
                need: tx.value,
            });
        }

        self.state.decrease_balance(&tx.from, tx.value)?;

        if let Some(ref to) = tx.to {
            self.state.increase_balance(to, tx.value);
        }

        self.state.increase_nonce(&tx.from);

        let receipt = Receipt {
            transaction_hash: transaction_hash(tx),
            success: true,
            gas_used: 21000,
            logs: Vec::new(),
        };

        self.gas_used += 21000;

        Ok(receipt)
    }

    /// Execute a block atomically.
    /// Executes against a cloned working state; only commits if all transactions
    /// succeed. On failure, the original state is unchanged.
    pub fn execute_block(&mut self, block: &Block) -> Result<ExecutionOutcome> {
        // Execute against a cloned working state for atomicity
        let mut working_state = self.state.clone();
        let mut receipts = Vec::new();
        let mut gas_used = 0u64;

        for tx in &block.transactions {
            match execute_tx(&mut working_state, tx) {
                Ok(receipt) => {
                    gas_used += receipt.gas_used;
                    receipts.push(receipt);
                }
                Err(e) => {
                    // Restore original state - block execution is atomic
                    // working_state is dropped, self.state remains unchanged
                    return Err(e);
                }
            }
        }

        // All transactions succeeded, commit working state
        let state_root = working_state.state_root();
        self.state = working_state;

        let tx_hashes: Vec<Hash> = block.transactions.iter().map(transaction_hash).collect();
        let tx_root = if tx_hashes.is_empty() {
            ZERO_HASH
        } else {
            merkle_root(&tx_hashes)
        };

        // Receipt root from actual receipt content hash, not tx hash
        let receipt_hashes: Vec<Hash> = receipts.iter().map(|r| r.receipt_hash()).collect();
        let receipt_root = if receipt_hashes.is_empty() {
            ZERO_HASH
        } else {
            merkle_root(&receipt_hashes)
        };

        Ok(ExecutionOutcome {
            receipts,
            tx_root,
            receipt_root,
            state_root,
            gas_used,
        })
    }

    pub fn into_state(self) -> StateDB {
        self.state
    }
}

/// Execute a single transaction against the given state (internal helper).
fn execute_tx(state: &mut StateDB, tx: &SignedTransaction) -> Result<Receipt> {
    let tx_nonce = state.nonce(&tx.from);

    if tx.nonce != tx_nonce {
        return Err(Error::InvalidNonce {
            expected: tx_nonce,
            got: tx.nonce,
        });
    }

    let sender_balance = state.balance(&tx.from);
    if sender_balance < tx.value {
        return Err(Error::InsufficientBalance {
            have: sender_balance,
            need: tx.value,
        });
    }

    state.decrease_balance(&tx.from, tx.value)?;

    if let Some(ref to) = tx.to {
        state.increase_balance(to, tx.value);
    }

    state.increase_nonce(&tx.from);

    let receipt = Receipt {
        transaction_hash: transaction_hash(tx),
        success: true,
        gas_used: 21000,
        logs: Vec::new(),
    };

    Ok(receipt)
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::types::{Account, Address, Header};

    fn create_test_tx(
        from: Address,
        to: Option<Address>,
        value: u128,
        nonce: u64,
    ) -> SignedTransaction {
        SignedTransaction::new(from, to, value, nonce, 21000, 1_000_000_000, vec![])
    }

    fn create_test_block(txs: Vec<SignedTransaction>) -> Block {
        Block::new(
            Header::new(
                ZERO_HASH,
                ZERO_HASH,
                ZERO_HASH,
                ZERO_HASH,
                0,
                0,
                0,
                0,
                0,
                [0u8; 20],
                vec![],
            ),
            txs,
        )
    }

    #[test]
    fn test_successful_transfer() {
        let mut state = StateDB::new();
        let from = [1u8; 20];
        let to = [2u8; 20];
        state.set_account(from, Account::new(100, 0));

        let tx = create_test_tx(from, Some(to), 30, 0);
        let mut executor = Executor::new(state);
        let receipt = executor.execute_transaction(&tx).unwrap();

        assert!(receipt.success);
        assert_eq!(executor.state().balance(&from), 70);
        assert_eq!(executor.state().balance(&to), 30);
        assert_eq!(executor.state().nonce(&from), 1);
    }

    #[test]
    fn test_insufficient_balance() {
        let mut state = StateDB::new();
        let from = [1u8; 20];
        let to = [2u8; 20];
        state.set_account(from, Account::new(50, 0));

        let tx = create_test_tx(from, Some(to), 100, 0);
        let executor = Executor::new(state);
        let mut exec = executor;
        let result = exec.execute_transaction(&tx);

        assert!(result.is_err());
    }

    #[test]
    fn test_invalid_nonce() {
        let mut state = StateDB::new();
        let from = [1u8; 20];
        let to = [2u8; 20];
        state.set_account(from, Account::new(100, 5));

        let tx = create_test_tx(from, Some(to), 10, 0);
        let executor = Executor::new(state);
        let mut exec = executor;
        let result = exec.execute_transaction(&tx);

        assert!(result.is_err());
    }

    #[test]
    fn test_execute_block() {
        let mut state = StateDB::new();
        let from = [1u8; 20];
        let to = [2u8; 20];
        state.set_account(from, Account::new(100, 0));

        let tx1 = create_test_tx(from, Some(to), 10, 0);
        let tx2 = create_test_tx(from, Some(to), 20, 1);

        let block = create_test_block(vec![tx1, tx2]);
        let executor = Executor::new(state);
        let mut exec = executor;
        let outcome = exec.execute_block(&block).unwrap();

        assert_eq!(outcome.receipts.len(), 2);
        assert!(outcome.receipts[0].success);
        assert!(outcome.receipts[1].success);
        assert_ne!(outcome.tx_root, ZERO_HASH);
        assert_ne!(outcome.receipt_root, ZERO_HASH);
        assert_eq!(exec.state().balance(&from), 70);
        assert_eq!(exec.state().nonce(&from), 2);
        assert_eq!(outcome.gas_used, 42000);
    }

    #[test]
    fn test_contract_creation_no_recipient() {
        let mut state = StateDB::new();
        let from = [1u8; 20];
        state.set_account(from, Account::new(100, 0));

        let tx = create_test_tx(from, None, 50, 0);
        let executor = Executor::new(state);
        let mut exec = executor;
        let result = exec.execute_transaction(&tx);

        assert!(result.is_ok());
        assert_eq!(exec.state().balance(&from), 50);
        assert_eq!(exec.state().nonce(&from), 1);
    }

    #[test]
    fn test_block_atomic_on_failure() {
        // If block fails, original state should be unchanged
        let mut state = StateDB::new();
        let from = [1u8; 20];
        let to = [2u8; 20];
        state.set_account(from, Account::new(100, 0));

        let tx1 = create_test_tx(from, Some(to), 10, 0);
        let tx2 = create_test_tx(from, Some(to), 200, 1); // exceeds balance

        let block = create_test_block(vec![tx1, tx2]);
        let executor = Executor::new(state);
        let mut exec = executor;
        let result = exec.execute_block(&block);

        assert!(result.is_err());
        // Original state unchanged
        assert_eq!(exec.state().balance(&from), 100);
        assert_eq!(exec.state().nonce(&from), 0);
    }

    #[test]
    fn test_large_value_transfer() {
        // Test u128 value transfer (larger than u64::MAX / 2)
        let mut state = StateDB::new();
        let from = [1u8; 20];
        let to = [2u8; 20];
        let large_value: u128 = u64::MAX as u128 + 1;
        state.set_account(from, Account::new(large_value, 0));

        let tx = create_test_tx(from, Some(to), large_value, 0);
        let mut executor = Executor::new(state);
        let receipt = executor.execute_transaction(&tx).unwrap();

        assert!(receipt.success);
        assert_eq!(executor.state().balance(&from), 0);
        assert_eq!(executor.state().balance(&to), large_value);
    }
}
