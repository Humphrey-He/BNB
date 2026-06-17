//! Consensus module for PoA (Proof of Authority) block production.
//!
//! This module implements a simple single-validator PoA consensus where
//! one validator produces blocks on a fixed interval.

use std::sync::Arc;
use std::time::Duration;
use tokio::sync::RwLock;
use tokio::time::interval;

use crate::crypto::{hash, merkle_root, transaction_hash};
use crate::error::Result;
use crate::mempool::Mempool;
use crate::state::StateDB;
use crate::types::{Address, Block, Header, Receipt, SignedTransaction, ZERO_HASH};

/// Genesis configuration for chain initialization
#[derive(Debug, Clone)]
pub struct Genesis {
    /// Initial validator address
    pub validator: Address,
    /// Initial block timestamp
    pub timestamp: u64,
    /// Block time in seconds
    pub block_time: u64,
    /// Gas limit per block
    pub gas_limit: u64,
}

impl Default for Genesis {
    fn default() -> Self {
        Self {
            validator: [0u8; 20], // All zeros = default validator
            timestamp: 0,
            block_time: 5, // 5 second block time
            gas_limit: 30_000_000,
        }
    }
}

/// Consensus engine for PoA block production
pub struct Consensus {
    /// Consensus configuration
    genesis: Genesis,
    /// State database
    db: Arc<RwLock<StateDB>>,
    /// Transaction mempool
    mempool: Arc<RwLock<Mempool>>,
}

impl Consensus {
    /// Create a new consensus engine
    pub fn new(genesis: Genesis, db: Arc<RwLock<StateDB>>, mempool: Arc<RwLock<Mempool>>) -> Self {
        Self {
            genesis,
            db,
            mempool,
        }
    }

    /// Initialize genesis block if not exists
    pub async fn init_genesis(&self) -> Result<()> {
        let db = self.db.read().await;

        // Check if chain already initialized
        if db.get_block_number()? > 0 {
            tracing::info!("Chain already initialized");
            return Ok(());
        }

        drop(db);

        // Create genesis block
        let genesis_block = self.create_genesis_block()?;

        // Store genesis block
        {
            let mut db = self.db.write().await;
            db.put_block(&genesis_block)?;
        }

        tracing::info!(
            "Genesis block initialized: {:?}",
            genesis_block.header.hash()
        );
        Ok(())
    }

    /// Create the genesis block
    fn create_genesis_block(&self) -> Result<Block> {
        let header = Header::new(
            ZERO_HASH,              // parent_hash
            ZERO_HASH,              // state_root (empty state)
            hash(&[]),              // tx_root (no transactions)
            hash(&[]),              // receipt_root (no receipts)
            self.genesis.timestamp, // timestamp
            0,                      // number (genesis = 0)
            0,                      // gas_used
            self.genesis.gas_limit, // gas_limit
            0,                      // nonce
            self.genesis.validator, // proposer
            vec![],                 // extra data
        );

        Ok(Block::new(header, vec![]))
    }

    /// Start the block production loop
    pub async fn start_block_production(&self) -> Result<()> {
        let mut ticker = interval(Duration::from_secs(self.genesis.block_time));
        let block_number = {
            let db = self.db.read().await;
            db.get_block_number()?
        };

        tracing::info!(
            "Starting block production: validator={}, block_time={}s",
            hex::encode(self.genesis.validator),
            self.genesis.block_time
        );

        // Start from current block + 1
        let mut next_block = block_number + 1;

        loop {
            ticker.tick().await;

            match self.produce_block(next_block).await {
                Ok(block) => {
                    tracing::info!(
                        "Block produced: #{} hash={} txs={}",
                        block.header.number,
                        hex::encode(block.header.hash()),
                        block.transactions.len()
                    );
                    next_block += 1;
                }
                Err(e) => {
                    tracing::error!("Block production failed: {}", e);
                    // Retry on next tick
                }
            }
        }
    }

    /// Produce a single block
    pub async fn produce_block(&self, number: u64) -> Result<Block> {
        // Get parent block hash
        let parent_hash = if number > 0 {
            let db = self.db.read().await;
            let parent = db
                .get_block_by_number(number - 1)?
                .ok_or_else(|| crate::Error::Custom("Parent block not found".into()))?;
            parent.header.hash()
        } else {
            ZERO_HASH
        };

        // Select transactions from mempool
        let transactions = self.select_transactions().await?;

        // Execute transactions and compute state root
        let (receipts, state_root, gas_used) = self.execute_transactions(&transactions).await?;

        // Compute tx root
        let tx_hashes: Vec<_> = transactions.iter().map(transaction_hash).collect();
        let tx_root = if tx_hashes.is_empty() {
            ZERO_HASH
        } else {
            merkle_root(&tx_hashes)
        };

        // Compute receipt root
        let receipt_hashes: Vec<_> = receipts.iter().map(|r| r.receipt_hash()).collect();
        let receipt_root = if receipt_hashes.is_empty() {
            ZERO_HASH
        } else {
            merkle_root(&receipt_hashes)
        };

        // Build block header
        let header = Header::new(
            parent_hash,            // parent_hash
            state_root,             // state_root
            tx_root,                // tx_root
            receipt_root,           // receipt_root
            chrono_now(),           // timestamp
            number,                 // number
            gas_used,               // gas_used
            self.genesis.gas_limit, // gas_limit
            0,                      // nonce (PoA doesn't use nonce for sealing)
            self.genesis.validator, // proposer
            vec![],                 // extra data
        );

        let block = Block::new(header, transactions);

        // Store block
        {
            let mut db = self.db.write().await;
            db.put_block(&block)?;
        }

        Ok(block)
    }

    /// Execute transactions and return receipts, state root, and gas used
    async fn execute_transactions(
        &self,
        transactions: &[SignedTransaction],
    ) -> Result<(Vec<Receipt>, crate::types::Hash, u64)> {
        let mut receipts = Vec::new();
        let mut gas_used = 0u64;

        // Clone state for execution
        let mut state = {
            let db = self.db.read().await;
            db.clone_to_exec()
        };

        // Execute each transaction
        for tx in transactions {
            match Self::execute_tx(&mut state, tx) {
                Ok(receipt) => {
                    gas_used += receipt.gas_used;
                    receipts.push(receipt);
                }
                Err(e) => {
                    tracing::warn!("Transaction execution failed: {}", e);
                    // Continue with other transactions in production, but for MVP we stop
                    break;
                }
            }
        }

        // Compute state root
        let state_root = state.state_root();

        Ok((receipts, state_root, gas_used))
    }

    /// Execute a single transaction
    fn execute_tx(state: &mut StateDB, tx: &SignedTransaction) -> Result<Receipt> {
        // Check nonce
        let tx_nonce = state.nonce(&tx.from);
        if tx.nonce != tx_nonce {
            return Err(crate::Error::InvalidNonce {
                expected: tx_nonce,
                got: tx.nonce,
            });
        }

        // Check balance
        let sender_balance = state.balance(&tx.from);
        if sender_balance < tx.value {
            return Err(crate::Error::InsufficientBalance {
                have: sender_balance,
                need: tx.value,
            });
        }

        // Transfer value
        state.decrease_balance(&tx.from, tx.value)?;
        if let Some(ref to) = tx.to {
            state.increase_balance(to, tx.value);
        }

        // Increment nonce
        state.increase_nonce(&tx.from);

        let receipt = Receipt {
            transaction_hash: transaction_hash(tx),
            success: true,
            gas_used: 21000,
            logs: Vec::new(),
        };

        Ok(receipt)
    }

    /// Select transactions from mempool
    async fn select_transactions(&self) -> Result<Vec<SignedTransaction>> {
        let mut mempool = self.mempool.write().await;
        let mut selected = Vec::new();
        let mut total_gas = 0u64;

        // Sort by fee priority and select until gas limit
        mempool.sort_by_fee();

        let db = self.db.read().await;

        while let Some(tx) = mempool.pop_transaction() {
            // Check gas limit
            let tx_gas = tx.gas_limit;
            if total_gas + tx_gas > self.genesis.gas_limit {
                break;
            }

            // Validate nonce
            if let Some(account) = db.get_account(&tx.from) {
                if tx.nonce != account.nonce {
                    tracing::debug!(
                        "Skipping tx: nonce mismatch (expected {}, got {})",
                        account.nonce,
                        tx.nonce
                    );
                    continue;
                }
            } else if tx.nonce != 0 {
                tracing::debug!("Skipping tx: account doesn't exist, nonce != 0");
                continue;
            }

            // Basic signature check (in production, verify properly)
            if tx.signature.is_empty() {
                tracing::debug!("Skipping tx: no signature");
                continue;
            }

            selected.push(tx);
            total_gas += tx_gas;
        }

        Ok(selected)
    }

    /// Get current validator
    pub fn validator(&self) -> Address {
        self.genesis.validator
    }

    /// Get block time
    pub fn block_time(&self) -> u64 {
        self.genesis.block_time
    }
}

/// Get current timestamp (placeholder - in production use proper timestamp)
fn chrono_now() -> u64 {
    std::time::SystemTime::now()
        .duration_since(std::time::UNIX_EPOCH)
        .unwrap()
        .as_secs()
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_genesis_default() {
        let genesis = Genesis::default();
        assert_eq!(genesis.block_time, 5);
        assert_eq!(genesis.gas_limit, 30_000_000);
    }

    #[test]
    fn test_chrono_now() {
        let now = chrono_now();
        assert!(now > 0);
    }
}
