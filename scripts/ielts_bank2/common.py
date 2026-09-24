"""Shared helpers for the batch 2 IELTS content."""

W1 = 'ONE WORD ONLY'
WN = 'ONE WORD AND/OR A NUMBER'


def ABC(*texts):
    """Lettered options, A first."""
    return [{'id': chr(65 + i), 'text': t} for i, t in enumerate(texts)]


MAP_OPTIONS = [{'id': letter, 'text': f'Location {letter}'} for letter in 'ABCDEF']
