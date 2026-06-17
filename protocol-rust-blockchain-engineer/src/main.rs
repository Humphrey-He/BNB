//! Binary entry point for the blockchain node.

use std::sync::Arc;
use tokio::sync::RwLock;
use tracing_subscriber::{layer::SubscriberExt, util::SubscriberInitExt};

use verifiable_rust_chain_node::{
    start_rpc_server, Consensus, Genesis, Mempool, RpcState, StateDB,
};

#[tokio::main]
async fn main() -> anyhow::Result<()> {
    // Initialize tracing/logging
    tracing_subscriber::registry()
        .with(
            tracing_subscriber::EnvFilter::try_from_default_env()
                .unwrap_or_else(|_| "info,verifiable_rust_chain_node=debug".into()),
        )
        .with(tracing_subscriber::fmt::layer())
        .init();

    tracing::info!("Starting verifiable-rust-chain-node");

    // Initialize components
    let db = Arc::new(RwLock::new(StateDB::new()));
    let mempool = Arc::new(RwLock::new(Mempool::new()));

    // Create genesis config
    let genesis = Genesis {
        validator: [0u8; 20], // Default validator
        timestamp: std::time::SystemTime::now()
            .duration_since(std::time::UNIX_EPOCH)
            .unwrap()
            .as_secs(),
        block_time: 5, // 5 second block time
        gas_limit: 30_000_000,
    };

    // Create consensus engine
    let consensus = Consensus::new(genesis.clone(), db.clone(), mempool.clone());

    // Initialize genesis block
    consensus.init_genesis().await?;

    // Clone db for RPC state
    let rpc_db = db.clone();

    // Start RPC server in background.
    // RPC_PORT defaults to 8081 to avoid colliding with the Go API server
    // (web3-blockchain-backend-engineer) which uses 8080.
    let rpc_port: u16 = std::env::var("RPC_PORT")
        .ok()
        .and_then(|p| p.parse().ok())
        .unwrap_or(8081);
    let rpc_state = RpcState { db: rpc_db };
    let rpc_handle = tokio::spawn(async move {
        if let Err(e) = start_rpc_server(rpc_state, rpc_port).await {
            tracing::error!("RPC server error: {}", e);
        }
    });

    // Start block production in background
    let consensus_handle = tokio::spawn(async move {
        if let Err(e) = consensus.start_block_production().await {
            tracing::error!("Consensus error: {}", e);
        }
    });

    tracing::info!("Node started successfully");
    tracing::info!("RPC server: http://localhost:{}", rpc_port);
    tracing::info!("Genesis validator: 0x{}", hex::encode(genesis.validator));

    // Wait for both handles
    tokio::select! {
        result = rpc_handle => {
            if let Err(e) = result {
                tracing::error!("RPC task panicked: {}", e);
            }
        }
        result = consensus_handle => {
            if let Err(e) = result {
                tracing::error!("Consensus task panicked: {}", e);
            }
        }
    }

    Ok(())
}
