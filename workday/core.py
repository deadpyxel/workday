import datetime
from dataclasses import dataclass
from typing import Optional

from workday.validations import validate_code_format


@dataclass
class DailyRegistry:
    """DailyRegistry class."""

    cod: str = ""
    work_start: Optional[datetime.datetime] = None
    end_start: Optional[datetime.datetime] = None
    lunch_start: Optional[datetime.datetime] = None
    lunch_end: Optional[datetime.datetime] = None
    notes: Optional[str] = None

    def __post_init__(self) -> None:
        """Post initialization process.

        Validates the provided registry code, and, if None,
        assigns one conforming to the specification.
        """
        if self.cod:
            validate_code_format(self.cod)
        current_date = datetime.datetime.today()
        self.cod = current_date.strftime("%Y%m%d")
