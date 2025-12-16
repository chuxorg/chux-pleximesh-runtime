# Event Contracts

The `pkg/event` package defines shared event schemas that slices use to communicate state changes inside the mesh.
Events stored here must remain minimal, stable, and free of slice-specific behavior so they can be reused safely.

Implementation details, handlers, or slice logic never live in this directory.
