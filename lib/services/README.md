# Services

This directory contains all business services used by ASRL.

Each service should follow the Service Framework lifecycle.

initialize()

↓

execute()

↓

summary()

↓

cleanup()

Commands should never implement business logic directly.

Commands only invoke services.