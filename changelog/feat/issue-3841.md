### Fixed
- statusgen's held scan now lets a clean final Evidence entry supersede an earlier entry's prose line that holds named rows ("rows 1-6 and 8 pass; row 7 could-not-check"), when every row it names is a row of that entry and the final entry re-ran every row of that entry clean; prose that names no row, or names rows in a form the scan cannot read, still refuses.
