//! Cryptographic utilities.

use sha2::{Digest, Sha256};

use crate::types::{Address, Hash, SignedTransaction, ZERO_ADDRESS, ZERO_HASH};

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

/// Derive address from public key (simplified - just hash the key).
pub fn address_from_pubkey(pubkey: &[u8]) -> Address {
    let h = hash(pubkey);
    let mut addr = ZERO_ADDRESS;
    addr.copy_from_slice(&h[12..32]);
    addr
}

/// Verify ECDSA signature (placeholder - real impl would use k256 crate).
pub fn verify_signature(_pubkey: &[u8], signature: &[u8], _message: &[u8]) -> bool {
    // Placeholder: check signature is non-empty
    // Real implementation would use k256::ecdsa::VerifyingKey
    !signature.is_empty()
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
        assert!(!verify_signature(&[], &[], &[]));
        assert!(!verify_signature(&[1, 2, 3], &[], &[1, 2, 3]));
    }

    #[test]
    fn test_verify_signature_valid() {
        assert!(verify_signature(&[1, 2, 3], &[1, 2, 3], &[1, 2, 3]));
    }
}
