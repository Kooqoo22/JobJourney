BEGIN;

TRUNCATE users, job_applications, application_events, application_documents, email_tokens, refresh_tokens RESTART IDENTITY CASCADE;

INSERT INTO users (email, password_hash, auth_provider, full_name, avatar_url, timezone, is_verified, is_banned, banned_at, ban_reason, role, last_login_at, created_at, updated_at, deleted_at)
VALUES
    ('verified@example.com',   crypt('Password123', gen_salt('bf', 10)), 'local',  'Vera Verified', NULL, 'Asia/Jakarta',    TRUE,  FALSE, NULL,                     NULL,                    'user',  NOW() - INTERVAL '2 days',  NOW() - INTERVAL '200 days', NOW() - INTERVAL '2 days',  NULL),
    ('unverified@example.com', crypt('Password123', gen_salt('bf', 10)), 'local',  'Uma Unverified', NULL, 'Asia/Jakarta',   FALSE, FALSE, NULL,                     NULL,                    'user',  NULL,                       NOW() - INTERVAL '1 day',    NOW() - INTERVAL '1 day',   NULL),
    ('banned@example.com',     crypt('Password123', gen_salt('bf', 10)), 'local',  'Ben Banned', NULL, 'Asia/Jakarta',       TRUE,  TRUE,  NOW() - INTERVAL '5 days', 'Violation of terms',    'user',  NOW() - INTERVAL '10 days', NOW() - INTERVAL '150 days', NOW() - INTERVAL '5 days',  NULL),
    ('deleted@example.com',    crypt('Password123', gen_salt('bf', 10)), 'local',  'Dan Deleted', NULL, 'Asia/Jakarta',      TRUE,  FALSE, NULL,                     NULL,                    'user',  NOW() - INTERVAL '30 days', NOW() - INTERVAL '120 days', NOW() - INTERVAL '30 days', NOW() - INTERVAL '20 days'),
    ('admin@example.com',      crypt('Password123', gen_salt('bf', 10)), 'local',  'Ada Admin', NULL, 'Asia/Jakarta',        TRUE,  FALSE, NULL,                     NULL,                    'admin', NOW() - INTERVAL '1 hour',  NOW() - INTERVAL '300 days', NOW() - INTERVAL '1 hour',  NULL),
    ('google@example.com',     NULL,                                      'google', 'Gina Google', 'https://i.pravatar.cc/150?u=google', 'America/New_York', TRUE, FALSE, NULL,          NULL,                    'user',  NOW() - INTERVAL '3 days',  NOW() - INTERVAL '90 days',  NOW() - INTERVAL '3 days',  NULL),
    ('reset@example.com',      crypt('Password123', gen_salt('bf', 10)), 'local',  'Rita Reset', NULL, 'Asia/Jakarta',       TRUE,  FALSE, NULL,                     NULL,                    'user',  NOW() - INTERVAL '7 days',  NOW() - INTERVAL '60 days',  NOW() - INTERVAL '7 days',  NULL),
    ('power@example.com',      crypt('Password123', gen_salt('bf', 10)), 'local',  'Pete Power', NULL, 'Asia/Jakarta',       TRUE,  FALSE, NULL,                     NULL,                    'user',  NOW() - INTERVAL '2 hours', NOW() - INTERVAL '365 days', NOW() - INTERVAL '2 hours', NULL);

INSERT INTO users (email, password_hash, auth_provider, full_name, avatar_url, timezone, is_verified, is_banned, banned_at, ban_reason, role, last_login_at, created_at, updated_at, deleted_at)
SELECT
    'user' || g || '@example.com',
    CASE WHEN prov = 'google' THEN NULL ELSE crypt('Password123', gen_salt('bf', 10)) END,
    prov::auth_provider,
    fname || ' ' || lname,
    CASE WHEN prov = 'google' THEN 'https://i.pravatar.cc/150?u=' || g ELSE NULL END,
    tz,
    NOT (b BETWEEN 65 AND 79),
    (b BETWEEN 80 AND 90),
    CASE WHEN (b BETWEEN 80 AND 90) THEN created + (random() * INTERVAL '30 days') ELSE NULL END,
    CASE WHEN (b BETWEEN 80 AND 90) THEN (ARRAY['Spam','Abuse','Violation of terms','Fraudulent activity','Multiple accounts'])[(1 + floor(random() * 5))::int] ELSE NULL END,
    'user',
    CASE WHEN NOT (b BETWEEN 65 AND 79) AND NOT (b BETWEEN 80 AND 90) THEN created + (random() * INTERVAL '60 days') ELSE NULL END,
    created,
    created + (random() * INTERVAL '90 days'),
    CASE WHEN b >= 91 THEN created + (random() * INTERVAL '100 days') ELSE NULL END
FROM (
    SELECT
        g,
        floor(random() * 100)::int AS b,
        (ARRAY['local','local','local','local','local','local','local','google'])[(1 + floor(random() * 8))::int] AS prov,
        (NOW() - (random() * INTERVAL '700 days')) AS created,
        (ARRAY['Asia/Jakarta','Asia/Makassar','America/New_York','Europe/London','Asia/Singapore','Australia/Sydney','Asia/Tokyo'])[(1 + floor(random() * 7))::int] AS tz,
        (ARRAY['Andi','Budi','Citra','Dewi','Eka','Fajar','Gita','Hadi','Indah','Joko','Kartika','Lina','Made','Nadia','Oki','Putri','Rian','Sari','Tono','Umar','Vina','Wawan','Yanti','Zaki'])[(1 + floor(random() * 24))::int] AS fname,
        (ARRAY['Santoso','Wijaya','Pratama','Kusuma','Halim','Nugroho','Saputra','Utami','Lestari','Hidayat','Permana','Anggraini','Setiawan','Maulana','Rahman'])[(1 + floor(random() * 15))::int] AS lname
    FROM generate_series(1, 200) AS g
) seed;

INSERT INTO job_applications (user_id, company_name, position_title, job_url, work_arrangement, employment_type, location, source, status, applied_date, salary_min, salary_max, currency, notes, is_archived, created_at, updated_at, deleted_at)
SELECT
    (SELECT id FROM users WHERE email = 'power@example.com'),
    'Company ' || initcap(st),
    'Senior Software Engineer',
    'https://careers.example.com/' || st,
    'remote',
    'full_time',
    'Jakarta',
    'LinkedIn',
    st::application_status,
    CASE WHEN st = 'wishlist' THEN NULL ELSE (NOW() - INTERVAL '60 days')::date END,
    90000, 140000, 'USD',
    'Power user reference application',
    (st IN ('accepted','rejected','withdrawn')),
    NOW() - INTERVAL '60 days',
    NOW() - INTERVAL '10 days',
    NULL
FROM unnest(ARRAY['wishlist','applied','screening','interview','offer','accepted','rejected','withdrawn','ghosted']) AS st;

INSERT INTO job_applications (user_id, company_name, position_title, job_url, work_arrangement, employment_type, location, source, status, applied_date, salary_min, salary_max, currency, notes, is_archived, created_at, updated_at, deleted_at)
VALUES (
    (SELECT id FROM users WHERE email = 'power@example.com'),
    'Deleted Company', 'Backend Engineer', NULL, 'hybrid', 'contract', 'Remote', 'Referral',
    'applied', (NOW() - INTERVAL '80 days')::date, NULL, NULL, NULL, 'Removed by user', FALSE,
    NOW() - INTERVAL '80 days', NOW() - INTERVAL '75 days', NOW() - INTERVAL '70 days'
);

INSERT INTO job_applications (user_id, company_name, position_title, job_url, work_arrangement, employment_type, location, source, status, applied_date, salary_min, salary_max, currency, notes, is_archived, created_at, updated_at, deleted_at)
SELECT
    user_id,
    company_name,
    position_title,
    job_url,
    work_arrangement,
    employment_type,
    location,
    source,
    status,
    CASE WHEN status = 'wishlist' THEN NULL ELSE created::date END,
    CASE WHEN r_sal < 0.6 THEN sal_min ELSE NULL END,
    CASE WHEN r_sal < 0.6 THEN sal_max ELSE NULL END,
    CASE WHEN r_sal < 0.6 THEN currency ELSE NULL END,
    notes,
    (r_arch < 0.15),
    created,
    CASE WHEN r_upd < 0.5 THEN created + (random() * (NOW() - created)) ELSE created END,
    CASE WHEN r_del < 0.10 THEN created + (random() * (NOW() - created)) ELSE NULL END
FROM (
    SELECT
        u.id AS user_id,
        u.created_at + (random() * (NOW() - u.created_at)) AS created,
        (ARRAY['wishlist','applied','applied','screening','interview','interview','offer','accepted','rejected','withdrawn','ghosted'])[(1 + floor(random() * 11))::int]::application_status AS status,
        (ARRAY['Tokopedia','Gojek','Traveloka','Bukalapak','Shopee','Grab','Google','Microsoft','Amazon','Meta','Netflix','Stripe','Atlassian','GitLab','Figma','Notion','Vercel','Datadog','Snowflake','Airbnb'])[(1 + floor(random() * 20))::int] AS company_name,
        (ARRAY['Software Engineer','Senior Software Engineer','Backend Engineer','Frontend Engineer','Full Stack Engineer','Data Engineer','DevOps Engineer','Product Manager','QA Engineer','Engineering Manager','Site Reliability Engineer','Mobile Engineer'])[(1 + floor(random() * 12))::int] AS position_title,
        CASE WHEN random() < 0.7 THEN 'https://jobs.example.com/' || floor(random() * 100000)::int ELSE NULL END AS job_url,
        CASE WHEN random() < 0.9 THEN (ARRAY['remote','onsite','hybrid'])[(1 + floor(random() * 3))::int]::work_arrangement ELSE NULL END AS work_arrangement,
        CASE WHEN random() < 0.9 THEN (ARRAY['full_time','part_time','contract','internship','freelance'])[(1 + floor(random() * 5))::int]::employment_type ELSE NULL END AS employment_type,
        (ARRAY['Jakarta','Bandung','Surabaya','Remote','Singapore','Yogyakarta','Bali'])[(1 + floor(random() * 7))::int] AS location,
        (ARRAY['LinkedIn','Referral','Company Website','Jobstreet','Glints','Indeed','Kalibrr','AngelList'])[(1 + floor(random() * 8))::int] AS source,
        (30000 + floor(random() * 90000))::numeric(14, 2) AS sal_min,
        (120000 + floor(random() * 80000))::numeric(14, 2) AS sal_max,
        (ARRAY['IDR','USD','SGD','EUR'])[(1 + floor(random() * 4))::int] AS currency,
        CASE WHEN random() < 0.4 THEN 'Applied via ' || (ARRAY['LinkedIn','referral','job board'])[(1 + floor(random() * 3))::int] ELSE NULL END AS notes,
        random() AS r_arch,
        random() AS r_del,
        random() AS r_upd,
        random() AS r_sal
    FROM users u
    CROSS JOIN LATERAL generate_series(1, floor(random() * 7)::int) AS n
    WHERE u.is_verified AND NOT u.is_banned AND u.deleted_at IS NULL
) app_seed;

INSERT INTO application_events (application_id, user_id, type, title, event_at, notes, remind_at, reminded_at, status_from, status_to, created_at, updated_at)
SELECT
    application_id,
    user_id,
    type,
    initcap(replace(type::text, '_', ' ')),
    event_at,
    CASE WHEN random() < 0.5 THEN 'Seed note for ' || type::text ELSE NULL END,
    CASE
        WHEN r_remind < 0.35 AND r_reminded < 0.5 THEN NOW() + (INTERVAL '1 day' + random() * INTERVAL '25 days')
        WHEN r_remind < 0.35 THEN event_at + INTERVAL '2 days'
        ELSE NULL
    END,
    CASE
        WHEN r_remind < 0.35 AND r_reminded >= 0.5 THEN event_at + INTERVAL '2 days' + INTERVAL '1 hour'
        ELSE NULL
    END,
    CASE WHEN type = 'status_changed' THEN (ARRAY['applied','screening','interview'])[(1 + floor(random() * 3))::int]::application_status ELSE NULL END,
    CASE WHEN type = 'status_changed' THEN (ARRAY['interview','offer','rejected','ghosted'])[(1 + floor(random() * 4))::int]::application_status ELSE NULL END,
    event_at,
    event_at
FROM (
    SELECT
        a.id AS application_id,
        a.user_id AS user_id,
        (ARRAY['applied','phone_screen','interview','assessment','offer','follow_up','deadline','note','status_changed'])[(1 + floor(random() * 9))::int]::event_type AS type,
        a.created_at + (random() * (NOW() - a.created_at)) AS event_at,
        random() AS r_remind,
        random() AS r_reminded
    FROM job_applications a
    CROSS JOIN LATERAL generate_series(1, floor(random() * 5)::int) AS n
    WHERE a.deleted_at IS NULL
) ev_seed;

INSERT INTO application_documents (application_id, user_id, file_url, file_name, mime_type, size_bytes, created_at, deleted_at)
SELECT
    a.id,
    a.user_id,
    'https://files.example.com/' || encode(gen_random_bytes(8), 'hex') || '.pdf',
    (ARRAY['resume.pdf','cover_letter.pdf','portfolio.pdf','offer_letter.pdf','transcript.pdf','reference.pdf'])[(1 + floor(random() * 6))::int],
    'application/pdf',
    (50000 + floor(random() * 2000000))::bigint,
    a.created_at + (random() * INTERVAL '5 days'),
    CASE WHEN random() < 0.15 THEN a.created_at + (random() * INTERVAL '20 days') ELSE NULL END
FROM job_applications a
CROSS JOIN LATERAL generate_series(1, floor(random() * 3)::int) AS n
WHERE a.deleted_at IS NULL;

INSERT INTO email_tokens (user_id, type, token_hash, expires_at, used_at, created_at)
SELECT u.id, 'verify', encode(gen_random_bytes(16), 'hex'), NOW() + INTERVAL '24 hours', NULL, u.created_at
FROM users u
WHERE u.auth_provider = 'local' AND u.is_verified = FALSE AND u.deleted_at IS NULL;

INSERT INTO email_tokens (user_id, type, token_hash, expires_at, used_at, created_at)
SELECT u.id, 'verify', encode(gen_random_bytes(16), 'hex'), u.created_at + INTERVAL '24 hours', u.created_at + INTERVAL '1 hour', u.created_at
FROM users u
WHERE u.auth_provider = 'local' AND u.is_verified = TRUE AND u.deleted_at IS NULL AND random() < 0.5;

INSERT INTO email_tokens (user_id, type, token_hash, expires_at, used_at, created_at)
SELECT u.id, 'reset', encode(gen_random_bytes(16), 'hex'), NOW() + INTERVAL '1 hour', NULL, NOW() - INTERVAL '10 minutes'
FROM users u
WHERE u.email = 'reset@example.com' OR (u.auth_provider = 'local' AND u.is_verified = TRUE AND u.deleted_at IS NULL AND random() < 0.1);

INSERT INTO email_tokens (user_id, type, token_hash, expires_at, used_at, created_at)
SELECT u.id, 'reset', encode(gen_random_bytes(16), 'hex'), u.created_at + INTERVAL '1 hour', u.created_at + INTERVAL '20 minutes', u.created_at
FROM users u
WHERE u.auth_provider = 'local' AND u.is_verified = TRUE AND u.deleted_at IS NULL AND random() < 0.1;

INSERT INTO refresh_tokens (user_id, token_hash, expires_at, revoked_at, created_at)
SELECT
    u.id,
    encode(gen_random_bytes(16), 'hex'),
    CASE WHEN b < 50 THEN NOW() + INTERVAL '168 hours' WHEN b < 75 THEN NOW() + INTERVAL '100 hours' ELSE NOW() - INTERVAL '10 hours' END,
    CASE WHEN b BETWEEN 50 AND 74 THEN NOW() - INTERVAL '2 hours' ELSE NULL END,
    NOW() - (random() * INTERVAL '5 days')
FROM users u
CROSS JOIN LATERAL generate_series(1, (1 + floor(random() * 3))::int) AS n
CROSS JOIN LATERAL (SELECT floor(random() * 100)::int AS b) bucket
WHERE u.is_verified AND NOT u.is_banned AND u.deleted_at IS NULL;

COMMIT;

SELECT 'users_total' AS metric, count(*) AS count FROM users
UNION ALL SELECT 'users_verified_active', count(*) FROM users WHERE is_verified AND NOT is_banned AND deleted_at IS NULL
UNION ALL SELECT 'users_unverified', count(*) FROM users WHERE NOT is_verified AND deleted_at IS NULL
UNION ALL SELECT 'users_banned', count(*) FROM users WHERE is_banned
UNION ALL SELECT 'users_soft_deleted', count(*) FROM users WHERE deleted_at IS NOT NULL
UNION ALL SELECT 'users_google', count(*) FROM users WHERE auth_provider = 'google'
UNION ALL SELECT 'job_applications_total', count(*) FROM job_applications
UNION ALL SELECT 'job_applications_active', count(*) FROM job_applications WHERE deleted_at IS NULL
UNION ALL SELECT 'job_applications_edited', count(*) FROM job_applications WHERE updated_at > created_at
UNION ALL SELECT 'job_applications_archived', count(*) FROM job_applications WHERE is_archived
UNION ALL SELECT 'job_applications_soft_deleted', count(*) FROM job_applications WHERE deleted_at IS NOT NULL
UNION ALL SELECT 'application_events_total', count(*) FROM application_events
UNION ALL SELECT 'application_events_pending_reminder', count(*) FROM application_events WHERE remind_at IS NOT NULL AND reminded_at IS NULL
UNION ALL SELECT 'application_documents_total', count(*) FROM application_documents
UNION ALL SELECT 'email_tokens_total', count(*) FROM email_tokens
UNION ALL SELECT 'refresh_tokens_active', count(*) FROM refresh_tokens WHERE revoked_at IS NULL AND expires_at > NOW()
UNION ALL SELECT 'refresh_tokens_revoked', count(*) FROM refresh_tokens WHERE revoked_at IS NOT NULL
UNION ALL SELECT 'refresh_tokens_expired', count(*) FROM refresh_tokens WHERE revoked_at IS NULL AND expires_at <= NOW();