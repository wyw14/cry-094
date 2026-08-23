CREATE TABLE users (
    id uuid PRIMARY KEY,
    email text NOT NULL UNIQUE,
    password_hash bytea NOT NULL,
    active boolean NOT NULL DEFAULT true,
    created_at timestamptz NOT NULL,
    version bigint NOT NULL DEFAULT 1 CHECK (version > 0)
);

CREATE TABLE refresh_tokens (
    id uuid PRIMARY KEY,
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    digest char(64) NOT NULL UNIQUE,
    expires_at timestamptz NOT NULL,
    revoked_at timestamptz,
    created_at timestamptz NOT NULL,
    CHECK (expires_at > created_at)
);
CREATE INDEX refresh_tokens_user_active_idx ON refresh_tokens(user_id, expires_at) WHERE revoked_at IS NULL;

CREATE TABLE teams (
    id uuid PRIMARY KEY,
    name text NOT NULL CHECK (char_length(name) >= 3),
    members jsonb NOT NULL DEFAULT '{}'::jsonb,
    version bigint NOT NULL DEFAULT 1 CHECK (version > 0),
    created_at timestamptz NOT NULL
);

CREATE TABLE script_libraries (
    id uuid PRIMARY KEY,
    team_id uuid NOT NULL REFERENCES teams(id) ON DELETE RESTRICT,
    name text NOT NULL,
    tags jsonb NOT NULL DEFAULT '{}'::jsonb,
    version bigint NOT NULL DEFAULT 1 CHECK (version > 0),
    created_at timestamptz NOT NULL,
    UNIQUE(team_id, name)
);

CREATE TABLE script_artifacts (
    id uuid PRIMARY KEY,
    library_id uuid NOT NULL REFERENCES script_libraries(id) ON DELETE RESTRICT,
    number integer NOT NULL CHECK (number > 0),
    filename text NOT NULL CHECK (filename ~ '\.(sh|ps1)$'),
    shell text NOT NULL CHECK (shell IN ('shell', 'powershell')),
    encoding text NOT NULL CHECK (encoding IN ('utf-8', 'utf-16le')),
    digest char(64) NOT NULL,
    storage_key text NOT NULL UNIQUE,
    size_bytes bigint NOT NULL CHECK (size_bytes BETWEEN 1 AND 1048576),
    uploaded_by uuid NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    uploaded_at timestamptz NOT NULL,
    parser_build text NOT NULL DEFAULT '',
    version bigint NOT NULL DEFAULT 1 CHECK (version > 0),
    UNIQUE(library_id, number),
    UNIQUE(library_id, digest)
);
CREATE INDEX script_artifacts_cursor_idx ON script_artifacts(library_id, uploaded_at, id);

CREATE TABLE analyses (
    id uuid PRIMARY KEY,
    library_id uuid NOT NULL REFERENCES script_libraries(id) ON DELETE RESTRICT,
    graph jsonb NOT NULL,
    execution_order text[] NOT NULL,
    issues jsonb NOT NULL,
    capability_checks jsonb NOT NULL,
    status text NOT NULL CHECK (status IN ('ready', 'blocked')),
    artifact_hash char(64) NOT NULL,
    parser_build text NOT NULL,
    completed_at timestamptz NOT NULL,
    UNIQUE(library_id, artifact_hash, parser_build)
);

CREATE TABLE precheck_plans (
    id uuid PRIMARY KEY,
    analysis_id uuid NOT NULL REFERENCES analyses(id) ON DELETE RESTRICT,
    team_id uuid NOT NULL REFERENCES teams(id) ON DELETE RESTRICT,
    state text NOT NULL CHECK (state IN ('draft','in_review','approved','rejected','signed')),
    steps jsonb NOT NULL,
    created_by uuid NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    reviewed_by uuid REFERENCES users(id) ON DELETE RESTRICT,
    review_reason text,
    content_digest char(64) NOT NULL,
    signature text,
    signer_key_id text,
    created_at timestamptz NOT NULL,
    reviewed_at timestamptz,
    signed_at timestamptz,
    version bigint NOT NULL DEFAULT 1 CHECK (version > 0),
    CHECK (created_by IS DISTINCT FROM reviewed_by),
    CHECK ((state IN ('approved','rejected','signed')) = (reviewed_by IS NOT NULL)),
    CHECK ((state = 'signed') = (signature IS NOT NULL AND signer_key_id IS NOT NULL))
);

CREATE TABLE audit_events (
    sequence bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    id uuid NOT NULL UNIQUE,
    team_id uuid NOT NULL REFERENCES teams(id) ON DELETE RESTRICT,
    actor_id uuid NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    source text NOT NULL,
    action text NOT NULL,
    resource_type text NOT NULL,
    resource_id text NOT NULL,
    before_data jsonb NOT NULL,
    after_data jsonb NOT NULL,
    reason text NOT NULL,
    occurred_at timestamptz NOT NULL
);
CREATE INDEX audit_events_resource_idx ON audit_events(team_id, resource_type, resource_id, sequence);

CREATE TABLE idempotency_keys (
    team_id uuid NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    operation text NOT NULL,
    key text NOT NULL,
    request_digest char(64) NOT NULL,
    response_status integer,
    response_body jsonb,
    expires_at timestamptz NOT NULL,
    PRIMARY KEY(team_id, operation, key)
);

CREATE TABLE outbox_events (
    id uuid PRIMARY KEY,
    topic text NOT NULL,
    payload jsonb NOT NULL,
    attempts integer NOT NULL DEFAULT 0 CHECK (attempts >= 0),
    next_attempt_at timestamptz NOT NULL,
    delivered_at timestamptz,
    dead_lettered_at timestamptz,
    created_at timestamptz NOT NULL
);
CREATE INDEX outbox_pending_idx ON outbox_events(next_attempt_at, id) WHERE delivered_at IS NULL AND dead_lettered_at IS NULL;
