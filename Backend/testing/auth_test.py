import logging
from datetime import datetime
from typing import Any, Dict

import pytest
import requests

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)


class AuthTester:
    def __init__(self, base_url: str = None):
        self.base_url = base_url or "http://localhost:6262"
        self.ws_url = f"ws://localhost:6262/api/ws"
        self.auth_tokens: Dict[str, str] = {}
        self.session = requests.Session()

    def register_user(self, username: str, password: str) -> Dict[str, Any]:
        """Register a new user and return the response"""
        try:
            response = self.session.post(
                f"{self.base_url}/api/auth/register",
                json={"username": username, "password": password}
            )
            response.raise_for_status()
            data = response.json()
            self.auth_tokens[username] = data["token"]
            return data
        except requests.exceptions.RequestException as e:
            logger.error(f"Registration failed for user {username}: {str(e)}")
            raise

    def login_user(self, username: str, password: str) -> Dict[str, Any]:
        """Login a user and return the response"""
        try:
            response = self.session.post(
                f"{self.base_url}/api/auth/login",
                json={"username": username, "password": password}
            )
            response.raise_for_status()
            data = response.json()
            self.auth_tokens[username] = data["token"]
            return data
        except requests.exceptions.RequestException as e:
            logger.error(f"Login failed for user {username}: {str(e)}")
            raise

    def logout_user(self, username: str) -> bool:
        """Logout a user and return success status"""
        try:
            headers = {"Authorization": f"Bearer {self.auth_tokens[username]}"}
            response = self.session.post(
                f"{self.base_url}/api/auth/logout",
                headers=headers
            )
            response.raise_for_status()
            del self.auth_tokens[username]
            return True
        except requests.exceptions.RequestException as e:
            logger.error(f"Logout failed for user {username}: {str(e)}")
            return False


@pytest.fixture
def tester(api_url):
    return AuthTester(api_url)


def test_user_registration(tester):
    """Test user registration functionality"""
    test_user = {
        "username": f"test_user_{datetime.now().timestamp()}",
        "password": "TestPass123!"
    }

    # Test successful registration
    response = tester.register_user(test_user["username"], test_user["password"])
    assert "token" in response
    assert "username" in response
    assert response["username"] == test_user["username"]

    # Test duplicate registration
    with pytest.raises(requests.exceptions.RequestException):
        tester.register_user(test_user["username"], test_user["password"])


def test_user_login_logout(tester):
    """Test user login and logout functionality"""
    test_user = {
        "username": f"test_user_{datetime.now().timestamp()}",
        "password": "TestPass123!"
    }

    # Register user first
    tester.register_user(test_user["username"], test_user["password"])

    # Test successful login
    login_response = tester.login_user(test_user["username"], test_user["password"])
    assert "token" in login_response
    assert "username" in login_response

    # Test invalid password
    with pytest.raises(requests.exceptions.RequestException):
        tester.login_user(test_user["username"], "WrongPassword123!")

    # Test successful logout
    assert tester.logout_user(test_user["username"]) is True


def test_invalid_auth_token(tester):
    """Test authentication with invalid token"""
    headers = {"Authorization": "Bearer invalid_token"}
    response = requests.get(f"{tester.base_url}/api/messages/recent", headers=headers)
    assert response.status_code == 401


if __name__ == "__main__":
    pytest.main([__file__])
