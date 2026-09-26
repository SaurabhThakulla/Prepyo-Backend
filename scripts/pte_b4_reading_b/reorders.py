"""Aggregator for all 50 PTE Reading bank 4b Re-order Paragraphs items."""

from scripts.pte_b4_reading_b.reorders_part1 import REORDER_ITEMS_PART1
from scripts.pte_b4_reading_b.reorders_part2 import REORDER_ITEMS_PART2

ALL_REORDERS = REORDER_ITEMS_PART1 + REORDER_ITEMS_PART2
assert len(ALL_REORDERS) == 50, f"Expected 50 reorder items, got {len(ALL_REORDERS)}"
