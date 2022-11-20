import datetime

from workday.exceptions import InvalidRegistryCodeError


def validate_code_format(code_str: str) -> bool:
    """Validates provided code agains expected format.

    By definition the registry codes can only be 8-length, numeric only,
    valid time data strings.

    Args:
        code_str (str): The registry code to be validated.

    Raises:
        InvalidRegistryCodeError: if provided code has length different than 8.
        InvalidRegistryCodeError: if provided code has non numeric characters.

    Returns:
        bool: True if the provided code has a valid format.
    """
    if len(code_str) != 8:
        raise InvalidRegistryCodeError(
            "Registry codes should have length of 8 characters"
        )
    if not code_str.isdigit():
        raise InvalidRegistryCodeError("Registry codes should only have numbers")
    expected_format = "%Y%m%d"

    return bool(datetime.datetime.strptime(code_str, expected_format))
