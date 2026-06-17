//! State database for storing account data and chain state.

use std::collections::HashMap;

use crate::crypto::{hash, header_hash, merkle_root};
use crate::error::{Error, Result};
use crate::types::{Account, Address, Block, Hash, Header, ZERO_HASH};

/// Outcome of a block import operation.
#[derive(Debug, Clone)]
pub enum ImportOutcome {
    /// Block is now canonical.
    Canonical,
    /// Block was stored but is not yet canonical (on a fork).
    Forked,
    /// Block was already known.
    Known,
}

/// Snapshot of account state at a specific block height.
#[derive(Debug, Clone)]
struct AccountSnapshot {
    accounts: HashMap<Address, Account>,
}

impl AccountSnapshot {
    fn new(accounts: &HashMap<Address, Account>) -> Self {
        Self {
            accounts: accounts.clone(),
        }
    }

    fn restore(&self, target: &mut HashMap<Address, Account>) {
        *target = self.accounts.clone();
    }
}

/// Reversion entry describing changes made by a single block.
#[derive(Debug, Clone)]
struct ReversionEntry {
    #[allow(dead_code)]
    block_number: u64,
    sender_changes: HashMap<Address, AccountChange>,
}

#[derive(Debug, Clone)]
struct AccountChange {
    previous_balance: u128,
    previous_nonce: u64,
}

/// StateDB holds the current state and provides access to chain data.
#[derive(Debug, Clone)]
pub struct StateDB {
    accounts: HashMap<Address, Account>,
    blocks: HashMap<Hash, Block>,
    /// Canonical chain: ordered list of canonical block hashes.
    canonical_chain: Vec<Hash>,
    /// Account snapshots at each canonical block height (for revert).
    snapshots: HashMap<u64, AccountSnapshot>,
    /// Reversion log: block_number -> changes made by that block.
    revert_log: HashMap<u64, ReversionEntry>,
    current_block: u64,
}

impl Default for StateDB {
    fn default() -> Self {
        Self::new()
    }
}

impl StateDB {
    pub fn new() -> Self {
        Self {
            accounts: HashMap::new(),
            blocks: HashMap::new(),
            canonical_chain: Vec::new(),
            snapshots: HashMap::new(),
            revert_log: HashMap::new(),
            current_block: 0,
        }
    }

    pub fn get_account(&self, addr: &Address) -> Option<&Account> {
        self.accounts.get(addr)
    }

    pub fn get_account_mut(&mut self, addr: &Address) -> Option<&mut Account> {
        self.accounts.get_mut(addr)
    }

    pub fn set_account(&mut self, addr: Address, account: Account) {
        self.accounts.insert(addr, account);
    }

    pub fn nonce(&self, addr: &Address) -> u64 {
        self.get_account(addr).map(|a| a.nonce).unwrap_or(0)
    }

    pub fn balance(&self, addr: &Address) -> u128 {
        self.get_account(addr).map(|a| a.balance).unwrap_or(0)
    }

    pub fn increase_nonce(&mut self, addr: &Address) {
        let account = self
            .accounts
            .entry(*addr)
            .or_insert_with(|| Account::new(0, 0));
        account.nonce += 1;
    }

    pub fn decrease_balance(&mut self, addr: &Address, amount: u128) -> Result<()> {
        let account = self
            .accounts
            .get_mut(addr)
            .ok_or(Error::AccountNotFound(*addr))?;
        if account.balance < amount {
            return Err(Error::InsufficientBalance {
                have: account.balance,
                need: amount,
            });
        }
        account.balance -= amount;
        Ok(())
    }

    pub fn increase_balance(&mut self, addr: &Address, amount: u128) {
        let account = self
            .accounts
            .entry(*addr)
            .or_insert_with(|| Account::new(0, 0));
        account.balance += amount;
    }

    pub fn state_root(&self) -> Hash {
        if self.accounts.is_empty() {
            return ZERO_HASH;
        }
        let mut account_hashes: Vec<Hash> = self
            .accounts
            .iter()
            .map(|(addr, account)| {
                let mut data = Vec::new();
                data.extend_from_slice(addr);
                data.extend_from_slice(&account.balance.to_le_bytes());
                data.extend_from_slice(&account.nonce.to_le_bytes());
                hash(&data)
            })
            .collect();
        account_hashes.sort();
        merkle_root(&account_hashes)
    }

    // ========================================================================
    // Block Storage Methods
    // ========================================================================

    /// Get block by number (canonical chain only).
    pub fn get_block_by_number(&self, number: u64) -> Result<Option<Block>> {
        if number as usize >= self.canonical_chain.len() {
            return Ok(None);
        }
        let hash = self.canonical_chain[number as usize];
        Ok(self.blocks.get(&hash).cloned())
    }

    /// Get block by hash.
    pub fn get_block_by_hash(&self, hash: &Hash) -> Result<Option<Block>> {
        Ok(self.blocks.get(hash).cloned())
    }

    /// Get current block number.
    pub fn get_block_number(&self) -> Result<u64> {
        Ok(self.current_block)
    }

    /// Set current block number.
    pub fn set_block_number(&mut self, number: u64) -> Result<()> {
        self.current_block = number;
        Ok(())
    }

    /// Store a block without changing canonical chain.
    /// Use `import_block` for fork-aware block import.
    pub fn put_block(&mut self, block: &Block) -> Result<()> {
        let hash = header_hash(&block.header);
        self.blocks.insert(hash, block.clone());
        if block.header.number > self.current_block {
            self.current_block = block.header.number;
        }
        Ok(())
    }

    /// Get block hash at a specific number (canonical chain).
    pub fn get_block_hash(&self, number: u64) -> Result<Option<Hash>> {
        if number as usize >= self.canonical_chain.len() {
            return Ok(None);
        }
        Ok(Some(self.canonical_chain[number as usize]))
    }

    /// Get canonical chain length.
    pub fn canonical_len(&self) -> usize {
        self.canonical_chain.len()
    }

    /// Get canonical chain hashes.
    pub fn canonical_chain(&self) -> &[Hash] {
        &self.canonical_chain
    }

    /// Check if a block hash is part of the canonical chain.
    pub fn is_canonical(&self, hash: &Hash) -> bool {
        self.canonical_chain.contains(hash)
    }

    /// Snapshot current account state for later revert.
    fn save_snapshot(&mut self, block_number: u64) {
        self.snapshots
            .insert(block_number, AccountSnapshot::new(&self.accounts));
    }

    /// Set a block as canonical and update the canonical chain.
    /// If the block is not already stored, returns an error.
    /// If the block extends the current canonical chain, extends it.
    /// If the block is on a different fork, performs a reorg.
    pub fn set_canonical_block(&mut self, block_hash: &Hash) -> Result<()> {
        let block = self
            .blocks
            .get(block_hash)
            .ok_or(Error::BlockNotFound(*block_hash))?
            .clone();

        // Build chain from genesis to this block by walking backwards
        let mut chain_hashes = Vec::new();
        let mut current: Header = block.header.clone();
        loop {
            chain_hashes.push(header_hash(&current));
            if current.number == 0 {
                break;
            }
            let parent = self.blocks.get(&current.parent_hash).ok_or_else(|| {
                Error::Custom(format!(
                    "Parent block {} not found while building chain",
                    hex::encode(current.parent_hash)
                ))
            })?;
            current = parent.header.clone();
        }
        chain_hashes.reverse();

        // If new chain starts at same height as current canonical chain and is
        // identical up to that point, just extend. Otherwise reorg.
        let reorg_needed = if chain_hashes.len() == self.canonical_chain.len() {
            // Same length - check if different
            chain_hashes != self.canonical_chain
        } else if chain_hashes.len() < self.canonical_chain.len() {
            // Shorter chain - can't become canonical
            return Err(Error::Custom(
                "Cannot set shorter chain as canonical".into(),
            ));
        } else {
            // Longer - reorg from divergence point
            true
        };

        if reorg_needed {
            // Find divergence point
            let divergence = self
                .canonical_chain
                .iter()
                .zip(chain_hashes.iter())
                .position(|(a, b)| a != b)
                .unwrap_or(0);

            // Revert blocks from current canonical chain from the divergence point onwards
            for i in (divergence..self.canonical_chain.len()).rev() {
                let _to_revert = self.canonical_chain[i];
                if let Some(revert_entry) = self.revert_log.remove(&(i as u64)) {
                    self.apply_reversion(&revert_entry);
                }
            }

            // Truncate canonical chain to divergence point
            self.canonical_chain.truncate(divergence);
        }

        // Add new chain segment to canonical chain
        let start_len = self.canonical_chain.len();
        for (i, &hash) in chain_hashes.iter().enumerate() {
            if !self.canonical_chain.contains(&hash) {
                self.canonical_chain.push(hash);
                let block_num = (start_len + i) as u64;
                self.save_snapshot(block_num);
            }
        }

        self.current_block = (self.canonical_chain.len() as u64).saturating_sub(1);
        Ok(())
    }

    /// Apply a reversion entry to restore previous account state.
    fn apply_reversion(&mut self, entry: &ReversionEntry) {
        for (addr, change) in &entry.sender_changes {
            if let Some(account) = self.accounts.get_mut(addr) {
                account.balance = change.previous_balance;
                account.nonce = change.previous_nonce;
            }
        }
    }

    /// Revert canonical chain to a specific block number.
    /// Removes blocks at heights > `block_number` from canonical chain
    /// and restores account state from snapshots.
    pub fn revert_to_block(&mut self, block_number: u64) -> Result<()> {
        if block_number as usize >= self.canonical_chain.len() {
            return Err(Error::Custom(format!(
                "Block {} does not exist in canonical chain",
                block_number
            )));
        }

        // Revert all blocks above the target
        while self.canonical_chain.len() as u64 > block_number + 1 {
            let idx = self.canonical_chain.len() - 1;
            let _to_revert = self.canonical_chain.remove(idx);
            if let Some(entry) = self.revert_log.remove(&(idx as u64)) {
                self.apply_reversion(&entry);
            }
        }

        // Restore account state from snapshot at block_number
        if let Some(snapshot) = self.snapshots.get(&block_number) {
            snapshot.restore(&mut self.accounts);
        }

        self.current_block = block_number;
        Ok(())
    }

    /// Log account changes made by a block for later reversion.
    pub fn log_block_changes(
        &mut self,
        block_number: u64,
        sender: Address,
        prev_balance: u128,
        prev_nonce: u64,
    ) {
        let entry = self
            .revert_log
            .entry(block_number)
            .or_insert_with(|| ReversionEntry {
                block_number,
                sender_changes: HashMap::new(),
            });
        entry.sender_changes.insert(
            sender,
            AccountChange {
                previous_balance: prev_balance,
                previous_nonce: prev_nonce,
            },
        );
    }

    /// Fork-aware block import: store block and update canonical chain if needed.
    /// Returns the import outcome.
    pub fn import_block(&mut self, block: &Block) -> Result<ImportOutcome> {
        let block_hash = header_hash(&block.header);

        // Already have this block
        if self.blocks.contains_key(&block_hash) {
            if self.is_canonical(&block_hash) {
                return Ok(ImportOutcome::Known);
            }
            // Check if it extends current canonical chain
            if block.header.number as usize == self.canonical_chain.len()
                && (block.header.number == 0
                    || block.header.parent_hash
                        == *self.canonical_chain.last().unwrap_or(&ZERO_HASH))
            {
                // Extend canonical chain
                self.put_block(block)?;
                self.canonical_chain.push(block_hash);
                self.save_snapshot(block.header.number);
                self.current_block = block.header.number;
                return Ok(ImportOutcome::Canonical);
            }
            return Ok(ImportOutcome::Forked);
        }

        // Validate parent exists (for non-genesis)
        if block.header.number > 0 && !self.blocks.contains_key(&block.header.parent_hash) {
            return Err(Error::Custom(format!(
                "Parent block {} not found for block {}",
                hex::encode(block.header.parent_hash),
                block.header.number
            )));
        }

        // Store the block
        self.put_block(block)?;

        // Determine if this should become canonical
        let extends_canonical = block.header.number as usize == self.canonical_chain.len()
            && (block.header.number == 0
                || block.header.parent_hash == *self.canonical_chain.last().unwrap_or(&ZERO_HASH));

        if extends_canonical {
            self.canonical_chain.push(block_hash);
            self.save_snapshot(block.header.number);
            self.current_block = block.header.number;
            Ok(ImportOutcome::Canonical)
        } else {
            // Check if this block is on a chain that could become canonical
            // (i.e., it extends a known block and is longer than current canonical)
            let chain_len = self.estimate_chain_length(block)?;
            if chain_len > self.canonical_chain.len() {
                // This chain is longer - set as canonical
                self.set_canonical_block(&block_hash)?;
                Ok(ImportOutcome::Canonical)
            } else {
                Ok(ImportOutcome::Forked)
            }
        }
    }

    /// Estimate total difficulty chain length by walking back to genesis.
    /// Since we don't store difficulty, this returns the chain height.
    fn estimate_chain_length(&self, block: &Block) -> Result<usize> {
        let mut count = 0;
        let mut current_hash = header_hash(&block.header);
        loop {
            count += 1;
            if self.blocks.get(&current_hash).map(|b| b.header.number) == Some(0) {
                break;
            }
            let hdr = self
                .blocks
                .get(&current_hash)
                .ok_or_else(|| Error::Custom("Block not found in chain".into()))?;
            current_hash = hdr.header.parent_hash;
        }
        Ok(count)
    }

    pub fn clone_to_exec(&self) -> StateDB {
        self.clone()
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_new_account() {
        let state = StateDB::new();
        let zero_addr = [0u8; 20];
        assert_eq!(state.balance(&zero_addr), 0);
        assert_eq!(state.nonce(&zero_addr), 0);
    }

    #[test]
    fn test_set_and_get_account() {
        let mut state = StateDB::new();
        let addr = [1u8; 20];
        state.set_account(addr, Account::new(100, 1));
        assert_eq!(state.balance(&addr), 100);
        assert_eq!(state.nonce(&addr), 1);
    }

    #[test]
    fn test_increase_nonce() {
        let mut state = StateDB::new();
        let addr = [1u8; 20];
        state.increase_nonce(&addr);
        assert_eq!(state.nonce(&addr), 1);
        state.increase_nonce(&addr);
        assert_eq!(state.nonce(&addr), 2);
    }

    #[test]
    fn test_transfer() {
        let mut state = StateDB::new();
        let from = [1u8; 20];
        let to = [2u8; 20];
        state.set_account(from, Account::new(100, 0));

        state.decrease_balance(&from, 30).unwrap();
        state.increase_balance(&to, 30);

        assert_eq!(state.balance(&from), 70);
        assert_eq!(state.balance(&to), 30);
    }

    #[test]
    fn test_insufficient_balance() {
        let mut state = StateDB::new();
        let addr = [1u8; 20];
        state.set_account(addr, Account::new(50, 0));

        let result = state.decrease_balance(&addr, 100);
        assert!(result.is_err());
    }

    #[test]
    fn test_state_root_empty() {
        let state = StateDB::new();
        assert_eq!(state.state_root(), ZERO_HASH);
    }

    #[test]
    fn test_state_root_with_accounts() {
        let mut state = StateDB::new();
        state.set_account([1u8; 20], Account::new(100, 0));
        state.set_account([2u8; 20], Account::new(200, 0));
        let root = state.state_root();
        assert_ne!(root, ZERO_HASH);
    }

    // ========================================================================
    // Fork-Aware Storage Tests
    // ========================================================================

    fn make_block(number: u64, parent_hash: Hash, extra: Vec<u8>) -> Block {
        use crate::crypto::hash;
        Block::new(
            Header::new(
                parent_hash,
                ZERO_HASH,
                hash(&[]),
                hash(&[]),
                number * 1000,
                number,
                0,
                30_000_000,
                0,
                [0u8; 20],
                extra,
            ),
            vec![],
        )
    }

    #[test]
    fn test_import_block_genesis() {
        let mut state = StateDB::new();
        let genesis = make_block(0, ZERO_HASH, vec![]);
        let hash = header_hash(&genesis.header);

        let outcome = state.import_block(&genesis).unwrap();
        assert!(matches!(outcome, ImportOutcome::Canonical));
        assert_eq!(state.canonical_len(), 1);
        assert_eq!(state.get_block_hash(0).unwrap(), Some(hash));
        assert!(state.is_canonical(&hash));
    }

    #[test]
    fn test_import_block_extends_canonical() {
        let mut state = StateDB::new();
        let genesis = make_block(0, ZERO_HASH, vec![0]);
        let genesis_hash = header_hash(&genesis.header);
        state.import_block(&genesis).unwrap();

        let block1 = make_block(1, genesis_hash, vec![1]);
        let hash1 = header_hash(&block1.header);
        let outcome = state.import_block(&block1).unwrap();
        assert!(matches!(outcome, ImportOutcome::Canonical));
        assert_eq!(state.canonical_len(), 2);
        assert_eq!(state.get_block_hash(1).unwrap(), Some(hash1));
    }

    #[test]
    fn test_import_block_fork() {
        let mut state = StateDB::new();
        let genesis = make_block(0, ZERO_HASH, vec![0]);
        let genesis_hash = header_hash(&genesis.header);
        state.import_block(&genesis).unwrap();

        // Canonical block at height 1
        let block1a = make_block(1, genesis_hash, vec![1]);
        let hash1a = header_hash(&block1a.header);
        state.import_block(&block1a).unwrap();
        assert_eq!(state.canonical_len(), 2);

        // Alternate block at height 1 (different extra → different hash)
        let block1b = make_block(1, genesis_hash, vec![2]);
        let hash1b = header_hash(&block1b.header);
        let outcome = state.import_block(&block1b).unwrap();
        // Same height as canonical tip but different hash → stored as fork
        assert!(matches!(outcome, ImportOutcome::Forked));
        assert!(state.is_canonical(&hash1a));
        assert!(!state.is_canonical(&hash1b));
        // Both blocks are accessible by hash
        assert!(state.get_block_by_hash(&hash1a).unwrap().is_some());
        assert!(state.get_block_by_hash(&hash1b).unwrap().is_some());
    }

    #[test]
    fn test_set_canonical_block_reorg() {
        let mut state = StateDB::new();

        // Build canonical chain: 0 -> 1 -> 2
        let b0 = make_block(0, ZERO_HASH, vec![0]);
        let h0 = header_hash(&b0.header);
        state.import_block(&b0).unwrap();

        let b1 = make_block(1, h0, vec![1]);
        let h1 = header_hash(&b1.header);
        state.import_block(&b1).unwrap();

        let b2 = make_block(2, h1, vec![2]);
        let _h2 = header_hash(&b2.header);
        state.import_block(&b2).unwrap();

        assert_eq!(state.canonical_len(), 3);

        // Build alternate fork: 0 -> 1' -> 2' (different extra → different hashes)
        let b1p = make_block(1, h0, vec![10]);
        let h1p = header_hash(&b1p.header);
        state.import_block(&b1p).unwrap();

        let b2p = make_block(2, h1p, vec![20]);
        let h2p = header_hash(&b2p.header);
        state.import_block(&b2p).unwrap();

        // Should still be on original canonical
        assert!(state.is_canonical(&_h2));

        // Make the alternate fork canonical (same length but different chain)
        state.set_canonical_block(&h2p).unwrap();

        // Now should be on alternate chain
        assert!(state.is_canonical(&h2p));
        assert!(!state.is_canonical(&_h2));
        assert_eq!(state.canonical_len(), 3);
    }

    #[test]
    fn test_revert_to_block() {
        let mut state = StateDB::new();

        let b0 = make_block(0, ZERO_HASH, vec![0]);
        let h0 = header_hash(&b0.header);
        state.import_block(&b0).unwrap();

        let b1 = make_block(1, h0, vec![1]);
        let h1 = header_hash(&b1.header);
        state.import_block(&b1).unwrap();

        let b2 = make_block(2, h1, vec![2]);
        let _h2 = header_hash(&b2.header);
        state.import_block(&b2).unwrap();

        // Add account state at block 2
        state.set_account([1u8; 20], Account::new(100, 0));

        // Revert to block 1
        state.revert_to_block(1).unwrap();
        assert_eq!(state.canonical_len(), 2);
        assert_eq!(state.get_block_hash(2).unwrap(), None);
        // Account state restored from snapshot
        assert_eq!(state.balance(&[1u8; 20]), 0);
    }

    #[test]
    fn test_get_block_by_hash_still_works_after_fork() {
        let mut state = StateDB::new();
        let genesis = make_block(0, ZERO_HASH, vec![0]);
        let genesis_hash = header_hash(&genesis.header);
        state.import_block(&genesis).unwrap();

        let block1a = make_block(1, genesis_hash, vec![1]);
        let hash1a = header_hash(&block1a.header);
        state.import_block(&block1a).unwrap();

        let block1b = make_block(1, genesis_hash, vec![2]);
        let hash1b = header_hash(&block1b.header);
        state.import_block(&block1b).unwrap();

        // Both blocks should still be accessible by hash
        assert!(state.get_block_by_hash(&hash1a).unwrap().is_some());
        assert!(state.get_block_by_hash(&hash1b).unwrap().is_some());
    }
}
