UPDATE questions
SET prompt = context_passage || ' In 20 seconds, please prepare your response. You will have 40 seconds to give your response.'
WHERE type_id IN ('pte-respond-to-situation', 'respond-to-situation')
  AND context_passage IS NOT NULL
  AND prompt = 'Listen to and read a description of a situation. You will have 20 seconds to think about your answer. Then you will hear a beep. You will have 40 seconds to answer the question. Please answer as completely as you can.';
