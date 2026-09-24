"""IELTS Academic Reading Bank Batch 3 Content (10 passages, 250 questions).

'lichens' and 'forecasting' replace 'urbanheat' and 'language', whose topics
were already live; they keep the replaced passages' places in the order.
"""

from scripts.content_b3.ielts_reading_p1_5 import PASSAGES_1_5
from scripts.content_b3.ielts_reading_p6_10 import PASSAGES_6_10
from scripts.content_b3.ielts_reading_replacements import FORECASTING, LICHENS

PASSAGES = PASSAGES_1_5[:1] + [LICHENS] + PASSAGES_1_5[1:] + PASSAGES_6_10[:3] + [FORECASTING] + PASSAGES_6_10[3:]
