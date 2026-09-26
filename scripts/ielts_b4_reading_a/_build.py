"""Small helpers that turn plain lists into the passage and option shapes used
by the batch-3 reading data (scripts/content_b3)."""


def paras(*texts):
    """Paragraph texts in order, labelled A, B, C..."""
    return [{"label": chr(65 + i), "text": t} for i, t in enumerate(texts)]


def opts(*texts):
    """Option texts in order, labelled A, B, C..."""
    return [{"id": chr(65 + i), "text": t} for i, t in enumerate(texts)]
