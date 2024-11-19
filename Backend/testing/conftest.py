import os
from pathlib import Path
import pytest
from dotenv import load_dotenv


def pytest_configure(config):
    env_path = Path(__file__).parent.parent / '.env'
    if env_path.exists():
        load_dotenv(env_path)
    else:
        env_path = Path(__file__).parent.parent / '.env.test'
        if env_path.exists():
            load_dotenv(env_path)


@pytest.fixture
def api_url():
    host = os.getenv('SERVER_HOST', 'localhost')
    port = os.getenv('SERVER_PORT', '6262')
    return f"http://{host}:{port}"


@pytest.fixture
def ws_url():
    host = os.getenv('SERVER_HOST', 'localhost')
    port = os.getenv('SERVER_PORT', '6262')
    return f"ws://{host}:{port}/api/ws"
