// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import {AccessControl} from "@openzeppelin/contracts/access/AccessControl.sol";
import {Pausable} from "@openzeppelin/contracts/utils/Pausable.sol";
import {ReentrancyGuard} from "@openzeppelin/contracts/utils/ReentrancyGuard.sol";
import {IERC20} from "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import {SafeERC20} from "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";

/**
 * @title RewardDistributor
 * @notice Manages reward distribution to vaults with timelock protection.
 * @dev Separates reward management from vault logic for security.
 */
contract RewardDistributor is AccessControl, Pausable, ReentrancyGuard {
    using SafeERC20 for IERC20;

    /// @notice Role for operators who can add rewards
    bytes32 public constant OPERATOR_ROLE = keccak256("OPERATOR_ROLE");

    /// @notice Role for timelock admin (can propose and execute)
    bytes32 public constant TIMELOCK_ADMIN_ROLE = keccak256("TIMELOCK_ADMIN_ROLE");

    /// @notice Minimum delay for timelock operations (1 hour)
    uint256 public constant MIN_DELAY = 1 hours;

    /// @notice Maximum delay for timelock operations (30 days)
    uint256 public constant MAX_DELAY = 30 days;

    /// @notice Grace period after timelock expires
    uint256 public constant GRACE_PERIOD = 7 days;

    /// @notice Emitted when reward rate is updated
    event RewardRateUpdated(uint256 oldRate, uint256 newRate);

    /// @notice Emitted when rewards are distributed
    event RewardsDistributed(address indexed vault, uint256 amount);

    /// @notice Emitted when timelock delay is updated
    event TimelockDelayUpdated(uint256 oldDelay, uint256 newDelay);

    /// @notice Emitted when pending timelock operation is created
    event TimelockPending(bytes32 indexed txHash, uint256 eta);

    /// @notice Emitted when timelock operation is executed
    event TimelockExecuted(bytes32 indexed txHash, uint256 eta);

    /// @notice Emitted when timelock operation is cancelled
    event TimelockCancelled(bytes32 indexed txHash);

    /// @notice The reward token
    IERC20 public immutable rewardToken;

    /// @notice Reward rate per second (for continuous distribution)
    uint256 public rewardRate;

    /// @notice Last update timestamp
    uint256 public lastUpdateTime;

    /// @notice Total rewards distributed
    uint256 public totalRewardsDistributed;

    /// @notice Mapping of pending timelock transactions (txHash -> eta)
    mapping(bytes32 => uint256) public pendingTimelockTxs;

    /// @notice Current timelock delay
    uint256 public timelockDelay;

    /**
     * @notice Constructor
     * @param rewardToken_ The reward ERC-20 token
     * @param initialAdmin Initial admin address
     */
    constructor(IERC20 rewardToken_, address initialAdmin) {
        require(address(rewardToken_) != address(0), "Zero reward token address");
        require(initialAdmin != address(0), "Zero admin address");

        rewardToken = rewardToken_;
        timelockDelay = MIN_DELAY;

        _grantRole(DEFAULT_ADMIN_ROLE, initialAdmin);
        _grantRole(TIMELOCK_ADMIN_ROLE, initialAdmin);
        _grantRole(OPERATOR_ROLE, initialAdmin);
    }

    /**
     * @dev Compute timelock transaction hash (consistent across schedule/execute/cancel)
     */
    function _timelockTxHash(
        address target,
        bytes calldata data,
        uint256 value
    ) internal pure returns (bytes32) {
        return keccak256(abi.encode(target, data, value));
    }

    /**
     * @notice Update the reward rate
     * @param newRate New reward rate per second
     */
    function setRewardRate(uint256 newRate) external onlyRole(DEFAULT_ADMIN_ROLE) {
        require(newRate <= 1e36, "Rate too high");

        uint256 oldRate = rewardRate;
        rewardRate = newRate;
        lastUpdateTime = block.timestamp;

        emit RewardRateUpdated(oldRate, newRate);
    }

    /**
     * @notice Add rewards to the pool
     * @param amount Amount of reward tokens to add
     */
    function addRewardPool(uint256 amount) external nonReentrant onlyRole(OPERATOR_ROLE) {
        require(amount > 0, "Zero amount");

        totalRewardsDistributed += amount;

        emit RewardsDistributed(address(this), amount);

        rewardToken.safeTransferFrom(msg.sender, address(this), amount);
    }

    /**
     * @notice Withdraw excess rewards (only via timelock)
     * @param to Recipient address
     * @param amount Amount to withdraw
     */
    function withdrawExcessReward(address to, uint256 amount) external nonReentrant onlyRole(TIMELOCK_ADMIN_ROLE) {
        require(to != address(0), "Zero address");
        require(amount > 0, "Zero amount");

        uint256 balance = rewardToken.balanceOf(address(this));
        require(balance >= amount, "Insufficient balance");

        rewardToken.safeTransfer(to, amount);
    }

    /**
     * @notice Update timelock delay
     * @param newDelay New delay in seconds
     */
    function setTimelockDelay(uint256 newDelay) external onlyRole(TIMELOCK_ADMIN_ROLE) {
        require(newDelay >= MIN_DELAY, "Delay too short");
        require(newDelay <= MAX_DELAY, "Delay too long");

        uint256 oldDelay = timelockDelay;
        timelockDelay = newDelay;

        emit TimelockDelayUpdated(oldDelay, newDelay);
    }

    /**
     * @notice Schedule a timelock transaction
     * @param target Target contract
     * @param data Calldata
     * @param value ETH value
     * @return txHash Transaction hash
     */
    function scheduleTimelock(
        address target,
        bytes calldata data,
        uint256 value
    ) external onlyRole(TIMELOCK_ADMIN_ROLE) returns (bytes32 txHash) {
        require(target != address(0), "Zero target");

        uint256 eta = block.timestamp + timelockDelay;
        txHash = _timelockTxHash(target, data, value);

        pendingTimelockTxs[txHash] = eta;

        emit TimelockPending(txHash, eta);
    }

    /**
     * @notice Execute a timelock transaction
     * @param target Target contract
     * @param data Calldata
     * @param value ETH value
     */
    function executeTimelock(
        address target,
        bytes calldata data,
        uint256 value
    ) external onlyRole(TIMELOCK_ADMIN_ROLE) nonReentrant {
        bytes32 txHash = _timelockTxHash(target, data, value);
        uint256 eta = pendingTimelockTxs[txHash];

        require(eta != 0, "Not scheduled");
        require(block.timestamp >= eta, "Too early");
        require(block.timestamp <= eta + GRACE_PERIOD, "Too late");

        delete pendingTimelockTxs[txHash];

        emit TimelockExecuted(txHash, eta);

        (bool success, ) = target.call{value: value}(data);
        require(success, "Call failed");
    }

    /**
     * @notice Cancel a pending timelock transaction
     * @param target Target contract
     * @param data Calldata
     * @param value ETH value
     */
    function cancelTimelock(
        address target,
        bytes calldata data,
        uint256 value
    ) external onlyRole(DEFAULT_ADMIN_ROLE) {
        bytes32 txHash = _timelockTxHash(target, data, value);

        require(pendingTimelockTxs[txHash] != 0, "Not pending");

        delete pendingTimelockTxs[txHash];

        emit TimelockCancelled(txHash);
    }

    /**
     * @notice Pause the distributor
     */
    function pause() external onlyRole(DEFAULT_ADMIN_ROLE) {
        _pause();
    }

    /**
     * @notice Unpause the distributor
     */
    function unpause() external onlyRole(DEFAULT_ADMIN_ROLE) {
        _unpause();
    }

    /**
     * @notice Check if a timelock transaction is pending
     * @param target Target address
     * @param data Calldata
     * @param value ETH value
     * @return eta Execution time, or 0 if not pending
     */
    function getPendingTx(
        address target,
        bytes calldata data,
        uint256 value
    ) external view returns (uint256 eta) {
        bytes32 txHash = _timelockTxHash(target, data, value);
        eta = pendingTimelockTxs[txHash];
    }
}
