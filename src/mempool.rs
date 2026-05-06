//! Mempool for pending transactions.

use std::collections::{BTreeSet, HashMap};

use crate::crypto::{transaction_hash, verify_signature};
use crate::error::{Error, Result};
use crate::types::{Address, Hash, SignedTransaction};

/// Ordering by nonce first, then by fee_per_gas (higher fee first).
#[derive(Debug, Clone)]
pub struct TxOrder {
    pub nonce: u64,
    pub fee_per_gas: u64,
    pub sender: Address,
    pub tx_hash: Hash,
}

impl PartialEq for TxOrder {
    fn eq(&self, other: &Self) -> bool {
        self.nonce == other.nonce
            && self.fee_per_gas == other.fee_per_gas
            && self.sender == other.sender
            && self.tx_hash == other.tx_hash
    }
}

impl Eq for TxOrder {}

impl PartialOrd for TxOrder {
    fn partial_cmp(&self, other: &Self) -> Option<std::cmp::Ordering> {
        Some(self.cmp(other))
    }
}

impl Ord for TxOrder {
    fn cmp(&self, other: &Self) -> std::cmp::Ordering {
        other
            .nonce
            .cmp(&self.nonce)
            .then_with(|| other.fee_per_gas.cmp(&self.fee_per_gas))
            .then_with(|| other.sender.cmp(&self.sender))
            .then_with(|| other.tx_hash.cmp(&self.tx_hash))
    }
}

/// Mempool for managing pending transactions.
#[derive(Debug, Clone)]
pub struct Mempool {
    by_hash: HashMap<Hash, SignedTransaction>,
    ordered: BTreeSet<TxOrder>,
}

impl Default for Mempool {
    fn default() -> Self {
        Self::new()
    }
}

impl Mempool {
    pub fn new() -> Self {
        Self {
            by_hash: HashMap::new(),
            ordered: BTreeSet::new(),
        }
    }

    pub fn has_tx(&self, tx_hash: &Hash) -> bool {
        self.by_hash.contains_key(tx_hash)
    }

    pub fn validate_tx(&self, tx: &SignedTransaction) -> Result<()> {
        if tx.nonce == 0 {
            return Err(Error::Mempool("nonce cannot be zero".to_string()));
        }
        if tx.gas_limit < 21000 {
            return Err(Error::Mempool(format!(
                "gas_limit {} less than minimum 21000",
                tx.gas_limit
            )));
        }
        if tx.max_fee_per_gas == 0 {
            return Err(Error::Mempool("max_fee_per_gas cannot be zero".to_string()));
        }
        let tx_hash = transaction_hash(tx);
        if !verify_signature(&tx.from, &tx.signature, &tx_hash) {
            return Err(Error::Mempool("invalid signature".to_string()));
        }
        for existing_tx in self.by_hash.values() {
            if existing_tx.from == tx.from
                && existing_tx.nonce == tx.nonce
                && tx.max_fee_per_gas <= existing_tx.max_fee_per_gas
            {
                return Err(Error::Mempool(
                    "replacement transaction must have higher fee".to_string(),
                ));
            }
        }
        Ok(())
    }

    pub fn insert(&mut self, tx: SignedTransaction) -> Result<()> {
        self.validate_tx(&tx)?;
        let tx_hash = transaction_hash(&tx);
        if self.by_hash.contains_key(&tx_hash) {
            return Err(Error::Mempool("duplicate transaction".to_string()));
        }
        let existing_for_sender_nonce: Option<Hash> = self
            .by_hash
            .values()
            .find(|existing| existing.from == tx.from && existing.nonce == tx.nonce)
            .map(transaction_hash);
        if let Some(old_hash) = existing_for_sender_nonce {
            if let Some(old_tx) = self.by_hash.remove(&old_hash) {
                let old_order = TxOrder {
                    nonce: old_tx.nonce,
                    fee_per_gas: old_tx.max_fee_per_gas,
                    sender: old_tx.from,
                    tx_hash: old_hash,
                };
                self.ordered.remove(&old_order);
            }
        }
        let order = TxOrder {
            nonce: tx.nonce,
            fee_per_gas: tx.max_fee_per_gas,
            sender: tx.from,
            tx_hash,
        };
        self.by_hash.insert(tx_hash, tx);
        self.ordered.insert(order);
        Ok(())
    }

    pub fn remove(&mut self, tx_hash: &Hash) -> Option<SignedTransaction> {
        if let Some(tx) = self.by_hash.remove(tx_hash) {
            let order = TxOrder {
                nonce: tx.nonce,
                fee_per_gas: tx.max_fee_per_gas,
                sender: tx.from,
                tx_hash: *tx_hash,
            };
            self.ordered.remove(&order);
            Some(tx)
        } else {
            None
        }
    }

    pub fn get_for_block(&mut self, account: &Address, max_txs: usize) -> Vec<SignedTransaction> {
        // Collect hashes first to avoid borrow conflict
        let hashes_to_remove: Vec<Hash> = self
            .by_hash
            .values()
            .filter(|tx| tx.from == *account)
            .map(transaction_hash)
            .collect();
        let mut result = Vec::new();
        for tx_hash in hashes_to_remove.into_iter().take(max_txs) {
            if let Some(removed) = self.remove(&tx_hash) {
                result.push(removed);
            }
        }
        result
    }

    pub fn update_after_block(&mut self, executed: &[SignedTransaction]) {
        for tx in executed {
            let tx_hash = transaction_hash(tx);
            if let Some(removed) = self.by_hash.remove(&tx_hash) {
                let order = TxOrder {
                    nonce: removed.nonce,
                    fee_per_gas: removed.max_fee_per_gas,
                    sender: removed.from,
                    tx_hash,
                };
                self.ordered.remove(&order);
            }
        }
    }

    pub fn evict_low_fee(&mut self, max_size: usize) {
        if self.by_hash.len() <= max_size {
            return;
        }
        while self.by_hash.len() > max_size {
            if let Some(to_remove) = self.ordered.iter().next().cloned() {
                self.by_hash.remove(&to_remove.tx_hash);
                self.ordered.remove(&to_remove);
            } else {
                break;
            }
        }
    }

    pub fn len(&self) -> usize {
        self.by_hash.len()
    }

    pub fn is_empty(&self) -> bool {
        self.by_hash.is_empty()
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    fn create_test_tx(
        from: Address,
        nonce: u64,
        max_fee_per_gas: u64,
        signature: Vec<u8>,
    ) -> SignedTransaction {
        SignedTransaction::new(
            from,
            Some([2u8; 20]),
            100,
            nonce,
            21000,
            max_fee_per_gas,
            signature,
        )
    }

    #[test]
    fn test_insert_and_has() {
        let mut mempool = Mempool::new();
        let tx = create_test_tx([1u8; 20], 1, 1_000_000_000, vec![1]);
        let tx_hash = transaction_hash(&tx);
        assert!(!mempool.has_tx(&tx_hash));
        mempool.insert(tx).unwrap();
        assert!(mempool.has_tx(&tx_hash));
    }

    #[test]
    fn test_dedup() {
        let mut mempool = Mempool::new();
        let tx = create_test_tx([1u8; 20], 1, 1_000_000_000, vec![1]);
        mempool.insert(tx.clone()).unwrap();
        assert!(mempool.insert(tx).is_err());
    }

    #[test]
    fn test_replacement_higher_fee() {
        let mut mempool = Mempool::new();
        let tx1 = create_test_tx([1u8; 20], 1, 1_000_000_000, vec![1]);
        mempool.insert(tx1).unwrap();
        let tx2 = create_test_tx([1u8; 20], 1, 2_000_000_000, vec![1]);
        mempool.insert(tx2).unwrap();
        assert_eq!(mempool.len(), 1);
        let txs: Vec<&SignedTransaction> = mempool.by_hash.values().collect();
        assert_eq!(txs[0].max_fee_per_gas, 2_000_000_000);
    }

    #[test]
    fn test_replacement_lower_fee_fails() {
        let mut mempool = Mempool::new();
        let tx1 = create_test_tx([1u8; 20], 1, 2_000_000_000, vec![1]);
        mempool.insert(tx1).unwrap();
        let tx2 = create_test_tx([1u8; 20], 1, 1_000_000_000, vec![1]);
        assert!(mempool.insert(tx2).is_err());
    }

    #[test]
    fn test_eviction_low_fee() {
        let mut mempool = Mempool::new();
        for i in 1..=5 {
            let tx = create_test_tx([1u8; 20], i, i * 100_000_000, vec![1]);
            mempool.insert(tx).unwrap();
        }
        mempool.evict_low_fee(3);
        assert_eq!(mempool.len(), 3);
    }

    #[test]
    fn test_update_after_block() {
        let mut mempool = Mempool::new();
        let tx1 = create_test_tx([1u8; 20], 1, 1_000_000_000, vec![1]);
        let tx_hash1 = transaction_hash(&tx1);
        mempool.insert(tx1.clone()).unwrap();
        mempool.update_after_block(&[tx1]);
        assert!(!mempool.has_tx(&tx_hash1));
    }
}
