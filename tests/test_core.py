import datetime

import pytest

from workday.core import DailyRegistry
from workday.exceptions import InvalidRegistryCodeError


def test_empty_registry_should_have_code(empty_registry: DailyRegistry) -> None:
    """Empty daily registries should at least be created with a code by default.

    Args:
        empty_registry (DailyRegistry): empty registry fixture.
    """
    assert empty_registry.cod
    assert isinstance(empty_registry.cod, str)


def test_new_registry_code_should_obey_format(empty_registry: DailyRegistry) -> None:
    """All Registries should obey the specified format.

    The designed format is YYYYMMDD or %Y%m%d.

    Args:
        empty_registry (DailyRegistry): empty registry fixture.
    """
    expected_format = "%Y%m%d"

    assert len(empty_registry.cod) == 8
    assert datetime.datetime.strptime(empty_registry.cod, expected_format)


@pytest.mark.parametrize(
    "invalid_str",
    [
        pytest.param("1999123a", id="has_letter"),
        pytest.param("?1111111", id="has_special_char"),
    ],
)
def test_codes_with_invalid_chars_raises_error(invalid_str: str) -> None:
    """Only numeric characters are allowed as part of registry code.

    Args:
        invalid_str (str): invalid characters sequence.
    """
    with pytest.raises(
        InvalidRegistryCodeError,
        match="Registry codes should only have numbers",
    ):
        DailyRegistry(cod=invalid_str)


@pytest.mark.parametrize(
    "invalid_str",
    [
        pytest.param("1", id="single_number"),
        pytest.param("1111", id="four_characters"),
        pytest.param("1111111111", id="ten_characters"),
    ],
)
def test_codes_with_wrong_length_raises_error(invalid_str: str) -> None:
    """Registry coulds should always be 8 numeric characters in length.

    Args:
        invalid_str (str): invalid character sequence.
    """
    with pytest.raises(
        InvalidRegistryCodeError,
        match="Registry codes should have length of 8 characters",
    ):
        DailyRegistry(cod=invalid_str)


@pytest.mark.parametrize(
    "invalid_str",
    [
        pytest.param("00000000", id="all_zeroes"),
        pytest.param("00010100", id="day_zero"),
        pytest.param("00010230", id="out_of_range"),
        pytest.param("00010132", id="extra_data"),
    ],
)
def test_codes_with_invalid_time_data_raises_error(invalid_str: str) -> None:
    """All Registry codes should be valid time data.

    Args:
        invalid_str (str): invalid character sequence.
    """
    expected_error_str = [
        f"time data '{invalid_str}' does not match format",
        "day is out of range for month",
        "unconverted data remains",
    ]
    error_matching_str = "|".join(expected_error_str)

    with pytest.raises(
        ValueError,
        match=error_matching_str,
    ):
        DailyRegistry(cod=invalid_str)
