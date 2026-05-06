//! Binary entry point for the blockchain node.

use tracing_subscriber::{layer::SubscriberExt, util::SubscriberInitExt};

fn main() {
    // Initialize tracing/logging
    tracing_subscriber::registry()
        .with(
            tracing_subscriber::EnvFilter::try_from_default_env()
                .unwrap_or_else(|_| "info,verifiable_rust_chain_node=debug".into()),
        )
        .with(tracing_subscriber::fmt::layer())
        .init();

    tracing::info!("Starting verifiable-rust-chain-node");

    // Placeholder - actual node startup would go here
    // This will be implemented in later weeks
}