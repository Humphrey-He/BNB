//! RPC module for the blockchain node.
//!
//! Provides HTTP RPC API for:
//! - `get_balance` - Query account balance
//! - `get_block` - Query block by number or hash
//! - `send_transaction` - Submit a signed transaction
//! - `get_transaction` - Query transaction by hash
//! - `get_block_number` - Query current block height
//! - `get_code` - Query contract code at address

use axum::{extract::State, http::StatusCode, response::IntoResponse, routing::post, Json, Router};
use serde::{Deserialize, Serialize};
use std::sync::Arc;
use tokio::sync::RwLock;

use crate::StateDB;

/// Shared application state for RPC handlers
#[derive(Clone)]
pub struct RpcState {
    pub db: Arc<RwLock<StateDB>>,
}

/// Initialize and start the RPC server
pub async fn start_rpc_server(state: RpcState, port: u16) -> std::result::Result<(), crate::Error> {
    let app = Router::new()
        .route("/rpc/get_balance", post(get_balance))
        .route("/rpc/get_block", post(get_block))
        .route("/rpc/get_block_by_number", post(get_block_by_number))
        .route("/rpc/send_transaction", post(send_transaction))
        .route("/rpc/get_transaction", post(get_transaction))
        .route("/rpc/get_block_number", post(get_block_number))
        .route("/rpc/get_code", post(get_code))
        .with_state(state);

    // Bind to loopback by default. In production, front this with a reverse
    // proxy (OpenResty) that binds the public port. Override with RPC_BIND=0.0.0.0
    // if you intentionally want to expose the RPC server directly.
    let bind_addr = std::env::var("RPC_BIND").unwrap_or_else(|_| "127.0.0.1".to_string());
    let addr = format!("{}:{}", bind_addr, port);
    let listener = tokio::net::TcpListener::bind(&addr).await?;
    tracing::info!("RPC server listening on {}", addr);

    axum::serve(listener, app).await?;
    Ok(())
}

// ============================================================================
// RPC Request/Response Types
// ============================================================================

#[derive(Debug, Deserialize)]
pub struct GetBalanceRequest {
    pub address: String,
    #[serde(default)]
    pub block_number: Option<u64>,
}

#[derive(Debug, Serialize)]
pub struct GetBalanceResponse {
    pub address: String,
    pub balance: String,
    pub block_number: u64,
}

#[derive(Debug, Deserialize)]
pub struct GetBlockRequest {
    #[serde(default)]
    pub block_number: Option<u64>,
    #[serde(default)]
    pub block_hash: Option<String>,
}

#[derive(Debug, Serialize)]
pub struct GetBlockResponse {
    pub number: u64,
    pub hash: String,
    pub parent_hash: String,
    pub state_root: String,
    pub tx_root: String,
    pub receipt_root: String,
    pub timestamp: u64,
    pub gas_used: u64,
    pub gas_limit: u64,
    pub proposer: String,
    pub tx_count: usize,
}

#[derive(Debug, Deserialize)]
pub struct SendTransactionRequest {
    pub from: String,
    pub to: Option<String>,
    pub value: String,
    pub nonce: u64,
    pub gas_limit: u64,
    pub max_fee_per_gas: String,
    pub signature: String,
}

#[derive(Debug, Serialize)]
pub struct SendTransactionResponse {
    pub tx_hash: String,
    pub nonce: u64,
}

#[derive(Debug, Deserialize)]
pub struct GetTransactionRequest {
    pub tx_hash: String,
}

#[derive(Debug, Serialize)]
pub struct GetTransactionResponse {
    pub tx_hash: String,
    pub block_number: Option<u64>,
    pub from: String,
    pub to: Option<String>,
    pub value: String,
    pub nonce: u64,
    pub success: bool,
    pub gas_used: u64,
}

#[derive(Debug, Deserialize)]
pub struct GetCodeRequest {
    pub address: String,
    #[serde(default)]
    pub block_number: Option<u64>,
}

#[derive(Debug, Serialize)]
pub struct GetCodeResponse {
    pub address: String,
    pub code: String,
    pub block_number: u64,
}

// ============================================================================
// RPC Handlers
// ============================================================================

/// GET_BALANCE - Query account balance
async fn get_balance(
    State(state): State<RpcState>,
    Json(req): Json<GetBalanceRequest>,
) -> std::result::Result<Json<GetBalanceResponse>, RpcError> {
    let address = parse_address(&req.address)?;

    let db = state.db.read().await;
    let account = db.get_account(&address);

    let balance = account
        .map(|a| a.balance.to_string())
        .unwrap_or_else(|| "0".to_string());
    let block_number = db
        .get_block_number()
        .map_err(|e| RpcError::Internal(e.to_string()))?;

    Ok(Json(GetBalanceResponse {
        address: req.address,
        balance,
        block_number,
    }))
}

/// GET_BLOCK - Query block by number or hash
async fn get_block(
    State(state): State<RpcState>,
    Json(req): Json<GetBlockRequest>,
) -> std::result::Result<Json<GetBlockResponse>, RpcError> {
    let db = state.db.read().await;

    let block = if let Some(hash) = req.block_hash {
        let hash_bytes = parse_hash(&hash)?;
        db.get_block_by_hash(&hash_bytes)
            .map_err(|e| RpcError::Internal(e.to_string()))?
    } else if let Some(number) = req.block_number {
        db.get_block_by_number(number)
            .map_err(|e| RpcError::Internal(e.to_string()))?
    } else {
        return Err(RpcError::InvalidParams(
            "Must provide either block_number or block_hash".into(),
        ));
    };

    match block {
        Some(block) => Ok(Json(GetBlockResponse {
            number: block.header.number,
            hash: hex::encode(block.header.hash()),
            parent_hash: hex::encode(block.header.parent_hash),
            state_root: hex::encode(block.header.state_root),
            tx_root: hex::encode(block.header.tx_root),
            receipt_root: hex::encode(block.header.receipt_root),
            timestamp: block.header.timestamp,
            gas_used: block.header.gas_used,
            gas_limit: block.header.gas_limit,
            proposer: hex::encode(block.header.proposer),
            tx_count: block.transactions.len(),
        })),
        None => Err(RpcError::NotFound("Block not found".into())),
    }
}

/// GET_BLOCK_BY_NUMBER - Query block by number (alias)
async fn get_block_by_number(
    State(state): State<RpcState>,
    Json(req): Json<GetBlockRequest>,
) -> std::result::Result<Json<GetBlockResponse>, RpcError> {
    let db = state.db.read().await;

    let block_number = req
        .block_number
        .ok_or_else(|| RpcError::InvalidParams("Must provide block_number".into()))?;

    let block = db
        .get_block_by_number(block_number)
        .map_err(|e| RpcError::Internal(e.to_string()))?;

    match block {
        Some(block) => Ok(Json(GetBlockResponse {
            number: block.header.number,
            hash: hex::encode(block.header.hash()),
            parent_hash: hex::encode(block.header.parent_hash),
            state_root: hex::encode(block.header.state_root),
            tx_root: hex::encode(block.header.tx_root),
            receipt_root: hex::encode(block.header.receipt_root),
            timestamp: block.header.timestamp,
            gas_used: block.header.gas_used,
            gas_limit: block.header.gas_limit,
            proposer: hex::encode(block.header.proposer),
            tx_count: block.transactions.len(),
        })),
        None => Err(RpcError::NotFound("Block not found".into())),
    }
}

/// SEND_TRANSACTION - Submit a signed transaction
async fn send_transaction(
    State(state): State<RpcState>,
    Json(req): Json<SendTransactionRequest>,
) -> std::result::Result<Json<SendTransactionResponse>, RpcError> {
    let _from = parse_address(&req.from)?;

    // In a real implementation, we would:
    // 1. Verify the signature
    // 2. Check nonce matches expected
    // 3. Add to mempool
    // For now, we just compute a tx hash

    let mut tx_data = Vec::new();
    tx_data.extend_from_slice(req.from.as_bytes());
    if let Some(ref to) = req.to {
        tx_data.extend_from_slice(to.as_bytes());
    }
    tx_data.extend_from_slice(req.value.as_bytes());
    tx_data.extend_from_slice(&req.nonce.to_le_bytes());
    let tx_hash = crate::crypto::hash(&tx_data);

    let db = state.db.read().await;

    // Get current block number for response
    let block_number = db
        .get_block_number()
        .map_err(|e| RpcError::Internal(e.to_string()))?;

    tracing::info!(
        "Transaction received: from={}, to={:?}, value={}, nonce={}, block={}",
        req.from,
        req.to,
        req.value,
        req.nonce,
        block_number
    );

    Ok(Json(SendTransactionResponse {
        tx_hash: hex::encode(tx_hash),
        nonce: req.nonce,
    }))
}

/// GET_TRANSACTION - Query transaction by hash
async fn get_transaction(
    State(state): State<RpcState>,
    Json(req): Json<GetTransactionRequest>,
) -> std::result::Result<Json<GetTransactionResponse>, RpcError> {
    let tx_hash = parse_hash(&req.tx_hash)?;

    let db = state.db.read().await;

    // Search for transaction in blocks
    // In a full implementation, we would maintain a tx hash -> (block, index) index
    let current_block = db
        .get_block_number()
        .map_err(|e| RpcError::Internal(e.to_string()))?;

    for block_num in 0..=current_block {
        if let Ok(Some(block)) = db.get_block_by_number(block_num) {
            for tx in &block.transactions {
                let mut tx_data = Vec::new();
                tx_data.extend_from_slice(&tx.from);
                tx_data.extend_from_slice(&tx.to.unwrap_or_default());
                tx_data.extend_from_slice(&tx.value.to_le_bytes());
                tx_data.extend_from_slice(&tx.nonce.to_le_bytes());
                let tx_hash_computed = crate::crypto::hash(&tx_data);
                if tx_hash_computed == tx_hash {
                    return Ok(Json(GetTransactionResponse {
                        tx_hash: req.tx_hash,
                        block_number: Some(block.header.number),
                        from: hex::encode(tx.from),
                        to: tx.to.map(hex::encode),
                        value: tx.value.to_string(),
                        nonce: tx.nonce,
                        success: true, // Would look up receipt
                        gas_used: 0,   // Would look up receipt
                    }));
                }
            }
        }
    }

    Err(RpcError::NotFound("Transaction not found".into()))
}

/// GET_BLOCK_NUMBER - Query current block height
async fn get_block_number(
    State(state): State<RpcState>,
) -> std::result::Result<Json<u64>, RpcError> {
    let db = state.db.read().await;
    let block_number = db
        .get_block_number()
        .map_err(|e| RpcError::Internal(e.to_string()))?;
    Ok(Json(block_number))
}

/// GET_CODE - Query contract code at address
async fn get_code(
    State(state): State<RpcState>,
    Json(req): Json<GetCodeRequest>,
) -> std::result::Result<Json<GetCodeResponse>, RpcError> {
    let address = parse_address(&req.address)?;

    let db = state.db.read().await;
    let account = db.get_account(&address);
    let block_number = db
        .get_block_number()
        .map_err(|e| RpcError::Internal(e.to_string()))?;

    let code = account
        .map(|a| {
            if a.code_hash == crate::types::ZERO_HASH {
                vec![]
            } else {
                // Would look up code from storage
                vec![]
            }
        })
        .unwrap_or_default();

    Ok(Json(GetCodeResponse {
        address: req.address,
        code: hex::encode(&code),
        block_number,
    }))
}

// ============================================================================
// Helper Functions
// ============================================================================

fn parse_address(addr: &str) -> std::result::Result<[u8; 20], RpcError> {
    let addr = addr.strip_prefix("0x").unwrap_or(addr);
    let bytes = hex::decode(addr).map_err(|e| RpcError::InvalidAddress(e.to_string()))?;
    if bytes.len() != 20 {
        return Err(RpcError::InvalidAddress(format!(
            "Expected 20 bytes, got {}",
            bytes.len()
        )));
    }
    let mut address = [0u8; 20];
    address.copy_from_slice(&bytes);
    Ok(address)
}

fn parse_hash(hash: &str) -> std::result::Result<[u8; 32], RpcError> {
    let hash = hash.strip_prefix("0x").unwrap_or(hash);
    let bytes = hex::decode(hash).map_err(|e| RpcError::InvalidHash(e.to_string()))?;
    if bytes.len() != 32 {
        return Err(RpcError::InvalidHash(format!(
            "Expected 32 bytes, got {}",
            bytes.len()
        )));
    }
    let mut h = [0u8; 32];
    h.copy_from_slice(&bytes);
    Ok(h)
}

// ============================================================================
// RPC Error Types
// ============================================================================

#[derive(Debug)]
pub enum RpcError {
    InvalidParams(String),
    InvalidAddress(String),
    InvalidHash(String),
    NotFound(String),
    Internal(String),
}

impl IntoResponse for RpcError {
    fn into_response(self) -> axum::response::Response {
        let (status, message) = match self {
            RpcError::InvalidParams(msg) => (StatusCode::BAD_REQUEST, msg),
            RpcError::InvalidAddress(msg) => (StatusCode::BAD_REQUEST, msg),
            RpcError::InvalidHash(msg) => (StatusCode::BAD_REQUEST, msg),
            RpcError::NotFound(msg) => (StatusCode::NOT_FOUND, msg),
            RpcError::Internal(msg) => (StatusCode::INTERNAL_SERVER_ERROR, msg),
        };

        let body = serde_json::json!({
            "error": message,
        });

        (status, Json(body)).into_response()
    }
}

impl From<crate::Error> for RpcError {
    fn from(e: crate::Error) -> Self {
        RpcError::Internal(e.to_string())
    }
}
