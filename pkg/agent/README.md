# Agent Contracts

The `pkg/agent` package is reserved for shared contracts that describe how agents interact with the mesh runtime.
Only lightweight, capability-agnostic types or interfaces belong here so multiple slices can depend on the same definitions without owning them.

Does not belong here: slice-specific implementations, helpers, or logic.
