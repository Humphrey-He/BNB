// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import {ERC20} from "@openzeppelin/contracts/token/ERC20/ERC20.sol";

/**
 * @title FeeOnTransferToken
 * @notice An ERC-20 token that deducts a fee on every transfer.
 * @dev Used to test vault handling of fee-on-transfer tokens.
 */
contract FeeOnTransferToken is ERC20 {
    uint256 public constant FEE_DENOMINATOR = 10000;
    uint256 public feePercentage; // basis points (e.g., 100 = 1%)

    /**
     * @param name_ Token name
     * @param symbol_ Token symbol
     * @param feePercentage_ Fee in basis points (100 = 1%)
     */
    constructor(
        string memory name_,
        string memory symbol_,
        uint256 feePercentage_
    ) ERC20(name_, symbol_) {
        require(feePercentage_ <= FEE_DENOMINATOR, "Fee too high");
        feePercentage = feePercentage_;
    }

    /**
     * @notice Transfers tokens with a fee deducted.
     * @param from Sender address
     * @param recipient Recipient address
     * @param amount Amount to transfer
     */
    function transferFrom(
        address from,
        address recipient,
        uint256 amount
    ) public override returns (bool) {
        uint256 fee = (amount * feePercentage) / FEE_DENOMINATOR;
        uint256 amountAfterFee = amount - fee;

        // Call parent implementation for actual transfer
        super.transferFrom(from, recipient, amountAfterFee);

        // Burn the fee portion using _burn
        if (fee > 0) {
            _burn(from, fee);
        }

        return true;
    }

    /**
     * @notice Transfers tokens with fee.
     * @param recipient Recipient address
     * @param amount Amount to transfer
     */
    function transfer(address recipient, uint256 amount) public override returns (bool) {
        uint256 fee = (amount * feePercentage) / FEE_DENOMINATOR;
        uint256 amountAfterFee = amount - fee;

        super.transfer(recipient, amountAfterFee);
        if (fee > 0) {
            _burn(msg.sender, fee);
        }

        return true;
    }

    /**
     * @notice Mint tokens for testing
     * @param to Address to mint to
     * @param amount Amount to mint
     */
    function mint(address to, uint256 amount) external {
        _mint(to, amount);
    }
}
