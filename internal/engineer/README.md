# Engineer Slice

The engineer slice will own orchestration needed to build or adapt mesh capabilities.
It is responsible for workflows that generate or modify artifacts on behalf of the runtime while respecting slice boundaries.

Belongs here:
- Engineer-facing workflows, request handlers, and state
- Interfaces tightly coupled to the engineer capability

Does not belong here:
- Generic helpers or types (keep those minimal and in pkg if absolutely necessary)
- Guardian or compliance logic
