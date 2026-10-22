-- 1. Hapus akun mhs yang lama agar bersih
DELETE FROM users WHERE username LIKE 'mhs%';

-- 2. Masukkan 60 akun baru
-- Teks $2a$10$vI... di bawah ini adalah enkripsi bcrypt untuk kata "password"
INSERT INTO users (username, email, password, role)
SELECT 
    'mhs' || LPAD(g::TEXT, 3, '0'), 
    'mhs' || LPAD(g::TEXT, 3, '0') || '@student.unair.ac.id', 
    '$2a$10$vI8aWBnW3fID.ZQ4/zo1G.q1lRps.9cGLcZEiGDMVr5yUP1KUOYTa', 
    'user'
FROM generate_series(1, 60) AS g;