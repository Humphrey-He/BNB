//! Cryptographic utilities.

use k256::ecdsa::{Signature as K256Signature, VerifyingKey};
use sha2::{Digest as Sha2Digest, Sha256};

use crate::types::{Address, Hash, Header, SignedTransaction, ZERO_ADDRESS, ZERO_HASH};

/// Calculate hash of byte slice.
pub fn hash(data: &[u8]) -> Hash {
    let mut hasher = Sha256::new();
    hasher.update(data);
    let result = hasher.finalize();
    let mut hash = ZERO_HASH;
    hash.copy_from_slice(&result);
    hash
}

/// Calculate hash of a transaction for signing.
/// Hashes: from || to || value || nonce || gas_limit || max_fee_per_gas
pub fn transaction_hash(tx: &SignedTransaction) -> Hash {
    let mut data = Vec::new();
    data.extend_from_slice(&tx.from);
    if let Some(ref to) = tx.to {
        data.extend_from_slice(to);
    } else {
        data.extend_from_slice(&ZERO_ADDRESS);
    }
    data.extend_from_slice(&tx.value.to_le_bytes());
    data.extend_from_slice(&tx.nonce.to_le_bytes());
    data.extend_from_slice(&tx.gas_limit.to_le_bytes());
    data.extend_from_slice(&tx.max_fee_per_gas.to_le_bytes());
    hash(&data)
}

/// Calculate hash of a block header.
pub fn header_hash(header: &Header) -> Hash {
    let bytes = bincode::serialize(header).expect("header serialization failed");
    hash(&bytes)
}

/// Derive address from public key using keccak256 (Ethereum's hash).
pub fn address_from_pubkey(pubkey: &[u8]) -> Address {
    let h = keccak256(pubkey);
    let mut addr = ZERO_ADDRESS;
    addr.copy_from_slice(&h[12..32]);
    addr
}

/// Compute keccak256 hash (Ethereum's hash function).
fn keccak256(data: &[u8]) -> Hash {
    use sha3::{Digest, Keccak256};
    let mut hasher = Keccak256::new();
    hasher.update(data);
    let result = hasher.finalize();
    let mut hash = ZERO_HASH;
    hash.copy_from_slice(&result);
    hash
}

/// Verify ECDSA signature using secp256k1.
/// Signature format: 65 bytes (r:32 bytes + s:32 bytes + v:1 byte)
/// Message: 32 bytes (sha256 hash of transaction data)
/// Address: 20 bytes (derived from recovered public key via keccak256)
pub fn verify_signature(pubkey: &[u8], signature: &[u8], message: &[u8]) -> bool {
    // Signature must be 65 bytes (64 bytes r,s + 1 byte recovery id)
    if signature.len() != 65 {
        return false;
    }

    // Address must be 20 bytes
    if pubkey.len() != 20 {
        return false;
    }

    // Message must be 32 bytes (hash)
    if message.len() != 32 {
        return false;
    }

    // Parse r and s from first 64 bytes
    let sig_bytes = &signature[..64];
    let Ok(sig) = K256Signature::from_slice(sig_bytes) else {
        return false;
    };

    // Recover verifying key from message and signature
    // Recovery id is encoded in the last byte of the 65-byte signature
    let Ok(recovery_id) = k256::ecdsa::RecoveryId::try_from(signature[64]) else {
        return false;
    };
    let Ok(verifying_key) = VerifyingKey::recover_from_msg(message, &sig, recovery_id) else {
        return false;
    };

    // Derive address from recovered public key
    // Public key is encoded point: 04 || x || y (uncompressed)
    // Address is keccak256(pubkey)[12:]
    let encoded_point = verifying_key.to_encoded_point(false);
    let pubkey_bytes = encoded_point.as_bytes();
    // Skip the 0x04 prefix byte and hash the x||y coordinates
    let pubkey_hash = keccak256(&pubkey_bytes[1..]);
    let mut recovered_address = ZERO_ADDRESS;
    recovered_address.copy_from_slice(&pubkey_hash[12..32]);

    recovered_address == pubkey
}

/// Compute merkle root from a list of hashes.
pub fn merkle_root(hashes: &[Hash]) -> Hash {
    if hashes.is_empty() {
        return ZERO_HASH;
    }
    if hashes.len() == 1 {
        return hashes[0];
    }

    let mut level: Vec<Hash> = hashes.to_vec();
    while level.len() > 1 {
        let mut next_level = Vec::new();
        for pair in level.chunks(2) {
            let mut combined = Vec::new();
            combined.extend_from_slice(&pair[0]);
            combined.extend_from_slice(if pair.len() > 1 { &pair[1] } else { &pair[0] });
            next_level.push(hash(&combined));
        }
        level = next_level;
    }
    level[0]
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_hash() {
        let h = hash(b"hello");
        assert_ne!(h, ZERO_HASH);
    }

    #[test]
    fn test_transaction_hash() {
        let tx = SignedTransaction::new(
            [1u8; 20],
            Some([2u8; 20]),
            100,
            0,
            21000,
            1_000_000_000,
            vec![],
        );
        let h = transaction_hash(&tx);
        assert_ne!(h, ZERO_HASH);
    }

    #[test]
    fn test_transaction_hash_stability() {
        // Same tx should produce same hash
        let tx1 = SignedTransaction::new(
            [1u8; 20],
            Some([2u8; 20]),
            100,
            0,
            21000,
            1_000_000_000,
            vec![1, 2, 3],
        );
        let tx2 = SignedTransaction::new(
            [1u8; 20],
            Some([2u8; 20]),
            100,
            0,
            21000,
            1_000_000_000,
            vec![1, 2, 3],
        );
        assert_eq!(transaction_hash(&tx1), transaction_hash(&tx2));
    }

    #[test]
    fn test_merkle_root_empty() {
        let h = merkle_root(&[]);
        assert_eq!(h, ZERO_HASH);
    }

    #[test]
    fn test_merkle_root_single() {
        let single = [0u8; 32];
        let h = merkle_root(&[single]);
        assert_eq!(h, single);
    }

    #[test]
    fn test_merkle_root_even() {
        let hashes = [[1u8; 32], [2u8; 32], [3u8; 32], [4u8; 32]];
        let h = merkle_root(&hashes);
        assert_ne!(h, ZERO_HASH);
        assert_ne!(h, hashes[0]);
    }

    #[test]
    fn test_merkle_root_odd() {
        let hashes = [[1u8; 32], [2u8; 32], [3u8; 32]];
        let h = merkle_root(&hashes);
        assert_ne!(h, ZERO_HASH);
    }

    #[test]
    fn test_address_from_pubkey_stability() {
        let pubkey = b"test public key";
        let addr1 = address_from_pubkey(pubkey);
        let addr2 = address_from_pubkey(pubkey);
        assert_eq!(addr1, addr2);
        assert_ne!(addr1, ZERO_ADDRESS);
    }

    #[test]
    fn test_verify_signature_empty() {
        // Empty signatures should fail
        assert!(!verify_signature(&[1u8; 20], &[], &[0u8; 32]));
        // Signature too short
        assert!(!verify_signature(&[1u8; 20], &[1, 2, 3], &[0u8; 32]));
    }

    #[test]
    fn test_verify_signature_valid() {
        use k256::ecdsa::signature::Signer;

        // Generate a real key pair
        let signing_key = k256::ecdsa::SigningKey::random(&mut rand_core::OsRng);
        let verifying_key = k256::ecdsa::VerifyingKey::from(&signing_key);

        // Derive the address from the verifying key
        let encoded = verifying_key.to_encoded_point(false);
        let pubkey_bytes = encoded.as_bytes();
        let pubkey_hash = keccak256(&pubkey_bytes[1..]); // skip 0x04 prefix
        let mut address = ZERO_ADDRESS;
        address.copy_from_slice(&pubkey_hash[12..32]);

        // Create a message (transaction hash)
        let message: [u8; 32] = [0x01u8; 32];

        // Sign the message
        let signature: k256::ecdsa::Signature = signing_key.sign(&message);

        // Convert to Ethereum format: r || s || v
        let mut eth_signature = Vec::with_capacity(65);
        eth_signature.extend_from_slice(&signature.to_bytes());

        // For recovery, we need to figure out which y-parity the r point has
        // The recovery id v = 0 if y of r is even, 1 if odd
        // We can determine this by checking if the recovered key matches
        // Try v=0 first, then v=1
        for v in [0u8, 1u8, 2u8, 3u8] {
            let mut test_sig = eth_signature.clone();
            test_sig.push(v);
            if verify_signature(&address, &test_sig, &message) {
                return; // Success
            }
        }
        panic!("Could not find valid recovery id for signature");
    }
}
