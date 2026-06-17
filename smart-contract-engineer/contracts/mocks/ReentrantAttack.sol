// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import {IERC20} from "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import {SafeERC20} from "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";

enum ReentrantAttackTarget {
    Deposit,
    Withdraw,
    Redeem
}

/**
 * @title ReentrantAttack
 * @notice A contract that attempts to re-enter the vault during deposit/withdraw/redeem.
 * @dev Used to test that the vault's ReentrancyGuard properly blocks re-entrancy attacks.
 */
contract ReentrantAttack {
    using SafeERC20 for IERC20;

    address public vault;
    address public asset;
    ReentrantAttackTarget public target;
    bool public attackInitiated;
    uint256 public reentryAttempts;

    constructor(address _vault, address _asset, ReentrantAttackTarget _target) {
        vault = _vault;
        asset = _asset;
        target = _target;
    }

    function attack(uint256 amount) external {
        attackInitiated = true;
        IERC20(asset).safeTransferFrom(msg.sender, address(this), amount);
        (bool success, ) = vault.call(abi.encodeWithSignature("deposit(uint256,address)", amount, address(this)));
        (success);
    }

    function attackWithdraw(uint256 assets, address owner) external {
        attackInitiated = true;
        (bool success, ) = vault.call(abi.encodeWithSignature("withdraw(uint256,address,address)", assets, address(this), owner));
        (success);
    }

    function attackRedeem(uint256 shares, address owner) external {
        attackInitiated = true;
        (bool success, ) = vault.call(abi.encodeWithSignature("redeem(uint256,address,address)", shares, address(this), owner));
        (success);
    }

    receive() external payable {
        reentryAttempts++;
        if (attackInitiated && reentryAttempts == 1) {
            if (target == ReentrantAttackTarget.Withdraw) {
                (bool success, ) = vault.call(abi.encodeWithSignature("withdraw(uint256,address,address)", 1, address(this), address(this)));
                (success);
            } else if (target == ReentrantAttackTarget.Redeem) {
                (bool success, ) = vault.call(abi.encodeWithSignature("redeem(uint256,address,address)", 1, address(this), address(this)));
                (success);
            }
        }
    }
}