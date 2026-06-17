// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import {ERC20} from "@openzeppelin/contracts/token/ERC20/ERC20.sol";
import {ERC20Permit} from "@openzeppelin/contracts/token/ERC20/extensions/ERC20Permit.sol";
import {SafeERC20} from "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";
import {AccessControl} from "@openzeppelin/contracts/access/AccessControl.sol";
import {Pausable} from "@openzeppelin/contracts/utils/Pausable.sol";
import {ReentrancyGuard} from "@openzeppelin/contracts/utils/ReentrancyGuard.sol";
import {IERC20} from "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import {Math} from "@openzeppelin/contracts/utils/math/Math.sol";

/**
 * @title SecureYieldVault
 * @notice A secure ERC-4626 inspired yield vault with share/asset model.
 * @dev Implements deposit, withdraw, mint, redeem with proper rounding.
 *      Integrates AccessControl, Pausable, ReentrancyGuard for security.
 *      Uses OpenZeppelin Math.mulDiv for safe arithmetic.
 */
contract SecureYieldVault is ERC20, ERC20Permit, AccessControl, Pausable, ReentrancyGuard {
    using SafeERC20 for IERC20;

    /// @notice Role for vault operators who can inject rewards
    bytes32 public constant OPERATOR_ROLE = keccak256("OPERATOR_ROLE");

    /// @notice Role for emergency administrators
    bytes32 public constant ADMIN_ROLE = keccak256("ADMIN_ROLE");

    /// @notice Emitted when user deposits assets
    event Deposited(address indexed sender, address indexed owner, uint256 assets, uint256 shares);

    /// @notice Emitted when user withdraws assets
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

    /// @notice Emitted when management fee is updated
    event FeeUpdated(uint256 oldFee, uint256 newFee);

    /// @notice The underlying asset token
    IERC20 public immutable asset;

    /// @notice Total deposited assets (principal) - does NOT include accumulated rewards
    uint256 public totalAssets;

    /// @notice Accumulated rewards that have been injected
    uint256 public accumulatedRewards;

    /// @notice Management fee in basis points (e.g., 100 = 1%)
    uint256 public managementFee;

    /// @notice Maximum fee allowed (1000 = 10%)
    uint256 public constant MAX_FEE = 1000;

    /**
     * @notice Constructor
     * @param asset_ The underlying ERC-20 asset
     * @param name_ Vault share token name
     * @param symbol_ Vault share token symbol
     */
    constructor(
        IERC20 asset_,
        string memory name_,
        string memory symbol_
    ) ERC20(name_, symbol_) ERC20Permit(name_) {
        require(address(asset_) != address(0), "Zero asset address");

        asset = asset_;

        _grantRole(DEFAULT_ADMIN_ROLE, msg.sender);
        _grantRole(ADMIN_ROLE, msg.sender);
    }

    modifier nonZeroAmount(uint256 amount) {
        require(amount > 0, "Zero amount");
        _;
    }

    /**
     * @notice Deposit assets and mint shares to the vault
     * @param assets Amount of assets to deposit
     * @param receiver Address to receive the shares
     * @return shares Amount of shares minted
     */
    function deposit(uint256 assets, address receiver)
        external
        nonReentrant
        whenNotPaused
        nonZeroAmount(assets)
        returns (uint256 shares)
    {
        // Measure actual received amount to handle fee-on-transfer tokens
        uint256 balanceBefore = asset.balanceOf(address(this));
        asset.safeTransferFrom(msg.sender, address(this), assets);
        uint256 received = asset.balanceOf(address(this)) - balanceBefore;

        require(received >= assets, "Fee deducted");

        shares = _convertToShares(received, Math.Rounding.Floor);
        require(shares > 0, "Zero shares");

        // CEI pattern: Checks -> Effects -> Interactions
        totalAssets += received;

        emit Deposited(msg.sender, receiver, received, shares);

        _mint(receiver, shares);
    }

    /**
     * @notice Mint exact shares by depositing assets
     * @param shares Amount of shares to mint
     * @param receiver Address to receive the shares
     * @return assets Amount of assets deposited
     */
    function mint(uint256 shares, address receiver)
        external
        nonReentrant
        whenNotPaused
        nonZeroAmount(shares)
        returns (uint256 assets)
    {
        assets = _convertToAssets(shares, Math.Rounding.Ceil);
        require(assets > 0, "Zero assets");

        uint256 balanceBefore = asset.balanceOf(address(this));
        asset.safeTransferFrom(msg.sender, address(this), assets);
        uint256 received = asset.balanceOf(address(this)) - balanceBefore;

        require(received >= assets, "Fee deducted");

        totalAssets += received;

        emit Deposited(msg.sender, receiver, received, shares);

        _mint(receiver, shares);
    }

    /**
     * @notice Withdraw assets by burning shares
     * @param assets Amount of assets to withdraw
     * @param receiver Address to receive the assets
     * @param owner Owner of the shares
     * @return shares Amount of shares burned
     */
    function withdraw(
        uint256 assets,
        address receiver,
        address owner
    ) external nonReentrant whenNotPaused nonZeroAmount(assets) returns (uint256 shares) {
        shares = _convertToShares(assets, Math.Rounding.Ceil);

        require(shares > 0, "Zero shares");
        require(balanceOf(owner) >= shares, "Insufficient balance");
        require(owner != address(0), "Invalid owner");

        // Calculate proportion of principal vs reward being withdrawn
        uint256 principalToWithdraw = (assets * totalAssets) / (totalAssets + accumulatedRewards);
        uint256 rewardToWithdraw = assets - principalToWithdraw;

        // CEI pattern
        totalAssets -= principalToWithdraw;
        accumulatedRewards -= rewardToWithdraw;

        if (msg.sender != owner) {
            _spendAllowance(owner, msg.sender, shares);
        }

        emit Withdrawn(msg.sender, receiver, owner, assets, shares);

        _burn(owner, shares);
        asset.safeTransfer(receiver, assets);
    }

    /**
     * @notice Redeem shares for assets
     * @param shares Amount of shares to redeem
     * @param receiver Address to receive the assets
     * @param owner Owner of the shares
     * @return assets Amount of assets returned
     */
    function redeem(
        uint256 shares,
        address receiver,
        address owner
    ) external nonReentrant whenNotPaused nonZeroAmount(shares) returns (uint256 assets) {
        require(balanceOf(owner) >= shares, "Insufficient balance");
        require(owner != address(0), "Invalid owner");

        assets = _convertToAssets(shares, Math.Rounding.Floor);
        require(assets > 0, "Zero assets");

        // Calculate proportion of principal vs reward being redeemed
        uint256 principalToRedeem = (assets * totalAssets) / (totalAssets + accumulatedRewards);
        uint256 rewardToRedeem = assets - principalToRedeem;

        totalAssets -= principalToRedeem;
        accumulatedRewards -= rewardToRedeem;

        if (msg.sender != owner) {
            _spendAllowance(owner, msg.sender, shares);
        }

        emit Withdrawn(msg.sender, receiver, owner, assets, shares);

        _burn(owner, shares);
        asset.safeTransfer(receiver, assets);
    }

    /**
     * @notice Emergency withdraw - burns shares proportionally and returns principal only
     * @param assets Amount of principal assets to withdraw
     * @param receiver Address to receive the assets
     */
    function emergencyWithdraw(uint256 assets, address receiver)
        external
        nonReentrant
        nonZeroAmount(assets)
        onlyRole(ADMIN_ROLE)
        whenPaused
    {
        uint256 shares = _convertToShares(assets, Math.Rounding.Floor);
        require(shares > 0, "Zero shares");
        require(balanceOf(msg.sender) >= shares, "Insufficient balance");

        // Withdraw only principal, no rewards
        totalAssets -= assets;

        emit EmergencyWithdrawn(msg.sender, assets, shares);

        _burn(msg.sender, shares);
        asset.safeTransfer(receiver, assets);
    }

    /**
     * @notice Inject rewards into the vault (called by operator)
     * @param amount Amount of reward tokens to add
     */
    function injectRewards(uint256 amount) external nonReentrant onlyRole(OPERATOR_ROLE) nonZeroAmount(amount) {
        uint256 balanceBefore = asset.balanceOf(address(this));
        asset.safeTransferFrom(msg.sender, address(this), amount);
        uint256 received = asset.balanceOf(address(this)) - balanceBefore;

        accumulatedRewards += received;

        emit RewardAdded(received);
    }

    /**
     * @notice Update management fee
     * @param newFee New fee in basis points
     */
    function setManagementFee(uint256 newFee) external onlyRole(ADMIN_ROLE) {
        require(newFee <= MAX_FEE, "Fee too high");

        uint256 oldFee = managementFee;
        managementFee = newFee;

        emit FeeUpdated(oldFee, newFee);
    }

    /**
     * @notice Pause the vault
     */
    function pause() external onlyRole(ADMIN_ROLE) {
        _pause();
    }

    /**
     * @notice Unpause the vault
     */
    function unpause() external onlyRole(ADMIN_ROLE) {
        _unpause();
    }

    /**
     * @notice Convert assets to shares
     * @param assets Amount of assets
     * @return shares Equivalent shares
     */
    function convertToShares(uint256 assets) external view returns (uint256) {
        return _convertToShares(assets, Math.Rounding.Floor);
    }

    /**
     * @notice Convert shares to assets
     * @param shares Amount of shares
     * @return assets Equivalent assets
     */
    function convertToAssets(uint256 shares) external view returns (uint256) {
        return _convertToAssets(shares, Math.Rounding.Floor);
    }

    /**
     * @notice Get total assets including accumulated rewards
     * @return Total assets under management
     */
    function totalAssetsWithRewards() external view returns (uint256) {
        return totalAssets + accumulatedRewards;
    }

    /**
     * @notice Get the balance of assets deposited by a user
     * @param owner Address of the account
     * @return User's asset balance
     */
    function balanceOfAssets(address owner) external view returns (uint256) {
        return _convertToAssets(balanceOf(owner), Math.Rounding.Floor);
    }

    /**
     * @dev Internal conversion using OZ Math.mulDiv for overflow protection
     */
    function _convertToShares(uint256 assets, Math.Rounding rounding) internal view returns (uint256) {
        uint256 supply = totalSupply();

        if (supply == 0) {
            return assets;
        }

        uint256 totalAssetsWithRewards_ = totalAssets + accumulatedRewards;

        if (totalAssetsWithRewards_ == 0) {
            return assets;
        }

        return Math.mulDiv(assets, supply, totalAssetsWithRewards_, rounding);
    }

    /**
     * @dev Internal conversion using OZ Math.mulDiv for overflow protection
     */
    function _convertToAssets(uint256 shares, Math.Rounding rounding) internal view returns (uint256) {
        uint256 supply = totalSupply();

        if (supply == 0) {
            return shares;
        }

        uint256 totalAssetsWithRewards_ = totalAssets + accumulatedRewards;

        return Math.mulDiv(shares, totalAssetsWithRewards_, supply, rounding);
    }
}
