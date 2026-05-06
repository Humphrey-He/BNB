//! Error types for the blockchain node.

use thiserror::Error;

#[derive(Error, Debug)]
pub enum Error {
    #[error("invalid transaction: {0}")]
    InvalidTransaction(String),

    #[error("insufficient balance: have {have}, need {need}")]
    InsufficientBalance { have: u128, need: u128 },

    #[error("invalid nonce: expected {expected}, got {got}")]
    InvalidNonce { expected: u64, got: u64 },

    #[error("block not found: {0:?}")]
    BlockNotFound(crate::types::Hash),

    #[error("account not found: {0:?}")]
    AccountNotFound(crate::types::Address),

    #[error("storage error: {0}")]
    Storage(String),

    #[error("serialization error: {0}")]
    Serialization(String),

    #[error("execution error: {0}")]
    Execution(String),

    #[error("mempool error: {0}")]
    Mempool(String),

    #[error("p2p error: {0}")]
    P2P(String),

    #[error("consensus error: {0}")]
    Consensus(String),
}

pub type Result<T> = std::result::Result<T, Error>;
