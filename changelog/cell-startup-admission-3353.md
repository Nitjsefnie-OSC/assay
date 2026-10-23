### Fixed
- Cell launches and completion no longer recursively sweep retained task history; explicit maintenance retains ownership, evidence and Git protections.
- Managed cache enrollment uses bounded accounting admission, with maintenance inventory outside launch custody and stale concurrent inventory refused.
- Resolved cell telemetry records preparation, scratch admission, final enrollment and confirmed child startup phases.

The existing cadence supervisor now triggers cache maintenance between completed passes and at the next due boundary, preserving unattended cache liveness after builds or accounting age expiry. Interactive mode retains explicit operator clean; startup and Finish remain bounded.
