// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import {ERC20} from "@openzeppelin/contracts/token/ERC20/ERC20.sol";

/**
 * @title MockERC20
 * @notice A standard ERC-20 token for testing purposes with mint and burn capabilities.
 * @dev Used as a mock asset for vault testing.
 */
contract MockERC20 is ERC20 {
    uint8 private _decimals;

    /**
     * @param name_ Token name
     * @param symbol_ Token symbol
     * @param decimals_ Token decimals (default: 18)
     */
    constructor(
        string memory name_,
        string memory symbol_,
        uint8 decimals_
    ) ERC20(name_, symbol_) {
        _decimals = decimals_;
    }

    /// @notice Returns the number of decimals used for token amounts
    function decimals() public view override returns (uint8) {
        return _decimals;
    }

    /**
     * @notice Mints new tokens to the specified address
     * @param to Address to receive the tokens
     * @param amount Amount of tokens to mint
     */
    function mint(address to, uint256 amount) external {
        _mint(to, amount);
    }

    /**
     * @notice Burns tokens from the specified address
     * @param from Address whose tokens will be burned
     * @param amount Amount of tokens to burn
     */
    function burn(address from, uint256 amount) external {
        _burn(from, amount);
    }
}
