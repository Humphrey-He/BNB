// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import {IERC20} from "@openzeppelin/contracts/token/ERC20/IERC20.sol";

/**
 * @title ReentrantReceiver
 * @notice A contract that can execute a callback during token transfers to test reentrancy.
 * @dev Used to test if vault properly guards against reentrancy attacks.
 */
contract ReentrantReceiver {
    address public targetToken;
    address public attacker;
    bool public shouldRevertOnCallback;
    bool public shouldReenter;

    uint256 public callbackCount;

    error ReentrancyBlocked();

    /**
     * @param _targetToken The ERC-20 token to use for reentrancy
     */
    constructor(address _targetToken) {
        targetToken = _targetToken;
    }

    /**
     * @notice Set the attacker address
     */
    function setAttacker(address _attacker) external {
        attacker = _attacker;
    }

    /**
     * @notice Configure whether to revert on callback
     */
    function setShouldRevertOnCallback(bool _shouldRevert) external {
        shouldRevertOnCallback = _shouldRevert;
    }

    /**
     * @notice Configure whether to re-enter during callback
     */
    function setShouldReenter(bool _shouldReenter) external {
        shouldReenter = _shouldReenter;
    }

    /**
     * @notice Callback function that can trigger reentrancy
     */
    function tokenReceived(address, uint256 amount, bytes calldata) external {
        callbackCount++;

        if (shouldRevertOnCallback) {
            revert("ReentrantReceiver: callback reverted");
        }

        if (shouldReenter && attacker != address(0)) {
            // Attempt to re-enter the vault
            bytes memory data = abi.encodeWithSignature(
                "withdraw(uint256)",
                amount
            );
            (bool success, ) = attacker.call(data);
            (success); // silence unused variable warning
        }
    }

    /**
     * @notice Reset callback counter
     */
    function resetCallbackCount() external {
        callbackCount = 0;
    }
}
