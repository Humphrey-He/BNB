//! Block and transaction validation.

use crate::crypto::{merkle_root, transaction_hash, verify_signature};
use crate::error::{Error, Result};
use crate::executor::Executor;
use crate::types::{Block, Hash, Header, ZERO_HASH};
use crate::StateDB;

/// Validator for blocks and headers.
#[derive(Debug, Clone)]
pub struct Validator;

impl Validator {
    pub fn new() -> Self {
        Self
    }

    pub fn validate_header(&self, header: &Header, expected_parent: Option<&Hash>) -> Result<()> {
        if header.timestamp == 0 {
            return Err(Error::Execution("timestamp cannot be zero".to_string()));
        }
        if header.gas_limit == 0 {
            return Err(Error::Execution("gas_limit cannot be zero".to_string()));
        }
        if header.gas_used > header.gas_limit {
            return Err(Error::Execution(format!(
                "gas_used {} exceeds gas_limit {}",
                header.gas_used, header.gas_limit
            )));
        }
        if header.extra.len() > 10000 {
            return Err(Error::Execution("extra data too large".to_string()));
        }
        if let Some(expected) = expected_parent {
            if header.parent_hash != *expected {
                return Err(Error::Execution(format!(
                    "parent hash mismatch: expected {:?}, got {:?}",
                    expected, header.parent_hash
                )));
            }
        }
        Ok(())
    }

    pub fn validate_block(
        &self,
        block: &Block,
        state: &StateDB,
        expected_state_root: &Hash,
    ) -> Result<()> {
        self.validate_header(&block.header, None)?;

        let tx_hashes: Vec<Hash> = block.transactions.iter().map(transaction_hash).collect();
        let computed_tx_root = if tx_hashes.is_empty() {
            ZERO_HASH
        } else {
            merkle_root(&tx_hashes)
        };

        if computed_tx_root != block.header.tx_root {
            return Err(Error::Execution(format!(
                "tx_root mismatch: expected {:?}, got {:?}",
                block.header.tx_root, computed_tx_root
            )));
        }

        for tx in &block.transactions {
            let tx_hash = transaction_hash(tx);
            if tx.signature.is_empty() {
                return Err(Error::Execution(
                    "transaction has empty signature".to_string(),
                ));
            }
            if !verify_signature(&tx.from, &tx.signature, &tx_hash) {
                return Err(Error::Execution(
                    "invalid transaction signature".to_string(),
                ));
            }
        }

        let mut executor = Executor::new(state.clone());
        let outcome = executor.execute_block(block)?;

        if outcome.state_root != *expected_state_root {
            return Err(Error::Execution(format!(
                "state_root mismatch: expected {:?}, got {:?}",
                expected_state_root, outcome.state_root
            )));
        }

        if outcome.receipt_root != block.header.receipt_root {
            return Err(Error::Execution(format!(
                "receipt_root mismatch: expected {:?}, got {:?}",
                block.header.receipt_root, outcome.receipt_root
            )));
        }

        if outcome.gas_used != block.header.gas_used {
            return Err(Error::Execution(format!(
                "gas_used mismatch: expected {}, got {}",
                block.header.gas_used, outcome.gas_used
            )));
        }

        Ok(())
    }

    pub fn check_parent_link(&self, header: &Header, parent_hash: &Hash) -> Result<()> {
        if header.parent_hash != *parent_hash {
            return Err(Error::Execution(format!(
                "parent link broken: header parent_hash {:?} != expected {:?}",
                header.parent_hash, parent_hash
            )));
        }
        Ok(())
    }

    pub fn check_merkle_proof(
        &self,
        expected_root: &Hash,
        items: &[Hash],
        _proof: &[Hash],
    ) -> bool {
        let computed_root = if items.is_empty() {
            ZERO_HASH
        } else {
            merkle_root(items)
        };
        computed_root == *expected_root
    }
}

impl Default for Validator {
    fn default() -> Self {
        Self::new()
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::types::Header;

    fn create_test_header(number: u64, parent_hash: Hash) -> Header {
        Header::new(
            parent_hash,
            ZERO_HASH,
            ZERO_HASH,
            ZERO_HASH,
            number * 1000,
            number,
            0,
            1000000,
            number,
            [0u8; 20],
            vec![],
        )
    }

    #[test]
    fn test_validate_header_valid() {
        let validator = Validator::new();
        let header = create_test_header(1, ZERO_HASH);
        assert!(validator.validate_header(&header, Some(&ZERO_HASH)).is_ok());
    }

    #[test]
    fn test_validate_header_zero_timestamp() {
        let validator = Validator::new();
        let mut header = create_test_header(1, ZERO_HASH);
        header.timestamp = 0;
        assert!(validator.validate_header(&header, None).is_err());
    }

    #[test]
    fn test_validate_header_gas_used_exceeds_limit() {
        let validator = Validator::new();
        let mut header = create_test_header(1, ZERO_HASH);
        header.gas_limit = 1000;
        header.gas_used = 1001;
        assert!(validator.validate_header(&header, None).is_err());
    }

    #[test]
    fn test_check_parent_link_valid() {
        let validator = Validator::new();
        let parent_hash = [0x42u8; 32];
        let header = create_test_header(1, parent_hash);
        assert!(validator.check_parent_link(&header, &parent_hash).is_ok());
    }

    #[test]
    fn test_check_parent_link_invalid() {
        let validator = Validator::new();
        let header = create_test_header(1, ZERO_HASH);
        assert!(validator.check_parent_link(&header, &[0x42u8; 32]).is_err());
    }

    #[test]
    fn test_check_merkle_proof_valid() {
        let validator = Validator::new();
        let items = [[1u8; 32], [2u8; 32]];
        let root = merkle_root(&items);
        assert!(validator.check_merkle_proof(&root, &items, &[]));
    }

    #[test]
    fn test_check_merkle_proof_invalid() {
        let validator = Validator::new();
        let items = [[1u8; 32], [2u8; 32]];
        assert!(!validator.check_merkle_proof(&[0x42u8; 32], &items, &[]));
    }

    #[test]
    fn test_validate_block_empty_block() {
        let validator = Validator::new();
        let state = StateDB::new();
        let block = Block::new(
            Header::new(
                ZERO_HASH,
                ZERO_HASH,
                ZERO_HASH,
                ZERO_HASH,
                1000,
                1,
                0,
                1000000,
                1,
                [0u8; 20],
                vec![],
            ),
            vec![],
        );
        let mut executor = Executor::new(state.clone());
        let outcome = executor.execute_block(&block).unwrap();
        assert!(validator
            .validate_block(&block, &state, &outcome.state_root)
            .is_ok());
    }
}
