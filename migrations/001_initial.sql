CREATE TABLE workflows (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    status TEXT NOT NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ
);

CREATE TABLE tasks (
    id TEXT NOT NULL,
    workflow_id TEXT NOT NULL REFERENCES workflows(id) ON DELETE CASCADE,

    name TEXT NOT NULL,
    task_type TEXT NOT NULL,
    payload JSONB,
    status TEXT NOT NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,

    PRIMARY KEY (workflow_id, id)
);

CREATE TABLE task_dependencies (
    workflow_id TEXT NOT NULL,
    task_id TEXT NOT NULL,
    depends_on_task_id TEXT NOT NULL,

    PRIMARY KEY (workflow_id, task_id, depends_on_task_id),

    FOREIGN KEY (workflow_id, task_id)
        REFERENCES tasks(workflow_id, id)
        ON DELETE CASCADE,

    FOREIGN KEY (workflow_id, depends_on_task_id)
        REFERENCES tasks(workflow_id, id)
        ON DELETE CASCADE
);

CREATE TABLE task_events (
    id BIGSERIAL PRIMARY KEY,

    workflow_id TEXT NOT NULL
        REFERENCES workflows(id)
        ON DELETE CASCADE,

    task_id TEXT,

    event_type TEXT NOT NULL,
    metadata JSONB,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);