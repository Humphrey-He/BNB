// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import {ERC20} from "@openzeppelin/contracts/token/ERC20/ERC20.sol";

/**
 * @title MaliciousERC20
 * @notice An ERC-20 token with non-standard return values.
 * @dev Some tokens (like USDT) don't return booleans on transfer/approve.
 *      This mock simulates that behavior to test SafeERC20 wrapper.
 */
contract MaliciousERC20 is ERC20 {
    bool public shouldRevert;
    bool public shouldReturnFalse;

    /**
     * @param name_ Token name
     * @param symbol_ Token symbol
     */
    constructor(
        string memory name_,
        string memory symbol_
    ) ERC20(name_, symbol_) {}

    /**
     * @notice Set whether transfers should revert
     * @param _shouldRevert If true, transfer will revert
     */
    function setShouldRevert(bool _shouldRevert) external {
        shouldRevert = _shouldRevert;
    }

    /**
     * @notice Set whether transfer should return false instead of true
     * @param _shouldReturnFalse If true, transfer returns false
     */
    function setShouldReturnFalse(bool _shouldReturnFalse) external {
        shouldReturnFalse = _shouldReturnFalse;
    }

    /**
     * @notice Transfer without proper return value (like old USDT)
     */
    function transfer(address, uint256) public override returns (bool) {
        if (shouldRevert) {
            revert("MaliciousERC20: transfer reverted");
        }
        return !shouldReturnFalse;
    }

    /**
     * @notice TransferFrom without proper return value
     */
    function transferFrom(
        address,
        address,
        uint256
    ) public override returns (bool) {
        if (shouldRevert) {
            revert("MaliciousERC20: transferFrom reverted");
        }
        return !shouldReturnFalse;
    }

    /**
     * @notice Approve without proper return value
     */
    function approve(address, uint256) public override returns (bool) {
        if (shouldRevert) {
            revert("MaliciousERC20: approve reverted");
        }
        return !shouldReturnFalse;
    }

    /**
     * @notice Mint tokens for testing
     */
    function mint(address to, uint256 amount) external {
        _mint(to, amount);
    }
}
