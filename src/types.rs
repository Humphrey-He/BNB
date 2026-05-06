//! Core types for the blockchain node.

use serde::{Deserialize, Serialize};

/// Address type (20 bytes)
pub type Address = [u8; 20];

/// Hash type (32 bytes)
pub type Hash = [u8; 32];

/// Zero hash constant.
pub const ZERO_HASH: Hash = [0u8; 32];

/// Zero address constant.
pub const ZERO_ADDRESS: Address = [0u8; 20];

/// Transaction data with signature.
#[derive(Debug, Clone, PartialEq, Eq, Serialize, Deserialize)]
pub struct SignedTransaction {
    /// Sender address.
    pub from: Address,
    /// Recipient address (None for contract creation).
    pub to: Option<Address>,
    /// Transfer amount.
    pub value: u128,
    /// Transaction nonce for ordering.
    pub nonce: u64,
    /// Gas limit for this transaction.
    pub gas_limit: u64,
    /// Max fee per gas (for EIP-1559 style fee market).
    pub max_fee_per_gas: u64,
    /// ECDSA signature bytes.
    pub signature: Vec<u8>,
}

impl SignedTransaction {
    /// Create a new signed transaction.
    pub fn new(
        from: Address,
        to: Option<Address>,
        value: u128,
        nonce: u64,
        gas_limit: u64,
        max_fee_per_gas: u64,
        signature: Vec<u8>,
    ) -> Self {
        Self {
            from,
            to,
            value,
            nonce,
            gas_limit,
            max_fee_per_gas,
            signature,
        }
    }
}

/// Block containing header and transactions.
#[derive(Debug, Clone, PartialEq, Eq, Serialize, Deserialize)]
pub struct Block {
    /// Block header.
    pub header: Header,
    /// List of transactions in this block.
    pub transactions: Vec<SignedTransaction>,
}

impl Block {
    /// Create a new block with header and transactions.
    pub fn new(header: Header, transactions: Vec<SignedTransaction>) -> Self {
        Self {
            header,
            transactions,
        }
    }
}

/// Block header with all merkle roots and metadata.
#[derive(Debug, Clone, PartialEq, Eq, Serialize, Deserialize)]
pub struct Header {
    /// Parent block hash.
    pub parent_hash: Hash,
    /// State merkle root after this block.
    pub state_root: Hash,
    /// Transactions merkle root.
    pub tx_root: Hash,
    /// Receipts merkle root.
    pub receipt_root: Hash,
    /// Block timestamp.
    pub timestamp: u64,
    /// Block number (height).
    pub number: u64,
    /// Gas used in this block.
    pub gas_used: u64,
    /// Gas limit for this block.
    pub gas_limit: u64,
    /// Block nonce for PoA.
    pub nonce: u64,
    /// Proposer address (validator who produced this block).
    pub proposer: Address,
    /// Extra data field.
    pub extra: Vec<u8>,
}

impl Header {
    /// Create a new header.
    #[allow(clippy::too_many_arguments)]
    pub fn new(
        parent_hash: Hash,
        state_root: Hash,
        tx_root: Hash,
        receipt_root: Hash,
        timestamp: u64,
        number: u64,
        gas_used: u64,
        gas_limit: u64,
        nonce: u64,
        proposer: Address,
        extra: Vec<u8>,
    ) -> Self {
        Self {
            parent_hash,
            state_root,
            tx_root,
            receipt_root,
            timestamp,
            number,
            gas_used,
            gas_limit,
            nonce,
            proposer,
            extra,
        }
    }
}

/// Transaction receipt after execution.
#[derive(Debug, Clone, PartialEq, Eq, Serialize, Deserialize)]
pub struct Receipt {
    /// Transaction hash this receipt is for.
    pub transaction_hash: Hash,
    /// Whether transaction succeeded.
    pub success: bool,
    /// Gas used during execution.
    pub gas_used: u64,
    /// Log entries emitted.
    pub logs: Vec<Log>,
}

impl Receipt {
    /// Compute hash of receipt content for merkle root.
    pub fn receipt_hash(&self) -> Hash {
        // Hash: transaction_hash || success || gas_used || logs_hash
        use sha2::Digest;
        let mut hasher = sha2::Sha256::new();
        hasher.update(self.transaction_hash);
        hasher.update([self.success as u8]);
        hasher.update(self.gas_used.to_le_bytes());
        let logs_hash = if self.logs.is_empty() {
            ZERO_HASH
        } else {
            let log_hashes: Vec<Hash> = self
                .logs
                .iter()
                .map(|log| {
                    let mut h = sha2::Sha256::new();
                    h.update(log.address);
                    for topic in &log.topics {
                        h.update(*topic);
                    }
                    h.update(&log.data);
                    let mut hash = ZERO_HASH;
                    hash.copy_from_slice(&h.finalize());
                    hash
                })
                .collect();
            let mut combined = Vec::new();
            for lh in &log_hashes {
                combined.extend_from_slice(lh);
            }
            let mut overall = ZERO_HASH;
            overall.copy_from_slice(&sha2::Sha256::digest(&combined));
            overall
        };
        hasher.update(logs_hash);
        let mut result = ZERO_HASH;
        result.copy_from_slice(&hasher.finalize());
        result
    }
}

/// Log entry emitted by contract.
#[derive(Debug, Clone, PartialEq, Eq, Serialize, Deserialize)]
pub struct Log {
    /// Contract address that emitted this log.
    pub address: Address,
    /// Log topics.
    pub topics: Vec<Hash>,
    /// Log data.
    pub data: Vec<u8>,
}

/// Account state in the state tree.
#[derive(Debug, Clone, PartialEq, Eq, Serialize, Deserialize)]
pub struct Account {
    /// Account balance.
    pub balance: u128,
    /// Account nonce.
    pub nonce: u64,
    /// Code hash (empty for EOA).
    pub code_hash: Hash,
    /// Storage root (empty for EOA).
    pub storage_root: Hash,
}

impl Account {
    /// Create a new EOA account with zero balance and nonce.
    pub fn new(balance: u128, nonce: u64) -> Self {
        Self {
            balance,
            nonce,
            code_hash: ZERO_HASH,
            storage_root: ZERO_HASH,
        }
    }
}

/// Genesis configuration for chain startup.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Genesis {
    /// Genesis block header.
    pub header: Header,
    /// Initial account states.
    pub accounts: Vec<(Address, Account)>,
}
