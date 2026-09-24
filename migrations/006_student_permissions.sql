INSERT INTO permissions (name, description) VALUES
('student:list', 'Melihat daftar semua student'),
('student:read:any', 'Melihat detail student milik siapa pun'),
('student:create', 'Mendaftarkan data student baru'),
('student:update:any', 'Mengubah data student milik siapa pun'),
('student:delete', 'Menghapus data student');

INSERT INTO role_permissions (role, permission) VALUES
('admin', 'student:list'), 
('admin', 'student:read:any'), 
('admin', 'student:create'), 
('admin', 'student:update:any'), 
('admin', 'student:delete'),
('staff', 'student:list'), 
('staff', 'student:read:any'), 
('staff', 'student:create');

ALTER TABLE students ADD COLUMN owner_id INTEGER;

UPDATE students SET owner_id = 1 WHERE owner_id IS NULL; 

ALTER TABLE students ALTER COLUMN owner_id SET NOT NULL;
ALTER TABLE students ADD CONSTRAINT fk_student_owner FOREIGN KEY (owner_id) REFERENCES users(id)[cite: 7];