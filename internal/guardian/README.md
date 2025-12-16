# Guardian Slice

This slice owns the guardrails that keep the mesh runtime operating within policy.
It will eventually host enforcement logic, policy evaluation, and safety monitors specific to guardian capabilities.

Belongs here:
- Guardian-specific orchestration and coordination logic
- Interfaces or structs tied directly to guardian responsibilities

Does not belong here:
- Shared utilities needed by multiple slices (those live under pkg when justified)
- Agent implementations or unrelated runtime features
