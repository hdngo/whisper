import asyncio
import logging
import statistics
import time
from concurrent.futures import ThreadPoolExecutor
from dataclasses import dataclass
from datetime import datetime
from typing import Any, Dict, List

import numpy as np
import pytest
from ws_test import WebSocketClient, WebSocketTester

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)


@dataclass
class PerformanceMetrics:
    connection_times: List[float]
    message_latencies: List[float]
    success_rate: float
    failed_operations: int
    total_operations: int

    def get_summary(self) -> Dict[str, Any]:
        """Generate a summary of performance metrics"""
        return {
            "connection_time_avg": round(statistics.mean(self.connection_times), 3),
            "connection_time_p95": round(np.percentile(self.connection_times, 95), 3),
            "message_latency_avg": round(statistics.mean(self.message_latencies), 3),
            "message_latency_p95": round(np.percentile(self.message_latencies, 95), 3),
            "success_rate": round(self.success_rate * 100, 2),
            "failed_operations": self.failed_operations,
            "total_operations": self.total_operations
        }


class LoadTester(WebSocketTester):
    def __init__(self, base_url: str = "http://localhost:6262"):
        super().__init__(base_url)
        self.metrics = PerformanceMetrics([], [], 0.0, 0, 0)
        self.executor = ThreadPoolExecutor(max_workers=50)

    async def create_test_user(self, index: int) -> tuple[str, str]:
        """Create a test user with unique credentials"""
        username = f"load_test_user_{index}_{datetime.now().timestamp()}"
        password = f"LoadTest123!_{index}"
        try:
            self.register_user(username, password)
            return username, password
        except Exception as e:
            logger.error(f"Failed to create test user {username}: {e}")
            raise

    async def measure_connection_time(self, username: str) -> float:
        """Measure time taken to establish WebSocket connection"""
        start_time = time.time()
        client = await self.setup_ws_client(username)
        connection_time = time.time() - start_time

        if client.connected:
            self.metrics.connection_times.append(connection_time)
            return connection_time
        else:
            self.metrics.failed_operations += 1
            raise Exception(f"Failed to connect for user {username}")

    async def measure_message_latency(self, sender: WebSocketClient, receiver: WebSocketClient, message: str) -> float:
        """Measure message delivery latency between two clients"""
        start_time = time.time()
        message_id = f"test_msg_{time.time()}"
        test_message = f"{message_id}:{message}"

        # Send message
        if not await sender.send_message(test_message):
            self.metrics.failed_operations += 1
            raise Exception(f"Failed to send message from {sender.username}")

        # Wait for message receipt
        max_wait = 5  # seconds
        start_wait = time.time()
        while time.time() - start_wait < max_wait:
            for msg in receiver.received_messages:
                if msg["type"] == "chat" and message_id in msg["payload"]["content"]:
                    latency = time.time() - start_time
                    self.metrics.message_latencies.append(latency)
                    return latency
            await asyncio.sleep(0.1)

        self.metrics.failed_operations += 1
        raise Exception("Message not received within timeout")

    async def run_concurrent_load_test(self, num_users: int, messages_per_user: int):
        """Run a concurrent load test with specified number of users and messages"""
        logger.info(f"Starting load test with {num_users} users, {messages_per_user} messages per user")

        # Create test users
        users = []
        for i in range(num_users):
            try:
                username, _ = await self.create_test_user(i)
                users.append(username)
            except Exception as e:
                logger.error(f"Failed to create user {i}: {e}")
                continue

        # Connect all users
        connection_tasks = [self.measure_connection_time(username) for username in users]
        await asyncio.gather(*connection_tasks, return_exceptions=True)

        # Send messages between users
        message_tasks = []
        for i in range(messages_per_user):
            for sender_username in users:
                if sender_username not in self.ws_clients:
                    continue
                sender = self.ws_clients[sender_username]

                # Send message to random receiver
                receiver_username = np.random.choice([u for u in users if u != sender_username])
                if receiver_username not in self.ws_clients:
                    continue
                receiver = self.ws_clients[receiver_username]

                task = self.measure_message_latency(
                    sender, receiver, f"Load test message {i}"
                )
                message_tasks.append(task)

        await asyncio.gather(*message_tasks, return_exceptions=True)

        # Calculate success rate
        total_ops = len(connection_tasks) + len(message_tasks)
        self.metrics.total_operations = total_ops
        self.metrics.success_rate = (total_ops - self.metrics.failed_operations) / total_ops

        # Cleanup
        await self.cleanup_ws_clients()
        return self.metrics.get_summary()


@pytest.fixture
def tester():
    return LoadTester()


@pytest.mark.asyncio
async def test_concurrent_connections(tester: LoadTester):
    """Test multiple concurrent WebSocket connections"""
    num_users = 50
    users = []

    # Create and connect multiple users
    for i in range(num_users):
        username = f"concurrent_user_{i}_{datetime.now().timestamp()}"
        password = f"TestPass123!_{i}"

        try:
            tester.register_user(username, password)
            client = await tester.setup_ws_client(username)
            users.append(username)
            assert client.connected is True
        except Exception as e:
            logger.error(f"Failed to setup user {username}: {e}")

    # Verify all users are connected
    connected_users = len([c for c in tester.ws_clients.values() if c.connected])
    assert connected_users >= num_users * 0.9  # Allow 10% failure rate

    # Cleanup
    await tester.cleanup_ws_clients()


@pytest.mark.asyncio
async def test_message_broadcast_stress(tester: LoadTester):
    """Test message broadcasting under load"""
    num_users = 20
    messages_per_user = 5

    metrics = await LoadTester().run_concurrent_load_test(num_users, messages_per_user)

    # Assert performance requirements
    assert metrics["success_rate"] >= 95.0  # 95% success rate
    assert metrics["message_latency_p95"] <= 1.0  # P95 latency under 1 second
    assert metrics["connection_time_p95"] <= 2.0  # P95 connection time under 2 seconds


@pytest.mark.asyncio
async def test_connection_handling_stress(tester: LoadTester):
    """Test rapid connect/disconnect cycles"""
    username = f"stress_test_user_{datetime.now().timestamp()}"
    password = "TestPass123!"

    # Register user
    tester.register_user(username, password)

    # Perform rapid connect/disconnect cycles
    cycles = 10
    for _ in range(cycles):
        # Connect
        client = await tester.setup_ws_client(username)
        assert client.connected is True

        # Send a message
        await client.send_message("Stress test message")

        # Disconnect
        await client.disconnect()
        assert client.connected is False

        # Small delay to prevent overwhelming the server
        await asyncio.sleep(0.1)


async def run_load_test_scenario():
    """Run a complete load test scenario"""
    tester = LoadTester()

    # Test scenarios
    scenarios = [
        (10, 20),   # Light load: 10 users, 20 messages each
        (25, 40),   # Medium load: 25 users, 40 messages each
        (50, 60),   # Heavy load: 50 users, 60 messages each
    ]

    results = []
    for num_users, messages_per_user in scenarios:
        logger.info(f"\nRunning load test with {num_users} users and {messages_per_user} messages per user")
        metrics = await tester.run_concurrent_load_test(num_users, messages_per_user)
        results.append({
            "scenario": f"{num_users} users x {messages_per_user} messages",
            "metrics": metrics
        })

        # Allow system to stabilize between scenarios
        await asyncio.sleep(5)

    # Print results
    logger.info("\nLoad Test Results:")
    for result in results:
        logger.info(f"\nScenario: {result['scenario']}")
        for metric, value in result['metrics'].items():
            logger.info(f"{metric}: {value}")

if __name__ == "__main__":
    asyncio.run(run_load_test_scenario())
