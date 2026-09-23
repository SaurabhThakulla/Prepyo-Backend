INSERT INTO users (email, name, role, nepal_region, timezone, referral_code)
VALUES ('admin@prepyo.online', 'Prepyo Admin', 'admin', 'Kathmandu', 'Asia/Kathmandu',
        'PREP-' || upper(substr(md5('admin@prepyo.online'), 1, 6)))
ON CONFLICT (email) DO UPDATE SET role = 'admin';
