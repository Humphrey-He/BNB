//! State database for storing account data.

use std::collections::HashMap;

use crate::crypto::{hash, merkle_root};
use crate::error::{Error, Result};
use crate::types::{Account, Address, Hash, ZERO_HASH};

#[derive(Debug, Clone)]
pub struct StateDB {
    accounts: HashMap<Address, Account>,
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
}
