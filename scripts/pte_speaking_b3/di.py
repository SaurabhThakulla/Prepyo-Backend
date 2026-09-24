"""50 Original Describe Image (DI) items for PTE Speaking Batch 3.

Contains 10 bar charts, 10 line graphs, 10 pie charts, 10 tables, and 10 process diagrams.
All data is invented for fictional academic and municipal scenarios.
Invented locations: Eldridge, Oakhaven, Valoria, Silverdale, Blythewood, Kingswell, Crestview.
"""

from scripts.pte_speaking_b3.svg_utils import (
    bar_svg,
    line_svg,
    pie_svg,
    table_svg,
    process_svg,
    data_uri,
)

# 10 Bar charts
BARS = [
    {
        "title": "Eldridge Reading Centre Visits by Month (Thousands)",
        "kind": "bar",
        "data": {
            "title": "Eldridge Reading Centre Visits (Thousands)",
            "unit": "Visits (k)",
            "categories": ["Jan", "Mar", "May", "Jul", "Sep", "Nov"],
            "series": {
                "Central Branch": [45, 52, 60, 75, 58, 48],
                "Westside Branch": [30, 35, 42, 50, 40, 32],
            },
        },
        "figure_data": "Central Branch: Jan 45k, Mar 52k, May 60k, Jul 75k, Sep 58k, Nov 48k; Westside Branch: Jan 30k, Mar 35k, May 42k, Jul 50k, Sep 40k, Nov 32k.",
        "model_answer": "The bar chart shows monthly visitor numbers for two reading branches in the town of Eldridge in thousands. The Central Branch consistently received higher visits than the Westside Branch across all six recorded months. Both centres experienced peak attendance in July, reaching 75,000 and 50,000 visits respectively, while January recorded the lowest visits. In conclusion, visitor traffic follows a clear summer peak before declining towards the winter months.",
        "explanation": "Identify the title and metric, highlight that Central Branch consistently leads, mention the July peak, and conclude with the seasonal pattern.",
    },
    {
        "title": "Oakhaven Refuse Sorting Rates across Municipalities (%)",
        "kind": "bar",
        "data": {
            "title": "Oakhaven Refuse Sorting Rates (%)",
            "unit": "Percentage (%)",
            "categories": ["District A", "District B", "District C", "District D", "District E"],
            "series": {
                "2020": [35, 42, 28, 50, 45],
                "2025": [48, 55, 38, 62, 54],
            },
        },
        "figure_data": "2020: District A 35%, District B 42%, District C 28%, District D 50%, District E 45%; 2025: District A 48%, District B 55%, District C 38%, District D 62%, District E 54%.",
        "model_answer": "The bar chart compares refuse sorting percentages across five districts in the municipality of Oakhaven between 2020 and 2025. Every district showed a measurable increase over the five-year period. District D achieved the highest rate in both years, rising from 50 percent to 62 percent, whereas District C maintained the lowest rates at 28 percent and 38 percent. Overall, waste separation practices improved consistently across all surveyed areas.",
        "explanation": "Compare the two survey years, highlight District D as the highest and District C as the lowest, and conclude with the general upward trend.",
    },
    {
        "title": "Valoria Public Bus Fleet Electrification (%)",
        "kind": "bar",
        "data": {
            "title": "Valoria Bus Fleet Electrification (%)",
            "unit": "Percentage (%)",
            "categories": ["Sector North", "Sector South", "Sector East", "Sector West"],
            "series": {
                "2022": [15, 22, 18, 10],
                "2024": [38, 45, 32, 28],
                "2026": [70, 82, 65, 58],
            },
        },
        "figure_data": "2022: North 15%, South 22%, East 18%, West 10%; 2024: North 38%, South 45%, East 32%, West 28%; 2026: North 70%, South 82%, East 65%, West 58%.",
        "model_answer": "The chart illustrates the percentage of electric buses across four transit sectors in the city of Valoria from 2022 to 2026. Sector South led the transition in every year, progressing from 22 percent to an impressive 82 percent by 2026. Conversely, Sector West lagged initially at 10 percent but grew significantly to 58 percent. Overall, electric bus adoption accelerated rapidly across the entire metropolitan area.",
        "explanation": "State the metric and four sectors, contrast Sector South with Sector West, and summarize the rapid electrification trend.",
    },
    {
        "title": "Silverdale Tertiary Enrollment by Faculty (Hundreds)",
        "kind": "bar",
        "data": {
            "title": "Silverdale Tertiary Enrollment (Hundreds)",
            "unit": "Students (hundreds)",
            "categories": ["Engineering", "Humanities", "Business", "Medicine", "Natural Sciences"],
            "series": {
                "Domestic": [24, 18, 30, 15, 20],
                "International": [12, 6, 18, 8, 10],
            },
        },
        "figure_data": "Domestic: Engineering 24, Humanities 18, Business 30, Medicine 15, Sciences 20; International: Engineering 12, Humanities 6, Business 18, Medicine 8, Sciences 10.",
        "model_answer": "The bar chart displays student enrollment across five faculties at Silverdale University in hundreds of students. Business enrolled the highest numbers of both domestic and international students, totaling 3,000 and 1,800 students respectively. In contrast, the Humanities faculty recorded the lowest international enrollment at only 600 students. In summary, business and engineering remain the dominant disciplines for both cohorts.",
        "explanation": "Identify the top faculty (Business) and lowest (Humanities international), compare domestic versus international figures, and summarize.",
    },
    {
        "title": "Blythewood Solar Energy Generation (Megawatt Hours)",
        "kind": "bar",
        "data": {
            "title": "Blythewood Solar Generation (MWh)",
            "unit": "MWh",
            "categories": ["Spring", "Summer", "Autumn", "Winter"],
            "series": {
                "Array Alpha": [420, 680, 390, 210],
                "Array Beta": [350, 590, 310, 180],
            },
        },
        "figure_data": "Array Alpha: Spring 420, Summer 680, Autumn 390, Winter 210; Array Beta: Spring 350, Summer 590, Autumn 310, Winter 180 MWh.",
        "model_answer": "This bar chart outlines seasonal power output in megawatt hours from two solar arrays in Blythewood. Both installations peaked during the summer season, with Array Alpha generating 680 MWh and Array Beta producing 590 MWh. Energy generation declined sharply in winter, reaching seasonal lows of 210 and 180 MWh. To conclude, solar output is heavily dependent on seasonal daylight variations.",
        "explanation": "Highlight the summer peaks, winter minimums, and compare Array Alpha's consistent lead over Array Beta.",
    },
    {
        "title": "Kingswell Clinic Consultation Delay Times (Minutes)",
        "kind": "bar",
        "data": {
            "title": "Kingswell Clinic Delay Times (Minutes)",
            "unit": "Minutes",
            "categories": ["Pediatrics", "Cardiology", "Orthopedics", "Dermatology", "General Practice"],
            "series": {
                "Morning": [18, 25, 30, 12, 22],
                "Afternoon": [28, 35, 42, 20, 38],
            },
        },
        "figure_data": "Morning: Pediatrics 18m, Cardiology 25m, Orthopedics 30m, Dermatology 12m, General 22m; Afternoon: Pediatrics 28m, Cardiology 35m, Orthopedics 42m, Dermatology 20m, General 38m.",
        "model_answer": "The bar chart compares patient consultation delay times across five medical departments at Kingswell Clinic during morning and afternoon shifts. In every department, afternoon delays were longer than morning delays. Orthopedics recorded the longest wait times, reaching 42 minutes in the afternoon, while Dermatology had the shortest wait at only 12 minutes in the morning. In conclusion, outpatient delays increase substantially as the day progresses.",
        "explanation": "Compare morning versus afternoon wait times, identify Orthopedics as highest and Dermatology as lowest, and summarize the daily pattern.",
    },
    {
        "title": "Crestview Municipal Budget Expenditure ($ Millions)",
        "kind": "bar",
        "data": {
            "title": "Crestview Budget Expenditure ($M)",
            "unit": "$ Millions",
            "categories": ["Transit", "Education", "Parks", "Public Safety", "Sanitation"],
            "series": {
                "Budgeted": [45, 60, 20, 50, 30],
                "Actual": [48, 58, 22, 54, 29],
            },
        },
        "figure_data": "Budgeted: Transit 45M, Education 60M, Parks 20M, Safety 50M, Sanitation 30M; Actual: Transit 48M, Education 58M, Parks 22M, Safety 54M, Sanitation 29M.",
        "model_answer": "This bar chart presents planned versus actual municipal expenditures for the district of Crestview in millions of dollars. Education received the largest expenditure, with 58 million dollars spent against a 60 million budget. Public Safety followed closely with 54 million in actual spending, whereas Parks accounted for the smallest allocation at 22 million. Overall, actual spending aligned closely with planned budgetary allocations across all sectors.",
        "explanation": "Note Education as the highest category and Parks as the lowest, compare budgeted versus actual figures, and draw a summary conclusion.",
    },
    {
        "title": "Greenvale Residential Thermal Power Sources (%)",
        "kind": "bar",
        "data": {
            "title": "Greenvale Thermal Power Sources (%)",
            "unit": "Percentage (%)",
            "categories": ["Zone 1", "Zone 2", "Zone 3", "Zone 4"],
            "series": {
                "Natural Gas": [55, 40, 30, 20],
                "District Heating": [25, 35, 45, 50],
                "Electric Heat": [20, 25, 25, 30],
            },
        },
        "figure_data": "Natural Gas: Zone 1 55%, Zone 2 40%, Zone 3 30%, Zone 4 20%; District: Zone 1 25%, Zone 2 35%, Zone 3 45%, Zone 4 50%; Electric: Zone 1 20%, Zone 2 25%, Zone 3 25%, Zone 4 30%.",
        "model_answer": "The chart illustrates heating energy sources across four zones in the municipality of Greenvale. Natural gas dominates in Zone 1 at 55 percent but falls to just 20 percent in Zone 4. Conversely, district heating rises steadily from 25 percent in Zone 1 to 50 percent in Zone 4. In summary, newer municipal zones rely significantly more on centralized district heating and electric systems than older sectors.",
        "explanation": "Contrast Natural Gas dominance in Zone 1 with District Heating growth in Zone 4, and conclude with the structural transition.",
    },
    {
        "title": "Port Haven Container Crane Productivity (Moves/Hour)",
        "kind": "bar",
        "data": {
            "title": "Port Haven Crane Productivity",
            "unit": "Moves per hour",
            "categories": ["Berth 1", "Berth 2", "Berth 3", "Berth 4"],
            "series": {
                "Standard Crane": [22, 24, 20, 25],
                "Automated Crane": [32, 36, 30, 38],
            },
        },
        "figure_data": "Standard: Berth 1 22, Berth 2 24, Berth 3 20, Berth 4 25; Automated: Berth 1 32, Berth 2 36, Berth 3 30, Berth 4 38 moves/hour.",
        "model_answer": "The bar chart compares container handling speed in moves per hour between standard and automated cranes at four berths in Port Haven. Automated cranes outperformed manual cranes across all four berths, achieving their highest rate of 38 moves per hour at Berth 4. Standard cranes averaged between 20 and 25 moves per hour. In conclusion, automated cargo handling improves loading throughput by over 40 percent.",
        "explanation": "Highlight the consistent efficiency advantage of automated cranes, identify the peak at Berth 4, and summarize the overall throughput gain.",
    },
    {
        "title": "Riverton Stream Hydrographic Flow Volume (m3/s)",
        "kind": "bar",
        "data": {
            "title": "Riverton Hydrographic Flow (m3/s)",
            "unit": "Cubic metres per second",
            "categories": ["Winter", "Spring", "Summer", "Autumn"],
            "series": {
                "Upper Reach": [35, 95, 25, 45],
                "Lower Reach": [55, 140, 40, 70],
            },
        },
        "figure_data": "Upper: Winter 35, Spring 95, Summer 25, Autumn 45; Lower: Winter 55, Spring 140, Summer 40, Autumn 70 m3/s.",
        "model_answer": "The bar chart displays seasonal water discharge rates in cubic metres per second for two sections of the Riverton waterway. Discharge peaked dramatically in spring, reaching 140 cubic metres per second at the Lower Reach and 95 at the Upper Reach, driven by snowmelt. Summer recorded the lowest flow rates of 25 and 40 cubic metres per second. To summarize, the river displays strong seasonal discharge variability.",
        "explanation": "Describe spring runoff peaks, identify summer low flows, compare upper and lower sections, and conclude.",
    },
]

# 10 Line graphs
LINES = [
    {
        "title": "Oakwood Air Particulate Monitoring Index (ug/m3)",
        "kind": "line",
        "data": {
            "title": "Oakwood Particulate Index (ug/m3)",
            "unit": "ug/m3",
            "categories": ["2018", "2019", "2020", "2021", "2022", "2023"],
            "series": {
                "Industrial Area": [48, 44, 38, 35, 32, 28],
                "Suburban Park": [22, 20, 18, 16, 15, 13],
            },
        },
        "figure_data": "Industrial: 2018 48, 2019 44, 2020 38, 2021 35, 2022 32, 2023 28; Suburban: 2018 22, 2019 20, 2020 18, 2021 16, 2022 15, 2023 13 ug/m3.",
        "model_answer": "The line graph tracks particulate concentrations in micrograms per cubic metre across two zones in Oakwood from 2018 to 2023. Both locations exhibited a steady and uninterrupted decline in pollution over the five-year timeframe. The industrial sector decreased from 48 to 28 micrograms, while the suburban park remained cleaner, falling from 22 to 13. In summary, air quality improved substantially across the municipality.",
        "explanation": "Highlight the steady downward trajectory in both zones, compare absolute values, and conclude with the air quality improvement.",
    },
    {
        "title": "Westhaven Municipal Desalination Output (Megalitres/Day)",
        "kind": "line",
        "data": {
            "title": "Westhaven Desalination Output (ML/day)",
            "unit": "Megalitres/day",
            "categories": ["2016", "2018", "2020", "2022", "2024"],
            "series": {
                "Plant Alpha": [40, 55, 75, 90, 110],
                "Plant Beta": [25, 35, 45, 60, 85],
            },
        },
        "figure_data": "Plant Alpha: 2016 40, 2018 55, 2020 75, 2022 90, 2024 110; Plant Beta: 2016 25, 2018 35, 2020 45, 2022 60, 2024 85 ML/day.",
        "model_answer": "The line graph traces daily potable water production in megalitres from two desalination facilities in Westhaven between 2016 and 2024. Output expanded continuously at both plants throughout the period. Plant Alpha consistently led production, rising from 40 to 110 megalitres per day, while Plant Beta more than tripled its output from 25 to 85. Overall, municipal desalination capacity grew more than twofold.",
        "explanation": "Identify the continuous growth trend, note that Plant Alpha remains higher, and highlight that total production more than doubled.",
    },
    {
        "title": "Highfield Farm Soil Moisture Levels (%)",
        "kind": "line",
        "data": {
            "title": "Highfield Soil Moisture Levels (%)",
            "unit": "Percentage (%)",
            "categories": ["May", "Jun", "Jul", "Aug", "Sep", "Oct"],
            "series": {
                "Mulched Plot": [32, 28, 25, 24, 27, 30],
                "Bare Soil Plot": [28, 22, 16, 14, 18, 24],
            },
        },
        "figure_data": "Mulched: May 32%, Jun 28%, Jul 25%, Aug 24%, Sep 27%, Oct 30%; Bare: May 28%, Jun 22%, Jul 16%, Aug 14%, Sep 18%, Oct 24%.",
        "model_answer": "The line graph illustrates monthly soil moisture percentages for mulched and bare agricultural test plots at Highfield Farm. Moisture levels declined to mid-summer troughs in August before recovering in autumn. The mulched plot maintained higher moisture retention throughout, dropping only to 24 percent compared to 14 percent for bare soil. In conclusion, organic mulching significantly prevents soil dehydration during peak summer heat.",
        "explanation": "Observe the U-shaped seasonal curve, contrast the mulched plot's superior retention against bare soil, and summarize.",
    },
    {
        "title": "Cafés and bookshops on a high street",
        "kind": "line",
        "data": {
            "title": "Cafés and Bookshops on Market Street",
            "unit": "Number of shops",
            "categories": ["2015", "2017", "2019", "2021", "2023"],
            "series": {
                "Cafés": [6, 8, 11, 14, 17],
                "Bookshops": [7, 6, 5, 4, 4],
            },
        },
        "figure_data": "Cafés: 2015 6, 2017 8, 2019 11, 2021 14, 2023 17; Bookshops: 2015 7, 2017 6, 2019 5, 2021 4, 2023 4.",
        "model_answer": "The line graph compares the number of cafés and bookshops on Market Street between 2015 and 2023. In 2015 there were slightly more bookshops, seven, than cafés, six. After that, the number of cafés rose steadily to reach seventeen in 2023, while bookshops fell gradually to four and then stayed level. The two lines crossed shortly after 2015. Overall, the street changed from a mix of both kinds of shop to one dominated by cafés.",
        "explanation": "Give the starting figures, describe the steady rise in cafés and the fall in bookshops, note where the lines cross, and summarise the overall change.",
    },
    {
        "title": "Valoria Metro Transit Ridership (Millions)",
        "kind": "line",
        "data": {
            "title": "Valoria Metro Ridership (Millions)",
            "unit": "Annual passengers (M)",
            "categories": ["2017", "2018", "2019", "2020", "2021", "2022", "2023"],
            "series": {
                "Blue Line": [45, 48, 52, 28, 35, 44, 50],
                "Red Line": [32, 35, 38, 20, 25, 32, 37],
            },
        },
        "figure_data": "Blue: 2017 45, 2018 48, 2019 52, 2020 28, 2021 35, 2022 44, 2023 50; Red: 2017 32, 2018 35, 2019 38, 2020 20, 2021 25, 2022 32, 2023 37.",
        "model_answer": "The line graph traces annual passenger numbers on two metro rail lines in Valoria from 2017 to 2023. Both lines grew steadily until 2019, experienced a sharp contraction in 2020 to 28 and 20 million riders, and rebounded robustly thereafter. By 2023, Blue Line ridership had almost recovered to pre-pandemic peaks at 50 million. To conclude, transit demand showed strong resilience following temporary disruption.",
        "explanation": "Describe pre-2020 growth, the steep 2020 drop, the subsequent recovery, and Blue Line's consistent lead over Red Line.",
    },
    {
        "title": "Stonebridge Wind Farm Capacity Factor (%)",
        "kind": "line",
        "data": {
            "title": "Stonebridge Wind Capacity Factor (%)",
            "unit": "Percentage (%)",
            "categories": ["Jan", "Mar", "May", "Jul", "Sep", "Nov"],
            "series": {
                "Offshore Station": [48, 42, 32, 24, 34, 46],
                "Onshore Station": [38, 32, 24, 16, 26, 36],
            },
        },
        "figure_data": "Offshore: Jan 48%, Mar 42%, May 32%, Jul 24%, Sep 34%, Nov 46%; Onshore: Jan 38%, Mar 32%, May 24%, Jul 16%, Sep 26%, Nov 36%.",
        "model_answer": "The line graph demonstrates monthly wind generator capacity factors at offshore and onshore stations in Stonebridge. Both stations achieved peak operating efficiency during the winter months of January and November, reaching 48 and 38 percent respectively. Performance fell to seasonal lows in July. Offshore turbines maintained an approximately ten percent efficiency advantage throughout the year due to stronger marine winds.",
        "explanation": "Highlight the winter peak vs summer trough, compare offshore superiority, and conclude with seasonal wind patterns.",
    },
    {
        "title": "Fairview River Dissolved Oxygen Metrics (mg/L)",
        "kind": "line",
        "data": {
            "title": "Fairview Dissolved Oxygen (mg/L)",
            "unit": "mg/L",
            "categories": ["2019", "2020", "2021", "2022", "2023"],
            "series": {
                "Upstream Control": [9.5, 9.6, 9.4, 9.5, 9.7],
                "Downstream Recovery": [5.2, 5.8, 6.5, 7.4, 8.2],
            },
        },
        "figure_data": "Upstream: 2019 9.5, 2020 9.6, 2021 9.4, 2022 9.5, 2023 9.7; Downstream: 2019 5.2, 2020 5.8, 2021 6.5, 2022 7.4, 2023 8.2 mg/L.",
        "model_answer": "This line graph records dissolved oxygen levels in milligrams per litre at two monitoring points along the Fairview River. Upstream control values remained stable near 9.5 mg/L across the five-year study. Downstream concentrations showed dramatic ecological recovery, rising steadily from a hypoxic 5.2 in 2019 to 8.2 in 2023. In conclusion, downstream water rehabilitation measures successfully narrowed the ecological gap.",
        "explanation": "Contrast stable upstream baseline with steep downstream recovery, and summarize the ecological improvement.",
    },
    {
        "title": "Bellview District Domestic Electricity Usage (kWh/Day)",
        "kind": "line",
        "data": {
            "title": "Bellview Domestic Electricity (kWh/day)",
            "unit": "kWh/day",
            "categories": ["Spring", "Early Summer", "Mid Summer", "Autumn", "Winter"],
            "series": {
                "Smart Meter Homes": [16, 18, 22, 15, 24],
                "Standard Homes": [20, 24, 30, 19, 32],
            },
        },
        "figure_data": "Smart: Spring 16, Early 18, Mid 22, Autumn 15, Winter 24; Standard: Spring 20, Early 24, Mid 30, Autumn 19, Winter 32 kWh/day.",
        "model_answer": "The line graph compares average household daily power consumption in kilowatt hours between homes with smart meters and standard meters in Bellview. Both cohorts peaked in winter and mid-summer due to heating and cooling demands. However, smart meter households consistently consumed 20 to 25 percent less electricity in every season, peaking at 24 kWh compared to 32 kWh for standard homes. Overall, real-time feedback encourages sustained energy conservation.",
        "explanation": "Identify the dual winter/summer seasonal peaks, quantify the smart meter efficiency savings, and summarize.",
    },
    {
        "title": "Silverlake Municipal Recycling Rate Trajectory (%)",
        "kind": "line",
        "data": {
            "title": "Silverlake Recycling Trajectory (%)",
            "unit": "Percentage (%)",
            "categories": ["2015", "2017", "2019", "2021", "2023"],
            "series": {
                "Paper and Cardboard": [45, 52, 60, 68, 74],
                "Plastics and Glass": [25, 30, 38, 48, 56],
            },
        },
        "figure_data": "Paper: 2015 45%, 2017 52%, 2019 60%, 2021 68%, 2023 74%; Plastics: 2015 25%, 2017 30%, 2019 38%, 2021 48%, 2023 56%.",
        "model_answer": "The line graph illustrates recovery rates for two recyclable waste streams in Silverlake between 2015 and 2023. Both categories displayed consistent, parallel upward trajectories. Paper and cardboard recycling increased from 45 to 74 percent, maintaining a steady lead throughout. Plastics and glass recycling more than doubled, progressing from 25 to 56 percent. In summary, municipal waste diversion programs achieved steady long-term success.",
        "explanation": "Highlight the parallel growth trends, compare paper's lead with plastics' rapid doubling, and summarize the overall policy success.",
    },
    {
        "title": "Crestmont High School Graduation Percentage (%)",
        "kind": "line",
        "data": {
            "title": "Crestmont Graduation Percentage (%)",
            "unit": "Percentage (%)",
            "categories": ["2018", "2019", "2020", "2021", "2022", "2023"],
            "series": {
                "Vocational Track": [72, 75, 78, 82, 85, 88],
                "Academic Track": [84, 85, 87, 88, 90, 92],
            },
        },
        "figure_data": "Vocational: 2018 72%, 2019 75%, 2020 78%, 2021 82%, 2022 85%, 2023 88%; Academic: 2018 84%, 2019 85%, 2020 87%, 2021 88%, 2022 90%, 2023 92%.",
        "model_answer": "The line graph traces secondary school graduation rates across vocational and academic pathways in Crestmont from 2018 to 2023. Both cohorts recorded gradual increases over the six-year period. While the academic pathway consistently achieved higher completion rates, reaching 92 percent, the vocational track experienced faster relative gains, climbing from 72 to 88 percent. In conclusion, the graduation gap between tracks narrowed over time.",
        "explanation": "Point out the upward trend in both cohorts, note academic track's higher level, and conclude with the narrowing achievement gap.",
    },
]

# 10 Pie charts
PIES = [
    {
        "title": "Eldridge Municipal Domestic Water Consumption Share (%)",
        "kind": "pie",
        "data": {
            "title": "Eldridge Domestic Water Usage (%)",
            "unit": "%",
            "slices": {
                "Showers and Bathing": 35,
                "Toilet Flushing": 25,
                "Washing Machines": 20,
                "Kitchen and Cooking": 12,
                "Outdoor Gardens": 8,
            },
        },
        "figure_data": "Bathing 35%, Toilets 25%, Washing Machines 20%, Kitchen 12%, Outdoor Gardens 8%.",
        "model_answer": "The pie chart outlines the breakdown of household water usage in the town of Eldridge. Bathing accounts for the largest proportion at 35 percent, followed by toilet flushing at 25 percent and clothes washing at 20 percent. Kitchen use and garden maintenance comprise the remaining 20 percent combined. In summary, personal hygiene and sanitation represent over half of total domestic water consumption.",
        "explanation": "Identify bathing as the dominant slice (35%), highlight outdoor gardens as the smallest (8%), and summarize sanitation dominance.",
    },
    {
        "title": "Oakhaven Municipal Land Allocation Breakdown (%)",
        "kind": "pie",
        "data": {
            "title": "Oakhaven Land Allocation (%)",
            "unit": "%",
            "slices": {
                "Residential Zones": 38,
                "Protected Parks": 24,
                "Commercial Hubs": 18,
                "Industrial Parks": 12,
                "Transport Corridors": 8,
            },
        },
        "figure_data": "Residential 38%, Parks 24%, Commercial 18%, Industrial 12%, Transport 8%.",
        "model_answer": "The pie chart illustrates land use distribution in the district of Oakhaven. Residential housing occupies the greatest share at 38 percent, followed by protected green spaces and parks at 24 percent. Commercial and industrial developments constitute 18 and 12 percent respectively, while transportation networks account for 8 percent. In conclusion, residential and parkland areas together make up over sixty percent of municipal land.",
        "explanation": "State the leading category (Residential 38%), contrast with transport (8%), and summarize the combined green/residential majority.",
    },
    {
        "title": "Valoria Commercial Cargo Fleet Propulsion Types (%)",
        "kind": "pie",
        "data": {
            "title": "Valoria Fleet Propulsion Types (%)",
            "unit": "%",
            "slices": {
                "Heavy Fuel Oil": 45,
                "Liquefied Gas": 25,
                "Bio-Methanol": 15,
                "Battery Hybrid": 10,
                "Wind Assist": 5,
            },
        },
        "figure_data": "Heavy Fuel Oil 45%, Liquefied Gas 25%, Bio-Methanol 15%, Hybrid 10%, Wind Assist 5%.",
        "model_answer": "The pie chart displays the distribution of engine fuel types across commercial maritime cargo ships registered in Valoria. Traditional heavy fuel oil represents the single largest share at 45 percent, while liquefied gas constitutes a quarter of the fleet. Cleaner alternatives including methanol, hybrids, and wind propulsion make up the remaining 30 percent. Overall, fossil fuels continue to dominate maritime commercial shipping.",
        "explanation": "Note Heavy Fuel Oil as the majority (45%), contrast with emergent wind/hybrid tech, and conclude on fossil fuel dominance.",
    },
    {
        "title": "Silverdale Mobile Computing Operating Platform Share (%)",
        "kind": "pie",
        "data": {
            "title": "Silverdale Mobile Platform Share (%)",
            "unit": "%",
            "slices": {
                "Platform Alpha": 52,
                "Platform Beta": 36,
                "Platform Gamma": 8,
                "Open Source Other": 4,
            },
        },
        "figure_data": "Alpha 52%, Beta 36%, Gamma 8%, Other 4%.",
        "model_answer": "The pie chart shows the market share of mobile operating systems in the region of Silverdale. Platform Alpha commands an absolute majority with 52 percent of the market, followed by Platform Beta at 36 percent. Minority platforms Gamma and open-source operating systems account for only 8 and 4 percent respectively. In conclusion, the mobile ecosystem is heavily dominated by a duopoly controlling 88 percent of devices.",
        "explanation": "Identify Platform Alpha's majority (>50%), point out Platform Beta's second position, and summarize the two-firm market dominance.",
    },
    {
        "title": "Blythewood Academic Reading Centre Expenditure (%)",
        "kind": "pie",
        "data": {
            "title": "Blythewood Reading Centre Budget (%)",
            "unit": "%",
            "slices": {
                "Electronic Subscriptions": 48,
                "Physical Books": 22,
                "Staffing Personnel": 18,
                "Facility Maintenance": 8,
                "Student Workshops": 4,
            },
        },
        "figure_data": "E-Subscriptions 48%, Print Books 22%, Staff 18%, Maintenance 8%, Workshops 4%.",
        "model_answer": "This pie chart outlines annual resource expenditure for the Blythewood campus reading centre. Digital journals and electronic database subscriptions represent nearly half of total spending at 48 percent. Physical print collections absorb 22 percent, while staff salaries take up 18 percent. Workshops account for the smallest fraction at 4 percent. To conclude, digital academic resources represent the primary budgetary commitment.",
        "explanation": "Highlight electronic subscriptions as the dominant expense (48%), contrast with physical books and workshops, and summarize.",
    },
    {
        "title": "Kingswell District Generation Portfolio Breakdown (%)",
        "kind": "pie",
        "data": {
            "title": "Kingswell Electricity Generation (%)",
            "unit": "%",
            "slices": {
                "Wind Power": 40,
                "Solar Photovoltaic": 25,
                "Hydroelectric": 18,
                "Natural Gas": 12,
                "Biomass": 5,
            },
        },
        "figure_data": "Wind 40%, Solar 25%, Hydro 18%, Natural Gas 12%, Biomass 5%.",
        "model_answer": "The pie chart presents the electricity generation mix for the Kingswell region. Renewable wind energy forms the largest component at 40 percent, complemented by solar photovoltaic power at 25 percent and hydroelectricity at 18 percent. Fossil gas represents only 12 percent, with biomass contributing 5 percent. In summary, renewable sources account for over eighty percent of the region's power generation.",
        "explanation": "Highlight Wind as the primary source (40%), note that renewables exceed 80% combined, and conclude on the clean energy mix.",
    },
    {
        "title": "Types of calls to a city helpline",
        "kind": "pie",
        "data": {
            "title": "Calls to the Crestview City Helpline (%)",
            "unit": "%",
            "slices": {"Waste collection": 31, "Road repairs": 24, "Parking": 19, "Noise": 15, "Other": 11},
        },
        "figure_data": "Waste collection 31%, Road repairs 24%, Parking 19%, Noise 15%, Other 11%.",
        "model_answer": "The pie chart shows the types of calls received by a city helpline. The largest share, 31 percent, concerned waste collection, followed by road repairs at 24 percent. Parking accounted for 19 percent and noise for 15 percent, while other issues made up the remaining 11 percent. Overall, more than half of all calls were about waste collection and roads.",
        "explanation": "Identify waste collection as the largest category and other issues as the smallest, and note that waste and roads together exceed half.",
    },
    {
        "title": "How time is spent at a youth summer camp",
        "kind": "pie",
        "data": {
            "title": "Greenfield Youth Camp: Share of Activity Time (%)",
            "unit": "%",
            "slices": {"Sport": 30, "Hiking": 25, "Crafts": 20, "Free time": 15, "Music": 10},
        },
        "figure_data": "Sport 30%, Hiking 25%, Crafts 20%, Free time 15%, Music 10%.",
        "model_answer": "The pie chart shows how time is divided between activities at a youth summer camp. Sport takes up the largest share at 30 percent, followed by hiking at 25 percent and crafts at 20 percent. Free time accounts for 15 percent, and music has the smallest share at 10 percent. In short, active outdoor pursuits make up more than half of the camp's time.",
        "explanation": "Name sport as the largest and music as the smallest share, and summarise that outdoor activities dominate.",
    },
    {
        "title": "Favourite kinds of holiday",
        "kind": "pie",
        "data": {
            "title": "Favourite Type of Holiday, Survey of 500 Adults (%)",
            "unit": "%",
            "slices": {"Beach": 38, "City break": 27, "Countryside": 18, "Cruise": 11, "Other": 6},
        },
        "figure_data": "Beach 38%, City break 27%, Countryside 18%, Cruise 11%, Other 6%.",
        "model_answer": "The pie chart shows the favourite type of holiday among 500 adults who took part in a survey. Beach holidays were the most popular, chosen by 38 percent, followed by city breaks at 27 percent. Eighteen percent preferred the countryside and 11 percent chose cruises, while other kinds of holiday made up just 6 percent. In summary, beach holidays and city breaks together were the choice of almost two thirds of the people asked.",
        "explanation": "Name beach holidays as the most popular and other holidays as the least, give the middle figures, and summarise the combined share of beach holidays and city breaks.",
    },
    {
        "title": "Portview Container Export Destination Profile (%)",
        "kind": "pie",
        "data": {
            "title": "Portview Export Destinations (%)",
            "unit": "%",
            "slices": {
                "East Asia": 44,
                "North America": 26,
                "Western Europe": 18,
                "Australasia": 8,
                "South America": 4,
            },
        },
        "figure_data": "East Asia 44%, North America 26%, Western Europe 18%, Australasia 8%, South America 4%.",
        "model_answer": "This pie chart illustrates destination regions for containerized maritime exports shipped from Portview. East Asia is the primary trade destination, receiving 44 percent of outgoing cargo, followed by North America at 26 percent and Western Europe at 18 percent. Trade with Australasia and South America totals 12 percent combined. In conclusion, transpacific trade routes dominate the port's outbound commercial logistics.",
        "explanation": "Identify East Asia as the largest destination (44%), North America second (26%), and summarize the geographic concentration.",
    },
]

# 10 Data Tables
TABLES = [
    {
        "title": "Continental Rail Travel Times and Fares",
        "kind": "table",
        "data": {
            "title": "Continental Rail Metrics",
            "headers": ["Corridor Route", "Distance (km)", "Travel Time (hrs)", "Standard Fare ($)"],
            "rows": [
                ["Eldridge to Silverdale", 320, 2.5, 45],
                ["Valoria to Crestview", 540, 4.0, 72],
                ["Oakhaven to Blythewood", 180, 1.5, 28],
                ["Kingswell to Greenvale", 680, 5.2, 95],
                ["Port Haven to Riverton", 410, 3.2, 60],
            ],
        },
        "figure_data": "Eldridge-Silverdale: 320km, 2.5h, $45; Valoria-Crestview: 540km, 4.0h, $72; Oakhaven-Blythewood: 180km, 1.5h, $28; Kingswell-Greenvale: 680km, 5.2h, $95; Port Haven-Riverton: 410km, 3.2h, $60.",
        "model_answer": "The table compares distance, journey duration, and ticket fares across five regional rail routes. Kingswell to Greenvale is the longest and most expensive route, spanning 680 kilometers, taking 5.2 hours, and costing 95 dollars. In contrast, Oakhaven to Blythewood is the shortest at 180 kilometers with a travel time of 1.5 hours and a fare of 28 dollars. Overall, ticket prices and transit times scale proportionally with route distance.",
        "explanation": "Highlight the longest/most expensive route, compare it with the shortest/cheapest route, and summarize the distance-cost correlation.",
    },
    {
        "title": "Forestry Timber Mechanical Density and Strength",
        "kind": "table",
        "data": {
            "title": "Timber Mechanical Properties",
            "headers": ["Species Name", "Density (kg/m3)", "Bending Strength (MPa)", "Hardness Rating (kN)"],
            "rows": [
                ["Alpine Larch", 590, 98, 4.2],
                ["Maritime Pine", 520, 85, 3.8],
                ["Silver Birch", 640, 110, 5.1],
                ["European Oak", 720, 125, 6.8],
                ["Norway Spruce", 460, 74, 3.2],
            ],
        },
        "figure_data": "Alpine Larch: 590 kg/m3, 98 MPa, 4.2 kN; Maritime Pine: 520, 85, 3.8; Silver Birch: 640, 110, 5.1; European Oak: 720, 125, 6.8; Norway Spruce: 460, 74, 3.2.",
        "model_answer": "The table presents physical and mechanical metrics for five structural timber species. European Oak recorded the highest values across all categories, with a density of 720 kg/m3, bending strength of 125 MPa, and hardness of 6.8 kN. Norway Spruce exhibited the lowest values, with a density of only 460 kg/m3 and strength of 74 MPa. In conclusion, timber density correlates directly with bending resistance and surface hardness.",
        "explanation": "Identify European Oak as highest across all metrics, Norway Spruce as lowest, and note the density-strength relationship.",
    },
    {
        "title": "Regional Agricultural Harvest Output (Tonnes/Hectare)",
        "kind": "table",
        "data": {
            "title": "Agricultural Crop Yields",
            "headers": ["Crop Variety", "2020 Yield", "2021 Yield", "2022 Yield", "2023 Yield"],
            "rows": [
                ["Winter Wheat", 7.2, 7.5, 6.8, 8.1],
                ["Spring Barley", 5.8, 6.1, 5.9, 6.4],
                ["Grain Maize", 9.4, 9.8, 8.9, 10.5],
                ["Oilseed Rape", 3.2, 3.4, 3.1, 3.6],
                ["Field Peas", 4.1, 4.3, 3.9, 4.5],
            ],
        },
        "figure_data": "Winter Wheat: 7.2, 7.5, 6.8, 8.1; Spring Barley: 5.8, 6.1, 5.9, 6.4; Grain Maize: 9.4, 9.8, 8.9, 10.5; Oilseed: 3.2, 3.4, 3.1, 3.6; Field Peas: 4.1, 4.3, 3.9, 4.5.",
        "model_answer": "The table tracks crop yields in tonnes per hectare across five agricultural crops from 2020 to 2023 in a farming region. Grain maize achieved the highest yields in every season, peaking at 10.5 tonnes per hectare in 2023. Oilseed rape registered the lowest yield throughout, ranging between 3.1 and 3.6 tonnes. In conclusion, all crops experienced a slight dip in 2022 before rebounding to four-year highs in 2023.",
        "explanation": "Compare the highest crop (Maize) and lowest (Oilseed), note the shared 2022 dip, and conclude with the 2023 recovery.",
    },
    {
        "title": "Municipal Meteorological Sensor Annual Summary",
        "kind": "table",
        "data": {
            "title": "Meteorological Sensor Metrics",
            "headers": ["Station Site", "Mean Temp (C)", "Annual Rainfall (mm)", "Wind Gust Max (km/h)"],
            "rows": [
                ["Valley Basin", 14.2, 620, 68],
                ["Coastal Bluff", 12.8, 890, 112],
                ["Mountain Ridge", 7.5, 1420, 138],
                ["Urban Core", 15.6, 580, 72],
                ["Forest Reserve", 11.4, 980, 84],
            ],
        },
        "figure_data": "Valley: 14.2C, 620mm, 68 km/h; Coastal: 12.8C, 890mm, 112 km/h; Mountain: 7.5C, 1420mm, 138 km/h; Urban: 15.6C, 580mm, 72 km/h; Forest: 11.4C, 980mm, 84 km/h.",
        "model_answer": "The table details annual meteorological readings across five monitoring stations in a region. The Urban Core recorded the highest mean temperature at 15.6 degrees Celsius, whereas the Mountain Ridge was the coldest at 7.5 degrees. However, the mountain site experienced the highest precipitation at 1,420 millimetres and strongest wind gusts of 138 km/h. To conclude, altitude and topography strongly influence local climate patterns.",
        "explanation": "Contrast temperature extremes between urban and mountain sites, highlight mountain precipitation/wind, and summarize.",
    },
    {
        "title": "Public swimming pools in Ashford",
        "kind": "table",
        "data": {
            "title": "Public Swimming Pools in Ashford",
            "headers": ["Pool", "Length (m)", "Adult Ticket ($)", "Weekly Visitors"],
            "rows": [
                ["Riverside", 50, "6.50", 4200],
                ["Lakeside", 33, "5.00", 3100],
                ["Hillview", 25, "4.00", 2600],
                ["Northgate", 25, "3.50", 1900],
            ],
        },
        "figure_data": "Riverside: 50 m, $6.50, 4,200 visitors; Lakeside: 33 m, $5.00, 3,100 visitors; Hillview: 25 m, $4.00, 2,600 visitors; Northgate: 25 m, $3.50, 1,900 visitors.",
        "model_answer": "The table compares four public swimming pools by length, ticket price and weekly visitors. Riverside is the largest pool at 50 metres, and although it is the most expensive at 6 dollars 50, it attracts the most visitors, 4,200 a week. Lakeside comes second on all three measures. Hillview and Northgate are both 25 metres long, but Northgate is the cheapest and has the fewest visitors, 1,900. Overall, the larger pools attract more swimmers despite their higher prices.",
        "explanation": "Compare the pools on each column, pick out Riverside as the largest, dearest and busiest, and conclude that size matters more than price.",
    },
    {
        "title": "Hospital Urgent Care Admission and Treatment Metrics",
        "kind": "table",
        "data": {
            "title": "Urgent Care Treatment Metrics",
            "headers": ["Triage Category", "Target Wait (min)", "Actual Wait (min)", "Admission Rate (%)"],
            "rows": [
                ["Immediate Resuscitation", 0, 1, 98],
                ["Very Urgent", 10, 14, 76],
                ["Urgent Care", 30, 42, 54],
                ["Standard Assessment", 60, 85, 28],
                ["Non-Urgent Clinic", 120, 145, 12],
            ],
        },
        "figure_data": "Resuscitation: Target 0m, Actual 1m, 98% admit; Very Urgent: 10m, 14m, 76%; Urgent: 30m, 42m, 54%; Standard: 60m, 85m, 28%; Non-Urgent: 120m, 145m, 12%.",
        "model_answer": "The table compares target versus actual wait times and admission rates across five triage categories at a hospital. Immediate resuscitation cases were seen almost instantly within one minute, resulting in a 98 percent admission rate. For lower-priority categories, actual waits exceeded targets, reaching 145 minutes for non-urgent cases with only 12 percent admitted. Overall, clinical acuity dictates both response speed and hospital admission probability.",
        "explanation": "Compare the top priority category with the non-urgent category, note wait time target overruns, and summarize clinical acuity.",
    },
    {
        "title": "Municipal Fast Charger Network Performance Data",
        "kind": "table",
        "data": {
            "title": "Fast Charger Performance Data",
            "headers": ["Hub Location", "Charger Plugs", "Power Rating (kW)", "Daily Sessions"],
            "rows": [
                ["Central Highway", 16, 350, 142],
                ["Airport Terminal", 12, 150, 98],
                ["Downtown Mall", 8, 150, 74],
                ["West Suburb Hub", 6, 75, 42],
                ["East Freight Depot", 10, 350, 88],
            ],
        },
        "figure_data": "Highway: 16 plugs, 350 kW, 142 sessions; Airport: 12 plugs, 150 kW, 98 sessions; Downtown: 8 plugs, 150 kW, 74 sessions; Suburb: 6 plugs, 75 kW, 42 sessions; Freight: 10 plugs, 350 kW, 88 sessions.",
        "model_answer": "The table summarizes operational specifications and daily utilization across five electric vehicle charging hubs in a district. The Central Highway hub provides the greatest capacity, featuring 16 ultra-fast 350 kW plugs and accommodating 142 daily charging sessions. The West Suburb hub is the smallest, with 6 chargers and 42 daily sessions. To conclude, charging infrastructure capacity and daily throughput are highest along major transit corridors.",
        "explanation": "Highlight Central Highway as the largest/most utilized hub, contrast with West Suburb, and summarize transit corridor demand.",
    },
    {
        "title": "Secondary School Extracurricular Participation Statistics",
        "kind": "table",
        "data": {
            "title": "Extracurricular Participation",
            "headers": ["Activity Society", "Active Members", "Weekly Hours", "Annual Budget ($)"],
            "rows": [
                ["Debating Society", 48, 4.0, 3200],
                ["Robotics Club", 36, 6.5, 7800],
                ["School Orchestra", 62, 5.0, 6500],
                ["Athletics Squad", 85, 8.0, 9200],
                ["Environmental Group", 42, 3.0, 2400],
            ],
        },
        "figure_data": "Debating: 48 members, 4.0h, $3200; Robotics: 36 members, 6.5h, $7800; Orchestra: 62 members, 5.0h, $6500; Athletics: 85 members, 8.0h, $9200; Environmental: 42 members, 3.0h, $2400.",
        "model_answer": "The table outlines membership, weekly commitment, and budget allocation for five extracurricular clubs at a secondary school. The Athletics Squad commands the largest membership of 85 students, greatest weekly dedication of 8 hours, and highest annual budget of 9,200 dollars. Conversely, the Environmental Group operates on the lowest budget of 2,400 dollars with 3 weekly hours. In summary, athletics and robotics demand the greatest institutional investment.",
        "explanation": "Identify Athletics as the highest across all metrics, contrast with the Environmental Group, and summarize resource distribution.",
    },
    {
        "title": "Coastal Ferry Route Performance and Patronage",
        "kind": "table",
        "data": {
            "title": "Ferry Route Performance Data",
            "headers": ["Ferry Route", "Crossing Time (min)", "Daily Sailings", "Annual Passengers (k)"],
            "rows": [
                ["Harbour Express", 18, 36, 820],
                ["Island Connector", 45, 14, 340],
                ["Peninsula Shuttle", 25, 24, 510],
                ["Southern Sound", 65, 8, 180],
                ["Eastern Inlet", 30, 20, 420],
            ],
        },
        "figure_data": "Harbour: 18m, 36 sailings, 820k pass; Island: 45m, 14 sailings, 340k; Peninsula: 25m, 24 sailings, 510k; Southern: 65m, 8 sailings, 180k; Eastern: 30m, 20 sailings, 420k.",
        "model_answer": "The table details crossing times, service frequencies, and passenger volumes across five coastal ferry corridors in a bay. Harbour Express is the busiest service, providing 36 daily departures and carrying 820,000 passengers annually over a swift 18-minute journey. Southern Sound is the longest and least frequented route, taking 65 minutes with only 8 daily sailings and 180,000 riders. Overall, shorter commuter routes generate significantly higher service frequencies and patronage.",
        "explanation": "Highlight Harbour Express as the highest frequency and patronage route, contrast with Southern Sound, and summarize.",
    },
    {
        "title": "Municipal Potable Water Quality Laboratory Readings",
        "kind": "table",
        "data": {
            "title": "Water Quality Laboratory Readings",
            "headers": ["Sampling Zone", "pH Value", "Turbidity (NTU)", "Residual Chlorine (mg/L)"],
            "rows": [
                ["Treatment Outlet", 7.4, 0.22, 1.2],
                ["North Reservoir", 7.3, 0.35, 0.9],
                ["Downtown Mains", 7.2, 0.42, 0.7],
                ["South Terminal", 7.1, 0.58, 0.5],
                ["Rural Extension", 7.0, 0.72, 0.4],
            ],
        },
        "figure_data": "Treatment: pH 7.4, 0.22 NTU, 1.2 mg/L; North: 7.3, 0.35, 0.9; Downtown: 7.2, 0.42, 0.7; South: 7.1, 0.58, 0.5; Rural: 7.0, 0.72, 0.4.",
        "model_answer": "The table records water chemistry and clarity metrics at five distribution points in a water network. Water leaving the treatment plant recorded optimal quality, with a neutral pH of 7.4, low turbidity of 0.22 NTU, and residual chlorine of 1.2 mg/L. As water traveled toward the Rural Extension, turbidity increased to 0.72 NTU while chlorine disinfectant dropped to 0.4 mg/L. In summary, water clarity and disinfectant levels gradually decline with distribution distance.",
        "explanation": "Note baseline values at Treatment Outlet, trace the gradient toward Rural Extension, and summarize the distance degradation effect.",
    },
]

# 10 Process diagrams
PROCESSES = [
    {
        "title": "How wooden furniture is made in a workshop",
        "kind": "process",
        "data": {
            "title": "Making Wooden Furniture",
            "steps": ["Timber dried", "Parts cut and shaped", "Joints glued", "Surfaces sanded", "Varnish applied"],
        },
        "figure_data": "Stage 1: Timber dried; Stage 2: Parts cut and shaped; Stage 3: Joints glued; Stage 4: Surfaces sanded; Stage 5: Varnish applied.",
        "model_answer": "The diagram shows the five stages of making wooden furniture in a workshop. First, the timber is dried, and then the parts are cut to size and shaped. Next, the joints are glued together. After that, the surfaces are sanded smooth, and finally varnish is applied to protect and finish the piece. Overall, the process turns raw timber into finished furniture in five straightforward steps.",
        "explanation": "Describe the five stages in order, using sequence words such as first, next, after that and finally.",
    },
    {
        "title": "Manufacturing Stages of Recycled Cardboard Packaging",
        "kind": "process",
        "data": {
            "title": "Cardboard Recycling Flow",
            "steps": ["Collection Sorting", "Hydraulic Pulping", "De-Inking Wash", "Sheet Pressing", "Roll Packaging"],
        },
        "figure_data": "Stage 1: Collection Sorting; Stage 2: Hydraulic Pulping; Stage 3: De-Inking Wash; Stage 4: Sheet Pressing; Stage 5: Roll Packaging.",
        "model_answer": "The process diagram shows the stages involved in converting post-consumer waste into recycled cardboard packaging. First, recovered paper is collected and sorted, then blended with water during hydraulic pulping. After de-inking and washing removes dyes, the refined slurry is pressed and dried into continuous paper sheets. The final step involves winding the material into finished industrial rolls. To summarize, the closed-loop process efficiently turns discarded paper into commercial packaging.",
        "explanation": "Describe each stage from collection to finished roll packaging, and conclude on the circular manufacturing cycle.",
    },
    {
        "title": "Municipal Glass Bottle Sorting and Cullet Processing Sequence",
        "kind": "process",
        "data": {
            "title": "Glass Bottle Recycling Cycle",
            "steps": ["Mechanical Crushing", "Color Sorting", "Contaminant Removal", "Furnace Smelting", "Container Molding"],
        },
        "figure_data": "Stage 1: Mechanical Crushing; Stage 2: Color Sorting; Stage 3: Contaminant Removal; Stage 4: Furnace Smelting; Stage 5: Container Molding.",
        "model_answer": "This diagram illustrates the industrial recycling sequence for container glass in a municipal processing facility. Collected glass bottles are first mechanically crushed into coarse cullet. High-speed optical sorters separate cullet by color before air suction and magnets remove ceramic and metallic contaminants. The clean cullet is smelted in a high-temperature furnace at 1500 degrees Celsius and molded into new bottles. Overall, recycling cullet significantly reduces smelting furnace energy consumption.",
        "explanation": "Explain mechanical crushing, optical color sorting, contaminant extraction, furnace melting, and container remolding.",
    },
    {
        "title": "Sedimentary Rock Stratification and Fossil Excavation Sequence",
        "kind": "process",
        "data": {
            "title": "Fossil Stratification Cycle",
            "steps": ["Organism Deposition", "Sediment Burial", "Lithification", "Tectonic Uplift", "Surface Erosion"],
        },
        "figure_data": "Stage 1: Organism Deposition; Stage 2: Sediment Burial; Stage 3: Lithification; Stage 4: Tectonic Uplift; Stage 5: Surface Erosion.",
        "model_answer": "The process diagram depicts the chronological stages of fossil preservation and geological exposure. An organism is initially deposited in an aquatic basin, followed by rapid burial beneath mineral sediments. Over millions of years, compaction and mineral cementing trigger lithification into solid rock. Tectonic forces subsequently uplift sedimentary strata, and surface weathering eventually exposes fossilized remains. In summary, fossil discovery requires a rare succession of burial, rock formation, and tectonic exposure.",
        "explanation": "Outline organism deposition, mineral burial, lithification, uplift, and erosional discovery.",
    },
    {
        "title": "Urban Rainwater Biofiltration Basin Sequence",
        "kind": "process",
        "data": {
            "title": "Rainwater Biofiltration Flow",
            "steps": ["Catchment Inflow", "Sediment Trap", "Engineered Soil", "Root Uptake", "Filtered Outflow"],
        },
        "figure_data": "Stage 1: Catchment Inflow; Stage 2: Sediment Trap; Stage 3: Engineered Soil; Stage 4: Root Uptake; Stage 5: Filtered Outflow.",
        "model_answer": "The flow diagram details the operational stages of an urban stormwater biofiltration swale. Surface runoff flows into the catchment inlet, where coarse debris settles in a forebay sediment trap. The water then percolates downward through engineered biofiltration soil while wetland plant roots absorb dissolved nitrogen and phosphorus. Finally, treated water enters perforated underdrains for clean discharge into local streams. To conclude, biofiltration naturally purifies urban runoff without chemical treatment.",
        "explanation": "Describe inflow, sedimentation, soil percolation, plant biological nutrient uptake, and clean discharge.",
    },
    {
        "title": "Industrial Flue Gas Desulfurization Sequence",
        "kind": "process",
        "data": {
            "title": "Flue Gas Desulfurization Flow",
            "steps": ["Exhaust Inflow", "Limestone Spray", "Chemical Reaction", "Gypsum Dewatering", "Clean Release"],
        },
        "figure_data": "Stage 1: Exhaust Inflow; Stage 2: Limestone Spray; Stage 3: Chemical Reaction; Stage 4: Gypsum Dewatering; Stage 5: Clean Release.",
        "model_answer": "The diagram charts the engineering stages of wet limestone flue gas desulfurization at a thermal power facility. Combustion exhaust gas is directed into a vertical scrubber tower where a limestone slurry spray contacts the flue stream. Sulfur dioxide reacts chemically with calcium carbonate and oxygen to synthesize synthetic gypsum. Hydrocyclones dewater the commercial-grade gypsum byproduct before cleansed gas vents into the atmosphere. In summary, the wet scrubbing process neutralizes acid gas emissions while manufacturing usable gypsum wallboard material.",
        "explanation": "Explain exhaust entry, limestone scrubbing, chemical synthesis of gypsum, dewatering, and cleansed gas release.",
    },
    {
        "title": "Municipal Composting Thermophilic Processing Stages",
        "kind": "process",
        "data": {
            "title": "Composting Processing Stages",
            "steps": ["Shredding Blending", "Windrow Piling", "Thermophilic Heat", "Screen Sifting", "Mature Curing"],
        },
        "figure_data": "Stage 1: Shredding Blending; Stage 2: Windrow Piling; Stage 3: Thermophilic Heat; Stage 4: Screen Sifting; Stage 5: Mature Curing.",
        "model_answer": "The flow chart illustrates the five primary steps in industrial municipal composting. Organic yard and food waste is first shredded and blended to balance carbon and nitrogen ratios. The material is arranged into aerated windrow piles where microbial activity generates thermophilic heat exceeding 55 degrees Celsius to destroy weed seeds and pathogens. After screen sifting removes oversize debris, compost cures into mature organic fertilizer. In conclusion, controlled biological decomposition transforms organic waste into valuable soil conditioner.",
        "explanation": "Describe shredding, windrow formation, microbial heat sanitation, sifting, and final curing.",
    },
    {
        "title": "Lithium Battery Hydrometallurgical Recycling Flow",
        "kind": "process",
        "data": {
            "title": "Battery Recycling Sequence",
            "steps": ["Discharge Shredding", "Magnetic Separation", "Acid Leaching", "Precipitation", "Cathode Synthesis"],
        },
        "figure_data": "Stage 1: Discharge Shredding; Stage 2: Magnetic Separation; Stage 3: Acid Leaching; Stage 4: Precipitation; Stage 5: Cathode Synthesis.",
        "model_answer": "This diagram illustrates the hydrometallurgical recycling process for spent lithium-ion battery cells. Spent batteries are completely discharged and shredded in an inert environment, followed by magnetic separation to recover steel casings. The remaining black mass undergoes acid leaching to dissolve critical metals into solution. Selective chemical precipitation isolates lithium, cobalt, and nickel salts, which are finally resynthesized into fresh battery cathode material. Overall, hydrometallurgical recycling recovers over ninety percent of critical battery minerals.",
        "explanation": "Trace discharging, shredding, magnetic separation, chemical acid leaching, and cathode remanufacturing.",
    },
    {
        "title": "Solar Photovoltaic Silicon Ingot Fabrication",
        "kind": "process",
        "data": {
            "title": "Silicon Ingot Fabrication",
            "steps": ["Quartz Reduction", "Polysilicon Refining", "Crucible Melting", "Crystal Pulling", "Wafer Slicing"],
        },
        "figure_data": "Stage 1: Quartz Reduction; Stage 2: Polysilicon Refining; Stage 3: Crucible Melting; Stage 4: Crystal Pulling; Stage 5: Wafer Slicing.",
        "model_answer": "The diagram displays the manufacturing sequence for producing monocrystalline solar silicon wafers. Quartz sand is reduced in an electric arc furnace to obtain metallurgical silicon, which is chemically refined into hyper-pure polysilicon. The refined material is melted inside a quartz crucible where the Czochralski process pulls a cylindrical monocrystalline silicon ingot. Finally, precision diamond wire saws slice the ingot into ultra-thin solar wafers. To summarize, complex thermal and chemical refining converts raw quartz into solar semiconductors.",
        "explanation": "Describe quartz reduction, polysilicon chemical purification, crystal growth, and diamond wire slicing.",
    },
    {
        "title": "Municipal Effluent Tertiary Treatment and Nutrient Removal",
        "kind": "process",
        "data": {
            "title": "Effluent Tertiary Treatment",
            "steps": ["Secondary Effluent", "Coagulant Dosing", "Disc Filtration", "UV Disinfection", "River Discharge"],
        },
        "figure_data": "Stage 1: Secondary Effluent; Stage 2: Coagulant Dosing; Stage 3: Disc Filtration; Stage 4: UV Disinfection; Stage 5: River Discharge.",
        "model_answer": "The flow diagram illustrates tertiary nutrient removal at a municipal wastewater treatment facility. Clarified secondary effluent is dosed with iron coagulant to precipitate dissolved phosphorus. The mixture passes through cloth disc filters to capture microscopic flocs, followed by ultraviolet light irradiation that destroys resistant viruses and bacteria. The purified water is safely returned to the receiving river. In conclusion, tertiary treatment effectively protects aquatic ecosystems by stripping remaining nutrients and pathogens.",
        "explanation": "Explain secondary effluent intake, phosphorus chemical dosing, disc filtration, UV disinfection, and safe environmental discharge.",
    },
]


def get_di_items():
    items = []
    # 10 bars
    for b in BARS:
        svg = bar_svg(b["data"]["title"], b["data"]["categories"], b["data"]["series"], b["data"]["unit"])
        items.append({
            "title": b["title"],
            "image_url": data_uri(svg),
            "figure_data": b["figure_data"],
            "model_answer": b["model_answer"],
            "explanation": b["explanation"],
        })
    # 10 lines
    for l in LINES:
        svg = line_svg(l["data"]["title"], l["data"]["categories"], l["data"]["series"], l["data"]["unit"])
        items.append({
            "title": l["title"],
            "image_url": data_uri(svg),
            "figure_data": l["figure_data"],
            "model_answer": l["model_answer"],
            "explanation": l["explanation"],
        })
    # 10 pies
    for p in PIES:
        svg = pie_svg(p["data"]["title"], p["data"]["slices"], p["data"]["unit"])
        items.append({
            "title": p["title"],
            "image_url": data_uri(svg),
            "figure_data": p["figure_data"],
            "model_answer": p["model_answer"],
            "explanation": p["explanation"],
        })
    # 10 tables
    for t in TABLES:
        svg = table_svg(t["data"]["title"], t["data"]["headers"], t["data"]["rows"])
        items.append({
            "title": t["title"],
            "image_url": data_uri(svg),
            "figure_data": t["figure_data"],
            "model_answer": t["model_answer"],
            "explanation": t["explanation"],
        })
    # 10 processes
    for pr in PROCESSES:
        svg = process_svg(pr["data"]["title"], pr["data"]["steps"])
        items.append({
            "title": pr["title"],
            "image_url": data_uri(svg),
            "figure_data": pr["figure_data"],
            "model_answer": pr["model_answer"],
            "explanation": pr["explanation"],
        })
    return items


assert len(get_di_items()) == 50
