"""Shared helpers for IELTS Listening Practice Tests 11-15 (batch 4)."""

FORM = 'Complete the form below. Write ONE WORD AND/OR A NUMBER for each answer.'
NOTES_W1 = 'Complete the notes below. Write ONE WORD ONLY for each answer.'
NOTES_WN = 'Complete the notes below. Write ONE WORD AND/OR A NUMBER for each answer.'
MCQ = 'Choose the correct letter, A, B or C.'
PLAN = 'Label the plan below. Choose the correct letter, A-F.'


def ABC(*texts):
    """Lettered options, A first."""
    return [{'id': chr(65 + i), 'text': t} for i, t in enumerate(texts)]


MAP_OPTIONS = [{'id': letter, 'text': f'Location {letter}'} for letter in 'ABCDEF']


def form(heading, questions, instructions=FORM):
    return dict(type_id='ielts-listening-completion', heading=heading, instructions=instructions, limit=1,
                questions=questions)


def mcq(questions):
    return dict(type_id='ielts-listening-mcq', heading='', instructions=MCQ, questions=questions)


def matching(heading, instructions, questions):
    return dict(type_id='ielts-listening-matching', heading=heading, instructions=instructions, questions=questions)
