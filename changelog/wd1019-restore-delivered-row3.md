### Fixed
- Restored the delivered consumer-routing Verify row in fourteen delivered briefs that medici-finance/assay#3428 had rewritten after delivery, so each brief's Verify table reads as the table its delivery carried and post-merge verification no longer trips the immutable-delivery comparison on that row. Three still-undelivered briefs keep their provisional rows.
