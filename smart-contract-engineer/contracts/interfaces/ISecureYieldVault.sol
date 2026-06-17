// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

/**
 * @title ISecureYieldVault
 * @notice Interface for a secure ERC-4626 inspired yield vault.
 * @dev Defines the core interface for deposit, withdraw, and share management.
 */
interface ISecureYieldVault {
    /// @notice Emitted when a user deposits assets
    event Deposited(address indexed sender, address indexed owner, uint256 assets, uint256 shares);

    /// @notice Emitted when a user withdraws assets
    event Withdrawn(
        address indexed sender,
        address indexed receiver,
        address indexed owner,
        uint256 assets,
        uint256 shares
    );

    /// @notice Emitted when reward is added to the vault
    event RewardAdded(uint256 amount);

    /// @notice Emitted when emergency withdrawal occurs
    event EmergencyWithdrawn(address indexed sender, uint256 assets, uint256 shares);

    /// @notice Emitted when fee is updated
    event FeeUpdated(uint256 oldFee, uint256 newFee);

    /**
     * @notice Deposit assets and mint shares to the vault.
     * @param assets Amount of assets to deposit
     * @param receiver Address to receive the shares
     * @return shares Amount of shares minted
     */
    function deposit(uint256 assets, address receiver) external returns (uint256 shares);

    /**
     * @notice Mint shares by depositing assets.
     * @param shares Amount of shares to mint
     * @param receiver Address to receive the shares
     * @return assets Amount of assets deposited
     */
    function mint(uint256 shares, address receiver) external returns (uint256 assets);

    /**
     * @notice Withdraw assets by burning shares.
     * @param assets Amount of assets to withdraw
     * @param receiver Address to receive the assets
     * @param owner Owner of the shares
     * @return shares Amount of shares burned
     */
    function withdraw(
        uint256 assets,
        address receiver,
        address owner
    ) external returns (uint256 shares);

    /**
     * @notice Redeem shares for assets.
     * @param shares Amount of shares to redeem
     * @param receiver Address to receive the assets
     * @param owner Owner of the shares
     * @return assets Amount of assets returned
     */
    function redeem(
        uint256 shares,
        address receiver,
        address owner
    ) external returns (uint256 assets);

    /**
     * @notice Convert assets to shares (floor rounding).
     * @param assets Amount of assets
     * @return shares Equivalent shares
     */
    function convertToShares(uint256 assets) external view returns (uint256 shares);

    /**
     * @notice Convert shares to assets (floor rounding).
     * @param shares Amount of shares
     * @return assets Equivalent assets
     */
    function convertToAssets(uint256 shares) external view returns (uint256 assets);

    /**
     * @notice Total deposited assets (principal only, excludes rewards).
     * @return Total assets
     */
    function totalAssets() external view returns (uint256);

    /**
     * @notice Total assets including accumulated rewards.
     * @return Total assets under management
     */
    function totalAssetsWithRewards() external view returns (uint256);

    /**
     * @notice Get the balance of assets deposited by a user.
     * @param owner Address of the account
     * @return User's asset balance
     */
    function balanceOfAssets(address owner) external view returns (uint256);

    /**
     * @notice Get the total shares for a user.
     * @param owner Address of the account
     * @return User's share balance
     */
    function balanceOf(address owner) external view returns (uint256);

    /**
     * @notice Total supply of shares.
     * @return Total share supply
     */
    function totalSupply() external view returns (uint256);

    /**
     * @notice Emergency withdraw - returns principal only, requires vault paused.
     * @param assets Amount of principal assets to withdraw
     * @param receiver Address to receive the assets
     */
    function emergencyWithdraw(uint256 assets, address receiver) external;

    /**
     * @notice Inject rewards into the vault.
     * @param amount Amount of reward tokens to add
     */
    function injectRewards(uint256 amount) external;

    /**
     * @notice Update management fee.
     * @param newFee New fee in basis points
     */
    function setManagementFee(uint256 newFee) external;

    /**
     * @notice Pause the vault.
     */
    function pause() external;

    /**
     * @notice Unpause the vault.
     */
    function unpause() external;

    /// @notice The underlying asset token
    function asset() external view returns (address);

    /// @notice Accumulated rewards
    function accumulatedRewards() external view returns (uint256);

    /// @notice Management fee
    function managementFee() external view returns (uint256);
}
