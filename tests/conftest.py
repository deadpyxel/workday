import pytest

from workday.core import DailyRegistry


@pytest.fixture()
def empty_registry() -> DailyRegistry:
    """Empty (new) DailyRegitry fixture.

    Returns:
        DailyRegistry: empty DailyRegistry instance
    """
    return DailyRegistry()
