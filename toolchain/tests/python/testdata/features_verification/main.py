import sys
import os

sys.path.append(os.getcwd())

from gen import *


def test_int_constants():
    assert MAX_PAGE_SIZE == 100
    assert MIN_PAGE_SIZE == 10
    assert isinstance(MAX_PAGE_SIZE, int)
    assert isinstance(MIN_PAGE_SIZE, int)


def test_string_constants():
    assert API_VERSION == "2.0.0"
    assert isinstance(API_VERSION, str)


def test_float_constants():
    assert PI_APPROX == 3.14159
    assert isinstance(PI_APPROX, float)


def test_bool_constants():
    assert DEBUG_MODE is False
    assert ENABLE_CACHE is True
    assert isinstance(DEBUG_MODE, bool)
    assert isinstance(ENABLE_CACHE, bool)


def test_pattern_functions():
    user_event = user_event_subject("user123", "login")
    assert user_event == "events.users.user123.login"

    cache_key_value = cache_key("sessions", "abc123")
    assert cache_key_value == "cache:sessions:abc123"

    endpoint = api_endpoint("2", "users", "42")
    assert endpoint == "/api/v2/users/42"


def test_pattern_interpolation():
    key_special = cache_key("user:data", "key/with/slashes")
    assert key_special == "cache:user:data:key/with/slashes"

    empty_parts = cache_key("", "")
    assert empty_parts == "cache::"

    numeric = user_event_subject("123", "456")
    assert numeric == "events.users.123.456"

    dup_segment = duplicated_segment("123", "abc")
    assert dup_segment == "users.123.abc.123"


if __name__ == "__main__":
    test_int_constants()
    test_string_constants()
    test_float_constants()
    test_bool_constants()
    test_pattern_functions()
    test_pattern_interpolation()
    print("Success")
