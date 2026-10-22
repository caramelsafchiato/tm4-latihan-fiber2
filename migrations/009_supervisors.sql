CREATE TABLE IF NOT EXISTS supervisors (
    id SERIAL PRIMARY KEY,
    name VARCHAR(120) NOT NULL,
    quota INTEGER NOT NULL CHECK (quota > 0),
    supervised_count INTEGER NOT NULL DEFAULT 0,
    version INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT supervisors_within_quota CHECK (supervised_count >= 0 AND supervised_count <= quota)
);

-- Constraint student_id sebagai PRIMARY KEY atau UNIQUE menjamin 1 mahasiswa HANYA punya 1 pembimbing
CREATE TABLE IF NOT EXISTS supervisions (
    student_id INTEGER PRIMARY KEY REFERENCES students(id) ON DELETE CASCADE,
    supervisor_id INTEGER NOT NULL REFERENCES supervisors(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS supervisions_supervisor_idx ON supervisions(supervisor_id);

-- Seeder dummy supervisor untuk uji beban
INSERT INTO supervisors (id, name, quota) VALUES (1, 'Dr. Ir. Pembimbing Utama, M.T.', 3)
ON CONFLICT (id) DO NOTHING;

INSERT INTO permissions (name, description) VALUES ('supervision:manage', 'Mengelola dosen pembimbing dan kuota')
ON CONFLICT (name) DO NOTHING;

INSERT INTO role_permissions (role_name, permission_name) VALUES ('admin', 'supervision:manage')
ON CONFLICT DO NOTHING;