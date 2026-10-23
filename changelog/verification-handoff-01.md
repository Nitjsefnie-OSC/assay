### Added
- Shared pure issue selection and advisory verification handoff contract in deskcore, with legacy observations retained as unknown.

### Fixed
- Addressed issues no longer enter ordinary scanner creation/reactivation or worker-ready discovery; receiving inboxes and independent gates retain authority.
- Routing labels preserve excluded-label retirement and completed-placeholder cleanup; malformed addresses stay held across full and work-ready views without extra history reads.
- The bounded handoff mutation runner uses its own schema and verifies unmodified named tests before applying any of its eighteen production mutations.
