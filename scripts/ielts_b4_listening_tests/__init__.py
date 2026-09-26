"""IELTS Listening Practice Tests 11-15 (migration 000096).

Original Prepyo practice material in the public IELTS Listening format; not
official, recalled or copied test content. Each test has four parts of ten
questions and each question carries `evidence`: the exact words in its part's
script that give the answer. scripts/generate_ielts_listening_tests_b4.py
checks everything and writes the SQL.
"""
from ielts_b4_listening_tests.test_11 import TEST as T11
from ielts_b4_listening_tests.test_12 import TEST as T12
from ielts_b4_listening_tests.test_13 import TEST as T13
from ielts_b4_listening_tests.test_14 import TEST as T14
from ielts_b4_listening_tests.test_15 import TEST as T15

TESTS = [T11, T12, T13, T14, T15]
