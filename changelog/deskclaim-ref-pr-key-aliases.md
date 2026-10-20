### Fixed

- `deskclaim-ref` now treats one PR claimed under two key spellings as one claim: `acquire` refuses while another spelling of the same PR and lane is live, settles same-instant races by claim date, `show` prints an `ALIAS-HELD` line after `FREE`, and an unreadable claim set exits 6. `deskdispatch` reports an `ALIAS-HELD` holder as a live holder instead of a claim-acquire error.
