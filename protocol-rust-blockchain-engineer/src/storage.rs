//! Storage abstraction with in-memory implementation.

use std::collections::{BTreeMap, HashMap};

use crate::crypto::header_hash;
use crate::error::Result;
use crate::types::{Account, Address, Block, Hash, Receipt};

/// Storage trait for blockchain data persistence.
/// Defines the interface that can be implemented with in-memory or RocksDB backend.
pub trait Storage: Send + Sync {
    fn put_block(&mut self, number: u64, block: &Block) -> Result<()>;
    fn get_block(&self, number: u64) -> Result<Option<Block>>;
    fn put_account_state(&mut self, addr: &Address, account: &Account) -> Result<()>;
    fn get_account_state(&self, addr: &Address) -> Result<Option<Account>>;
    fn put_tx_index(&mut self, tx_hash: &Hash, location: (u64, usize)) -> Result<()>;
    fn get_tx_index(&self, tx_hash: &Hash) -> Result<Option<(u64, usize)>>;
    fn put_receipt(&mut self, tx_hash: &Hash, receipt: &Receipt) -> Result<()>;
    fn get_receipt(&self, tx_hash: &Hash) -> Result<Option<Receipt>>;
    fn put_block_hash(&mut self, number: u64, hash: &Hash) -> Result<()>;
    fn get_block_hash(&self, number: u64) -> Result<Option<Hash>>;
    fn update_canonical_chain(&mut self, blocks: &[(u64, Hash)]) -> Result<()>;
}

/// In-memory storage implementation for testing and development.
#[derive(Debug, Clone, Default)]
pub struct StorageMem {
    blocks: HashMap<Hash, Block>,
    accounts: HashMap<Address, Account>,
    tx_index: HashMap<Hash, (u64, usize)>,
    receipts: HashMap<Hash, Receipt>,
    block_hashes: BTreeMap<u64, Hash>,
    canonical_chain: BTreeMap<u64, Hash>,
}

impl StorageMem {
    pub fn new() -> Self {
        Self {
            blocks: HashMap::new(),
            accounts: HashMap::new(),
            tx_index: HashMap::new(),
            receipts: HashMap::new(),
            block_hashes: BTreeMap::new(),
            canonical_chain: BTreeMap::new(),
        }
    }
}

impl Storage for StorageMem {
    fn put_block(&mut self, _number: u64, block: &Block) -> Result<()> {
        let hash = header_hash(&block.header);
        self.blocks.insert(hash, block.clone());
        Ok(())
    }

    fn get_block(&self, number: u64) -> Result<Option<Block>> {
        let hash = self.block_hashes.get(&number).copied();
        match hash {
            Some(h) => Ok(self.blocks.get(&h).cloned()),
            None => Ok(None),
        }
    }

    fn put_account_state(&mut self, addr: &Address, account: &Account) -> Result<()> {
        self.accounts.insert(*addr, account.clone());
        Ok(())
    }

    fn get_account_state(&self, addr: &Address) -> Result<Option<Account>> {
        Ok(self.accounts.get(addr).cloned())
    }

    fn put_tx_index(&mut self, tx_hash: &Hash, location: (u64, usize)) -> Result<()> {
        self.tx_index.insert(*tx_hash, location);
        Ok(())
    }

    fn get_tx_index(&self, tx_hash: &Hash) -> Result<Option<(u64, usize)>> {
        Ok(self.tx_index.get(tx_hash).copied())
    }

    fn put_receipt(&mut self, tx_hash: &Hash, receipt: &Receipt) -> Result<()> {
        self.receipts.insert(*tx_hash, receipt.clone());
        Ok(())
    }

    fn get_receipt(&self, tx_hash: &Hash) -> Result<Option<Receipt>> {
        Ok(self.receipts.get(tx_hash).cloned())
    }

    fn put_block_hash(&mut self, number: u64, hash: &Hash) -> Result<()> {
        self.block_hashes.insert(number, *hash);
        Ok(())
    }

    fn get_block_hash(&self, number: u64) -> Result<Option<Hash>> {
        Ok(self.block_hashes.get(&number).copied())
    }

    fn update_canonical_chain(&mut self, blocks: &[(u64, Hash)]) -> Result<()> {
        for (number, hash) in blocks {
            self.canonical_chain.insert(*number, *hash);
            self.block_hashes.insert(*number, *hash);
        }
        Ok(())
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::types::{Header, ZERO_HASH};

    fn create_test_block(number: u64) -> Block {
        Block::new(
            Header::new(
                ZERO_HASH,
                ZERO_HASH,
                ZERO_HASH,
                ZERO_HASH,
                number * 1000,
                number,
                0,
                0,
                0,
                [0u8; 20],
                vec![],
            ),
            vec![],
        )
    }

    fn create_test_account(balance: u128, nonce: u64) -> Account {
        Account::new(balance, nonce)
    }

    #[test]
    fn test_put_and_get_block() {
        let mut storage = StorageMem::new();
        let block = create_test_block(1);
        let hash = header_hash(&block.header);
        storage.put_block(1, &block).unwrap();
        storage.update_canonical_chain(&[(1, hash)]).unwrap();
        let retrieved = storage.get_block(1).unwrap();
        assert!(retrieved.is_some());
        assert_eq!(retrieved.unwrap().header.number, 1);
    }

    #[test]
    fn test_get_nonexistent_block() {
        let storage = StorageMem::new();
        assert!(storage.get_block(999).unwrap().is_none());
    }

    #[test]
    fn test_put_and_get_account() {
        let mut storage = StorageMem::new();
        let addr = [1u8; 20];
        let account = create_test_account(100, 5);
        storage.put_account_state(&addr, &account).unwrap();
        let retrieved = storage.get_account_state(&addr).unwrap();
        assert!(retrieved.is_some());
        assert_eq!(retrieved.unwrap().balance, 100);
    }

    #[test]
    fn test_put_and_get_tx_index() {
        let mut storage = StorageMem::new();
        let tx_hash = [0x42u8; 32];
        storage.put_tx_index(&tx_hash, (5, 10)).unwrap();
        assert_eq!(storage.get_tx_index(&tx_hash).unwrap(), Some((5, 10)));
    }

    #[test]
    fn test_put_and_get_receipt() {
        let mut storage = StorageMem::new();
        let tx_hash = [0x42u8; 32];
        let receipt = Receipt {
            transaction_hash: tx_hash,
            success: true,
            gas_used: 21000,
            logs: vec![],
        };
        storage.put_receipt(&tx_hash, &receipt).unwrap();
        assert!(storage.get_receipt(&tx_hash).unwrap().unwrap().success);
    }

    #[test]
    fn test_update_canonical_chain() {
        let mut storage = StorageMem::new();
        let blocks = vec![(1, [0x01u8; 32]), (2, [0x02u8; 32])];
        storage.update_canonical_chain(&blocks).unwrap();
        assert_eq!(storage.get_block_hash(1).unwrap(), Some([0x01u8; 32]));
    }
}
