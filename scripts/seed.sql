INSERT INTO users(id,email,password_hash,active,created_at) VALUES
('11111111-1111-4111-8111-111111111111','admin@example.test','$2a$10$7EqJtq98hPqEX7fNZaFWoO5sHnFS5tV3TjTRppq2YTSJwWgL6F5aK',true,'2026-01-01T00:00:00Z'),
('22222222-2222-4222-8222-222222222222','reviewer@example.test','$2a$10$7EqJtq98hPqEX7fNZaFWoO5sHnFS5tV3TjTRppq2YTSJwWgL6F5aK',true,'2026-01-01T00:00:00Z')
ON CONFLICT(id) DO NOTHING;

INSERT INTO teams(id,name,members,version,created_at) VALUES
('aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa','Platform Operations',
 '{"11111111-1111-4111-8111-111111111111":{"user_id":"11111111-1111-4111-8111-111111111111","role":"admin","joined_at":"2026-01-01T00:00:00Z"},"22222222-2222-4222-8222-222222222222":{"user_id":"22222222-2222-4222-8222-222222222222","role":"reviewer","joined_at":"2026-01-01T00:00:00Z"}}',1,'2026-01-01T00:00:00Z')
ON CONFLICT(id) DO NOTHING;
