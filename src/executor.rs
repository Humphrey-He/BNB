//! Transaction and block executor.

use crate::crypto::{merkle_root, transaction_hash};
use crate::error::Result;
use crate::types::{Block, Hash, Log, Receipt, SignedTransaction, ZERO_HASH};
use crate::StateDB;

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
            return Err(crate::Error::InvalidNonce {
                expected: tx_nonce,
                got: tx.nonce,
            });
        }

        let sender_balance = self.state.balance(&tx.from);
        if sender_balance < tx.value as u64 {
            return Err(crate::Error::InsufficientBalance {
                have: sender_balance,
                need: tx.value as u64,
            });
        }

        self.state.decrease_balance(&tx.from, tx.value as u64)?;

        if let Some(ref to) = tx.to {
            self.state.increase_balance(to, tx.value as u64);
        }

        self.state.increase_nonce(&tx.from);

        let receipt = Receipt {
            transaction_hash: transaction_hash(tx),
            success: true,
            gas_used: 21000,
            logs: Vec::new(),
        };

        Ok(receipt)
    }

    pub fn execute_block(&mut self, block: &Block) -> Result<(Vec<Receipt>, Hash, Hash)> {
        let mut receipts = Vec::new();

        for tx in &block.transactions {
            let receipt = self.execute_transaction(tx)?;
            receipts.push(receipt);
        }

        let tx_hashes: Vec<Hash> = block.transactions.iter().map(transaction_hash).collect();
        let tx_root = if tx_hashes.is_empty() {
            ZERO_HASH
        } else {
            merkle_root(&tx_hashes)
        };

        let receipt_hashes: Vec<Hash> = receipts.iter().map(|r| r.transaction_hash).collect();
        let receipt_root = if receipt_hashes.is_empty() {
            ZERO_HASH
        } else {
            merkle_root(&receipt_hashes)
        };

        Ok((receipts, tx_root, receipt_root))
    }

    pub fn into_state(self) -> StateDB {
        self.state
    }
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
        let (receipts, tx_root, receipt_root) = exec.execute_block(&block).unwrap();

        assert_eq!(receipts.len(), 2);
        assert!(receipts[0].success);
        assert!(receipts[1].success);
        assert_ne!(tx_root, ZERO_HASH);
        assert_ne!(receipt_root, ZERO_HASH);
        assert_eq!(exec.state().balance(&from), 70);
        assert_eq!(exec.state().nonce(&from), 2);
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
}
