-- Targeted corrections to historical seeds. Existing IDs and learner references stay intact.
UPDATE questions SET figure_data = replace(figure_data,
 'Canada is highest throughout but grows the least in absolute terms; India starts lowest and grows fastest after 2010; the gap between Canada and India narrows from 50 to 51 points but the ratio falls sharply.',
 'Canada is highest throughout; Mexico has the largest absolute increase (67 percentage points), followed by Canada (43) and India (42); India starts lowest; from 2010 to 2020 Mexico gains 41 points and India 36; the Canada-India gap widens slightly from 50 to 51 points, while the ratio falls sharply.')
WHERE id = 'ielts-wrt-fg-001' AND exam = 'IELTS' AND skill = 'writing';

UPDATE questions SET figure_data = replace(figure_data,
 'under a third of its 2010 figure', 'under a third of its 2015 figure')
WHERE id = 'ielts-wrt-fg-004' AND exam = 'IELTS' AND skill = 'writing';

UPDATE questions SET figure_data = replace(figure_data,
 'organic waste is the largest fraction everywhere and rises as national income falls',
 'organic waste is the largest fraction in Germany, Brazil and Kenya, while paper is largest in Japan; the chart contains no income data, so do not infer an income relationship')
WHERE id = 'ielts-wrt-fg-009' AND exam = 'IELTS' AND skill = 'writing';

UPDATE questions SET model_answer = replace(model_answer,
 'reflective materials, green roofs and wider tree cover',
 'reflective roofing, permeable surfaces and wider tree cover')
WHERE id = 'pte-wrt-001' AND exam = 'PTE' AND skill = 'writing';

-- The outweigh prompt asks for an opinion, so use the supported opinion category.
UPDATE questions SET type_id = 'ielts-writing-task2-opinion', type_name = 'Opinion / Agree or Disagree'
WHERE id = 'ielts-wrt-001' AND type_id = 'ielts-writing-task2' AND exam = 'IELTS' AND skill = 'writing';

INSERT INTO questions (id, exam_version_id, exam, supported_exams, skill, type_id, type_name,
 title, prompt, context_passage, model_answer, explanation, time_limit_seconds, difficulty, tags, points)
VALUES
('pte-wrt-swt-002', 'pte-2026-01', 'PTE', ARRAY['PTE'], 'writing', 'summarize-written-text', 'Summarize Written Text',
 'Libraries Beyond Book Lending',
 'Read the passage below and summarise it in ONE single sentence of between 5 and 75 words.',
 'Public libraries are sometimes described as institutions made obsolete by online information. This view overlooks the difference between information being available and people being able to use it. Residents without reliable internet connections can use library computers, while staff help visitors complete forms and assess unfamiliar websites. Libraries also provide quiet study areas and meeting spaces that do not require a purchase. These services are especially valuable to people whose homes offer little privacy or room to work. However, expanding digital support while maintaining book collections requires trained staff and dependable funding. A library cannot meet every local need simply by installing more computers. Its continuing value comes from combining access to resources with practical assistance and a welcoming shared space.',
 'Although online information has changed their role, public libraries remain valuable by combining digital access, practical assistance and shared study spaces, provided that they have trained staff and dependable funding.',
 'Link the contrast with online information to the main services and the staffing and funding condition; use one sentence of 5-75 words.',
 600, 'medium', ARRAY['PTE Writing', 'Summarize Written Text'], 10),
('pte-wrt-swt-003', 'pte-2026-01', 'PTE', ARRAY['PTE'], 'writing', 'summarize-written-text', 'Summarize Written Text',
 'Repairing Rather Than Replacing',
 'Read the passage below and summarise it in ONE single sentence of between 5 and 75 words.',
 'When a household appliance stops working, replacement can appear easier than repair. A new product has a clear price, whereas diagnosing a fault takes time and may reveal further costs. Yet replacing an entire machine because one component fails also discards materials and energy invested in manufacturing it. Repair services can extend product life and reduce this waste, but their success depends on more than consumer goodwill. Technicians need spare parts, service information and equipment that can be opened without destroying its casing. Manufacturers can help by designing replaceable components and publishing repair instructions. Consumers still need reliable advice about safety and likely repair costs. Repair is therefore most practical when product design and accessible services make it a predictable option rather than an uncertain experiment.',
 'Repair can extend appliance life and reduce manufacturing waste, but it becomes a practical alternative to replacement only when repairable designs, spare parts, technical information and reliable services make costs and safety predictable.',
 'Include the environmental benefit and the conditions that make repair practical, rather than claiming that every appliance should always be repaired.',
 600, 'medium', ARRAY['PTE Writing', 'Summarize Written Text'], 10),
('pte-wrt-swt-004', 'pte-2026-01', 'PTE', ARRAY['PTE'], 'writing', 'summarize-written-text', 'Summarize Written Text',
 'Restoring Urban Streams',
 'Read the passage below and summarise it in ONE single sentence of between 5 and 75 words.',
 'Many urban streams were enclosed in underground pipes to make room for roads and buildings. Although this approach freed land for construction, it removed habitats and separated residents from local waterways. Some cities are now bringing these streams back to the surface. Planted banks can provide shelter for wildlife, while paths along the water create places for walking and recreation. Open channels may also slow the movement of stormwater when they have space to spread into surrounding vegetation. These benefits are not automatic. A narrow concrete channel offers little habitat, and polluted water remains polluted even when it is visible. Restoration therefore requires attention to water quality, bank design and maintenance as well as the removal of pipes. Successful projects treat the stream as part of a wider urban ecosystem rather than merely as a decorative feature.',
 'Bringing buried urban streams back to the surface can restore habitats, recreation space and stormwater capacity, but these benefits depend on water quality, suitable bank design and ongoing maintenance rather than simply removing pipes.',
 'Combine the potential environmental and recreational benefits with the conditions for success; do not claim that uncovering a stream alone removes pollution.',
 600, 'medium', ARRAY['PTE Writing', 'Summarize Written Text'], 10),
('pte-wrt-swt-005', 'pte-2026-01', 'PTE', ARRAY['PTE'], 'writing', 'summarize-written-text', 'Summarize Written Text',
 'Learning Through Retrieval',
 'Read the passage below and summarise it in ONE single sentence of between 5 and 75 words.',
 'Students often prepare for examinations by reading their notes repeatedly. Familiar wording can make this activity feel productive, even when the learner would struggle to explain the material without looking at it. Retrieval practice takes a different approach: learners close their notes and try to recall an explanation, solve a problem or answer a question. This effort reveals gaps that rereading may conceal. Checking the answer afterward provides feedback and prevents errors from being repeatedly rehearsed. Short retrieval activities spread across several days also give learners opportunities to revisit material after some forgetting has occurred. However, recalling isolated facts is not a substitute for understanding connections between ideas. Effective study combines retrieval and corrective feedback with explanations and varied problems, so that learners practise both remembering information and applying it in unfamiliar situations.',
 'Retrieval practice exposes gaps that familiar notes can conceal, and when spaced over time and combined with corrective feedback, explanations and varied problems, it supports both remembering information and applying it beyond rehearsed questions.',
 'Explain why recall differs from rereading and retain the need for feedback, spacing and understanding; avoid presenting memorisation alone as sufficient.',
 600, 'medium', ARRAY['PTE Writing', 'Summarize Written Text'], 10)
ON CONFLICT (id) DO NOTHING;
