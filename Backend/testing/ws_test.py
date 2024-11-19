import asyncio
import json
import logging
from datetime import datetime
from typing import Any, Dict, List, Optional

import pytest
import requests
import websockets
from auth_test import AuthTester

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)


class WebSocketClient:
    def __init__(self, url: str, token: str, username: str):
        self.url = url
        self.token = token
        self.username = username
        self.websocket: Optional[websockets.WebSocketClientProtocol] = None
        self.received_messages: List[Dict[str, Any]] = []
        self.connected = False
        self.should_run = True

    async def connect(self) -> bool:
        """Establish WebSocket connection"""
        try:
            self.websocket = await websockets.connect(
                self.url,
                subprotocols=[f"access_token|{self.token}"]
            )
            self.connected = True
            return True
        except Exception as e:
            logger.error(f"WebSocket connection failed for {self.username}: {str(e)}")
            return False

    async def disconnect(self):
        """Close WebSocket connection"""
        if self.websocket:
            self.should_run = False
            await self.websocket.close()
            self.connected = False

    async def send_message(self, content: str) -> bool:
        """Send a message through WebSocket"""
        try:
            if self.websocket and self.connected:
                await self.websocket.send(content)
                return True
            return False
        except Exception as e:
            logger.error(f"Failed to send message for {self.username}: {str(e)}")
            return False

    async def listen(self):
        """Listen for incoming messages"""
        try:
            while self.should_run and self.websocket:
                message = await self.websocket.recv()
                parsed_message = json.loads(message)
                self.received_messages.append(parsed_message)
                logger.info(f"Received message for {self.username}: {parsed_message}")
        except websockets.exceptions.ConnectionClosed:
            logger.info(f"WebSocket connection closed for {self.username}")
            self.connected = False
        except Exception as e:
            logger.error(f"Error in WebSocket listener for {self.username}: {str(e)}")
            self.connected = False


class WebSocketTester(AuthTester):
    def __init__(self, base_url: str = "http://localhost:6262"):
        super().__init__(base_url)
        self.ws_clients: Dict[str, WebSocketClient] = {}

    async def setup_ws_client(self, username: str) -> WebSocketClient:
        """Create and connect a WebSocket client"""
        if username not in self.auth_tokens:
            raise ValueError(f"User {username} not authenticated")

        client = WebSocketClient(self.ws_url, self.auth_tokens[username], username)
        self.ws_clients[username] = client
        connected = await client.connect()
        if connected:
            asyncio.create_task(client.listen())
        return client

    async def cleanup_ws_clients(self):
        """Disconnect all WebSocket clients"""
        for client in self.ws_clients.values():
            await client.disconnect()
        self.ws_clients.clear()


@pytest.fixture
def tester():
    return WebSocketTester()


@pytest.mark.asyncio
async def test_websocket_connection(tester: WebSocketTester):
    """Test WebSocket connection establishment"""
    username = f"ws_test_user_{datetime.now().timestamp()}"
    password = "TestPass123!"

    # Register and login user
    tester.register_user(username, password)

    # Setup WebSocket client
    client = await tester.setup_ws_client(username)
    assert client.connected is True

    # Cleanup
    await client.disconnect()
    assert client.connected is False


@pytest.mark.asyncio
async def test_message_broadcast(tester: WebSocketTester):
    """Test message broadcasting between multiple clients"""
    # Create two test users
    users = [
        (f"ws_test_user1_{datetime.now().timestamp()}", "TestPass123!"),
        (f"ws_test_user2_{datetime.now().timestamp()}", "TestPass123!")
    ]

    # Register and connect users
    for username, password in users:
        tester.register_user(username, password)
        await tester.setup_ws_client(username)

    # Send test message from first user
    test_message = "Hello, World!"
    client1 = tester.ws_clients[users[0][0]]
    await client1.send_message(test_message)

    # Wait for message propagation
    await asyncio.sleep(1)

    # Verify message received by second user
    client2 = tester.ws_clients[users[1][0]]
    received_messages = [
        msg for msg in client2.received_messages
        if msg["type"] == "chat" and msg["payload"]["content"] == test_message
    ]
    assert len(received_messages) > 0

    # Cleanup
    await tester.cleanup_ws_clients()


@pytest.mark.asyncio
async def test_user_presence(tester: WebSocketTester):
    """Test user presence notifications"""
    username = f"ws_test_user_{datetime.now().timestamp()}"
    password = "TestPass123!"

    # Register first user
    tester.register_user(username, password)
    client = await tester.setup_ws_client(username)

    # Wait for connection messages
    await asyncio.sleep(1)

    # Verify join message
    join_messages = [
        msg for msg in client.received_messages
        if msg["type"] == "join" and msg["payload"]["username"] == username
    ]
    assert len(join_messages) > 0

    # Verify users list message
    users_messages = [
        msg for msg in client.received_messages
        if msg["type"] == "users"
    ]
    assert len(users_messages) > 0
    assert username in users_messages[-1]["payload"]

    # Cleanup
    await tester.cleanup_ws_clients()


@pytest.mark.asyncio
async def test_message_persistence(tester: WebSocketTester):
    """Test message persistence and history retrieval"""
    username = f"ws_test_user_{datetime.now().timestamp()}"
    password = "TestPass123!"

    # Register and connect user
    tester.register_user(username, password)
    headers = {"Authorization": f"Bearer {tester.auth_tokens[username]}"}

    # Send test message
    client = await tester.setup_ws_client(username)
    test_message = "Test persistence message"
    await client.send_message(test_message)

    # Wait for message to be stored
    await asyncio.sleep(1)

    # Retrieve recent messages
    response = requests.get(f"{tester.base_url}/api/messages/recent", headers=headers)
    assert response.status_code == 200

    messages = response.json()
    assert len(messages) > 0
    assert any(msg["content"] == test_message for msg in messages)

    # Cleanup
    await tester.cleanup_ws_clients()

if __name__ == "__main__":
    pytest.main([__file__, "-v"])
